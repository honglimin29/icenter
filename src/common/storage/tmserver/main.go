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

package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"icenter/src/common"
	"icenter/src/common/blog"
	"icenter/src/common/storage/tmserver/app"
	"icenter/src/common/storage/tmserver/app/options"
	_ "icenter/src/common/storage/tmserver/core/command"
	"icenter/src/common/types"
	"icenter/src/common/util"

	"github.com/spf13/pflag"
	// init all cmds
)

func main() {
	common.SetIdentification(types.CC_MODULE_TXC)
	runtime.GOMAXPROCS(runtime.NumCPU())

	blog.InitLogs()
	defer blog.CloseLogs()

	op := options.NewServerOption()
	op.AddFlags(pflag.CommandLine)

	util.InitFlags()

	if err := app.Run(context.Background(), op); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		blog.Errorf("process stoped by %v", err)
		blog.CloseLogs()
		os.Exit(1)
	}
}
