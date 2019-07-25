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

package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"icenter/src/common"
	"icenter/src/common/backbone"
	cc "icenter/src/common/backbone/configcenter"
	"icenter/src/common/blog"
	"icenter/src/common/rdapi"
	"icenter/src/common/storage/dal/mongo"
	"icenter/src/common/storage/dal/redis"
	"icenter/src/common/types"
	"icenter/src/common/version"
	"icenter/src/source_controller/coreservice/app/options"
	coresvr "icenter/src/source_controller/coreservice/service"
)

// CoreServer the core server
type CoreServer struct {
	Core    *backbone.Engine
	Config  options.Config
	Service coresvr.CoreServiceInterface
}

func (t *CoreServer) onCoreServiceConfigUpdate(previous, current cc.ProcessConfig) {

	t.Config.Mongo = mongo.ParseConfigFromKV("mongodb", current.ConfigMap)
	t.Config.Redis = redis.ParseConfigFromKV("redis", current.ConfigMap)

	blog.V(3).Infof("the new cfg:%#v the origin cfg:%#v", t.Config, current.ConfigMap)

}

// Run main function
func Run(ctx context.Context, op *options.ServerOption) error {
	svrInfo, err := newServerInfo(op)
	if err != nil {
		return fmt.Errorf("wrap server info failed, err: %v", err)
	}

	coreSvr := new(CoreServer)
	coreService := coresvr.New()
	coreSvr.Service = coreService

	webhandler := coreService.WebService()
	webhandler.ServiceErrorHandler(rdapi.ServiceErrorHandler)

	input := &backbone.BackboneParameter{
		ConfigUpdate: coreSvr.onCoreServiceConfigUpdate,
		ConfigPath:   op.ServConf.ExConfig,
		Regdiscv:     op.ServConf.RegDiscover,
		SrvInfo:      svrInfo,
	}

	engine, err := backbone.NewBackbone(ctx, input)
	if err != nil {
		return fmt.Errorf("new backbone failed, err: %v", err)
	}

	var configReady bool
	for sleepCnt := 0; sleepCnt < common.APPConfigWaitTime; sleepCnt++ {
		// redis not found
		if "" == coreSvr.Config.Redis.Address {
			time.Sleep(time.Second)
			continue
		}
		// Mongo not found
		if "" == coreSvr.Config.Mongo.Address {
			time.Sleep(time.Second)
			continue
		}
		configReady = true
		break

	}

	if false == configReady {
		return fmt.Errorf("Configuration item not found")
	}

	coreSvr.Core = engine
	err = coreService.SetConfig(coreSvr.Config, engine, engine.CCErr, engine.Language)
	if err != nil {
		return err
	}
	if err := backbone.StartServer(ctx, engine, webhandler); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
	}
	return nil
}

func newServerInfo(op *options.ServerOption) (*types.ServerInfo, error) {
	ip, err := op.ServConf.GetAddress()
	if err != nil {
		return nil, err
	}

	port, err := op.ServConf.GetPort()
	if err != nil {
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	info := &types.ServerInfo{
		IP:       ip,
		Port:     port,
		HostName: hostname,
		Scheme:   "http",
		Version:  version.GetVersion(),
		Pid:      os.Getpid(),
	}
	return info, nil
}
