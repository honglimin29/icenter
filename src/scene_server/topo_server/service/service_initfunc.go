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

package service

import (
	"net/http"

	"icenter/src/common"
)

func (s *Service) initHealth() {
	s.addAction(http.MethodGet, "/healthz", s.Health, nil)
}

func (s *Service) initAssociation() {

	// mainline topo methods
	s.addAction(http.MethodPost, "/topo/model/mainline", s.CreateMainLineObject, nil)
	s.addAction(http.MethodDelete, "/topo/model/mainline/owners/{owner_id}/objectids/{bk_obj_id}", s.DeleteMainLineObject, nil)
	s.addAction(http.MethodGet, "/topo/model/{owner_id}", s.SearchMainLineObjectTopo, nil)
	s.addAction(http.MethodGet, "/topo/model/{owner_id}/{cls_id}/{bk_obj_id}", s.SearchObjectByClassificationID, nil)
	s.addAction(http.MethodGet, "/topo/inst/{owner_id}/{bk_biz_id}", s.SearchBusinessTopo, nil)
	// TODO: delete this api, it's not used by front.
	s.addAction(http.MethodGet, "/topo/inst/child/{owner_id}/{obj_id}/{app_id}/{inst_id}", s.SearchMainLineChildInstTopo, nil)

	// association type methods
	s.addAction(http.MethodPost, "/topo/association/type/action/search/batch", s.SearchObjectAssoWithAssoKindList, nil)
	s.addAction(http.MethodPost, "/topo/association/type/action/search", s.SearchAssociationType, nil)
	s.addAction(http.MethodPost, "/topo/association/type/action/create", s.CreateAssociationType, nil)
	s.addAction(http.MethodPut, "/topo/association/type/{id}/action/update", s.UpdateAssociationType, nil)
	s.addAction(http.MethodDelete, "/topo/association/type/{id}/action/delete", s.DeleteAssociationType, nil)

	// object association methods
	s.addAction(http.MethodPost, "/object/association/action/search", s.SearchObjectAssociation, nil)
	s.addAction(http.MethodPost, "/object/association/action/create", s.CreateObjectAssociation, nil)
	s.addAction(http.MethodPut, "/object/association/{id}/action/update", s.UpdateObjectAssociation, nil)
	s.addAction(http.MethodDelete, "/object/association/{id}/action/delete", s.DeleteObjectAssociation, nil)

	// inst association methods
	s.addAction(http.MethodPost, "/inst/association/action/search", s.SearchAssociationInst, nil)
	s.addAction(http.MethodPost, "/inst/association/action/create", s.CreateAssociationInst, nil)
	s.addAction(http.MethodDelete, "/inst/association/{association_id}/action/delete", s.DeleteAssociationInst, nil)

	// topo search methods
	s.addAction(http.MethodPost, "/inst/association/search/owner/{owner_id}/object/{bk_obj_id}", s.SearchInstByAssociation, nil)
	s.addAction(http.MethodPost, "/inst/association/topo/search/owner/{owner_id}/object/{bk_obj_id}/inst/{inst_id}", s.SearchInstTopo, nil)

	// ATTENTION: the following methods is not recommended
	s.addAction(http.MethodPost, "/inst/search/topo/owner/{owner_id}/object/{bk_object_id}/inst/{inst_id}", s.SearchInstChildTopo, nil)
	s.addAction(http.MethodPost, "/inst/association/action/{bk_obj_id}/import", s.ImportInstanceAssociation, nil)

}

func (s *Service) initAuditLog() {

	s.addAction(http.MethodPost, "/audit/search", s.AuditQuery, nil)
	s.addAction(http.MethodPost, "/object/{bk_obj_id}/audit/search", s.InstanceAuditQuery, nil)
}

func (s *Service) initCompatiblev2() {
	s.addAction(http.MethodPost, "/app/searchAll", s.SearchAllApp, nil)

	s.addAction(http.MethodPut, "/openapi/set/multi/{appid}", s.UpdateMultiSet, nil)
	s.addAction(http.MethodDelete, "/openapi/set/multi/{appid}", s.DeleteMultiSet, nil)
	s.addAction(http.MethodDelete, "/openapi/set/setHost/{appid}", s.DeleteSetHost, nil)

	s.addAction(http.MethodPut, "/openapi/module/multi/{"+common.BKAppIDField+"}", s.UpdateMultiModule, nil)
	s.addAction(http.MethodPost, "/openapi/module/searchByApp/{"+common.BKAppIDField+"}", s.SearchModuleByApp, nil)
	s.addAction(http.MethodPost, "/openapi/module/searchByProperty/{"+common.BKAppIDField+"}", s.SearchModuleBySetProperty, nil)
	s.addAction(http.MethodPost, "/openapi/module/multi", s.AddMultiModule, nil)
	s.addAction(http.MethodDelete, "/openapi/module/multi/{"+common.BKAppIDField+"}", s.DeleteMultiModule, nil)

}

func (s *Service) initBusiness() {
	s.addAction(http.MethodPost, "/app/{owner_id}", s.CreateBusiness, nil)
	s.addAction(http.MethodDelete, "/app/{owner_id}/{app_id}", s.DeleteBusiness, nil)
	s.addAction(http.MethodPut, "/app/{owner_id}/{app_id}", s.UpdateBusiness, nil)
	s.addAction(http.MethodPut, "/app/status/{flag}/{owner_id}/{app_id}", s.UpdateBusinessStatus, nil)
	s.addAction(http.MethodPost, "/app/search/{owner_id}", s.SearchBusiness, nil)
	s.addAction(http.MethodPost, "/app/default/{owner_id}/search", s.SearchArchivedBusiness, nil)
	s.addAction(http.MethodPost, "/app/default/{owner_id}", s.CreateDefaultBusiness, nil)
	s.addAction(http.MethodGet, "/topo/internal/{owner_id}/{app_id}", s.GetInternalModule, nil)
	// find reduced business list with only few fields for business itself.
	s.addAction(http.MethodGet, "/app/with_reduced", s.SearchReducedBusinessList, nil)

}

func (s *Service) initModule() {
	s.addAction(http.MethodPost, "/module/{app_id}/{set_id}", s.CreateModule, nil)
	s.addAction(http.MethodDelete, "/module/{app_id}/{set_id}/{module_id}", s.DeleteModule, nil)
	s.addAction(http.MethodPut, "/module/{app_id}/{set_id}/{module_id}", s.UpdateModule, nil)
	s.addAction(http.MethodPost, "/module/search/{owner_id}/{app_id}/{set_id}", s.SearchModule, nil)

}

func (s *Service) initSet() {
	s.addAction(http.MethodPost, "/set/{app_id}", s.CreateSet, nil)
	s.addAction(http.MethodDelete, "/set/{app_id}/{set_id}", s.DeleteSet, nil)
	s.addAction(http.MethodDelete, "/set/{app_id}/batch", s.DeleteSets, nil)
	s.addAction(http.MethodPut, "/set/{app_id}/{set_id}", s.UpdateSet, nil)
	s.addAction(http.MethodPost, "/set/search/{owner_id}/{app_id}", s.SearchSet, nil)

}

func (s *Service) initInst() {
	s.addAction(http.MethodPost, "/inst/{owner_id}/{bk_obj_id}", s.CreateInst, nil)
	s.addAction(http.MethodDelete, "/inst/{owner_id}/{bk_obj_id}/{inst_id}", s.DeleteInst, nil)
	s.addAction(http.MethodDelete, "/inst/{owner_id}/{bk_obj_id}/batch", s.DeleteInsts, nil)
	s.addAction(http.MethodPut, "/inst/{owner_id}/{bk_obj_id}/{inst_id}", s.UpdateInst, nil)
	s.addAction(http.MethodPut, "/inst/{owner_id}/{bk_obj_id}/batch/update", s.UpdateInsts, nil)
	s.addAction(http.MethodPost, "/inst/search/{owner_id}/{bk_obj_id}", s.SearchInsts, nil)
	s.addAction(http.MethodPost, "/inst/search/owner/{owner_id}/object/{bk_obj_id}/detail", s.SearchInstAndAssociationDetail, nil)
	s.addAction(http.MethodPost, "/inst/search/owner/{owner_id}/object/{bk_obj_id}", s.SearchInstByObject, nil)
	s.addAction(http.MethodPost, "/inst/search/{owner_id}/{bk_obj_id}/{inst_id}", s.SearchInstByInstID, nil)

}

func (s *Service) initObjectAttribute() {
	s.addAction(http.MethodPost, "/objectattr", s.CreateObjectAttribute, nil)
	s.addAction(http.MethodPost, "/objectattr/search", s.SearchObjectAttribute, nil)
	s.addAction(http.MethodPut, "/objectattr/{id}", s.UpdateObjectAttribute, nil)
	s.addAction(http.MethodDelete, "/objectattr/{id}", s.DeleteObjectAttribute, nil)
}

func (s *Service) initObjectClassification() {
	s.addAction(http.MethodPost, "/object/classification", s.CreateClassification, nil)
	s.addAction(http.MethodPost, "/object/classification/{owner_id}/objects", s.SearchClassificationWithObjects, nil)
	s.addAction(http.MethodPost, "/object/classifications", s.SearchClassification, nil)
	s.addAction(http.MethodPut, "/object/classification/{id}", s.UpdateClassification, nil)
	s.addAction(http.MethodDelete, "/object/classification/{id}", s.DeleteClassification, nil)
}

func (s *Service) initObjectObjectUnique() {
	s.addAction(http.MethodPost, "/object/{bk_obj_id}/unique/action/create", s.CreateObjectUnique, nil)
	s.addAction(http.MethodPut, "/object/{bk_obj_id}/unique/{id}/action/update", s.UpdateObjectUnique, nil)
	s.addAction(http.MethodDelete, "/object/{bk_obj_id}/unique/{id}/action/delete", s.DeleteObjectUnique, nil)
	s.addAction(http.MethodGet, "/object/{bk_obj_id}/unique/action/search", s.SearchObjectUnique, nil)
}

func (s *Service) initObjectGroup() {
	s.addAction(http.MethodPost, "/objectatt/group/new", s.CreateObjectGroup, nil)
	s.addAction(http.MethodPut, "/objectatt/group/update", s.UpdateObjectGroup, nil)
	s.addAction(http.MethodDelete, "/objectatt/group/groupid/{id}", s.DeleteObjectGroup, nil)
	s.addAction(http.MethodPut, "/objectatt/group/property", s.UpdateObjectAttributeGroupProperty, s.ParseUpdateObjectAttributeGroupPropertyInput)
	s.addAction(http.MethodDelete, "/objectatt/group/owner/{owner_id}/object/{bk_object_id}/propertyids/{property_id}/groupids/{group_id}", s.DeleteObjectAttributeGroup, nil)
	s.addAction(http.MethodPost, "/objectatt/group/property/owner/{owner_id}/object/{bk_obj_id}", s.SearchGroupByObject, nil)
}

func (s *Service) initObject() {
	s.addAction(http.MethodPost, "/object/batch", s.CreateObjectBatch, nil)
	s.addAction(http.MethodPost, "/object/search/batch", s.SearchObjectBatch, nil)
	s.addAction(http.MethodPost, "/object", s.CreateObject, nil)
	s.addAction(http.MethodPost, "/objects", s.SearchObject, nil)
	s.addAction(http.MethodPost, "/objects/topo", s.SearchObjectTopo, nil)
	s.addAction(http.MethodPut, "/object/{id}", s.UpdateObject, nil)
	s.addAction(http.MethodDelete, "/object/{id}", s.DeleteObject, nil)

}
func (s *Service) initPrivilegeGroup() {
	s.addAction(http.MethodPost, "/topo/privilege/group/{bk_supplier_account}", s.CreateUserGroup, nil)
	s.addAction(http.MethodDelete, "/topo/privilege/group/{bk_supplier_account}/{group_id}", s.DeleteUserGroup, nil)
	s.addAction(http.MethodPut, "/topo/privilege/group/{bk_supplier_account}/{group_id}", s.UpdateUserGroup, nil)
	s.addAction(http.MethodPost, "/topo/privilege/group/{bk_supplier_account}/search", s.SearchUserGroup, nil)
}

func (s *Service) initPrivilegeRole() {
	s.addAction(http.MethodPost, "/topo/privilege/{bk_supplier_account}/{bk_obj_id}/{bk_property_id}", s.CreatePrivilege, s.ParseCreateRolePrivilegeOriginData)
	s.addAction(http.MethodGet, "/topo/privilege/{bk_supplier_account}/{bk_obj_id}/{bk_property_id}", s.GetPrivilege, nil)
}

func (s *Service) initPrivilege() {
	s.addAction(http.MethodPost, "/topo/privilege/group/detail/{bk_supplier_account}/{group_id}", s.UpdateUserGroupPrivi, nil)
	s.addAction(http.MethodGet, "/topo/privilege/group/detail/{bk_supplier_account}/{group_id}", s.GetUserGroupPrivi, nil)
	s.addAction(http.MethodGet, "/topo/privilege/user/detail/{bk_supplier_account}/{user_name}", s.GetUserPrivi, nil)
}

func (s *Service) initGraphics() {
	s.addAction(http.MethodPost, "/objects/topographics/scope_type/{scope_type}/scope_id/{scope_id}/action/search", s.SelectObjectTopoGraphics, nil)
	s.addPublicAction(http.MethodPost, "/objects/topographics/scope_type/{scope_type}/scope_id/{scope_id}/action/update", s.UpdateObjectTopoGraphics, s.ParseOriginGraphicsUpdateInput)
}
func (s *Service) initIdentifier() {
	s.addAction(http.MethodPost, "/identifier/{obj_type}/search", s.SearchIdentifier, s.ParseSearchIdentifierOriginData)
}

func (s *Service) initService() {
	s.initHealth()
	s.initAssociation()
	s.initAuditLog()
	s.initCompatiblev2()
	s.initBusiness()
	s.initInst()
	s.initModule()
	s.initSet()
	s.initObject()
	s.initObjectAttribute()
	s.initObjectClassification()
	s.initObjectGroup()
	s.initPrivilegeGroup()
	s.initPrivilegeRole()
	s.initPrivilege()
	s.initGraphics()
	s.initIdentifier()
	s.initObjectObjectUnique()

	s.initBusinessObject()
	s.initBusinessClassification()
	s.initBusinessObjectAttribute()
	s.initBusinessObjectUnique()
	s.initBusinessObjectAttrGroup()
	s.initBusinessAssociation()
	s.initBusinessGraphics()
	s.initBusinessInst()

}
