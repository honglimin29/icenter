/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package authcenter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"icenter/src/apimachinery/flowctrl"
	"icenter/src/apimachinery/rest"
	"icenter/src/apimachinery/util"
	"icenter/src/auth/authcenter/permit"
	"icenter/src/auth/meta"
	"icenter/src/common/blog"
	commonutil "icenter/src/common/util"
)

const (
	authAppCodeHeaderKey   string = "X-BK-APP-CODE"
	authAppSecretHeaderKey string = "X-BK-APP-SECRET"
	cmdbUser               string = "user"
	cmdbUserID             string = "system"
)

// ParseConfigFromKV returns a new config
func ParseConfigFromKV(prefix string, configmap map[string]string) (AuthConfig, error) {
	var err error
	var cfg AuthConfig
	enable, exist := configmap[prefix+".enable"]
	if !exist {
		return AuthConfig{}, nil
	}

	cfg.Enable, err = strconv.ParseBool(enable)
	if err != nil {
		return AuthConfig{}, errors.New(`invalid auth "enable" value`)
	}

	if !cfg.Enable {
		return AuthConfig{}, nil
	}

	enableSync, exist := configmap[prefix+".enableSync"]
	if exist && len(enableSync) > 0 {
		cfg.EnableSync, err = strconv.ParseBool(enableSync)
		if err != nil {
			return AuthConfig{}, errors.New(`invalid auth "enable" value`)
		}
	}

	address, exist := configmap[prefix+".address"]
	if !exist {
		return cfg, errors.New(`missing "address" configuration for auth center`)
	}

	cfg.Address = strings.Split(strings.Replace(address, " ", "", -1), ",")
	if len(cfg.Address) == 0 {
		return cfg, errors.New(`invalid "address" configuration for auth center`)
	}
	for i := range cfg.Address {
		if !strings.HasSuffix(cfg.Address[i], "/") {
			cfg.Address[i] = cfg.Address[i] + "/"
		}
	}

	cfg.AppSecret, exist = configmap[prefix+".appSecret"]
	if !exist {
		return cfg, errors.New(`missing "appSecret" configuration for auth center`)
	}

	if len(cfg.AppSecret) == 0 {
		return cfg, errors.New(`invalid "appSecret" configuration for auth center`)
	}

	cfg.AppCode, exist = configmap[prefix+".appCode"]
	if !exist {
		return cfg, errors.New(`missing "appCode" configuration for auth center`)
	}

	if len(cfg.AppCode) == 0 {
		return cfg, errors.New(`invalid "appCode" configuration for auth center`)
	}

	cfg.SystemID = SystemIDCMDB

	return cfg, nil
}

// NewAuthCenter create a instance to handle resources with blueking's AuthCenter.
func NewAuthCenter(tls *util.TLSClientConfig, cfg AuthConfig) (*AuthCenter, error) {
	blog.V(5).Infof("new auth center client with parameters tls: %+v, cfg: %+v", tls, cfg)
	if !cfg.Enable {
		return new(AuthCenter), nil
	}
	client, err := util.NewClient(tls)
	if err != nil {
		return nil, err
	}

	c := &util.Capability{
		Client: client,
		Discover: &acDiscovery{
			servers: cfg.Address,
		},
		Throttle: flowctrl.NewRateLimiter(1000, 1000),
		Mock: util.MockInfo{
			Mocked: false,
		},
	}

	header := http.Header{}
	header.Set("Content-Type", "application/json")
	header.Set("Accept", "application/json")
	header.Set(authAppCodeHeaderKey, cfg.AppCode)
	header.Set(authAppSecretHeaderKey, cfg.AppSecret)

	return &AuthCenter{
		Config: cfg,
		authClient: &authClient{
			client:      rest.NewRESTClient(c, ""),
			Config:      cfg,
			basicHeader: header,
		},
	}, nil
}

// AuthCenter means BlueKing's authorize center,
// which is also a open source product.
type AuthCenter struct {
	Config AuthConfig
	// http client instance
	client rest.ClientInterface
	// http header info
	header     http.Header
	authClient *authClient
}

func (ac *AuthCenter) Enabled() bool {
	return ac.Config.Enable
}

func (ac *AuthCenter) Authorize(ctx context.Context, a *meta.AuthAttribute) (decision meta.Decision, err error) {
	if !ac.Config.Enable {
		return meta.Decision{Authorized: true}, nil
	}

	// filter out SkipAction, which set by api server to skip authorization
	noSkipResources := make([]meta.ResourceAttribute, 0)
	for _, resource := range a.Resources {
		if resource.Action == meta.SkipAction {
			continue
		}
		noSkipResources = append(noSkipResources, resource)
	}
	a.Resources = noSkipResources
	if len(noSkipResources) == 0 {
		blog.V(5).Infof("Authorize skip. auth attribute: %+v", a)
		return meta.Decision{Authorized: true}, nil
	}

	batchresult, err := ac.AuthorizeBatch(ctx, a.User, a.Resources...)
	noAuth := make([]string, 0)
	for i, item := range batchresult {
		if !item.Authorized {
			noAuth = append(noAuth, fmt.Sprintf("resource [%v] permission deny by reason: %s", a.Resources[i].Type, item.Reason))
		}
	}

	if len(noAuth) > 0 {
		return meta.Decision{
			Authorized: false,
			Reason:     fmt.Sprintf("%v", noAuth),
		}, nil
	}

	return meta.Decision{Authorized: true}, nil
}

func (ac *AuthCenter) AuthorizeBatch(ctx context.Context, user meta.UserInfo, resources ...meta.ResourceAttribute) (decisions []meta.Decision, err error) {
	rid := commonutil.ExtractRequestIDFromContext(ctx)
	decisions = make([]meta.Decision, len(resources))
	if !ac.Config.Enable {
		for i := range decisions {
			decisions[i].Authorized = true
		}
		return decisions, nil
	}

	header := http.Header{}
	header.Set(AuthSupplierAccountHeaderKey, user.SupplierAccount)

	// this two index array record the resources's action original index.
	// used for recover the order of decisions.
	sysInputIndexes := make([]int, 0)
	sysInputExactIndexes := make([]int, 0)
	bizInputIndexes := make(map[int64][]int)
	bizInputExactIndexes := make(map[int64][]int)

	sysInput := AuthBatch{
		Principal: Principal{
			Type: cmdbUser,
			ID:   user.UserName,
		},
		ScopeInfo: ScopeInfo{
			ScopeType: ScopeTypeIDSystem,
			ScopeID:   SystemIDCMDB,
		},
		ResourceActions: make([]ResourceAction, 0),
	}
	sysExactInput := sysInput

	businessesInputs := make(map[int64]AuthBatch)
	businessesExactInputs := make(map[int64]AuthBatch)
	for index, rsc := range resources {
		action, err := adaptorAction(&rsc)
		if err != nil {
			blog.Errorf("auth batch, but adaptor action:%s failed, err: %v, rid: %s", rsc.Action, err, rid)
			return nil, err
		}

		// pick out skip resource at first.
		if permit.ShouldSkipAuthorize(&rsc) {
			// this resource should be skipped, do not need to verify in auth center.
			decisions[index].Authorized = true
			blog.V(5).Infof("skip authorization for resource: %+v, rid: %s", rsc, rid)
			continue
		}

		info, err := adaptor(&rsc)
		if err != nil {
			blog.Errorf("auth batch, but adaptor resource type:%s failed, err: %v, rid: %s", rsc.Basic.Type, err, rid)
			return nil, err
		}

		// modify special resource
		if rsc.Type == meta.MainlineModel || rsc.Type == meta.ModelTopology {
			blog.V(4).Infof("force convert scope type to global for resource type: %s, rid: %s", rsc.Type, rid)
			rsc.BusinessID = 0
		}

		if rsc.BusinessID > 0 {
			// this is a business resource.
			var tmpInputs map[int64]AuthBatch
			var tmpIndexes map[int64][]int
			if len(info.ResourceID) > 0 {
				tmpInputs = businessesExactInputs
				tmpIndexes = bizInputExactIndexes
			} else {
				tmpInputs = businessesInputs
				tmpIndexes = bizInputIndexes

			}

			if _, exist := tmpInputs[rsc.BusinessID]; !exist {
				tmpInputs[rsc.BusinessID] = AuthBatch{
					Principal: Principal{
						Type: cmdbUser,
						ID:   user.UserName,
					},
					ScopeInfo: ScopeInfo{
						ScopeType: ScopeTypeIDBiz,
						ScopeID:   strconv.FormatInt(rsc.BusinessID, 10),
					},
				}
				// initialize the business input indexes.
				tmpIndexes[rsc.BusinessID] = make([]int, 0)
			}

			a := tmpInputs[rsc.BusinessID]
			a.ResourceActions = append(a.ResourceActions, ResourceAction{
				ActionID:     action,
				ResourceType: info.ResourceType,
				ResourceID:   info.ResourceID,
			})
			tmpInputs[rsc.BusinessID] = a

			// record it's resource index
			indexes := tmpIndexes[rsc.BusinessID]
			indexes = append(indexes, index)
			tmpIndexes[rsc.BusinessID] = indexes
		} else {

			if len(info.ResourceID) > 0 {
				sysExactInput.ResourceActions = append(sysExactInput.ResourceActions, ResourceAction{
					ActionID:     action,
					ResourceType: info.ResourceType,
					ResourceID:   info.ResourceID,
				})

				// record it's system resource index
				sysInputExactIndexes = append(sysInputExactIndexes, index)
			} else {
				sysInput.ResourceActions = append(sysInput.ResourceActions, ResourceAction{
					ActionID:     action,
					ResourceType: info.ResourceType,
				})

				// record it's system resource index
				sysInputIndexes = append(sysInputIndexes, index)
			}

		}
	}

	// it's time to get the auth status from auth center now.
	// get biz resource auth status at first.
	// any business inputs
	for biz, rsc := range businessesInputs {
		statuses, err := ac.authClient.verifyAnyResourceBatch(ctx, header, &rsc)
		if err != nil {
			return nil, fmt.Errorf("get any resource[%s/%s] auth status failed, err: %v", rsc.ScopeType, rsc.ScopeID, err)
		}

		if len(statuses) != len(rsc.ResourceActions) {
			return nil, fmt.Errorf("got mismatch any biz authorize response from auth center, want: %d, got: %d", len(rsc.ResourceActions), len(statuses))
		}

		// update the decisions
		for index, status := range statuses {
			if rsc.ResourceActions[index].ResourceType != status.ResourceType ||
				string(rsc.ResourceActions[index].ActionID) != status.ActionID {
				return nil, fmt.Errorf("got any business auth mismatch info from auth center, with resource type[%s:%s], action[%s:%s]",
					rsc.ResourceActions[index].ResourceType, status.ResourceType, rsc.ResourceActions[index].ActionID, status.ActionID)
			}
			decisions[bizInputIndexes[biz][index]].Authorized = status.IsPass
		}
	}

	// exact business inputs
	for biz, rsc := range businessesExactInputs {
		statuses, err := ac.authClient.verifyExactResourceBatch(ctx, header, &rsc)
		if err != nil {
			return nil, fmt.Errorf("get exact resource[%s/%s] auth status failed, err: %v", rsc.ScopeType, rsc.ScopeID, err)
		}

		if len(statuses) != len(rsc.ResourceActions) {
			return nil, fmt.Errorf("got mismatch exact biz authorize response from auth center, want: %d, got: %d", len(rsc.ResourceActions), len(statuses))
		}

		// update the decisions
		for index, status := range statuses {
			if rsc.ResourceActions[index].ResourceType != status.ResourceType ||
				string(rsc.ResourceActions[index].ActionID) != status.ActionID {
				return nil, fmt.Errorf("got exact business auth mismatch info from auth center, with resource type[%s:%s], action[%s:%s]",
					rsc.ResourceActions[index].ResourceType, status.ResourceType, rsc.ResourceActions[index].ActionID, status.ActionID)
			}
			decisions[bizInputExactIndexes[biz][index]].Authorized = status.IsPass
		}
	}

	if len(sysInput.ResourceActions) != 0 {
		// get system resource auth status secondly.
		statuses, err := ac.authClient.verifyAnyResourceBatch(ctx, header, &sysInput)
		if err != nil {
			return nil, fmt.Errorf("get any system resource[%s/%s] auth status failed, err: %v", sysInput.ScopeType, sysInput.ScopeID, err)
		}

		if len(statuses) != len(sysInput.ResourceActions) {
			return nil, fmt.Errorf("got mismatch any system authorize response from auth center, want: %d, got: %d", len(sysInput.ResourceActions), len(statuses))
		}

		// update the system auth decisions
		for index, status := range statuses {
			if sysInput.ResourceActions[index].ResourceType != status.ResourceType ||
				string(sysInput.ResourceActions[index].ActionID) != status.ActionID {
				return nil, fmt.Errorf("got any system auth mismatch info from auth center, with resource type[%s:%s], action[%s:%s]",
					sysInput.ResourceActions[index].ResourceType, status.ResourceType,
					sysInput.ResourceActions[index].ActionID, status.ActionID)
			}
			decisions[sysInputIndexes[index]].Authorized = status.IsPass
		}
	}

	if len(sysExactInput.ResourceActions) != 0 {
		// get system resource auth status secondly.
		statuses, err := ac.authClient.verifyExactResourceBatch(ctx, header, &sysExactInput)
		if err != nil {
			return nil, fmt.Errorf("get exact system resource[%s/%s] auth status failed, err: %v", sysInput.ScopeType, sysInput.ScopeID, err)
		}

		if len(statuses) != len(sysExactInput.ResourceActions) {
			return nil, fmt.Errorf("got mismatch exact authorize response from auth center, want: %d, got: %d", len(sysExactInput.ResourceActions), len(statuses))
		}

		// update the system auth decisions
		for index, status := range statuses {
			if sysExactInput.ResourceActions[index].ResourceType != status.ResourceType ||
				string(sysExactInput.ResourceActions[index].ActionID) != status.ActionID {
				return nil, fmt.Errorf("got exact system auth mismatch info from auth center, with resource type[%s:%s], action[%s:%s]",
					sysExactInput.ResourceActions[index].ResourceType, status.ResourceType,
					sysExactInput.ResourceActions[index].ActionID, status.ActionID)
			}
			decisions[sysInputExactIndexes[index]].Authorized = status.IsPass
		}
	}

	return decisions, nil

}

func (ac *AuthCenter) GetAnyAuthorizedBusinessList(ctx context.Context, user meta.UserInfo) ([]int64, error) {
	if !ac.Config.Enable {
		return make([]int64, 0), nil
	}
	info := &Principal{
		Type: cmdbUser,
		ID:   user.UserName,
	}

	var appList []string
	var err error

	appList, err = ac.authClient.GetAnyAuthorizedScopes(ctx, ScopeTypeIDBiz, info)
	if err != nil {
		return nil, err
	}

	businessIDs := make([]int64, 0)
	for _, app := range appList {
		id, err := strconv.ParseInt(app, 10, 64)
		if err != nil {
			return businessIDs, err
		}
		businessIDs = append(businessIDs, id)
	}

	return businessIDs, nil
}

// get a user's authorized read business list.
func (ac *AuthCenter) GetExactAuthorizedBusinessList(ctx context.Context, user meta.UserInfo) ([]int64, error) {
	if !ac.Config.Enable {
		return make([]int64, 0), nil
	}

	option := &ListAuthorizedResources{
		Principal: Principal{
			Type: cmdbUser,
			ID:   user.UserName,
		},
		ScopeInfo: ScopeInfo{
			ScopeType: ScopeTypeIDSystem,
			ScopeID:   SystemIDCMDB,
		},
		TypeActions: []TypeAction{
			{
				ActionID:     Get,
				ResourceType: SysBusinessInstance,
			},
		},
		DataType: "array",
		Exact:    true,
	}
	appListRsc, err := ac.authClient.GetAuthorizedResources(ctx, option)
	if err != nil {
		return nil, err
	}

	businessIDs := make([]int64, 0)
	for _, appRsc := range appListRsc {
		for _, appList := range appRsc.ResourceIDs {
			for _, app := range appList {
				id, err := strconv.ParseInt(app.ResourceID, 10, 64)
				if err != nil {
					return businessIDs, err
				}
				businessIDs = append(businessIDs, id)
			}
		}
	}

	return businessIDs, nil
}
func (ac *AuthCenter) AdminEntrance(ctx context.Context, user meta.UserInfo) ([]string, error) {
	info := &Principal{
		Type: cmdbUser,
		ID:   user.UserName,
	}

	var systemList []string
	var err error
	if ac.Config.Enable {
		systemList, err = ac.authClient.GetAnyAuthorizedScopes(ctx, ScopeTypeIDSystem, info)
		if err != nil {
			return nil, err
		}
	}

	return systemList, nil
}

func (ac *AuthCenter) GetAuthorizedAuditList(ctx context.Context, user meta.UserInfo, businessID int64) ([]AuthorizedResource, error) {
	scopeInfo := ScopeInfo{}
	var resourceType ResourceTypeID
	if businessID > 0 {
		scopeInfo.ScopeType = ScopeTypeIDBiz
		scopeInfo.ScopeID = strconv.FormatInt(businessID, 10)
		resourceType = BizAuditLog
	} else {
		scopeInfo.ScopeType = ScopeTypeIDSystem
		scopeInfo.ScopeID = SystemIDCMDB
		resourceType = SysAuditLog
	}

	info := &ListAuthorizedResources{
		Principal: Principal{
			Type: cmdbUser,
			ID:   user.UserName,
		},
		ScopeInfo: scopeInfo,
		TypeActions: []TypeAction{
			{
				ActionID:     Get,
				ResourceType: resourceType,
			},
		},
		DataType: "array",
		Exact:    true,
	}

	var authorizedAudits []AuthorizedResource
	var err error
	if ac.Config.Enable {
		authorizedAudits, err = ac.authClient.GetAuthorizedResources(ctx, info)
		if err != nil {
			return nil, err
		}
	}

	return authorizedAudits, nil
}

func (ac *AuthCenter) RegisterResource(ctx context.Context, rs ...meta.ResourceAttribute) error {
	rid := commonutil.ExtractRequestIDFromContext(ctx)

	if ac.Config.Enable == false {
		blog.V(5).Infof("auth disabled, auth config: %+v, rid: %s", ac.Config, rid)
		return nil
	}

	if len(rs) == 0 {
		return errors.New("no resource to be registered")
	}

	registerInfo, err := ac.DryRunRegisterResource(ctx, rs...)
	if err != nil {
		return err
	}

	header := http.Header{}
	header.Set(AuthSupplierAccountHeaderKey, rs[0].SupplierAccount)

	if err := ac.authClient.registerResource(ctx, header, registerInfo); err != nil {
		if err == ErrDuplicated {
			return nil
		}
		return err
	}

	return nil
}

func (ac *AuthCenter) DryRunRegisterResource(ctx context.Context, rs ...meta.ResourceAttribute) (*RegisterInfo, error) {
	rid := commonutil.ExtractRequestIDFromContext(ctx)
	user := commonutil.ExtractRequestUserFromContext(ctx)
	if len(user) == 0 {
		user = cmdbUserID
	}

	if ac.Config.Enable == false {
		blog.V(5).Infof("auth disabled, auth config: %+v, rid: %s", ac.Config, rid)
		return nil, nil
	}

	if len(rs) <= 0 {
		blog.V(5).Infof("no resource should be register, rid: %s", rid)
		return nil, nil
	}
	info := RegisterInfo{}
	info.CreatorType = cmdbUser
	info.CreatorID = user
	info.Resources = make([]ResourceEntity, 0)
	for _, r := range rs {
		if len(r.Basic.Type) == 0 {
			return nil, errors.New("invalid resource attribute with empty object")
		}
		scope, err := ac.getScopeInfo(&r)
		if err != nil {
			return nil, err
		}

		rscInfo, err := adaptor(&r)
		if err != nil {
			return nil, fmt.Errorf("adaptor resource info failed, err: %v", err)
		}
		entity := ResourceEntity{}
		entity.ScopeInfo.ScopeID = scope.ScopeID
		entity.ScopeInfo.ScopeType = scope.ScopeType
		entity.ResourceType = rscInfo.ResourceType
		entity.ResourceID = rscInfo.ResourceID
		entity.ResourceName = rscInfo.ResourceName

		// TODO replace register with batch create or update interface, currently is register one by one.
		info.Resources = append(info.Resources, entity)
	}
	return &info, nil
}

func (ac *AuthCenter) DeregisterResource(ctx context.Context, rs ...meta.ResourceAttribute) error {
	if !ac.Config.Enable {
		return nil
	}
	if len(rs) <= 0 {
		// not resource should be deregister
		return nil
	}
	info := DeregisterInfo{}
	header := http.Header{}
	for _, r := range rs {
		if len(r.Basic.Type) == 0 {
			return errors.New("invalid resource attribute with empty object")
		}

		scope, err := ac.getScopeInfo(&r)
		if err != nil {
			return err
		}

		rscInfo, err := adaptor(&r)
		if err != nil {
			return fmt.Errorf("adaptor resource info failed, err: %v", err)
		}

		entity := ResourceEntity{}
		entity.ScopeID = scope.ScopeID
		entity.ScopeType = scope.ScopeType
		entity.ResourceType = rscInfo.ResourceType
		entity.ResourceID = rscInfo.ResourceID
		entity.ResourceName = rscInfo.ResourceName

		info.Resources = append(info.Resources, entity)

		header.Set(AuthSupplierAccountHeaderKey, r.SupplierAccount)
	}

	return ac.authClient.deregisterResource(ctx, header, &info)
}

func (ac *AuthCenter) UpdateResource(ctx context.Context, r *meta.ResourceAttribute) error {
	if !ac.Config.Enable {
		return nil
	}

	if len(r.Basic.Type) == 0 || len(r.Basic.Name) == 0 {
		return errors.New("invalid resource attribute with empty object or object name")
	}

	scope, err := ac.getScopeInfo(r)
	if err != nil {
		return err
	}

	rscInfo, err := adaptor(r)
	if err != nil {
		return fmt.Errorf("adaptor resource info failed, err: %v", err)
	}
	info := &UpdateInfo{
		ScopeInfo:    *scope,
		ResourceInfo: *rscInfo,
	}

	header := http.Header{}
	header.Set(AuthSupplierAccountHeaderKey, r.SupplierAccount)
	return ac.authClient.updateResource(ctx, header, info)
}

func (ac *AuthCenter) Get(ctx context.Context) error {
	panic("implement me")
}

func (ac *AuthCenter) ListResources(ctx context.Context, r *meta.ResourceAttribute) ([]meta.BackendResource, error) {
	if !ac.Config.Enable {
		return nil, nil
	}

	scopeInfo, err := ac.getScopeInfo(r)
	if err != nil {
		return nil, err
	}
	resourceType, err := convertResourceType(r.Type, r.BusinessID)
	if err != nil {
		return nil, err
	}
	header := http.Header{}
	resourceID, err := GenerateResourceID(*resourceType, r)
	if err != nil {
		return nil, err
	}
	blog.V(5).Infof("GenerateResourceID result: %+v", resourceID)
	searchCondition := SearchCondition{
		ScopeInfo:    *scopeInfo,
		ResourceType: *resourceType,
	}
	if resourceID != nil && len(resourceID) > 0 {
		searchCondition.ParentResources = resourceID[:len(resourceID)-1]
	}
	result, err := ac.authClient.ListResources(ctx, header, searchCondition)
	return result, err
}

func (ac *AuthCenter) getScopeInfo(r *meta.ResourceAttribute) (*ScopeInfo, error) {
	s := new(ScopeInfo)
	// TODO: this operation may be wrong, because some api filters does not
	// fill the business id field, so these api should be normalized.
	if r.BusinessID > 0 {
		s.ScopeType = ScopeTypeIDBiz
		s.ScopeID = strconv.FormatInt(r.BusinessID, 10)
	} else {
		s.ScopeType = ScopeTypeIDSystem
		s.ScopeID = SystemIDCMDB
	}
	return s, nil
}

type acDiscovery struct {
	// auth's servers address, must prefixed with http:// or https://
	servers []string
	index   int
	sync.Mutex
}

func (s *acDiscovery) GetServers() ([]string, error) {
	s.Lock()
	defer s.Unlock()

	num := len(s.servers)
	if num == 0 {
		return []string{}, errors.New("oops, there is no server can be used")
	}

	if s.index < num-1 {
		s.index = s.index + 1
		return append(s.servers[s.index-1:], s.servers[:s.index-1]...), nil
	} else {
		s.index = 0
		return append(s.servers[num-1:], s.servers[:num-1]...), nil
	}
}

func (ac *AuthCenter) RawDeregisterResource(ctx context.Context, scope ScopeInfo, rs ...meta.BackendResource) error {
	if !ac.Config.Enable {
		return nil
	}
	if len(rs) <= 0 {
		// not resource should be deregister
		return nil
	}
	info := DeregisterInfo{}
	header := http.Header{}
	for _, r := range rs {
		entity := ResourceEntity{}
		entity.ScopeID = scope.ScopeID
		entity.ScopeType = scope.ScopeType
		entity.ResourceType = ResourceTypeID(r[len(r)-1].ResourceType)
		resourceID := make([]RscTypeAndID, 0)
		for _, item := range r {
			resourceID = append(resourceID, RscTypeAndID{
				ResourceType: ResourceTypeID(item.ResourceType),
				ResourceID:   item.ResourceID,
			})
		}
		entity.ResourceID = resourceID

		info.Resources = append(info.Resources, entity)
	}

	return ac.authClient.deregisterResource(ctx, header, &info)
}
