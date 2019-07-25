/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.,
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the ",License",); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an ",AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"icenter/src/common/mapstr"
	"icenter/src/common/metadata"
	"icenter/src/source_controller/coreservice/core"
)

func (s *coreService) CreateManyModelClassification(params core.ContextParams, pathParams, queryParams ParamsGetter, data mapstr.MapStr) (interface{}, error) {

	inputDatas := metadata.CreateManyModelClassifiaction{}
	if err := data.MarshalJSONInto(&inputDatas); nil != err {
		return nil, err
	}
	return s.core.ModelOperation().CreateManyModelClassification(params, inputDatas)
}

func (s *coreService) CreateOneModelClassification(params core.ContextParams, pathParams, queryParams ParamsGetter, data mapstr.MapStr) (interface{}, error) {

	inputData := metadata.CreateOneModelClassification{}
	if err := data.MarshalJSONInto(&inputData); nil != err {
		return nil, err
	}
	return s.core.ModelOperation().CreateOneModelClassification(params, inputData)
}

func (s *coreService) SetOneModelClassificaition(params core.ContextParams, pathParams, queryParams ParamsGetter, data mapstr.MapStr) (interface{}, error) {

	inputData := metadata.SetOneModelClassification{}
	if err := data.MarshalJSONInto(&inputData); nil != err {
		return nil, err
	}

	return s.core.ModelOperation().SetOneModelClassification(params, inputData)
}

func (s *coreService) SetManyModelClassificaiton(params core.ContextParams, pathParams, queryParams ParamsGetter, data mapstr.MapStr) (interface{}, error) {

	inputDatas := metadata.SetManyModelClassification{}
	if err := data.MarshalJSONInto(&inputDatas); nil != err {
		return nil, err
	}
	return s.core.ModelOperation().SetManyModelClassification(params, inputDatas)
}

func (s *coreService) UpdateModelClassification(params core.ContextParams, pathParams, queryParams ParamsGetter, data mapstr.MapStr) (interface{}, error) {

	inputData := metadata.UpdateOption{}
	if err := data.MarshalJSONInto(&inputData); nil != err {
		return nil, err
	}
	return s.core.ModelOperation().UpdateModelClassification(params, inputData)
}

func (s *coreService) DeleteModelClassification(params core.ContextParams, pathParams, queryParams ParamsGetter, data mapstr.MapStr) (interface{}, error) {

	inputData := metadata.DeleteOption{}
	if err := data.MarshalJSONInto(&inputData); nil != err {
		return nil, err
	}
	return s.core.ModelOperation().DeleteModelClassification(params, inputData)
}

func (s *coreService) CascadeDeleteModelClassification(params core.ContextParams, pathParams, queryParams ParamsGetter, data mapstr.MapStr) (interface{}, error) {

	inputData := metadata.DeleteOption{}
	if err := data.MarshalJSONInto(&inputData); nil != err {
		return nil, err
	}
	return s.core.ModelOperation().CascadeDeleteModeClassification(params, inputData)
}

func (s *coreService) SearchModelClassification(params core.ContextParams, pathParams, queryParams ParamsGetter, data mapstr.MapStr) (interface{}, error) {

	inputData := metadata.QueryCondition{}
	if err := data.MarshalJSONInto(&inputData); nil != err {
		return nil, err
	}

	dataResult, err := s.core.ModelOperation().SearchModelClassification(params, inputData)
	if nil != err {
		return dataResult, err
	}

	// translate language
	for index := range dataResult.Info {
		dataResult.Info[index].ClassificationName = s.TranslateClassificationName(params.Lang, &dataResult.Info[index])
	}

	return dataResult, err
}
