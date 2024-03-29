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

package registerdiscover

import (
	"context"
	"fmt"
	"icenter/src/common/blog"
	"sort"
	"strconv"
	"time"

	"icenter/src/common/backbone/service_mange/zk"
	"icenter/src/common/zkclient"

	gozk "github.com/samuel/go-zookeeper/zk"
)

// ZkRegDiscv do register and discover by zookeeper
type ZkRegDiscv struct {
	zkcli          *zkclient.ZkClient
	cancel         context.CancelFunc
	rootCxt        context.Context
	sessionTimeOut time.Duration
}

// NewZkRegDiscv create a object of ZkRegDiscv
func NewZkRegDiscv(client *zk.ZkClient) *ZkRegDiscv {
	ctx, ctxCancel := client.WithCancel()
	return &ZkRegDiscv{
		zkcli:          client.Client(),
		sessionTimeOut: client.SessionTimeOut(),
		cancel:         ctxCancel,
		rootCxt:        ctx,
	}
}

// RegisterAndWatch create ephemeral node for the service and watch it. if it exit, register again
func (zkRD *ZkRegDiscv) RegisterAndWatch(path string, data []byte) error {
	blog.Infof("register server and watch it. path(%s), data(%s)", path, string(data))
	go func() {
		var registerPath string
		var watchEvn <-chan gozk.Event
		var err error

		timer1 := time.NewTicker(5 * time.Second)
		watchCtx := zkRD.rootCxt
		for {
			select {
			case <-timer1.C:
				if len(registerPath) > 0 {
					// watchEvn will let us known any change on registerPath
					continue
				}

				registerPath, err = zkRD.zkcli.CreateEphAndSeqEx(path, data)
				if err != nil {
					blog.Errorf("fail to register server node(%s). err:%s", path, err.Error())
					continue
				}
				_, _, watchEvn, err = zkRD.zkcli.ExistW(registerPath)
				if err != nil {
					blog.Errorf("fail to watch register node(%s), err:%s\n", registerPath, err.Error())

					// clear register path, so that it can register to a new path
					zkRD.zkcli.Del(registerPath, -1)
					registerPath = ""
					continue
				}
			case <-watchCtx.Done():
				blog.Infof("watch register node(%s) done, now exist service register.\n", path)
				zkRD.zkcli.Del(registerPath, -1)
				return
			case <-watchEvn:
				blog.Infof("watch register node(%s) exist changed, event(%v)\n", path, watchEvn)
				zkRD.zkcli.Del(registerPath, -1)
				registerPath = ""
				continue
			}
		}
	}()

	blog.Infof("finish register server node(%s) and watch it\n", path)
	return nil
}

// GetServNodes get server nodes by path
func (zkRD *ZkRegDiscv) GetServNodes(path string) ([]string, error) {
	return zkRD.zkcli.GetChildren(path)
}

// Ping to ping server
func (zkRD *ZkRegDiscv) Ping() error {
	return zkRD.zkcli.Ping()
}

// Discover watch the children
func (zkRD *ZkRegDiscv) Discover(path string) (<-chan *DiscoverEvent, error) {
	fmt.Printf("begin to discover by watch children of path(%s)\n", path)
	discvCtx := zkRD.rootCxt

	env := make(chan *DiscoverEvent, 1)

	go zkRD.loopDiscover(discvCtx, path, env)

	return env, nil
}

func (zkRD *ZkRegDiscv) loopDiscover(discvCtx context.Context, path string, env chan *DiscoverEvent) {
	for {
		_, watchEnv, err := zkRD.zkcli.WatchChildren(path)
		if err != nil {
			fmt.Printf("fail to watch children for path(%s), err:%s\n", path, err.Error())
			if zkclient.ErrNoNode == err {
				fmt.Printf("children node(%s) is not exist, will watch after 5s\n", path)
				time.Sleep(5 * time.Second)
				continue
			}

			fmt.Printf("unknow err accur when watch children:%v,will watch after 10s", err)
			time.Sleep(10 * time.Second)
			continue
		}

		// write into discoverEvent channel
		env <- zkRD.getServerInfoByPath(path)

		select {
		case <-discvCtx.Done():
			fmt.Printf("discover path(%s) done\n", path)
			return
		case <-watchEnv:
			fmt.Printf("watch found the children of path(%s) change\n", path)
		}
	}
}

func (zkRD *ZkRegDiscv) getServerInfoByPath(path string) *DiscoverEvent {
	for {
		discvEnv := &DiscoverEvent{
			Err: nil,
			Key: path,
		}
		servNodes, err := zkRD.zkcli.GetChildren(path)
		if err != nil {
			if zkclient.ErrNoNode == err {
				fmt.Printf("children node(%s) is not exist, will watch after 5s\n", path)
				time.Sleep(5 * time.Second)
				continue
			}
			if err == zkclient.ErrConnectionClosed {
				err = zkRD.zkcli.ConnectEx(zkRD.sessionTimeOut)
				fmt.Printf("connect zookeeper error:%v,will try connect after 10s", err)
				time.Sleep(10 * time.Second)
				continue
			}
			fmt.Printf("get children node(%s) error:%v,will watch after 10s", path, err)
			time.Sleep(10 * time.Second)
			continue
		}
		discvEnv.Nodes = append(discvEnv.Nodes, servNodes...)
		// sort server node
		servNodes = zkRD.sortNode(servNodes)

		isGetNodeInfoErr := false
		// get server info
		for _, node := range servNodes {
			servPath := path + "/" + node
			servInfo, err := zkRD.zkcli.Get(servPath)
			if err != nil {
				if err == zkclient.ErrNodeExists {
					continue
				}
				if err == zkclient.ErrConnectionClosed {
					err = zkRD.zkcli.ConnectEx(zkRD.sessionTimeOut)
					fmt.Printf("connect zookeeper error:%v,will try connect after 5s", err)
					time.Sleep(5 * time.Second)
					isGetNodeInfoErr = true
					break
				}
				isGetNodeInfoErr = true
				fmt.Printf("fail to get server info from zookeeper by path(%s), err:%s\n", servPath, err.Error())
				break
			}

			discvEnv.Server = append(discvEnv.Server, servInfo)
		}
		if isGetNodeInfoErr {
			continue
		}

		return discvEnv
	}

}

func (zkRD *ZkRegDiscv) sortNode(nodes []string) []string {
	var sortPart []int
	mapSortNode := make(map[int]string)
	for _, chNode := range nodes {
		if len(chNode) <= 10 {
			fmt.Printf("node(%s) is less then 10, there is not the seq number\n", chNode)
			continue
		}

		p, err := strconv.Atoi(chNode[len(chNode)-10:])
		if err != nil {
			fmt.Printf("fail to conv string to seq number for node(%s), err:%s\n", chNode, err.Error())
			continue
		}

		sortPart = append(sortPart, p)
		mapSortNode[p] = chNode
	}

	sort.Ints(sortPart)

	var children []string
	for _, part := range sortPart {
		children = append(children, mapSortNode[part])
	}

	return children
}
