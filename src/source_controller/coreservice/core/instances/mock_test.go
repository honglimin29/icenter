/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.,
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the ",License",); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an ",AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package instances_test

import (
	"context"
	"testing"
	"time"

	"icenter/src/common/errors"
	"icenter/src/common/language"
	"icenter/src/common/metadata"
	"icenter/src/common/storage/dal/mongo"
	"icenter/src/common/storage/dal/mongo/local"
	"icenter/src/source_controller/coreservice/core"
	"icenter/src/source_controller/coreservice/core/instances"

	"github.com/stretchr/testify/require"
)

type mockDependences struct {
}

// IsInstanceExist used to check if the  instances  asst exist
func (s *mockDependences) IsInstAsstExist(ctx core.ContextParams, objID string, instID uint64) (exists bool, err error) {
	return false, nil
}

// DeleteInstAsst used to delete inst asst
func (s *mockDependences) DeleteInstAsst(ctx core.ContextParams, objID string, instID uint64) error {
	return nil
}

// SelectObjectAttWithParams select object att with params
func (s *mockDependences) SelectObjectAttWithParams(ctx core.ContextParams, objID string) (attribute []metadata.Attribute, err error) {
	return nil, nil
}

// SearchUnique search unique attribute
func (s *mockDependences) SearchUnique(ctx core.ContextParams, objID string) (uniqueAttr []metadata.ObjectUnique, err error) {
	return nil, nil
}

func newInstances(t *testing.T) core.InstanceOperation {

	db, err := local.NewMgo("mongodb://cc:cc@localhost:27010,localhost:27011,localhost:27012,localhost:27013/cmdb", time.Minute)
	require.NoError(t, err)
	return instances.New(db, &mockDependences{})
}

var defaultCtx = func() core.ContextParams {
	err, _ := errors.New("../../../../../resources/errors/")
	lan, _ := language.New("../../../../../resources/language/")
	return core.ContextParams{
		Context:         context.Background(),
		ReqID:           "test_req_id",
		SupplierAccount: "test_owner",
		User:            "test_user",
		Error:           err.CreateDefaultCCErrorIf("en"),
		Lang:            lan.CreateDefaultCCLanguageIf("en"),
	}
}()
