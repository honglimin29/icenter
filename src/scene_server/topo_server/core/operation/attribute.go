/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package operation

import (
	"context"
	"icenter/src/apimachinery"
	"icenter/src/auth/extensions"
	"icenter/src/common"
	"icenter/src/common/blog"
	"icenter/src/common/condition"
	"icenter/src/common/mapstr"
	"icenter/src/common/metadata"
	"icenter/src/scene_server/topo_server/core/inst"
	"icenter/src/scene_server/topo_server/core/model"
	"icenter/src/scene_server/topo_server/core/types"
)

// AttributeOperationInterface attribute operation methods
type AttributeOperationInterface interface {
	CreateObjectAttribute(params types.ContextParams, data mapstr.MapStr) (model.AttributeInterface, error)
	DeleteObjectAttribute(params types.ContextParams, cond condition.Condition) error
	FindObjectAttributeWithDetail(params types.ContextParams, cond condition.Condition) ([]*metadata.ObjAttDes, error)
	FindObjectAttribute(params types.ContextParams, cond condition.Condition) ([]model.AttributeInterface, error)
	UpdateObjectAttribute(params types.ContextParams, data mapstr.MapStr, attID int64) error

	SetProxy(modelFactory model.Factory, instFactory inst.Factory, obj ObjectOperationInterface, asst AssociationOperationInterface, grp GroupOperationInterface)
}

// NewAttributeOperation create a new attribute operation instance
func NewAttributeOperation(client apimachinery.ClientSetInterface, authManager *extensions.AuthManager) AttributeOperationInterface {
	return &attribute{
		clientSet:   client,
		authManager: authManager,
	}
}

type attribute struct {
	clientSet    apimachinery.ClientSetInterface
	authManager  *extensions.AuthManager
	modelFactory model.Factory
	instFactory  inst.Factory
	obj          ObjectOperationInterface
	asst         AssociationOperationInterface
	grp          GroupOperationInterface
}

func (a *attribute) SetProxy(modelFactory model.Factory, instFactory inst.Factory, obj ObjectOperationInterface, asst AssociationOperationInterface, grp GroupOperationInterface) {
	a.modelFactory = modelFactory
	a.instFactory = instFactory
	a.obj = obj
	a.asst = asst
	a.grp = grp
}

func (a *attribute) CreateObjectAttribute(params types.ContextParams, data mapstr.MapStr) (model.AttributeInterface, error) {

	var err error
	att := a.modelFactory.CreateAttribute(params)

	err = att.Parse(data)
	if nil != err {
		blog.Errorf("[operation-attr] failed to parse the attribute data (%#v), error info is %s", data, err.Error())
		return nil, err
	}

	// check the object id
	objID := att.Attribute().ObjectID
	err = a.obj.IsValidObject(params, objID)
	if nil != err {
		return nil, params.Err.New(common.CCErrTopoObjectAttributeCreateFailed, err.Error())
	}

	// check is the group exist
	cond := condition.CreateCondition()
	cond.Field(common.BKObjIDField).Eq(att.Attribute().ObjectID)
	cond.Field(common.BKPropertyGroupIDField).Eq(att.Attribute().PropertyGroup)
	groupResult, err := a.grp.FindObjectGroup(params, cond)
	if nil != err {
		blog.Errorf("[operation-attr] failed to search the attribute group data (%#v), error info is %s", cond.ToMapStr(), err.Error())
		return nil, err
	}
	// create the default group
	if 0 == len(groupResult) {

		group := metadata.Group{
			IsDefault:  true,
			GroupIndex: -1,
			GroupName:  common.BKBizDefault,
			GroupID:    common.BKBizDefault,
			ObjectID:   att.Attribute().ObjectID,
			OwnerID:    att.Attribute().OwnerID,
		}
		if nil != params.MetaData {
			group.Metadata = *params.MetaData
		}

		data := mapstr.NewFromStruct(group, "field")
		if _, err := a.grp.CreateObjectGroup(params, data); nil != err {
			blog.Errorf("[operation-obj] failed to create the default group, err: %s", err.Error())
			return nil, params.Err.Error(common.CCErrTopoObjectGroupCreateFailed)
		}
	}

	// create a new one
	err = att.Create()
	if nil != err {
		blog.Errorf("[operation-attr] failed to save the attribute data (%#v), error info is %s", data, err.Error())
		return nil, err
	}

	return att, nil
}

func (a *attribute) DeleteObjectAttribute(params types.ContextParams, cond condition.Condition) error {

	attrItems, err := a.FindObjectAttribute(params, cond)
	if nil != err {
		blog.Errorf("[operation-attr] failed to find the attributes by the cond(%v), err: %v", cond.ToMapStr(), err)
		return params.Err.New(common.CCErrTopoObjectAttributeDeleteFailed, err.Error())
	}

	// auth: check authorization
	// var objID string
	// for idx, attrItem := range attrItems {
	// 	oID := attrItem.Attribute().ObjectID
	// 	if idx == 0 && objID != oID {
	// 		return params.Err.New(common.CCErrTopoObjectAttributeDeleteFailed, "can't attributes of multiple model per request")
	// 	}
	// }
	// if err := a.authManager.AuthorizeByObjectID(params.Context, params.Header, meta.Update, objID); err != nil {
	// 	return params.Err.New(common.CCErrCommAuthorizeFailed, err.Error())
	// }

	for _, attrItem := range attrItems {
		// delete the attribute
		rsp, err := a.clientSet.CoreService().Model().DeleteModelAttr(context.Background(), params.Header, attrItem.Attribute().ObjectID, &metadata.DeleteOption{Condition: cond.ToMapStr()})
		if nil != err {
			blog.Errorf("[operation-attr] delete object attribute failed, request object controller with err: %v", err)
			return params.Err.Error(common.CCErrCommHTTPDoRequestFailed)
		}

		if !rsp.Result {
			blog.Errorf("[operation-attr] failed to delete the attribute by condition(%v), err: %s", cond.ToMapStr(), rsp.ErrMsg)
			return params.Err.New(rsp.Code, rsp.ErrMsg)
		}
	}

	return nil
}

func (a *attribute) FindObjectAttributeWithDetail(params types.ContextParams, cond condition.Condition) ([]*metadata.ObjAttDes, error) {
	attrs, err := a.FindObjectAttribute(params, cond)
	if nil != err {
		return nil, err
	}
	results := make([]*metadata.ObjAttDes, 0)
	for _, attr := range attrs {
		result := &metadata.ObjAttDes{Attribute: *attr.Attribute()}

		attribute := attr.Attribute()
		grpCond := condition.CreateCondition()
		grpCond.Field(metadata.GroupFieldGroupID).Eq(attribute.PropertyGroup)
		grpCond.Field(metadata.GroupFieldSupplierAccount).Eq(attribute.OwnerID)
		grpCond.Field(metadata.GroupFieldObjectID).Eq(attribute.ObjectID)
		grps, err := a.grp.FindObjectGroup(params, grpCond)
		if nil != err {
			return nil, err
		}

		for _, grp := range grps {
			// should be only one
			result.PropertyGroupName = grp.Group().GroupName
		}

		results = append(results, result)
	}

	return results, nil
}

func (a *attribute) FindObjectAttribute(params types.ContextParams, cond condition.Condition) ([]model.AttributeInterface, error) {
	fCond := cond.ToMapStr()
	if nil != params.MetaData {
		fCond.Merge(metadata.PublicAndBizCondition(*params.MetaData))
		fCond.Remove(metadata.BKMetadata)
	} else {
		fCond.Merge(metadata.BizLabelNotExist)
	}

	rsp, err := a.clientSet.CoreService().Model().ReadModelAttrByCondition(context.Background(), params.Header, &metadata.QueryCondition{Condition: fCond})
	if nil != err {
		blog.Errorf("[operation-attr] failed to request object controller, error info is %s", err.Error())
		return nil, params.Err.Error(common.CCErrCommHTTPDoRequestFailed)
	}

	if !rsp.Result {
		blog.Errorf("[operation-attr] failed to search attribute by the condition(%#v), error info is %s", fCond, rsp.ErrMsg)
		return nil, params.Err.New(rsp.Code, rsp.ErrMsg)
	}

	return model.CreateAttribute(params, a.clientSet, rsp.Data.Info), nil
}

func (a *attribute) UpdateObjectAttribute(params types.ContextParams, data mapstr.MapStr, attID int64) error {

	input := metadata.UpdateOption{
		Condition: condition.CreateCondition().Field(common.BKFieldID).Eq(attID).ToMapStr(),
		Data:      data,
	}

	rsp, err := a.clientSet.CoreService().Model().UpdateModelAttrsByCondition(context.Background(), params.Header, &input)
	if nil != err {
		blog.Errorf("[operation-attr] failed to request object controller, error info is %s", err.Error())
		return params.Err.Error(common.CCErrCommHTTPDoRequestFailed)
	}

	if !rsp.Result {
		blog.Errorf("[operation-attr] failed to update the attribute by the attr-id(%d), error info is %s", attID, rsp.ErrMsg)
		return params.Err.New(rsp.Code, rsp.ErrMsg)
	}

	return nil
}
