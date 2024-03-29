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
	"encoding/json"

	"icenter/src/common"
	"icenter/src/common/mapstr"
	"icenter/src/common/metadata"
	"icenter/src/scene_server/topo_server/core/types"
)

func (s *Service) ParseSearchIdentifierOriginData(data []byte) (mapstr.MapStr, error) {
	rst := new(metadata.SearchIdentifierParam)
	err := json.Unmarshal(data, &rst)
	if nil != err {
		return nil, err
	}
	result := mapstr.MapStr{}
	result.Set("origin", rst)
	return result, nil
}

func (s *Service) SearchIdentifier(params types.ContextParams, pathParams, queryParams ParamsGetter, data mapstr.MapStr) (interface{}, error) {
	param, ok := data["origin"].(*metadata.SearchIdentifierParam)
	if !ok {
		return nil, params.Err.New(common.CCErrCommParamsIsInvalid, "param not set")
	}
	retval, err := s.Core.IdentifierOperation().SearchIdentifier(params, pathParams("obj_type"), param)
	if err != nil {
		return nil, err
	}
	return retval.Data, nil
}
