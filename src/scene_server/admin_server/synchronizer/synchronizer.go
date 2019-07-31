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

package synchronizer

import (
	"context"
	"fmt"

	"icenter/src/apimachinery"
	"icenter/src/auth"
	"icenter/src/auth/authcenter"
	"icenter/src/auth/extensions"
	"icenter/src/common/blog"
	"icenter/src/scene_server/admin_server/synchronizer/handler"
	"icenter/src/scene_server/admin_server/synchronizer/meta"
)

// AuthSynchronizer stores all related resource
type AuthSynchronizer struct {
	AuthConfig  authcenter.AuthConfig
	clientSet   apimachinery.ClientSetInterface
	ctx         context.Context
	Workers     *[]Worker
	WorkerQueue chan meta.WorkRequest
	Producer    *Producer
}

// NewSynchronizer new a synchronizer object
func NewSynchronizer(ctx context.Context, authConfig *authcenter.AuthConfig, clientSet apimachinery.ClientSetInterface) *AuthSynchronizer {
	return &AuthSynchronizer{ctx: ctx, AuthConfig: *authConfig, clientSet: clientSet}
}

// Run do start synchronize
func (d *AuthSynchronizer) Run() error {
	if d.AuthConfig.Enable == false {
		blog.Info("authConfig is disabled, exit now")
		return nil
	}

	blog.Infof("auth synchronize start...")

	// init queue
	d.WorkerQueue = make(chan meta.WorkRequest, 1000)

	// make fake handler
	blog.Infof("new auth client with config: %+v", d.AuthConfig)
	authorize, err := auth.NewAuthorize(nil, d.AuthConfig)
	if err != nil {
		blog.Errorf("new auth client failed, err: %+v", err)
		return fmt.Errorf("new auth client failed, err: %+v", err)
	}
	authManager := extensions.NewAuthManager(d.clientSet, authorize)
	workerHandler := handler.NewIAMHandler(d.clientSet, authManager)

	// init worker
	workers := make([]Worker, 3)
	for w := 1; w <= 3; w++ {
		worker := NewWorker(w, d.WorkerQueue, workerHandler)
		workers = append(workers, *worker)
		worker.Start()
	}
	d.Workers = &workers

	// init producer
	d.Producer = NewProducer(d.clientSet, authManager, d.WorkerQueue)
	d.Producer.Start()
	blog.Infof("auth synchronize started")
	return nil
}
