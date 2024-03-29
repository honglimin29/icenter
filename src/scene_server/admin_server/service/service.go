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

package service

import (
	"context"

	"icenter/src/auth/authcenter"
	"icenter/src/common"
	"icenter/src/common/backbone"
	"icenter/src/common/errors"
	"icenter/src/common/metadata"
	"icenter/src/common/metric"
	"icenter/src/common/rdapi"
	"icenter/src/common/storage/dal"
	"icenter/src/common/types"
	"icenter/src/scene_server/admin_server/app/options"

	"github.com/emicklei/go-restful"
)

type Service struct {
	*backbone.Engine
	db           dal.RDB
	ccApiSrvAddr string
	ctx          context.Context
	Config       options.Config
	authCenter   *authcenter.AuthCenter
}

func NewService(ctx context.Context) *Service {
	return &Service{
		ctx: ctx,
	}
}

func (s *Service) SetDB(db dal.RDB) {
	s.db = db
}

func (s *Service) SetAuthcenter(authCenter *authcenter.AuthCenter) {
	s.authCenter = authCenter
}

func (s *Service) SetApiSrvAddr(ccApiSrvAddr string) {
	s.ccApiSrvAddr = ccApiSrvAddr
}

func (s *Service) WebService() *restful.Container {
	container := restful.NewContainer()

	api := new(restful.WebService)
	getErrFunc := func() errors.CCErrorIf {
		return s.CCErr
	}
	api.Path("/migrate/v3").Filter(rdapi.AllGlobalFilter(getErrFunc)).Produces(restful.MIME_JSON)

	api.Route(api.POST("/authcenter/init").To(s.InitAuthCenter))
	api.Route(api.POST("/migrate/{distribution}/{ownerID}").To(s.migrate))
	api.Route(api.POST("/migrate/system/hostcrossbiz/{ownerID}").To(s.SetSystemConfiguration))
	api.Route(api.POST("/clear").To(s.clear))
	api.Route(api.GET("/healthz").To(s.Healthz))

	container.Add(api)

	healthzAPI := new(restful.WebService).Produces(restful.MIME_JSON)
	healthzAPI.Route(healthzAPI.GET("/healthz").To(s.Healthz))
	container.Add(healthzAPI)

	return container
}

func (s *Service) Healthz(req *restful.Request, resp *restful.Response) {
	meta := metric.HealthMeta{IsHealthy: true}

	// zk health status
	zkItem := metric.HealthItem{IsHealthy: true, Name: types.CCFunctionalityServicediscover}
	if err := s.Engine.Ping(); err != nil {
		zkItem.IsHealthy = false
		zkItem.Message = err.Error()
	}
	meta.Items = append(meta.Items, zkItem)

	// mongodb
	meta.Items = append(meta.Items, metric.NewHealthItem(types.CCFunctionalityMongo, s.db.Ping()))

	for _, item := range meta.Items {
		if item.IsHealthy == false {
			meta.IsHealthy = false
			meta.Message = "admin server is unhealthy"
			break
		}
	}

	info := metric.HealthInfo{
		Module:     types.CC_MODULE_MIGRATE,
		HealthMeta: meta,
		AtTime:     metadata.Now(),
	}

	answer := metric.HealthResponse{
		Code:    common.CCSuccess,
		Data:    info,
		OK:      meta.IsHealthy,
		Result:  meta.IsHealthy,
		Message: meta.Message,
	}
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteEntity(answer)
}
