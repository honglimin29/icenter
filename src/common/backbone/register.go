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

package backbone

import (
	"github.com/gin-gonic/gin/json"

	"icenter/src/common/backbone/service_mange/zk"
	"icenter/src/common/registerdiscover"
	"icenter/src/common/types"
	"icenter/src/framework/core/errors"
)

type ServiceRegisterInterface interface {
	// Ping to ping server
	Ping() error
	// register local server info, it can only be called for once.
	Register(path string, c types.ServerInfo) error
}

func NewServiceRegister(client *zk.ZkClient) (ServiceRegisterInterface, error) {
	s := new(serviceRegister)
	s.client = registerdiscover.NewRegDiscoverEx(client)
	return s, nil
}

type serviceRegister struct {
	client *registerdiscover.RegDiscover
}

func (s *serviceRegister) Register(path string, c types.ServerInfo) error {
	if c.IP == "0.0.0.0" {
		return errors.New("register ip can not be 0.0.0.0")
	}

	js, err := json.Marshal(c)
	if err != nil {
		return err
	}

	return s.client.RegisterAndWatchService(path, js)
}

// Ping to ping server
func (s *serviceRegister) Ping() error {
	return s.client.Ping()
}
