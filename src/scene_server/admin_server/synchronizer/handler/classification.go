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

package handler

import (
	"context"
	"fmt"

	"icenter/src/auth/extensions"
	authmeta "icenter/src/auth/meta"
	"icenter/src/common/blog"
	"icenter/src/scene_server/admin_server/synchronizer/meta"
	"icenter/src/scene_server/admin_server/synchronizer/utils"
)

// HandleModuleSync do sync module of one business
func (ih *IAMHandler) HandleClassificationSync(task *meta.WorkRequest) error {
	businessSimplify := task.Data.(extensions.BusinessSimplify)
	header := utils.NewAPIHeaderByBusiness(&businessSimplify)

	// step1 get instances by business from core service
	bizID := businessSimplify.BKAppIDField
	classifications, err := ih.authManager.CollectClassificationByBusinessIDs(context.Background(), *header, bizID)
	if err != nil {
		blog.Errorf("collect classifications by business id failed, err: %+v", err)
		return err
	}
	if len(classifications) == 0 {
		blog.Infof("no classifications found for business: %d", bizID)
		return nil
	}
	resources := ih.authManager.MakeResourcesByClassifications(*header, authmeta.EmptyAction, bizID, classifications...)
	if len(resources) == 0 && len(classifications) > 0 {
		blog.Errorf("make iam resource for classifications %+v return empty", classifications)
		return nil
	}

	// step2 get classifications by business from iam
	rs := &authmeta.ResourceAttribute{
		Basic: authmeta.Basic{
			Type: authmeta.ModelClassification,
		},
		SupplierAccount: "",
		BusinessID:      businessSimplify.BKAppIDField,
		Layers:          make([]authmeta.Item, 0),
	}

	taskName := fmt.Sprintf("sync classifications for business: %d", businessSimplify.BKAppIDField)
	iamIDPrefix := ""
	skipDeregister := false
	return ih.diffAndSync(taskName, rs, iamIDPrefix, resources, skipDeregister)
}
