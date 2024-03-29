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

package command

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"icenter/src/common"
	"icenter/src/common/blog"
	"icenter/src/common/storage/dal"
	"icenter/src/common/util"
)

func export(ctx context.Context, db dal.RDB, opt *option) error {
	file, err := os.Create(opt.position)
	if nil != err {
		return err
	}
	defer file.Close()
	defer file.Sync()

	topo, err := getBKTopo(ctx, db, opt)
	if nil != err {
		return err
	}

	err = topo.BizTopo.walk(func(node *Node) error {
		node.Data = util.CopyMap(node.Data, nil,
			[]string{
				common.BKInstParentStr,
				common.BKChildStr,
				common.BKAppIDField,
				common.BKSetIDField,
				common.BKModuleIDField,
				common.BKInstIDField,
				common.BKOwnerIDField,
				common.BKSupplierIDField,
				common.CreateTimeField,
				common.LastTimeField,
				"_id",
			},
		)
		return nil
	})
	if err != nil {
		blog.Errorf("walk biz topo failed, err: %+v", err)
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(topo)
	if nil != err {
		return fmt.Errorf("encode topo error: %s", err.Error())
	}

	return nil
}
