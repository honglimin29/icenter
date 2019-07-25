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

package actions

import (
	"github.com/emicklei/go-restful"
	"icenter/src/common/http/httpserver"
)

// Action restful action struct
type Action struct {
	Verb          string               // Verb identifying the action ("GET", "POST", "WATCH", PROXY", etc).
	Path          string               // The path of the action
	Params        []*restful.Parameter // List of parameters associated with the action.
	Handler       restful.RouteFunction
	FilterHandler []restful.FilterFunction
	Version       string //api 版本号，为空表示没有版本
}

var acts = []*httpserver.Action{}

// GetAPIAction fetch api actions
func GetAPIAction() []*httpserver.Action {
	return acts
}
