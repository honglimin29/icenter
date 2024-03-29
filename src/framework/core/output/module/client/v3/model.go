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

package v3

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/tidwall/gjson"

	"icenter/src/framework/common"
	"icenter/src/framework/core/log"
	"icenter/src/framework/core/types"
)

type ModelGetter interface {
	Model() ModelInterface
}
type ModelInterface interface {
	CreateObject(data types.MapStr) (int64, error)
	DeleteObject(cond common.Condition) error
	UpdateObject(data types.MapStr, cond common.Condition) error
	SearchObjects(cond common.Condition) ([]types.MapStr, error)
	SearchObjectTopo(cond common.Condition) ([]types.MapStr, error)
}

type Model struct {
	cli *Client
}

func newModel(cli *Client) *Model {
	return &Model{
		cli: cli,
	}
}

// CreateObject create a new model object
func (m *Model) CreateObject(data types.MapStr) (int64, error) {

	targetURL := fmt.Sprintf("%s/api/v3/object", m.cli.GetAddress())

	rst, err := m.cli.httpCli.POST(targetURL, nil, data.ToJSON())
	if nil != err {
		return 0, err
	}

	gs := gjson.ParseBytes(rst)

	// check result
	if !gs.Get("result").Bool() {
		return 0, errors.New(gs.Get("bk_error_msg").String())
	}

	// parse id
	id, err := strconv.ParseInt(gs.Get("data.id").String(), 10, 64)
	if err != nil {
		return 0, errors.New(gs.Get("bk_error_msg").String())
	}

	return id, nil
}

// DeleteObject delete a object by condition
func (m *Model) DeleteObject(cond common.Condition) error {

	data := cond.ToMapStr()
	id, err := data.Int("id")
	if nil != err {
		return err
	}

	targetURL := fmt.Sprintf("%s/api/v3/object/%d", m.cli.GetAddress(), id)
	log.Infof("targetURL %s", targetURL)
	rst, err := m.cli.httpCli.DELETE(targetURL, nil, nil)
	if nil != err {
		return err
	}

	gs := gjson.ParseBytes(rst)

	// check result
	if !gs.Get("result").Bool() {
		return errors.New(gs.Get("bk_error_msg").String())
	}

	return nil
}

// UpdateObject update a object by condition
func (m *Model) UpdateObject(data types.MapStr, cond common.Condition) error {

	dataCond := cond.ToMapStr()
	id, err := dataCond.Int("id")
	if nil != err {
		return err
	}

	targetURL := fmt.Sprintf("%s/api/v3/object/%d", m.cli.GetAddress(), id)

	rst, err := m.cli.httpCli.PUT(targetURL, nil, data.ToJSON())
	if nil != err {
		return err
	}

	gs := gjson.ParseBytes(rst)

	// check result
	if !gs.Get("result").Bool() {
		return errors.New(gs.Get("bk_error_msg").String())
	}
	return nil
}

// SearchObjects search some objects by condition
func (m *Model) SearchObjects(cond common.Condition) ([]types.MapStr, error) {

	data := cond.ToMapStr()

	targetURL := fmt.Sprintf("%s/api/v3/objects", m.cli.GetAddress())

	rst, err := m.cli.httpCli.POST(targetURL, nil, data.ToJSON())
	if nil != err {
		return nil, err
	}

	gs := gjson.ParseBytes(rst)

	// check result
	if !gs.Get("result").Bool() {
		return nil, errors.New(gs.Get("bk_error_msg").String())
	}

	dataStr := gs.Get("data").String()
	if 0 == len(dataStr) {
		return nil, errors.New("data is empty")
	}

	resultMap := make([]types.MapStr, 0)
	err = json.Unmarshal([]byte(dataStr), &resultMap)
	return resultMap, err
}

// SearchObjectTopo search object topo by condition
func (m *Model) SearchObjectTopo(cond common.Condition) ([]types.MapStr, error) {

	data := cond.ToMapStr()

	targetURL := fmt.Sprintf("%s/api/v3/objects/topo", m.cli.GetAddress())

	rst, err := m.cli.httpCli.POST(targetURL, nil, data.ToJSON())
	if nil != err {
		return nil, err
	}

	gs := gjson.ParseBytes(rst)

	// check result
	if !gs.Get("result").Bool() {
		return nil, errors.New(gs.Get("bk_error_msg").String())
	}

	dataStr := gs.Get("data").String()
	if 0 == len(dataStr) {
		return nil, errors.New("data is empty")
	}

	resultMap := make([]types.MapStr, 0)
	err = json.Unmarshal([]byte(dataStr), &resultMap)
	return resultMap, err

}
