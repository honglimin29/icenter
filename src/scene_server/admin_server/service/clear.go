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
	"net/http"

	"github.com/emicklei/go-restful"

	"icenter/src/common"
	"icenter/src/common/blog"
	"icenter/src/common/metadata"
	"icenter/src/common/storage/dal"
	"icenter/src/common/util"
	"icenter/src/common/version"
)

// clear drop tables common.AllTables from db
func (s *Service) clear(req *restful.Request, resp *restful.Response) {
	rHeader := req.Request.Header
	defErr := s.CCErr.CreateDefaultCCErrorIf(util.GetLanguage(rHeader))

	if version.CCRunMode == version.CCRunModeProduct {
		resp.WriteError(http.StatusInternalServerError, &metadata.RespError{Msg: defErr.Error(common.CCErrCommMigrateFailed)})
		return
	}

	err := clearDatabase(s.db)
	if nil != err {
		blog.Errorf("clear error: %v", err)
		resp.WriteError(http.StatusInternalServerError, &metadata.RespError{Msg: defErr.Error(common.CCErrCommMigrateFailed)})
		return
	}

	resp.WriteEntity(metadata.NewSuccessResp(nil))
}

func clearDatabase(db dal.RDB) error {
	// clear mongodb
	for _, tablename := range common.AllTables {
		db.DropTable(tablename)
	}

	// TODO clear redis

	return nil
}
