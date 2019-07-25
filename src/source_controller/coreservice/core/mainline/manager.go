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

package mainline

import (
	"context"
	"fmt"
	"strconv"

	"icenter/src/common"
	"icenter/src/common/blog"
	"icenter/src/common/mapstr"
	"icenter/src/common/metadata"
	"icenter/src/common/storage/dal"
	"icenter/src/common/universalsql/mongo"
	"icenter/src/common/util"
	"icenter/src/source_controller/coreservice/core"
)

type InstanceMainline struct {
	dbProxy   dal.RDB
	bkBizID   int64
	modelTree *metadata.TopoModelNode

	modelIDs        []string
	objectParentMap map[string]string

	businessInstances []mapstr.MapStr
	setInstances      []mapstr.MapStr
	moduleInstances   []mapstr.MapStr
	mainlineInstances []mapstr.MapStr

	instanceMap      map[string]*metadata.TopoInstance
	allTopoInstances []*metadata.TopoInstance

	root *metadata.TopoInstanceNode
}

func NewInstanceMainline(proxy dal.DB, bkBizID int64) (*InstanceMainline, error) {
	im := &InstanceMainline{
		dbProxy:           proxy,
		bkBizID:           bkBizID,
		objectParentMap:   map[string]string{},
		setInstances:      make([]mapstr.MapStr, 0),
		moduleInstances:   make([]mapstr.MapStr, 0),
		mainlineInstances: make([]mapstr.MapStr, 0),
		allTopoInstances:  make([]*metadata.TopoInstance, 0),
		instanceMap:       map[string]*metadata.TopoInstance{},
	}
	return im, nil
}

func (im *InstanceMainline) SetModelTree(modelTree *metadata.TopoModelNode) {
	// step1
	im.modelTree = modelTree
}

func (im *InstanceMainline) LoadModelParentMap() {
	// step2
	im.modelIDs = im.modelTree.LeftestObjectIDList()
	for idx, objectID := range im.modelIDs {
		if idx == 0 {
			continue
		}
		im.objectParentMap[objectID] = im.modelIDs[idx-1]
	}
	blog.V(5).Infof("LoadModelParentMap mainline models: %+v, objectParentMap: %+v", im.modelIDs, im.objectParentMap)
}

func (im *InstanceMainline) LoadSetInstances() error {
	ctx := core.ContextParams{}

	// set instance list of target business
	mongoCondition := mongo.NewCondition()
	mongoCondition.Element(&mongo.Eq{Key: common.BKAppIDField, Val: im.bkBizID})

	err := im.dbProxy.Table(common.BKTableNameBaseSet).Find(mongoCondition.ToMapStr()).All(ctx, &im.setInstances)
	if err != nil {
		blog.Errorf("get set instances by business:%d failed, %+v", im.bkBizID, err)
		return fmt.Errorf("get set instances by business:%d failed, %+v", im.bkBizID, err)
	}
	blog.V(5).Infof("get set instances by business:%d result: %+v", im.bkBizID, im.setInstances)
	return nil
}

func (im *InstanceMainline) LoadModuleInstances() error {
	// module instance list of target business
	ctx := core.ContextParams{}
	mongoCondition := mongo.NewCondition()
	mongoCondition.Element(&mongo.Eq{Key: common.BKAppIDField, Val: im.bkBizID})

	err := im.dbProxy.Table(common.BKTableNameBaseModule).Find(mongoCondition.ToMapStr()).All(ctx, &im.moduleInstances)
	if err != nil {
		blog.Errorf("get module instances by business:%d failed, %+v", im.bkBizID, err)
		return fmt.Errorf("get module instances by business:%d failed, %+v", im.bkBizID, err)
	}
	blog.V(5).Infof("get module instances by business:%d result: %+v", im.bkBizID, im.moduleInstances)
	return nil
}

func (im *InstanceMainline) LoadMainlineInstances() error {
	// load other mainline instance(except business,set,module) list of target business
	var err error
	ctx := core.ContextParams{}
	condCheckModel := mongo.NewCondition()
	_, metaCond := condCheckModel.Embed(metadata.BKMetadata)
	_, labelCond := metaCond.Embed(metadata.BKLabel)
	labelCond.Element(&mongo.Eq{Key: common.BKAppIDField, Val: strconv.FormatInt(im.bkBizID, 10)})
	condCheckModel.Element(&mongo.In{Key: common.BKObjIDField, Val: im.modelIDs})
	cond := condCheckModel.ToMapStr()
	err = im.dbProxy.Table(common.BKTableNameBaseInst).Find(cond).All(ctx, &im.mainlineInstances)
	if err != nil {
		blog.Errorf("get other mainline instances by business:%d failed, %+v", im.bkBizID, err)
		return fmt.Errorf("get other mainline instances by business:%d failed, %+v", im.bkBizID, err)
	}
	blog.V(5).Infof("get other mainline instances by business:%d result: %+v", im.bkBizID, im.mainlineInstances)
	return nil
}

func (im *InstanceMainline) ConstructBizTopoInstance(withDetail bool) error {
	// enqueue business instance to allTopoInstances, instanceMap
	ctx := core.ContextParams{}
	bizTopoInstance := &metadata.TopoInstance{
		ObjectID:         common.BKInnerObjIDApp,
		InstanceID:       im.bkBizID,
		ParentInstanceID: 0,
		Detail:           map[string]interface{}{},
	}

	if withDetail == true {
		// get business detail here
		mongoCondition := mongo.NewCondition()
		mongoCondition.Element(&mongo.Eq{Key: common.BKAppIDField, Val: im.bkBizID})

		err := im.dbProxy.Table(common.BKTableNameBaseApp).Find(mongoCondition.ToMapStr()).All(ctx, &im.businessInstances)
		if err != nil {
			blog.Errorf("get business instances by business:%d failed, err: %+v", im.bkBizID, err)
			return fmt.Errorf("get business instances by business:%d failed, err: %+v", im.bkBizID, err)
		}
		blog.V(5).Infof("SearchMainlineInstanceTopo businessInstances: %+v", im.businessInstances)
		if len(im.businessInstances) == 0 {
			blog.Errorf("business instances by business:%d not found", im.bkBizID)
			return fmt.Errorf("business instances by business:%d not found", im.bkBizID)
		}
		bizTopoInstance.Detail = im.businessInstances[0]
	}

	im.allTopoInstances = append(im.allTopoInstances, bizTopoInstance)
	im.instanceMap[bizTopoInstance.Key()] = bizTopoInstance
	return nil
}

func (im *InstanceMainline) OrganizeSetInstance(withDetail bool) error {
	for _, instance := range im.setInstances {
		instanceID, err := util.GetInt64ByInterface(instance[common.BKSetIDField])
		if err != nil {
			blog.Errorf("parse instanceID:%+v to int64 failed, %+v", instance[common.BKSetIDField], err)
			return fmt.Errorf("parse instanceID:%+v to int64 failed, %+v", instance[common.BKSetIDField], err)
		}
		parentInstanceID, err := util.GetInt64ByInterface(instance[common.BKInstParentStr])
		if err != nil {
			blog.Errorf("parse instanceID:%+v to int64 failed, %+v", instance[common.BKInstParentStr], err)
			return fmt.Errorf("parse instanceID:%+v to int64 failed, %+v", instance[common.BKInstParentStr], err)
		}

		defaultFieldValue, err := util.GetInt64ByInterface(instance[common.BKDefaultField])
		if err != nil {
			blog.Errorf("parse set instance default field failed, default: %+v, err: %+v", instance[common.BKDefaultField], err)
			return fmt.Errorf("parse set instance default field failed, default: %+v, err: %+v", instance[common.BKDefaultField], err)
		}

		topoInstance := &metadata.TopoInstance{
			Default:          defaultFieldValue,
			ObjectID:         common.BKInnerObjIDSet,
			InstanceID:       instanceID,
			ParentInstanceID: parentInstanceID,
			Detail:           map[string]interface{}{},
		}
		if withDetail == true {
			topoInstance.Detail = instance
		}
		im.allTopoInstances = append(im.allTopoInstances, topoInstance)
		im.instanceMap[topoInstance.Key()] = topoInstance
	}
	return nil
}

func (im *InstanceMainline) OrganizeModuleInstance(withDetail bool) error {
	for _, instance := range im.moduleInstances {
		instanceID, err := util.GetInt64ByInterface(instance[common.BKModuleIDField])
		if err != nil {
			blog.Errorf("parse instanceID:%+v to int64 failed, %+v", instance[common.BKModuleIDField], err)
			return fmt.Errorf("parse instanceID:%+v to int64 failed, %+v", instance[common.BKModuleIDField], err)
		}
		parentInstanceID, err := util.GetInt64ByInterface(instance[common.BKInstParentStr])
		if err != nil {
			blog.Errorf("parse instanceID:%+v to int64 failed, %+v", instance[common.BKInstParentStr], err)
			return fmt.Errorf("parse instanceID:%+v to int64 failed, %+v", instance[common.BKInstParentStr], err)
		}

		defaultFieldValue, err := util.GetInt64ByInterface(instance[common.BKDefaultField])
		if err != nil {
			blog.Errorf("parse module instance default field failed, default: %+v, err: %+v", instance[common.BKDefaultField], err)
			return fmt.Errorf("parse module instance default field failed, default: %+v, err: %+v", instance[common.BKDefaultField], err)
		}

		topoInstance := &metadata.TopoInstance{
			Default:          defaultFieldValue,
			ObjectID:         common.BKInnerObjIDModule,
			InstanceID:       instanceID,
			ParentInstanceID: parentInstanceID,
			Detail:           map[string]interface{}{},
		}
		if withDetail == true {
			topoInstance.Detail = instance
		}
		im.allTopoInstances = append(im.allTopoInstances, topoInstance)
		im.instanceMap[topoInstance.Key()] = topoInstance
	}
	return nil
}

func (im *InstanceMainline) OrganizeMainlineInstance(withDetail bool) error {
	for _, instance := range im.mainlineInstances {
		instanceID, err := util.GetInt64ByInterface(instance[common.BKInstIDField])
		if err != nil {
			blog.Errorf("parse instanceID:%+v to int64 failed, %+v", instance[common.BKInstIDField], err)
			return fmt.Errorf("parse instanceID:%+v to int64 failed, %+v", instance[common.BKInstIDField], err)
		}
		parentInstanceID, err := util.GetInt64ByInterface(instance[common.BKInstParentStr])
		if err != nil {
			blog.Errorf("parse instanceID:%+v to int64 failed, %+v", instance[common.BKInstParentStr], err)
			return fmt.Errorf("parse instanceID:%+v to int64 failed, %+v", instance[common.BKInstParentStr], err)
		}
		topoInstance := &metadata.TopoInstance{
			ObjectID:         instance[common.BKObjIDField].(string),
			InstanceID:       instanceID,
			ParentInstanceID: parentInstanceID,
			Detail:           map[string]interface{}{},
		}
		if withDetail == true {
			topoInstance.Detail = instance
		}
		im.allTopoInstances = append(im.allTopoInstances, topoInstance)
		im.instanceMap[topoInstance.Key()] = topoInstance
	}
	return nil
}

func (im *InstanceMainline) CheckAndFillingMissingModels(withDetail bool) error {
	ctx := core.ContextParams{}
	// prepare loop that make sure all node's parent are exist in allTopoInstances
	for _, topoInstance := range im.allTopoInstances {
		blog.V(5).Infof("topo instance: %+v", topoInstance)
		if topoInstance.ParentInstanceID == 0 {
			continue
		}
		var parentKey string
		if topoInstance.ObjectID == common.BKInnerObjIDSet && topoInstance.Default == 1 {
			// `空闲机池` 是一种特殊的set，它用来包含空闲机和故障机两个模块，它的父节点直接是业务（不论是否有自定义层级）
			// 这类特殊情况的结点是业务，不需要重复获取，ConstructInstanceTopoTree 会做进一步处理
			parentKey = fmt.Sprintf("%s:%d", common.BKInnerObjIDApp, topoInstance.ParentInstanceID)
		} else {
			parentObjectID := im.objectParentMap[topoInstance.ObjectID]
			parentKey = fmt.Sprintf("%s:%d", parentObjectID, topoInstance.ParentInstanceID)
		}
		// check whether parent instance exist, if not, try to get it at best.
		_, exist := im.instanceMap[parentKey]
		if exist == true {
			continue
		}
		blog.Warnf("get parent of %+v with key=%s failed, not Found", topoInstance, parentKey)
		// There is a bug in legacy code that business before mainline model be created in cc_ObjectBase table has no bk_biz_id field
		// and therefore find parentInstance failed.
		// In this case current algorithm degenerate in to o(n) query cost.

		mongoCondition := mongo.NewCondition()
		mongoCondition.Element(&mongo.Eq{Key: common.BKInstIDField, Val: topoInstance.ParentInstanceID})

		missedInstances := make([]mapstr.MapStr, 0)
		err := im.dbProxy.Table(common.BKTableNameBaseInst).Find(mongoCondition.ToMapStr()).All(ctx, &missedInstances)
		if err != nil {
			blog.Errorf("get common instances with ID:%d failed, %+v", topoInstance.ParentInstanceID, err)
			return err
		}
		blog.V(5).Infof("get missed instances by id:%d results: %+v", topoInstance.ParentInstanceID, missedInstances)
		if len(missedInstances) == 0 {
			if topoInstance.ObjectID == common.BKInnerObjIDSet &&
				im.bkBizID == topoInstance.ParentInstanceID {
				// `空闲机池` 是一种特殊的set，它用来包含空闲机和故障机两个模块，它的父节点直接是业务（不论是否有自定义层级）
				// 这类特殊情况的结点是业务，不需要重复获取，ConstructInstanceTopoTree 会做进一步处理
				continue
			} else {
				blog.Errorf("found unexpected count of missedInstances: %+v", missedInstances)
				return fmt.Errorf("SearchMainlineInstanceTopo found %d missedInstances with instanceID=%d", len(missedInstances), topoInstance.ParentInstanceID)
			}
		}
		if len(missedInstances) > 1 {
			blog.Errorf("found too many(%d) missedInstances: %+v by id: %d", len(missedInstances), missedInstances, topoInstance.ParentInstanceID)
			return fmt.Errorf("found too many(%d) missedInstances: %+v by id: %d", len(missedInstances), missedInstances, topoInstance.ParentInstanceID)
		}
		instance := missedInstances[0]
		instanceID, err := util.GetInt64ByInterface(instance[common.BKInstIDField])
		if err != nil {
			blog.Errorf("parse instanceID:%+v to int64 failed, %+v", instance[common.BKInstIDField], err)
			return fmt.Errorf("parse instanceID:%+v to int64 failed, %+v", instance[common.BKInstIDField], err)
		}

		var parentInstanceID int64
		parentValue, existed := instance[common.BKInstParentStr]
		if existed == true {
			parentInstanceID, err = util.GetInt64ByInterface(parentValue)
			if err != nil {
				blog.Errorf("parse instanceID:%+v to int64 failed, %+v", instance[common.BKInstParentStr], err)
				return fmt.Errorf("parse instanceID:%+v to int64 failed, %+v", instance[common.BKInstParentStr], err)
			}
		} else {
			// `空闲机池` 是一种特殊的set，它用来包含空闲机和故障机两个模块，它的父节点直接是业务（不论是否有自定义层级）
			// 这类特殊情况的结点是业务，不需要重复获取，ConstructInstanceTopoTree 会做进一步处理
			if topoInstance.ObjectID == common.BKInnerObjIDSet && im.bkBizID == topoInstance.ParentInstanceID {
				continue
			}
			blog.Errorf("construct biz topo tree, instance doesn't have field %s, instance: %+v, err: %+v", common.BKInstParentStr, instance, err)
			return fmt.Errorf("construct biz topo tree, instance doesn't have field %s, instance: %+v, err: %+v", common.BKInstParentStr, instance, err)
		}
		blog.V(7).Infof("model: %s, instance: %d, parent: %d", topoInstance.ObjectID, topoInstance.InstanceID, parentInstanceID)

		topoInstance := &metadata.TopoInstance{
			ObjectID:         util.GetStrByInterface(instance[common.BKObjIDField]),
			InstanceID:       instanceID,
			ParentInstanceID: parentInstanceID,
			Detail:           map[string]interface{}{},
		}
		if withDetail == true {
			topoInstance.Detail = instance
		}
		im.allTopoInstances = append(im.allTopoInstances, topoInstance)
		im.instanceMap[topoInstance.Key()] = topoInstance
	}
	return nil
}

func (im *InstanceMainline) ConstructInstanceTopoTree(withDetail bool) error {
	topoInstanceNodeMap := map[string]*metadata.TopoInstanceNode{}
	for index := 0; index < len(im.allTopoInstances); index++ {
		topoInstance := im.allTopoInstances[index]
		blog.V(5).Infof("topoInstance: %+v", topoInstance)
		if topoInstance.ParentInstanceID == 0 {
			continue
		}

		parentObjectID := im.objectParentMap[topoInstance.ObjectID]
		parentKey := fmt.Sprintf("%s:%d", parentObjectID, topoInstance.ParentInstanceID)
		if _, exist := topoInstanceNodeMap[parentKey]; exist == false {
			parentInstance, exist := im.instanceMap[parentKey]
			if exist == false {
				// 空闲机池 是一种特殊的set，它用来包含空闲机和故障机两个模块，它的父节点直接是业务（不论是否有自定义层级）
				if topoInstance.ObjectID == common.BKInnerObjIDSet && im.bkBizID == topoInstance.ParentInstanceID {
					parentObjectID = common.BKInnerObjIDApp
					parentKey = fmt.Sprintf("%s:%d", parentObjectID, im.bkBizID)
					parentInstance, exist = im.instanceMap[parentKey]
				}
				if exist == false {
					cond := map[string]interface{}{
						common.BKObjIDField:  parentObjectID,
						common.BKInstIDField: topoInstance.ParentInstanceID,
					}
					inst := mapstr.MapStr{}
					if err := im.dbProxy.Table(common.BKTableNameBaseInst).Find(cond).One(context.Background(), &inst); err != nil {
						if im.dbProxy.IsNotFoundError(err) == false {
							blog.Errorf("get mainline instances failed, filter: %+v, err: %+v", cond, err)
							return fmt.Errorf("get other mainline instances failed, filer: %+v, err: %+v", cond, err)
						} else {
							im.mainlineInstances = append(im.mainlineInstances, inst)
							blog.Errorf("unexpected err, parent instance not found, instance: %+v", topoInstance)
							continue
						}
					}
					parentValue, existed := inst[common.BKInstParentStr]
					if existed == false {
						blog.Errorf("get mainline instances failed, field %s not in db data, data: %+v", common.BKInstParentStr, inst)
						return fmt.Errorf("get mainline instances failed, field %s not in db data, data: %+v", common.BKInstParentStr, inst)
					}
					parentParentID, err := util.GetInt64ByInterface(parentValue)
					if err != nil {
						blog.Errorf("get mainline instances failed, field %s parse into int failed, data: %+v, err: %+v", common.BKInstParentStr, inst, err)
						return fmt.Errorf("get mainline instances failed, field %s parse into int failed, data: %+v, err: %+v", common.BKInstParentStr, inst, err)
					}
					parentInstance = &metadata.TopoInstance{
						ObjectID:         parentObjectID,
						InstanceID:       topoInstance.ParentInstanceID,
						ParentInstanceID: parentParentID,
						Detail:           inst,
					}
					im.instanceMap[parentKey] = parentInstance
					im.allTopoInstances = append(im.allTopoInstances, parentInstance)
				}
			}
			topoInstanceNode := &metadata.TopoInstanceNode{
				ObjectID:   parentInstance.ObjectID,
				InstanceID: parentInstance.InstanceID,
				Detail:     parentInstance.Detail,
				Children:   []*metadata.TopoInstanceNode{},
			}
			topoInstanceNodeMap[parentKey] = topoInstanceNode
		}

		parentInstanceNode := topoInstanceNodeMap[parentKey]

		// extract tree root node pointer
		if parentInstanceNode.ObjectID == common.BKInnerObjIDApp {
			im.root = parentInstanceNode
		}

		if _, exist := topoInstanceNodeMap[topoInstance.Key()]; exist == false {
			childTopoInstanceNode := &metadata.TopoInstanceNode{
				ObjectID:   topoInstance.ObjectID,
				InstanceID: topoInstance.InstanceID,
				Detail:     topoInstance.Detail,
				Children:   []*metadata.TopoInstanceNode{},
			}
			topoInstanceNodeMap[topoInstance.Key()] = childTopoInstanceNode
		}
		childTopoInstanceNode, _ := topoInstanceNodeMap[topoInstance.Key()]
		parentInstanceNode.Children = append(parentInstanceNode.Children, childTopoInstanceNode)
	}
	return nil
}

func (im *InstanceMainline) GetInstanceMap() map[string]*metadata.TopoInstance {
	return im.instanceMap
}

func (im *InstanceMainline) GetRoot() *metadata.TopoInstanceNode {
	return im.root
}
