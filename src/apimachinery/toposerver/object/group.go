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

package object

import (
	"context"
	"fmt"
	"net/http"

	"icenter/src/common/metadata"
)

func (t *object) CreatePropertyGroup(ctx context.Context, h http.Header, dat metadata.Group) (resp *metadata.Response, err error) {
	resp = new(metadata.Response)
	subPath := "/objectatt/group/new"

	err = t.client.Post().
		WithContext(ctx).
		Body(dat).
		SubResource(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
	return
}

func (t *object) UpdatePropertyGroup(ctx context.Context, h http.Header, cond *metadata.PropertyGroupCondition) (resp *metadata.Response, err error) {
	resp = new(metadata.Response)
	subPath := "/objectatt/group/update"

	err = t.client.Put().
		WithContext(ctx).
		Body(cond).
		SubResource(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
	return
}

func (t *object) DeletePropertyGroup(ctx context.Context, groupID string, h http.Header) (resp *metadata.Response, err error) {
	resp = new(metadata.Response)
	subPath := fmt.Sprintf("/objectatt/group/groupid/%s", groupID)

	err = t.client.Delete().
		WithContext(ctx).
		Body(nil).
		SubResource(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
	return
}

func (t *object) UpdatePropertyGroupObjectAtt(ctx context.Context, h http.Header, data metadata.PropertyGroupObjectAtt) (resp *metadata.Response, err error) {
	resp = new(metadata.Response)
	subPath := "/objectatt/group/property"

	err = t.client.Put().
		WithContext(ctx).
		Body(data).
		SubResource(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
	return
}

func (t *object) DeletePropertyGroupObjectAtt(ctx context.Context, ownerID string, objID string, propertyID string, groupID string, h http.Header) (resp *metadata.Response, err error) {
	resp = new(metadata.Response)
	subPath := fmt.Sprintf("/objectatt/group/owner/%s/object/%s/propertyids/%s/groupids/%s", ownerID, objID, propertyID, groupID)

	err = t.client.Delete().
		WithContext(ctx).
		Body(nil).
		SubResource(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
	return
}

func (t *object) SelectPropertyGroupByObjectID(ctx context.Context, ownerID string, objID string, h http.Header, data map[string]interface{}) (resp *metadata.Response, err error) {
	resp = new(metadata.Response)
	subPath := fmt.Sprintf("/objectatt/group/property/owner/%s/object/%s", ownerID, objID)

	err = t.client.Post().
		WithContext(ctx).
		Body(data).
		SubResource(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
	return
}
