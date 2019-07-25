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

package api

import (
	"crypto/tls"
	"encoding/json"

	restful "github.com/emicklei/go-restful"
	redis "gopkg.in/redis.v5"
	"icenter/src/common"
	"icenter/src/common/blog"
	"icenter/src/common/conf"
	"icenter/src/common/core/cc/actions"
	"icenter/src/common/core/cc/config"
	"icenter/src/common/core/cc/wactions"
	"icenter/src/common/errors"
	cchttp "icenter/src/common/http"
	"icenter/src/common/http/httpserver"
	"icenter/src/common/http/httpserver/webserver"
	"icenter/src/common/language"
)

type APIRequest struct {
	AppID    string      `json:"appID"`
	Operator string      `json:"operator"`
	Request  interface{} `json:"request"`
}

type APIRsp struct {
	HTTPCode int         `json:"-"`
	Result   bool        `json:"result"`
	Code     int         `json:"bk_error_code"`
	Message  interface{} `json:"bk_error_msg"`
	Data     interface{} `json:"data"`
}

type BKAPIRsp struct {
	HTTPCode int         `json:"-"`
	Result   bool        `json:"result"`
	Code     int         `json:"bk_error_code"`
	Message  interface{} `json:"bk_error_msg"`
	Data     interface{} `json:"data"`
}

type APIResource struct {
	ConfigData   []byte
	Config       string
	URL          string
	IsCliSSL     bool
	CliTLS       *tls.Config
	Actions      []*httpserver.Action
	GlobalFilter func(req *restful.Request, resp *restful.Response, fchain *restful.FilterChain)
	Wactions     []*webserver.Action
	//MetaCli      storage.DI
	//InstCli      storage.DI
	CacheCli  *redis.Client
	Error     errors.CCErrorIf
	HostCtrl  func() string
	ObjCtrl   func() string
	ProcCtrl  func() string
	EventCtrl func() string
	AuditCtrl func() string
	HostAPI   func() string
	TopoAPI   func() string
	ProcAPI   func() string
	EventAPI  func() string
	APIAddr   func() string
	AddrSrv   AddrSrv
	Lang      language.CCLanguageIf
}

// AddrSrv get server address interface
type AddrSrv interface {
	GetServer(servType string) (string, error)
}

var api = APIResource{
	Config: "",
	URL:    "",
}

func GetAPIResource() *APIResource {
	return &api
}

func NewAPIResource() *APIResource {
	return &api
}

func (a *APIResource) SetConfig(conf *config.CCAPIConfig) {
	a.Config = conf.ExConfig
}

func (a *APIResource) IsClientSSL() bool {
	return a.IsCliSSL
}

func (a *APIResource) GetClientSSL() *tls.Config {
	return a.CliTLS
}

func (a *APIResource) InitAction() {
	apiAction := actions.GetAPIAction()
	a.Actions = append(a.Actions, apiAction...)
}

func (a *APIResource) InitWaction() {
	apiAction := wactions.GetAPIAction()
	a.Wactions = append(a.Wactions, apiAction...)
}

func (a *APIResource) PreProcess(data []byte) (string, error) {
	var req APIRequest
	if err := json.Unmarshal(data, &req); err != nil {
		blog.Errorf("fail to parse json, error:%s. data = %s.", err.Error(), string(data))
		return "", cchttp.InternalError(common.CC_ERR_Comm_JSON_DECODE, common.CC_ERR_Comm_JSON_DECODE_STR)
	}

	d, err := json.Marshal(req.Request)
	if err != nil {
		blog.Errorf("fail to encode json, error:%s", err.Error())
		return "", cchttp.InternalError(common.CC_ERR_Comm_JSON_ENCODE, common.CC_ERR_Comm_JSON_ENCODE_STR)
	}

	return string(d), nil
}

func (a *APIResource) ParseConfig() (map[string]string, error) {
	ccapiConfig := new(conf.Config)
	if "" != a.Config {
		ccapiConfig.InitConfig(a.Config)
	} else {
		ccapiConfig.ParseConf(a.ConfigData)
	}

	return ccapiConfig.Configmap, nil
}

func (a *APIResource) ParseConf(data []byte) (map[string]string, error) {
	ccapiConfig := new(conf.Config)
	ccapiConfig.ParseConf(data)
	a.ConfigData = data

	return ccapiConfig.Configmap, nil
}

// CreateAPIRspStr create api rsp str
func (a *APIResource) CreateAPIRspStr(errcode int, info interface{}) (string, error) {
	rsp := BKAPIRsp{
		Result:  true,
		Code:    0,
		Message: nil,
		Data:    nil,
	}

	if common.CCSuccess != errcode {
		rsp.Result = false
		rsp.Code = errcode
		rsp.Message = info
	} else {
		rsp.Message = common.CCSuccessStr
		rsp.Data = info
	}

	s, err := json.Marshal(rsp)

	return string(s), err
}

// CreateAPIRspErrStrWithData create api rsp str return errorno, errormsg, errdata
func (a *APIResource) CreateAPIRspErrStrWithData(errcode int, strmsg, errdata interface{}) (string, error) {
	rsp := BKAPIRsp{
		Result:  true,
		Code:    0,
		Message: nil,
		Data:    nil,
	}

	rsp.Result = false
	rsp.Code = errcode
	rsp.Message = strmsg
	rsp.Data = errdata

	s, err := json.Marshal(rsp)

	return string(s), err
}

func (a *APIResource) CreateBKAPIRspStr(errcode int, info interface{}) (string, error) {
	rsp := BKAPIRsp{
		Result:  true,
		Code:    0,
		Message: nil,
		Data:    nil,
	}

	if 0 != errcode {
		rsp.Result = false
		rsp.Code = errcode
		rsp.Message = info
	} else {
		rsp.Data = info
	}

	s, err := json.Marshal(rsp)

	return string(s), err
}

// RunAutoAction call the callback function when the server starts
func (a *APIResource) RunAutoAction(config map[string]string) error {
	autoAction := actions.GetAutoAction()
	chErr := make(chan error, len(autoAction))
	for _, a := range autoAction {
		action := a
		blog.Debug("Start excetion auto  action %s ", action.Name)
		go func() {
			err := action.Run(config)
			if err != nil {
				chErr <- err
			}
		}()
	}

	return <-chErr
}
