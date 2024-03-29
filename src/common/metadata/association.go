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

package metadata

import (
	"icenter/src/common/mapstr"
)

const (
	// AssociationFieldObjectID the association data field definition
	AssociationFieldObjectID = "bk_obj_id"
	// AssociationFieldAsstID the association data field bk_obj_asst_id
	AssociationFieldAsstID = "bk_obj_asst_id"
	// AssociationFieldObjectAttributeID the association data field definition
	//AssociationFieldObjectAttributeID = "bk_object_att_id"
	// AssociationFieldSupplierAccount the association data field definition
	AssociationFieldSupplierAccount = "bk_supplier_account"
	// AssociationFieldAssociationForward the association data field definition
	// AssociationFieldAssociationForward = "bk_asst_forward"
	// AssociationFieldAssociationObjectID the association data field definition
	AssociationFieldAssociationObjectID = "bk_asst_obj_id"
	// AssociationFieldAssociationName the association data field definition
	// AssociationFieldAssociationName = "bk_asst_name"
	// AssociationFieldAssociationId auto incr id
	AssociationFieldAssociationId   = "id"
	AssociationFieldAssociationKind = "bk_asst_id"
)

type SearchAssociationTypeRequest struct {
	BasePage  `json:"page"`
	Condition map[string]interface{} `json:"condition"`
}

type SearchAssociationTypeResult struct {
	BaseResp `json:",inline"`
	Data     struct {
		Count int                `json:"count"`
		Info  []*AssociationKind `json:"info"`
	} `json:"data"`
}

type CreateAssociationTypeResult struct {
	BaseResp `json:",inline"`
	Data     RspID `json:"data"`
}

type UpdateAssociationTypeRequest struct {
	AsstName  string `field:"bk_asst_name" json:"bk_asst_name" bson:"bk_asst_name"`
	SrcDes    string `field:"src_des" json:"src_des" bson:"src_des"`
	DestDes   string `field:"dest_des" json:"dest_des" bson:"dest_des"`
	Direction string `field:"direction" json:"direction" bson:"direction"`
}

type UpdateAssociationTypeResult struct {
	BaseResp `json:",inline"`
	Data     string `json:"data"`
}

type DeleteAssociationTypeResult struct {
	BaseResp `json:",inline"`
	Data     string `json:"data"`
}

type SearchAssociationObjectRequest struct {
	Condition mapstr.MapStr `json:"condition"`
}

type SearchAssociationObjectResult struct {
	BaseResp `json:",inline"`
	Data     []*Association `json:"data"`
}

type CreateAssociationObjectResult struct {
	BaseResp `json:",inline"`
	Data     RspID `json:"data"`
}

type UpdateAssociationObjectRequest struct {
	AsstName string `field:"bk_asst_name" json:"bk_asst_name" bson:"bk_asst_name"`
}

type UpdateAssociationObjectResult struct {
	BaseResp `json:",inline"`
	Data     string `json:"data"`
}

type DeleteAssociationObjectResult struct {
	BaseResp `json:",inline"`
	Data     string `json:"data"`
}

type SearchAssociationInstRequestCond struct {
	ObjectAsstID string  `field:"bk_obj_asst_id" json:"bk_obj_asst_id,omitempty" bson:"bk_obj_asst_id,omitempty"`
	AsstID       string  `field:"bk_asst_id" json:"bk_asst_id,omitempty" bson:"bk_asst_id,omitempty"`
	ObjectID     string  `field:"bk_obj_id" json:"bk_obj_id,omitempty" bson:"bk_obj_id,omitempty"`
	AsstObjID    string  `field:"bk_asst_obj_id" json:"bk_asst_obj_id,omitempty" bson:"bk_asst_obj_id,omitempty"`
	InstID       []int64 `field:"bk_inst_id" json:"bk_inst_id,omitempty" bson:"bk_inst_id,omitempty"`
	AsstInstID   []int64 `field:"bk_asst_inst_id" json:"bk_asst_inst_id,omitempty" bson:"bk_asst_inst_id,omitempty"`
	BothObjectID string  `field:"both_obj_id" json:"both_obj_id" bson:"both_obj_id"`
	BothInstID   []int64 `field:"both_inst_id" json:"both_inst_id" bson:"both_inst_id"`
}

type SearchAssociationInstRequest struct {
	Condition mapstr.MapStr `json:"condition"` // construct condition mapstr by condition.Condition
}

type SearchAssociationInstResult struct {
	BaseResp `json:",inline"`
	Data     []*InstAsst `json:"data"`
}

type CreateAssociationInstRequest struct {
	ObjectAsstID string `field:"bk_obj_asst_id" json:"bk_obj_asst_id,omitempty" bson:"bk_obj_asst_id,omitempty"`
	InstID       int64  `field:"bk_inst_id" json:"bk_inst_id,omitempty" bson:"bk_inst_id,omitempty"`
	AsstInstID   int64  `field:"bk_asst_inst_id" json:"bk_asst_inst_id,omitempty" bson:"bk_asst_inst_id,omitempty"`
}
type CreateAssociationInstResult struct {
	BaseResp `json:",inline"`
	Data     RspID `json:"data"`
}

type DeleteAssociationInstRequest struct {
	Condition mapstr.MapStr `json:"condition"`
}

type DeleteAssociationInstResult struct {
	BaseResp `json:",inline"`
	Data     string `json:"data"`
}

type AssociationKindIDs struct {
	// the association kind ids.
	AsstIDs []string `json:"asst_ids"`
}

type ListAssociationsWithAssociationKindResult struct {
	BaseResp `json:",inline"`
	Data     AssociationList `json:"data"`
}

type AssociationList struct {
	Associations []AssociationDetail `json:"associations"`
}

type AssociationDetail struct {
	// the ID of the association kind.
	AssociationKindID string        `json:"bk_asst_id"`
	Associations      []Association `json:"assts"`
}

// 关联类型
type AssociationDirection string

const (
	NoneDirection       AssociationDirection = "none"
	DestinationToSource AssociationDirection = "src_to_dest"
	SourceToDestination AssociationDirection = "dest_to_src"
	Bidirectional       AssociationDirection = "bidirectional"
)

type AssociationKind struct {
	ID int64 `field:"id" json:"id" bson:"id"`
	// a unique association id created by user.
	AssociationKindID string `field:"bk_asst_id" json:"bk_asst_id" bson:"bk_asst_id"`
	// a memorable name for this association kind, could be a chinese name, a english name etc.
	AssociationKindName string `field:"bk_asst_name" json:"bk_asst_name" bson:"bk_asst_name"`
	// the owner that this association type belongs to.
	OwnerID string `field:"bk_supplier_account" json:"bk_supplier_account" bson:"bk_supplier_account"`
	// the describe for the relationship from source object to the target(destination) object, which will be displayed
	// when the topology is constructed between objects.
	SourceToDestinationNote string `field:"src_des" json:"src_des" bson:"src_des"`
	// the describe for the relationship from the target(destination) object to source object, which will be displayed
	// when the topology is constructed between objects.
	DestinationToSourceNote string `field:"dest_des" json:"dest_des" bson:"dest_des"`
	// the association direction between two objects.
	Direction AssociationDirection `field:"direction" json:"direction" bson:"direction"`
	// whether this is a pre-defined kind.
	IsPre *bool `field:"ispre" json:"ispre" bson:"ispre"`
	//	define the metadata of association kind
	Metadata `field:"metadata" json:"metadata" bson:"metadata"`
}

func (cli *AssociationKind) Parse(data mapstr.MapStr) (*AssociationKind, error) {
	// TODO support parse metadata params
	err := mapstr.SetValueToStructByTags(cli, data)
	if nil != err {
		return nil, err
	}

	return cli, err
}

type AssociationOnDeleteAction string
type AssociationMapping string

const (
	// this is a default action, which is do nothing when a association between object is deleted.
	NoAction AssociationOnDeleteAction = "none"
	// delete related source object instances when the association is deleted.
	DeleteSource AssociationOnDeleteAction = "delete_src"
	// delete related destination object instances when the association is deleted.
	DeleteDestinatioin AssociationOnDeleteAction = "delete_dest"

	// the source object can be related with only one destination object
	OneToOneMapping AssociationMapping = "1:1"
	// the source object can be related with multiple destination objects
	OneToManyMapping AssociationMapping = "1:n"
	// multiple source object can be related with multiple destination objects
	ManyToManyMapping AssociationMapping = "n:n"
)

// Association defines the association between two objects.
type Association struct {
	ID      int64  `field:"id" json:"id" bson:"id"`
	OwnerID string `field:"bk_supplier_account" json:"bk_supplier_account" bson:"bk_supplier_account"`

	// the unique id belongs to  this association, should be generated with rules as follows:
	// "$ObjectID"_"$AsstID"_"$AsstObjID"
	AssociationName string `field:"bk_obj_asst_id" json:"bk_obj_asst_id" bson:"bk_obj_asst_id"`
	// the alias name of this association, which is a substitute name in the association kind $AsstKindID
	AssociationAliasName string `field:"bk_obj_asst_name" json:"bk_obj_asst_name" bson:"bk_obj_asst_name"`

	// describe which object this association is defined for.
	ObjectID string `field:"bk_obj_id" json:"bk_obj_id" bson:"bk_obj_id"`
	// describe where the Object associate with.
	AsstObjID string `field:"bk_asst_obj_id" json:"bk_asst_obj_id" bson:"bk_asst_obj_id"`
	// the association kind used by this association.
	AsstKindID string `field:"bk_asst_id" json:"bk_asst_id" bson:"bk_asst_id"`

	// defined which kind of association can be used between the source object and destination object.
	Mapping AssociationMapping `field:"mapping" json:"mapping" bson:"mapping"`
	// describe the action when this association is deleted.
	OnDelete AssociationOnDeleteAction `field:"on_delete" json:"on_delete" bson:"on_delete"`
	// describe whether this association is a pre-defined association or not,
	// if true, it means this association is used by cmdb itself.
	IsPre *bool `field:"ispre" json:"ispre" bson:"ispre"`

	ClassificationID string `field:"bk_classification_id" json:"-" bson:"-"`
	ObjectIcon       string `field:"bk_obj_icon" json:"-" bson:"-"`
	ObjectName       string `field:"bk_obj_name" json:"-" bson:"-"`

	//	define the metadata of assocication
	Metadata `field:"metadata" json:"metadata" bson:"metadata"`
}

// return field means which filed is set but is forbidden to update.
func (a *Association) CanUpdate() (field string, can bool) {
	if a.ID != 0 {
		return "id", false
	}

	if len(a.OwnerID) != 0 {
		return "bk_supplier_account", false
	}

	if len(a.AssociationName) != 0 {
		return "bk_obj_asst_id", false
	}

	if len(a.ObjectID) != 0 {
		return "bk_obj_id", false
	}

	if len(a.AsstObjID) != 0 {
		return "bk_asst_obj_id", false
	}

	if len(a.Mapping) != 0 {
		return "mapping", false
	}

	if a.IsPre != nil {
		return "ispre", false
	}

	// only on delete, association kind id, alias name can be update.
	return "", true
}

// Parse load the data from mapstr attribute into attribute instance
func (cli *Association) Parse(data mapstr.MapStr) (*Association, error) {
	//TODO support parse metadata params
	err := mapstr.SetValueToStructByTags(cli, data)
	if nil != err {
		return nil, err
	}

	return cli, err
}

// ToMapStr to mapstr
func (cli *Association) ToMapStr() mapstr.MapStr {
	return mapstr.SetValueToMapStrByTags(cli)
}

// InstAsst an association definition between instances.
type InstAsst struct {
	// sequence ID
	ID int64 `field:"id" json:"id"`
	// inst id associate to ObjectID
	InstID int64 `field:"bk_inst_id" json:"bk_inst_id" bson:"bk_inst_id"`
	// association source ObjectID
	ObjectID string `field:"bk_obj_id" json:"bk_obj_id" bson:"bk_obj_id"`
	// inst id associate to AsstObjectID
	AsstInstID int64 `field:"bk_asst_inst_id" json:"bk_asst_inst_id"  bson:"bk_asst_inst_id"`
	// association target ObjectID
	AsstObjectID string `field:"bk_asst_obj_id" json:"bk_asst_obj_id" bson:"bk_asst_obj_id"`
	// bk_supplier_account
	OwnerID string `field:"bk_supplier_account" json:"bk_supplier_account" bson:"bk_supplier_account"`
	// association id between two object
	ObjectAsstID string `field:"bk_obj_asst_id" json:"bk_obj_asst_id" bson:"bk_obj_asst_id"`
	// association kind id
	AssociationKindID string `field:"bk_asst_id" json:"bk_asst_id" bson:"bk_asst_id"`

	//	define the metadata of assocication kind
	Metadata `field:"metadata" json:"metadata" bson:"metadata"`
}

func (asst InstAsst) GetInstID(objID string) (instID int64, ok bool) {
	switch objID {
	case asst.ObjectID:
		return asst.InstID, true
	case asst.AsstObjectID:
		return asst.AsstInstID, true
	default:
		return 0, false
	}
}

type InstNameAsst struct {
	ID         string `json:"id"`
	ObjID      string `json:"bk_obj_id"`
	ObjIcon    string `json:"bk_obj_icon"`
	InstID     int64  `json:"bk_inst_id"`
	ObjectName string `json:"bk_obj_name"`
	InstName   string `json:"bk_inst_name"`
	AssoID     int64  `json:"asso_id"`
	// AsstName   string                 `json:"bk_asst_name"`
	// AsstID   string                 `json:"bk_asst_id"`
	InstInfo map[string]interface{} `json:"inst_info,omitempty"`
}

// Parse load the data from mapstr attribute into attribute instance
func (cli *InstAsst) Parse(data mapstr.MapStr) (*InstAsst, error) {

	err := mapstr.SetValueToStructByTags(cli, data)
	if nil != err {
		return nil, err
	}

	return cli, err
}

// ToMapStr to mapstr
func (cli *InstAsst) ToMapStr() mapstr.MapStr {
	return mapstr.SetValueToMapStrByTags(cli)
}

// MainlineObjectTopo the mainline object topo
type MainlineObjectTopo struct {
	ObjID      string `field:"bk_obj_id" json:"bk_obj_id"`
	ObjName    string `field:"bk_obj_name" json:"bk_obj_name"`
	OwnerID    string `field:"bk_supplier_account" json:"bk_supplier_account"`
	NextObj    string `field:"bk_next_obj" json:"bk_next_obj"`
	NextName   string `field:"bk_next_name" json:"bk_next_name"`
	PreObjID   string `field:"bk_pre_obj_id" json:"bk_pre_obj_id"`
	PreObjName string `field:"bk_pre_obj_name" json:"bk_pre_obj_name"`
}

// Parse load the data from mapstr attribute into attribute instance
func (cli *MainlineObjectTopo) Parse(data mapstr.MapStr) (*MainlineObjectTopo, error) {

	err := mapstr.SetValueToStructByTags(cli, data)
	if nil != err {
		return nil, err
	}

	return cli, err
}

// ToMapStr to mapstr
func (cli *MainlineObjectTopo) ToMapStr() mapstr.MapStr {
	return mapstr.SetValueToMapStrByTags(cli)
}

// TopoInst 实例拓扑结构
type TopoInst struct {
	InstID   int64  `json:"bk_inst_id"`
	InstName string `json:"bk_inst_name"`
	ObjID    string `json:"bk_obj_id"`
	ObjName  string `json:"bk_obj_name"`
	Default  int    `json:"default"`
}

// TopoInstRst 拓扑实例
type TopoInstRst struct {
	TopoInst `json:",inline"`
	Child    []*TopoInstRst `json:"child"`
}

// ConditionItem subcondition
type ConditionItem struct {
	Field    string      `json:"field,omitempty"`
	Operator string      `json:"operator,omitempty"`
	Value    interface{} `json:"value,omitempty"`
}

// AssociationParams  association params
type AssociationParams struct {
	Page      BasePage                   `json:"page,omitempty"`
	Fields    map[string][]string        `json:"fields,omitempty"`
	Condition map[string][]ConditionItem `json:"condition,omitempty"`
}

// ResponeImportAssociation  import association result
type ResponeImportAssociation struct {
	BaseResp `json:",inline"`
	Data     ResponeImportAssociationData `json:"data"`
}

type RowMsgData struct {
	Row int    `json:"row"`
	Msg string `json:"message"`
}

type ResponeImportAssociationData struct {
	ErrMsgMap []RowMsgData `json:"err_msg"`
}

// ResponeImportAssociation  import association result
type RequestImportAssociation struct {
	AssociationInfoMap map[int]ExcelAssocation `json:"association_info"`
}
