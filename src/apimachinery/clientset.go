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

package apimachinery

import (
	"icenter/src/apimachinery/coreservice"
	"icenter/src/apimachinery/discovery"
	"icenter/src/apimachinery/flowctrl"
	"icenter/src/apimachinery/healthz"
	"icenter/src/apimachinery/objcontroller"
	"icenter/src/apimachinery/util"
)

type ClientSetInterface interface {
	ObjectController() objcontroller.ObjControllerClientInterface
	CoreService() coreservice.CoreServiceClientInterface

	Healthz() healthz.HealthzInterface
}

func NewApiMachinery(c *util.APIMachineryConfig, discover discovery.DiscoveryInterface) (ClientSetInterface, error) {
	client, err := util.NewClient(c.TLSConfig)
	if err != nil {
		return nil, err
	}

	flowcontrol := flowctrl.NewRateLimiter(c.QPS, c.Burst)
	return NewClientSet(client, discover, flowcontrol), nil
}

func NewClientSet(client util.HttpClient, discover discovery.DiscoveryInterface, throttle flowctrl.RateLimiter) ClientSetInterface {
	return &ClientSet{
		version:  "v3",
		client:   client,
		discover: discover,
		throttle: throttle,
	}
}

func NewMockClientSet() *ClientSet {
	return &ClientSet{
		version:  "unit_test",
		client:   nil,
		discover: discovery.NewMockDiscoveryInterface(),
		throttle: flowctrl.NewMockRateLimiter(),
		Mock:     util.MockInfo{Mocked: true},
	}
}

type ClientSet struct {
	version  string
	client   util.HttpClient
	discover discovery.DiscoveryInterface
	throttle flowctrl.RateLimiter
	Mock     util.MockInfo
}

func (cs *ClientSet) Healthz() healthz.HealthzInterface {
	c := &util.Capability{
		Client:   cs.client,
		Throttle: cs.throttle,
	}
	return healthz.NewHealthzClient(c, cs.discover)
}

func (cs *ClientSet) CoreService() coreservice.CoreServiceClientInterface {
	c := &util.Capability{
		Client:   cs.client,
		Discover: cs.discover.CoreService(),
		Throttle: cs.throttle,
		Mock:     cs.Mock,
	}
	return coreservice.NewCoreServiceClient(c, cs.version)
}

func (cs *ClientSet) ObjectController() objcontroller.ObjControllerClientInterface {
	c := &util.Capability{
		Client:   cs.client,
		Discover: cs.discover.ObjectCtrl(),
		Throttle: cs.throttle,
		Mock:     cs.Mock,
	}
	return objcontroller.NewObjectControllerInterface(c, cs.version)
}
