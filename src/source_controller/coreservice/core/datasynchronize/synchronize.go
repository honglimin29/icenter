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

package datasynchronize

import (
	"icenter/src/common"
	"icenter/src/common/blog"
	"icenter/src/common/mapstr"
	"icenter/src/common/metadata"
	"icenter/src/common/storage/dal"
	"icenter/src/source_controller/coreservice/core"
)

type SynchronizeManager struct {
	dbProxy   dal.RDB
	dependent OperationDependences
}

// New create a new model manager instance
func New(dbProxy dal.RDB, dependent OperationDependences) core.DataSynchronizeOperation {
	return &SynchronizeManager{
		dbProxy:   dbProxy,
		dependent: dependent,
	}
}

func (s *SynchronizeManager) SynchronizeInstanceAdapter(ctx core.ContextParams, syncData *metadata.SynchronizeParameter) ([]metadata.ExceptionResult, error) {
	syncDataAdpater := NewSynchronizeInstanceAdapter(syncData, s.dbProxy)
	err := syncDataAdpater.PreSynchronizeFilter(ctx)
	if err != nil {
		blog.Errorf("SynchronizeInstanceAdapter error, err:%s,rid:%s", err.Error(), ctx.ReqID)
		return nil, err
	}
	syncDataAdpater.SaveSynchronize(ctx)
	return syncDataAdpater.GetErrorStringArr(ctx)

}

func (s *SynchronizeManager) SynchronizeModelAdapter(ctx core.ContextParams, syncData *metadata.SynchronizeParameter) ([]metadata.ExceptionResult, error) {
	syncDataAdpater := NewSynchronizeModelAdapter(syncData, s.dbProxy)
	err := syncDataAdpater.PreSynchronizeFilter(ctx)
	if err != nil {
		return nil, err
	}
	syncDataAdpater.SaveSynchronize(ctx)
	return syncDataAdpater.GetErrorStringArr(ctx)

}

func (s *SynchronizeManager) SynchronizeAssociationAdapter(ctx core.ContextParams, syncData *metadata.SynchronizeParameter) ([]metadata.ExceptionResult, error) {
	syncDataAdpater := NewSynchronizeAssociationAdapter(syncData, s.dbProxy)
	err := syncDataAdpater.PreSynchronizeFilter(ctx)
	if err != nil {
		return nil, err
	}
	syncDataAdpater.SaveSynchronize(ctx)
	return syncDataAdpater.GetErrorStringArr(ctx)

}

func (s *SynchronizeManager) Find(ctx core.ContextParams, input *metadata.SynchronizeFindInfoParameter) ([]mapstr.MapStr, uint64, error) {
	adapter := NewSynchronizeFindAdapter(input, s.dbProxy)
	return adapter.Find(ctx)
}

func (s *SynchronizeManager) ClearData(ctx core.ContextParams, input *metadata.SynchronizeClearDataParameter) error {

	adapter := NewClearData(s.dbProxy, input)
	if input.Sign == "" {
		blog.Errorf("clearData parameter synchronize_flag illegal, input:%#v,rid:%s", input, ctx.ReqID)
		return ctx.Error.Errorf(common.CCErrCommParamsNeedSet, "synchronize_flag")
	}

	if !input.Legality(common.SynchronizeSignPrefix) {
		blog.Errorf("clearData parameter illegal, input:%#v,rid:%s", input, ctx.ReqID)
		return ctx.Error.Errorf(common.CCErrCommParamsInvalid, input.Sign)
	}
	adapter.clearData(ctx)
	return nil
}
