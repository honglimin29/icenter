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

package mongo_test

import (
	"testing"

	"icenter/src/common/mapstr"

	"icenter/src/common"
	"icenter/src/common/condition"
	"icenter/src/common/metadata"
	"icenter/src/common/universalsql/mongo"

	"github.com/stretchr/testify/require"
)

func TestNewConditionFromMapStrWithCustomType(t *testing.T) {

	target := mongo.NewCondition()
	target.Element(&mongo.Eq{Key: "custom_type", Val: common.DataStatusDisabled})

	sql, err := target.ToSQL()
	require.NoError(t, err)
	t.Logf("target sql:%s", sql)

	recoverSql, err := mongo.NewConditionFromMapStr(target.ToMapStr())
	require.NoError(t, err)
	sql, err = recoverSql.ToSQL()
	require.NoError(t, err)
	t.Logf("recover sql:%s", sql)
}
func TestNewConditionFromMapStr(t *testing.T) {

	target := mongo.NewCondition()
	target.Element(&mongo.Eq{Key: "testelementeq", Val: "testeqval"})
	target.And(&mongo.Lt{Key: "testandlt", Val: "testandltval"})
	target.Or(&mongo.Lt{Key: "testorlt", Val: "testorltval"})
	target.Element(&mongo.In{Key: "testelementin", Val: []string{"testelementin"}})
	_, embed := target.Embed("testembedname")
	embed.Or(&mongo.Gt{Key: "testembedgt", Val: "testembedgtval"})
	embed.And(&mongo.Gt{Key: "testembedgt", Val: "testembedgtval"})
	embed.Element(&mongo.Lt{Key: "testembedeq", Val: "testembedeqval"})
	embed.Element(&mongo.Lt{Key: "testembedeq2", Val: "testembedeqval2"})
	_, subembed := embed.Embed("subembed")
	subembed.Element(&mongo.Eq{Key: "subembedkey", Val: "subembedkeyval"})

	sql, _ := target.ToSQL()
	t.Logf("target sql:%s", sql)

	recoverSql, err := mongo.NewConditionFromMapStr(target.ToMapStr())
	require.NoError(t, err)
	sql, _ = recoverSql.ToSQL()
	t.Logf("recover sql:%s", sql)
}

func TestMgCondition(t *testing.T) {
	target := mongo.NewCondition()
	target.Element(
		mongo.Field("name.first").Nin([]string{"test1", "test2"}).In([]string{"test3", "test4"}),
		mongo.Field("age").Lte(75).Gte(15),
		mongo.Field("name.last").Eq("yang"),
	)
	sql, _ := target.ToSQL()
	t.Logf("%s", sql)

	target.And(
		mongo.Field("").Lt(75).Gte(15),
		mongo.Field("").In([]string{"red", "green"}),
	)
	sql, _ = target.ToSQL()
	t.Logf("%s", sql)

	target.Or(
		mongo.Field("").All(5),
		mongo.Field("age").Size(3).All([]int{6, 7, 8}),
	)
	sql, _ = target.ToSQL()
	t.Logf("%s", sql)

	target.Nor(
		mongo.Field("age").Lt(75).Gte(15),
		mongo.Field("family").In([]string{"wang", "yang"}),
	)
	sql, _ = target.ToSQL()
	t.Logf("%s", sql)

	target.Not(
		mongo.Field("age").Lt(75).Gte(15),
		mongo.Field("family").In([]string{"li", "yang"}),
	)
	sql, _ = target.ToSQL()
	t.Logf("%s", sql)
}

func TestIssue1708(t *testing.T) {
	testData := metadata.QueryCondition{
		Condition: mapstr.MapStr{
			"bk_group_id": "default",
			"bk_obj_id":   "1",
			"id":          mapstr.MapStr{"$nin": []int{0}},
			"metadata":    mapstr.MapStr{"label": nil},
		},
	}

	cond, err := mongo.NewConditionFromMapStr(testData.Condition)
	require.NoError(t, err)
	t.Logf("t:%#v", cond.ToMapStr())

}

func TestIssue1738(t *testing.T) {
	cond := mongo.NewCondition()
	cond.Element(&mongo.Eq{Key: "bk_set_name", Val: nil})
	cond.Element(&mongo.Eq{Key: "bk_set_id", Val: nil})
	cond.Element(&mongo.Eq{Key: "bk_biz_id", Val: nil})
	cond.Element(&mongo.Eq{Key: "bk_parent_id", Val: 2})
	cond.Element(&mongo.In{Key: "bk_parent_in_nil", Val: nil})
	cond.Element(&mongo.Nin{Key: "bk_parent_nin_nil", Val: nil})
	cond.Element(&mongo.Neq{Key: "bk_data_status", Val: "disabled"})

	result, err := cond.ToSQL()
	require.NoError(t, err)
	t.Logf("sql:%s", result)

	inputMapStr := cond.ToMapStr()

	outCond := mongo.NewCondition()
	for i := 0; i <= 1; i++ {
		outCond, err = mongo.NewConditionFromMapStr(inputMapStr)
		require.NoError(t, err)
	}

	result, err = outCond.ToSQL()
	require.NoError(t, err)
	t.Logf("sql_1738:%s", result)
}

func TestNewConditionFromMapStrFromCommonCondition(t *testing.T) {
	type tmpStruct struct {
		A int
	}

	target := condition.CreateCondition()
	target.Field("eq").Eq(1)
	target.Field("int_arr").Eq([]int{1, 2, 4})
	target.Field("str_arr").Eq([]string{"1", "2", "4"})
	target.Field("struct").Eq(tmpStruct{A: 1})
	target.Field("struct_arr").Eq([]tmpStruct{tmpStruct{A: 1}})

	_, err := mongo.NewConditionFromMapStr(target.ToMapStr())
	require.NoError(t, err)
}
