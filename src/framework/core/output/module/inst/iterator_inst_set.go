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

package inst

import (
	"icenter/src/framework/common"
	//"icenter/src/framework/core/log"
	"icenter/src/framework/core/output/module/client"
	"icenter/src/framework/core/output/module/model"
	"icenter/src/framework/core/types"
	//"fmt"
	"io"
)

// SetIterator the iterator interface for the set
type SetIterator interface {
	Next() (SetInterface, error)
	ForEach(callbackItem func(item SetInterface) error) error
}
type iteratorInstSet struct {
	targetModel model.Model
	cond        common.Condition
	buffer      []types.MapStr
	bufIdx      int
}

func NewIteratorInstSet(target model.Model, cond common.Condition) (SetIterator, error) {

	iter := &iteratorInstSet{
		targetModel: target,
		cond:        cond,
		buffer:      make([]types.MapStr, 0),
	}

	iter.cond.SetLimit(DefaultLimit)
	iter.cond.SetStart(iter.bufIdx)

	existItems, err := client.GetClient().CCV3(client.Params{SupplierAccount: target.GetSupplierAccount()}).Set().SearchSets(iter.cond)
	if nil != err {
		return nil, err
	}
	//fmt.Println("set next", existItems, iter.cond.ToMapStr())

	iter.buffer = append(iter.buffer, existItems...)

	return iter, nil

}

func (cli *iteratorInstSet) Next() (SetInterface, error) {

	if len(cli.buffer) == cli.bufIdx {

		cli.cond.SetStart(cli.bufIdx)

		existItems, err := client.GetClient().CCV3(client.Params{SupplierAccount: cli.targetModel.GetSupplierAccount()}).Set().SearchSets(cli.cond)
		if nil != err {
			return nil, err
		}

		if 0 == len(existItems) {
			cli.bufIdx = 0
			return nil, io.EOF
		}

		cli.buffer = append(cli.buffer, existItems...)
	}

	tmpItem := cli.buffer[cli.bufIdx]
	cli.bufIdx++

	returnItem := &set{
		target: cli.targetModel,
		datas:  tmpItem,
	}

	return returnItem, nil
}

func (cli *iteratorInstSet) ForEach(callbackItem func(item SetInterface) error) error {
	for {

		item, err := cli.Next()
		if nil != err {
			if io.EOF == err {
				return nil
			}
			return err
		}

		if nil == item {
			return nil
		}

		err = callbackItem(item)
		if nil != err {
			return err
		}
	}

}
