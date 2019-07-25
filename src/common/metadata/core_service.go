/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.,
 * Copyright (C) 2017,-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the ",License",); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an ",AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package metadata

import (
	"fmt"

	"icenter/src/common"
	"icenter/src/common/blog"
	"icenter/src/common/mapstr"
	"icenter/src/common/util"
)

// CreateModelAttributeGroup used to create a new group for some attributes
type CreateModelAttributeGroup struct {
	Data Group `json:"data"`
}

// SetModelAttributeGroup used to create a new group for  some attributes, if it is exists, then update it
type SetModelAttributeGroup CreateModelAttributeGroup

// CreateManyModelClassifiaction create many input params
type CreateManyModelClassifiaction struct {
	Data []Classification `json:"datas"`
}

// CreateOneModelClassification create one model classification
type CreateOneModelClassification struct {
	Data Classification `json:"data"`
}

// SetManyModelClassification set many input params
type SetManyModelClassification CreateManyModelClassifiaction

// SetOneModelClassification set one input params
type SetOneModelClassification CreateOneModelClassification

// DeleteModelClassificationResult delete the model classification result
type DeleteModelClassificationResult struct {
	BaseResp `json:",inline"`
	Data     DeletedCount `json:"data"`
}

// CreateModel create model params
type CreateModel struct {
	Spec       Object      `json:"spec"`
	Attributes []Attribute `json:"attributes"`
}

// SetModel define SetMode method input params
type SetModel CreateModel

// SearchModelInfo search  model params
type SearchModelInfo struct {
	Spec       Object      `json:"spec"`
	Attributes []Attribute `json:"attributes"`
}

// CreateModelAttributes create model attributes
type CreateModelAttributes struct {
	Attributes []Attribute `json:"attributes"`
}

type SetModelAttributes CreateModelAttributes

type CreateModelAttrUnique struct {
	Data ObjectUnique `json:"data"`
}

type UpdateModelAttrUnique struct {
	Data UpdateUniqueRequest `json:"data"`
}

type DeleteModelAttrUnique struct {
	Metadata `field:"metadata" json:"metadata" bson:"metadata"`
}

type CreateModelInstance struct {
	Data mapstr.MapStr `json:"data"`
}

type CreateManyModelInstance struct {
	Datas []mapstr.MapStr `json:"datas"`
}

type SetModelInstance CreateModelInstance
type SetManyModelInstance CreateManyModelInstance

type CreateAssociationKind struct {
	Data AssociationKind `json:"data"`
}

type CreateManyAssociationKind struct {
	Datas []AssociationKind `json:"datas"`
}
type SetAssociationKind CreateAssociationKind
type SetManyAssociationKind CreateManyAssociationKind

type CreateModelAssociation struct {
	Spec Association `json:"spec"`
}

type SetModelAssociation CreateModelAssociation

type CreateOneInstanceAssociation struct {
	Data InstAsst `json:"data"`
}
type CreateManyInstanceAssociation struct {
	Datas []InstAsst `json:"datas"`
}

type Dimension struct {
	AppID int64 `json:"bk_biz_id"`
}

type SetOneInstanceAssociation CreateOneInstanceAssociation
type SetManyInstanceAssociation CreateManyInstanceAssociation

type TopoModelNode struct {
	Children []*TopoModelNode
	ObjectID string
}

type SearchTopoModelNodeResult struct {
	BaseResp `json:",inline"`
	Data     TopoModelNode `json:"data"`
}

// LeftestObjectIDList extrac leftest node's id of each level, arrange as a list
// it's useful in model mainline topo case, as bk_mainline relationship degenerate to a list.
func (tn *TopoModelNode) LeftestObjectIDList() []string {
	objectIDs := make([]string, 0)
	node := tn
	for {
		objectIDs = append(objectIDs, node.ObjectID)
		if len(node.Children) == 0 {
			break
		}
		node = node.Children[0]
	}
	return objectIDs
}

type TopoInstanceNode struct {
	Children   []*TopoInstanceNode
	ObjectID   string
	InstanceID int64
	Detail     map[string]interface{}
}

type SearchTopoInstanceNodeResult struct {
	BaseResp `json:",inline"`
	Data     TopoInstanceNode `json:"data"`
}

func (node *TopoInstanceNode) Name() string {
	var name string
	var exist bool
	var val interface{}
	switch node.ObjectID {
	case common.BKInnerObjIDSet:
		val, exist = node.Detail[common.BKSetNameField]
	case common.BKInnerObjIDApp:
		val, exist = node.Detail[common.BKAppNameField]
	case common.BKInnerObjIDModule:
		val, exist = node.Detail[common.BKModuleNameField]
	default:
		val, exist = node.Detail[common.BKInstNameField]
	}

	if exist == true {
		name = util.GetStrByInterface(val)
	} else {
		blog.V(7).Infof("extract topo instance node:%+v name failed", *node)
		name = fmt.Sprintf("%s:%d", node.ObjectID, node.InstanceID)
	}
	return name
}

func (node *TopoInstanceNode) TraversalFindModule(targetID int64) []*TopoInstanceNode {
	// ex: module1 ==> reverse([bizID, mainline1, ..., mainline2, set1, module1])
	return node.TraversalFindNode(common.BKInnerObjIDModule, targetID)
}

func (node *TopoInstanceNode) TraversalFindNode(objectType string, targetID int64) []*TopoInstanceNode {
	if common.GetObjByType(node.ObjectID) == objectType && node.InstanceID == targetID {
		return []*TopoInstanceNode{node}
	}

	for _, child := range node.Children {
		path := child.TraversalFindNode(objectType, targetID)
		if len(path) > 0 {
			path = append(path, node)
			return path
		}
	}

	return []*TopoInstanceNode{}
}

type TopoInstance struct {
	ObjectID         string
	InstanceID       int64
	ParentInstanceID int64
	Detail           map[string]interface{}
	Default          int64
}

// Key generate a unique key for instance(as instances's of different object type maybe conflict)
func (ti *TopoInstance) Key() string {
	return fmt.Sprintf("%s:%d", ti.ObjectID, ti.InstanceID)
}

// TransferHostsCrossBusinessRequest Transfer host across business request parameter
type TransferHostsCrossBusinessRequest struct {
	SrcApplicationID int64   `json:"src_bk_biz_id"`
	DstApplicationID int64   `json:"dst_bk_biz_id"`
	HostIDArr        []int64 `json:"bk_host_id"`
	DstModuleIDArr   []int64 `json:"bk_module_ids"`
}

// HostModuleRelationRequest gethost module relation request parameter
type HostModuleRelationRequest struct {
	ApplicationID int64   `json:"bk_biz_id"`
	SetIDArr      []int64 `json:"bk_set_ids"`
	HostIDArr     []int64 `json:"bk_host_ids"`
	ModuleIDArr   []int64 `json:"bk_module_ids"`
}

// Empty empty struct
func (h *HostModuleRelationRequest) Empty() bool {
	if h.ApplicationID != 0 {
		return false
	}
	if len(h.SetIDArr) != 0 {
		return false
	}
	if len(h.ModuleIDArr) != 0 {
		return false
	}

	if len(h.HostIDArr) != 0 {
		return false
	}
	return true
}

// DeleteHostRequest delete host from application
type DeleteHostRequest struct {
	ApplicationID int64   `json:"bk_biz_id"`
	HostIDArr     []int64 `json:"bk_host_ids"`
}
