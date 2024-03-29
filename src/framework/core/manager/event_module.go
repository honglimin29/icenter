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

package manager

import (
	"icenter/src/framework/core/log"
	"icenter/src/framework/core/output"
	"icenter/src/framework/core/types"

	"errors"
)

type eventModule struct {
	outputerMgr output.Manager
}

func (cli *eventModule) parse(data types.MapStr) ([]*types.Event, error) {

	dataArr, err := data.MapStrArray("data")
	if nil != err {
		return nil, err
	}

	tm, err := data.Time("action_time")
	if nil != err {
		log.Error("failed to get action time")
		return nil, err
	}

	action := data.String("action")
	if 0 == len(action) {
		log.Error("the event action is not set")
		return nil, errors.New("the event action is not set")
	}

	eves := make([]*types.Event, 0)
	for _, dataItem := range dataArr {

		curModule, err := dataItem.MapStr("cur_data")

		if nil != err {
			log.Errorf("failed to get the curr data, %s", err.Error())
			return nil, err
		}

		preModule, err := dataItem.MapStr("pre_data")

		if nil != err {
			log.Errorf("failed to get the curr data, %s", err.Error())
			return nil, err
		}

		ev := &types.Event{}
		ev.SetCurrData(curModule)
		ev.SetPreData(preModule)
		ev.SetAction(action)
		ev.SetActionTime(*tm)
		eves = append(eves, ev)

	}

	return eves, nil
}
