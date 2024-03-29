package orm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"orm/internal/errs"
	"orm/internal/valuer"
	"testing"
)

// union
func TestSelector_Union(t *testing.T) {
	db := memoryDB(t)
	type Order struct {
		Id        int
		UsingCol1 string
		UsingCol2 string
	}
	type OrderDetail struct {
		OrderId int
		ItemId  int
	}
	type Pay struct {
		Id      int
		OrderId int
		Price   int
	}
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "union",
			q: func() QueryBuilder {
				selector1 := NewSelector[OrderDetail](db)
				selector2 := NewSelector[Order](db)
				return selector1.Union(selector2)
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `order_detail` UNION SELECT * FROM `order`;",
			},
		},
		{
			name: "union all",
			q: func() QueryBuilder {
				selector1 := NewSelector[OrderDetail](db)
				selector2 := NewSelector[Order](db)
				return selector1.UnionAll(selector2)
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `order_detail` UNION ALL SELECT * FROM `order`;",
			},
		},
		{
			name: "union where",
			q: func() QueryBuilder {
				selector1 := NewSelector[OrderDetail](db).Where(C("OrderId").EQ(1))
				selector2 := NewSelector[Order](db).Where(C("Id").EQ(2))
				return selector1.Union(selector2)
			}(),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `order_detail` WHERE `order_id` = ? UNION SELECT * FROM `order` WHERE `id` = ?;",
				Args: []any{1, 2},
			},
		},
		{
			name: "union where all",
			q: func() QueryBuilder {
				selector1 := NewSelector[OrderDetail](db).Where(C("OrderId").EQ(1))
				selector2 := NewSelector[Order](db).Where(C("Id").EQ(2))
				return selector1.UnionAll(selector2)
			}(),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `order_detail` WHERE `order_id` = ? UNION ALL SELECT * FROM `order` WHERE `id` = ?;",
				Args: []any{1, 2},
			},
		},
		{
			name: "union join ",
			q: func() QueryBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				selector1 := NewSelector[Order](db).From(t1.Join(t2).On(t1.C("Id").EQ(t2.C("OrderId"))))

				selector2 := NewSelector[Pay](db).Select().Where(C("Price").EQ(3))
				return selector1.Union(selector2)
			}(),
			wantQuery: &Query{
				SQL:  "SELECT * FROM (`order` AS `t1` JOIN `order_detail` AS `t2` ON `t1`.`id` = `t2`.`order_id`) UNION SELECT * FROM `pay` WHERE `price` = ?;",
				Args: []any{3},
			},
		},
		{
			name: "union subquery ",
			q: func() QueryBuilder {
				sub := NewSelector[OrderDetail](db).Select(C("OrderId")).AsSubquery("sub")
				selector1 := NewSelector[Order](db).Where(C("Id").In(sub))
				selector2 := NewSelector[Pay](db).Select().Where(C("Price").EQ(3))
				return selector1.Union(selector2)
			}(),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `order` WHERE `id` IN (SELECT `order_id` FROM `order_detail`) UNION SELECT * FROM `pay` WHERE `price` = ?;",
				Args: []any{3},
			},
		},
		//{
		//	name: "subquery union",
		//	q: func() QueryBuilder {
		//		sub := NewSelector[OrderDetail](db).Select(C("OrderId")).AsSubquery("sub")
		//		selector1 := NewSelector[Order](db).Where(C("Id").In(sub))
		//		selector2 := NewSelector[Pay](db).Select().Where(C("Price").EQ(3))
		//		sub = selector1.Union(selector2).AsSubquery("sub2")
		//		return NewSelector[Order](db).From(sub)
		//	}(),
		//	wantQuery: &Query{
		//		SQL:  "SELECT * FROM (SELECT * FROM `order` WHERE `id` IN (SELECT `order_id` FROM `order_detail`) UNION SELECT * FROM `pay` WHERE `price` = ?) AS `sub2`;",
		//		Args: []any{3},
		//	},
		//},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}

}

// Join 和 Subquery 混合使用
func TestSelector_SubqueryAndJoin(t *testing.T) {
	db := memoryDB(t)
	type Order struct {
		Id        int
		UsingCol1 string
		UsingCol2 string
	}
	type OrderDetail struct {
		OrderId int
		ItemId  int
	}
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			// 虽然泛型是 Order，但是我们传入 OrderDetail
			name: "table and join",
			q: func() QueryBuilder {
				t1 := TableOf(&Order{})
				sub := NewSelector[OrderDetail](db).AsSubquery("sub")
				return NewSelector[Order](db).Select(sub.C("ItemId")).
					From(t1.Join(sub).On(t1.C("Id").EQ(sub.C("OrderId"))))
			}(),
			wantQuery: &Query{
				SQL: "SELECT `sub`.`item_id` FROM (`order` JOIN (SELECT * FROM `order_detail`) AS `sub` ON `id` = `sub`.`order_id`);",
			},
		},
		{
			name: "table and left join",
			q: func() QueryBuilder {
				t1 := TableOf(&Order{})
				sub := NewSelector[OrderDetail](db).AsSubquery("sub")
				return NewSelector[Order](db).Select(sub.C("ItemId")).
					From(t1.LeftJoin(sub).On(t1.C("Id").EQ(sub.C("OrderId"))))
			}(),
			wantQuery: &Query{
				SQL: "SELECT `sub`.`item_id` FROM (`order` LEFT JOIN (SELECT * FROM `order_detail`) AS `sub` ON `id` = `sub`.`order_id`);",
			},
		},
		{
			name: "join and join",
			q: func() QueryBuilder {
				sub1 := NewSelector[Order](db).AsSubquery("sub1")
				sub2 := NewSelector[OrderDetail](db).AsSubquery("sub2")
				return NewSelector[Order](db).From(sub1.Join(sub2).On(sub1.C("Id").EQ(sub2.C("OrderId"))))
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM ((SELECT * FROM `order`) AS `sub1` JOIN (SELECT * FROM `order_detail`) AS `sub2` ON `sub1`.`id` = `sub2`.`order_id`);",
			},
		},
		{
			name: "join and join using",
			q: func() QueryBuilder {
				sub1 := NewSelector[OrderDetail](db).AsSubquery("sub1")
				sub2 := NewSelector[OrderDetail](db).AsSubquery("sub2")
				return NewSelector[Order](db).From(sub1.RightJoin(sub2).Using("Id"))
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM ((SELECT * FROM `order_detail`) AS `sub1` RIGHT JOIN (SELECT * FROM `order_detail`) AS `sub2` USING (`id`));",
			},
		},
		{
			name: "join sub sub",
			q: func() QueryBuilder {
				sub1 := NewSelector[OrderDetail](db).AsSubquery("sub1")
				sub2 := NewSelector[OrderDetail](db).From(sub1).AsSubquery("sub2")
				t1 := TableOf(&Order{}).As("o1")
				return NewSelector[Order](db).From(sub2.LeftJoin(t1).On(sub2.C("OrderId").EQ(t1.C("Id"))))
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM ((SELECT * FROM (SELECT * FROM `order_detail`) AS `sub1`) AS `sub2` LEFT JOIN `order` AS `o1` ON `sub2`.`order_id` = `o1`.`id`);",
			},
		},
		{
			name: "join sub sub using",
			q: func() QueryBuilder {
				sub1 := NewSelector[OrderDetail](db).AsSubquery("sub1")
				sub2 := NewSelector[OrderDetail](db).From(sub1).AsSubquery("sub2")
				t1 := TableOf(&Order{}).As("o1")
				return NewSelector[Order](db).From(sub2.Join(t1).Using("Id"))
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM ((SELECT * FROM (SELECT * FROM `order_detail`) AS `sub1`) AS `sub2` JOIN `order` AS `o1` USING (`id`));",
			},
		},
		{
			name: "invalid field",
			q: func() QueryBuilder {
				t1 := TableOf(&Order{})
				sub := NewSelector[OrderDetail](db).AsSubquery("sub")
				return NewSelector[Order](db).Select(sub.C("Invalid")).From(t1.Join(sub).On(t1.C("Id").EQ(sub.C("OrderId")))).Where()
			}(),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			name: "invalid field in predicates",
			q: func() QueryBuilder {
				t1 := TableOf(&Order{})
				sub := NewSelector[OrderDetail](db).AsSubquery("sub")
				return NewSelector[Order](db).Select(sub.C("ItemId")).From(t1.Join(sub).On(t1.C("Id").EQ(sub.C("Invalid")))).Where()
			}(),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			name: "invalid field in aggregate function",
			q: func() QueryBuilder {
				t1 := TableOf(&Order{})
				sub := NewSelector[OrderDetail](db).AsSubquery("sub")
				return NewSelector[Order](db).Select(Max("Invalid")).From(t1.Join(sub).On(t1.C("Id").EQ(sub.C("OrderId")))).Where()
			}(),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			name: "not selected",
			q: func() QueryBuilder {
				t1 := TableOf(&Order{})
				sub := NewSelector[OrderDetail](db).Select(C("OrderId")).AsSubquery("sub")
				return NewSelector[Order](db).Select(sub.C("ItemId")).From(t1.Join(sub).On(t1.C("Id").EQ(sub.C("OrderId"))))
			}(),
			wantErr: errs.NewErrUnknownField("ItemId"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}

}

func TestSelector_Subquery(t *testing.T) {
	db := memoryDB(t)
	type Order struct {
		Id        int
		UsingCol1 string
		UsingCol2 string
	}

	type OrderDetail struct {
		OrderId int
		ItemId  int
	}
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "from as",
			q: func() QueryBuilder {
				sub := NewSelector[OrderDetail](db).AsSubquery("sub")
				return NewSelector[Order](db).From(sub)
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM (SELECT * FROM `order_detail`) AS `sub`;",
			},
		},
		{
			name: "in",
			q: func() QueryBuilder {
				sub := NewSelector[OrderDetail](db).Select(C("OrderId")).AsSubquery("sub")
				return NewSelector[Order](db).Where(C("Id").In(sub))
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `order` WHERE `id` IN (SELECT `order_id` FROM `order_detail`);",
			},
		},
		{
			name: "GT",
			q: func() QueryBuilder {
				sub := NewSelector[OrderDetail](db).Select(C("OrderId")).AsSubquery("sub")
				return NewSelector[Order](db).Where(C("Id").GT(sub))
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `order` WHERE `id` > (SELECT `order_id` FROM `order_detail`);",
			},
		},
		{
			name: "exist",
			q: func() QueryBuilder {
				sub := NewSelector[OrderDetail](db).Select(C("OrderId")).AsSubquery("sub")
				return NewSelector[Order](db).Where(Exists(sub))
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `order` WHERE  EXIST (SELECT `order_id` FROM `order_detail`);",
			},
		},
		{
			name: "not exist",
			q: func() QueryBuilder {
				sub := NewSelector[OrderDetail](db).Select(C("OrderId")).AsSubquery("sub")
				return NewSelector[Order](db).Where(Not(Exists(sub)))
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `order` WHERE  NOT ( EXIST (SELECT `order_id` FROM `order_detail`));",
			},
		},
		{
			name: "all",
			q: func() QueryBuilder {
				sub := NewSelector[OrderDetail](db).Select(C("OrderId")).AsSubquery("sub")
				return NewSelector[Order](db).Where(C("Id").GT(All(sub)))
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `order` WHERE `id` > ALL (SELECT `order_id` FROM `order_detail`);",
			},
		},
		{
			name: "some",
			q: func() QueryBuilder {
				sub := NewSelector[OrderDetail](db).Select(C("OrderId")).AsSubquery("sub")
				return NewSelector[Order](db).Where(C("Id").GT(Some(sub)))
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `order` WHERE `id` > SOME (SELECT `order_id` FROM `order_detail`);",
			},
		},
		{
			name: "some and any",
			q: func() QueryBuilder {
				sub := NewSelector[OrderDetail](db).Select(C("OrderId")).AsSubquery("sub")
				return NewSelector[Order](db).Where(C("Id").GT(Some(sub)).And(C("Id").LT(Any(sub))))
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `order` WHERE (`id` > SOME (SELECT `order_id` FROM `order_detail`)) AND (`id` < ANY (SELECT `order_id` FROM `order_detail`));",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}

}

func TestSelector_Join(t *testing.T) {
	db := memoryDB(t)

	type Order struct {
		Id        int
		UsingCol1 string
		UsingCol2 string
	}

	type OrderDetail struct {
		OrderId int
		ItemId  int

		UsingCol1 string
		UsingCol2 string
	}

	type Item struct {
		Id int
	}

	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			// 虽然泛型是 Order，但是我们传入 OrderDetail
			name: "specify table",
			q:    NewSelector[Order](db).From(TableOf(&OrderDetail{})),
			wantQuery: &Query{
				SQL: "SELECT * FROM `order_detail`;",
			},
		},
		{
			name: "join",
			q: func() QueryBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{})
				return NewSelector[Order](db).
					From(t1.Join(t2).On(t1.C("Id").EQ(t2.C("OrderId"))))
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM (`order` AS `t1` JOIN `order_detail` ON `t1`.`id` = `order_id`);",
			},
		},
		{
			name: "multiple join",
			q: func() QueryBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := TableOf(&Item{}).As("t3")
				return NewSelector[Order](db).
					From(t1.Join(t2).
						On(t1.C("Id").EQ(t2.C("OrderId"))).
						Join(t3).On(t2.C("ItemId").EQ(t3.C("Id"))))
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM ((`order` AS `t1` JOIN `order_detail` AS `t2` ON `t1`.`id` = `t2`.`order_id`) JOIN `item` AS `t3` ON `t2`.`item_id` = `t3`.`id`);",
			},
		},
		{
			name: "use join tab col",
			q: func() QueryBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := TableOf(&Item{}).As("t3")
				return NewSelector[Order](db).Select(t1.C("Id"), t1.C("UsingCol1"), t2.C("ItemId"), t3.C("Id")).
					From(t1.Join(t2).
						On(t1.C("Id").EQ(t2.C("OrderId"))).
						Join(t3).On(t2.C("ItemId").EQ(t3.C("Id"))))
			}(),
			wantQuery: &Query{
				SQL: "SELECT `t1`.`id`,`t1`.`using_col1`,`t2`.`item_id`,`t3`.`id` FROM ((`order` AS `t1` JOIN `order_detail` AS `t2` ON `t1`.`id` = `t2`.`order_id`) JOIN `item` AS `t3` ON `t2`.`item_id` = `t3`.`id`);",
			},
		},
		{
			name: "left multiple join",
			q: func() QueryBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := TableOf(&Item{}).As("t3")
				return NewSelector[Order](db).
					From(t1.LeftJoin(t2).
						On(t1.C("Id").EQ(t2.C("OrderId"))).
						LeftJoin(t3).On(t2.C("ItemId").EQ(t3.C("Id"))))
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM ((`order` AS `t1` LEFT JOIN `order_detail` AS `t2` ON `t1`.`id` = `t2`.`order_id`) LEFT JOIN `item` AS `t3` ON `t2`.`item_id` = `t3`.`id`);",
			},
		},
		{
			name: "right multiple join",
			q: func() QueryBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := TableOf(&Item{}).As("t3")
				return NewSelector[Order](db).
					From(t1.RightJoin(t2).
						On(t1.C("Id").EQ(t2.C("OrderId"))).
						RightJoin(t3).On(t2.C("ItemId").EQ(t3.C("Id"))))
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM ((`order` AS `t1` RIGHT JOIN `order_detail` AS `t2` ON `t1`.`id` = `t2`.`order_id`) RIGHT JOIN `item` AS `t3` ON `t2`.`item_id` = `t3`.`id`);",
			},
		},
		{
			name: "join multiple using",
			q: func() QueryBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{})
				return NewSelector[Order](db).
					From(t1.Join(t2).Using("UsingCol1", "UsingCol2"))
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM (`order` AS `t1` JOIN `order_detail` USING (`using_col1`,`using_col2`));",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
}

func TestSelector_OffsetLimit(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "offset only",
			q:    NewSelector[TestModel](db).Offset(10),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` OFFSET ?;",
				Args: []any{10},
			},
		},
		{
			name: "limit only",
			q:    NewSelector[TestModel](db).Limit(10),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` LIMIT ?;",
				Args: []any{10},
			},
		},
		{
			name: "limit offset",
			q:    NewSelector[TestModel](db).Limit(20).Offset(10),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` LIMIT ? OFFSET ?;",
				Args: []any{20, 10},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
}

func TestSelector_Having(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			// 调用了，但是啥也没传
			name: "none",
			q:    NewSelector[TestModel](db).GroupBy(C("Age")).Having(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` GROUP BY `age`;",
			},
		},
		{
			// 单个条件
			name: "single",
			q: NewSelector[TestModel](db).GroupBy(C("Age")).
				Having(C("FirstName").EQ("Deng")),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` GROUP BY `age` HAVING `first_name` = ?;",
				Args: []any{"Deng"},
			},
		},
		{
			// 多个条件
			name: "multiple",
			q: NewSelector[TestModel](db).GroupBy(C("Age")).
				Having(C("FirstName").EQ("Deng"), C("LastName").EQ("Ming")),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` GROUP BY `age` HAVING (`first_name` = ?) AND (`last_name` = ?);",
				Args: []any{"Deng", "Ming"},
			},
		},
		{
			// 聚合函数
			name: "avg",
			q: NewSelector[TestModel](db).GroupBy(C("Age")).
				Having(Avg("Age").EQ(18)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` GROUP BY `age` HAVING AVG(`age`) = ?;",
				Args: []any{18},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
}

func TestSelector_GroupBy(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			// 调用了，但是啥也没传
			name: "none",
			q:    NewSelector[TestModel](db).GroupBy(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},
		{
			// 单个
			name: "single",
			q:    NewSelector[TestModel](db).GroupBy(C("Age")),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` GROUP BY `age`;",
			},
		},
		{
			// 多个
			name: "multiple",
			q:    NewSelector[TestModel](db).GroupBy(C("Age"), C("FirstName")),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` GROUP BY `age`,`first_name`;",
			},
		},
		{
			// 不存在
			name:    "invalid column",
			q:       NewSelector[TestModel](db).GroupBy(C("Invalid")),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
}

func TestSelector_Select(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			// 没有指定
			name: "all",
			q:    NewSelector[TestModel](db),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},
		{
			name:    "invalid column",
			q:       NewSelector[TestModel](db).Select(Avg("Invalid")),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			name: "partial columns",
			q:    NewSelector[TestModel](db).Select(C("Id"), C("FirstName")),
			wantQuery: &Query{
				SQL: "SELECT `id`,`first_name` FROM `test_model`;",
			},
		},
		{
			name: "avg",
			q:    NewSelector[TestModel](db).Select(Avg("Age")),
			wantQuery: &Query{
				SQL: "SELECT AVG(`age`) FROM `test_model`;",
			},
		},
		{
			name: "raw expression",
			q:    NewSelector[TestModel](db).Select(Raw("COUNT(DISTINCT `first_name`)")),
			wantQuery: &Query{
				SQL: "SELECT COUNT(DISTINCT `first_name`) FROM `test_model`;",
			},
		},
		// 别名
		{
			name: "alias",
			q: NewSelector[TestModel](db).
				Select(C("Id").As("my_id"),
					Avg("Age").As("avg_age")),
			wantQuery: &Query{
				SQL: "SELECT `id` AS `my_id`,AVG(`age`) AS `avg_age` FROM `test_model`;",
			},
		},
		// WHERE 忽略别名
		{
			name: "where ignore alias",
			q: NewSelector[TestModel](db).
				Where(C("Id").As("my_id").LT(100)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `id` < ?;",
				Args: []any{100},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
}

func TestSelector_Build(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			// From 都不调用
			name: "no from",
			q:    NewSelector[TestModel](db),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},
		{
			// 单一简单条件
			name: "single and simple predicate",
			q:    NewSelector[TestModel](db).Where(C("Id").EQ(1)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `id` = ?;",
				Args: []any{1},
			},
		},
		{
			// 多个 predicate
			name: "multiple predicates",
			q: NewSelector[TestModel](db).
				Where(C("Age").GT(18), C("Age").LT(35)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` > ?) AND (`age` < ?);",
				Args: []any{18, 35},
			},
		},
		{
			// 使用 AND
			name: "and",
			q: NewSelector[TestModel](db).
				Where(C("Age").GT(18).And(C("Age").LT(35))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` > ?) AND (`age` < ?);",
				Args: []any{18, 35},
			},
		},
		{
			// 使用 OR
			name: "or",
			q: NewSelector[TestModel](db).
				Where(C("Age").GT(18).Or(C("Age").LT(35))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` > ?) OR (`age` < ?);",
				Args: []any{18, 35},
			},
		},
		{
			// 使用 NOT
			name: "not",
			q:    NewSelector[TestModel](db).Where(Not(C("Age").GT(18))),
			wantQuery: &Query{
				// NOT 前面有两个空格，因为我们没有对 NOT 进行特殊处理
				SQL:  "SELECT * FROM `test_model` WHERE  NOT (`age` > ?);",
				Args: []any{18},
			},
		},
		{
			// 非法列
			name:    "invalid column",
			q:       NewSelector[TestModel](db).Where(Not(C("Invalid").GT(18))),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			// 使用 RawExpr
			name: "raw expression",
			q: NewSelector[TestModel](db).
				Where(Raw("`age` < ?", 18).AsPredicate()),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `age` < ?;",
				Args: []any{18},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
}

func TestSelector_GetMulti_baseType(t *testing.T) {
	mockDB, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = mockDB.Close() }()
	db, err := OpenDB("mysql", mockDB)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name      string
		queryRes  func(t *testing.T) any
		mockErr   error
		mockOrder func(mock sqlmock.Sqlmock)
		wantErr   error
		wantVal   any
	}{
		// 返回原生基本类型
		// int
		{
			name: "res int",
			queryRes: func(t *testing.T) any {
				queryer := NewSelector[int](db).Select(C("Age")).From(TableOf(&TestModel{}))
				result, err := queryer.GetMulti(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"age"}).AddRow(10).
					AddRow(18).AddRow(22)
				mock.ExpectQuery("SELECT `age` FROM `test_model`;").
					WillReturnRows(rows)
			},
			wantVal: func() (res []*int) {
				vals := []int{10, 18, 22}
				for i := 0; i < len(vals); i++ {
					res = append(res, &vals[i])
				}
				return
			}(),
		},
		{
			name: "res int32",
			queryRes: func(t *testing.T) any {
				queryer := NewSelector[int32](db).Select(C("Age")).From(TableOf(&TestModel{}))
				result, err := queryer.GetMulti(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"age"}).AddRow(10).
					AddRow(18).AddRow(22)
				mock.ExpectQuery("SELECT `age` FROM `test_model`;").
					WillReturnRows(rows)
			},
			wantVal: func() (res []*int32) {
				vals := []int32{10, 18, 22}
				for i := 0; i < len(vals); i++ {
					res = append(res, &vals[i])
				}
				return
			}(),
		},
		// int64
		{
			name: "avg res int64",
			queryRes: func(t *testing.T) any {
				queryer := NewSelector[int64](db).Select(C("Age")).From(TableOf(&TestModel{}))
				result, err := queryer.GetMulti(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"age"}).AddRow(10).
					AddRow(18).AddRow(22)
				mock.ExpectQuery("SELECT `age` FROM `test_model`;").
					WillReturnRows(rows)
			},
			wantVal: func() (res []*int64) {
				vals := []int64{10, 18, 22}
				for i := 0; i < len(vals); i++ {
					res = append(res, &vals[i])
				}
				return
			}(),
		},
		// float32
		{
			name: "avg res float32",
			queryRes: func(t *testing.T) any {
				queryer := NewSelector[float32](db).Select(C("Age")).From(TableOf(&TestModel{}))
				result, err := queryer.GetMulti(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"age"}).AddRow(10.2).AddRow(18.8)
				mock.ExpectQuery("SELECT `age` FROM `test_model`;").
					WillReturnRows(rows)
			},
			wantVal: func() (res []*float32) {
				vals := []float32{10.2, 18.8}
				for i := 0; i < len(vals); i++ {
					res = append(res, &vals[i])
				}
				return
			}(),
		},
		// float64
		{
			name: "avg res float64",
			queryRes: func(t *testing.T) any {
				queryer := NewSelector[float64](db).Select(C("Age")).From(TableOf(&TestModel{}))
				result, err := queryer.GetMulti(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"age"}).AddRow(10.2).AddRow(18.8)
				mock.ExpectQuery("SELECT `age` FROM `test_model`;").
					WillReturnRows(rows)
			},
			wantVal: func() (res []*float64) {
				vals := []float64{10.2, 18.8}
				for i := 0; i < len(vals); i++ {
					res = append(res, &vals[i])
				}
				return
			}(),
		},
		// byte
		{
			name: "res byte",
			queryRes: func(t *testing.T) any {
				queryer := NewSelector[byte](db).Select(C("FirstName")).
					From(TableOf(&TestModel{}))
				result, err := queryer.GetMulti(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"first_name"}).AddRow('D').AddRow('a')
				mock.ExpectQuery("SELECT `first_name` FROM `test_model`;").
					WillReturnRows(rows)
			},
			wantVal: func() (res []*byte) {
				vals := []byte{'D', 'a'}
				for i := 0; i < len(vals); i++ {
					res = append(res, &vals[i])
				}
				return
			}(),
		},
		// bytes
		{
			name: "res bytes",
			queryRes: func(t *testing.T) any {
				queryer := NewSelector[[]byte](db).Select(C("FirstName")).
					From(TableOf(&TestModel{}))
				result, err := queryer.GetMulti(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"first_name"}).AddRow([]byte("Li")).AddRow([]byte("Liu"))
				mock.ExpectQuery("SELECT `first_name` FROM `test_model`;").
					WillReturnRows(rows)
			},
			wantVal: func() (res []*[]byte) {
				vals := [][]byte{[]byte("Li"), []byte("Liu")}
				for i := 0; i < len(vals); i++ {
					res = append(res, &vals[i])
				}
				return
			}(),
		},
		// string
		{
			name: "res string",
			queryRes: func(t *testing.T) any {
				queryer := NewSelector[string](db).Select(C("FirstName")).
					From(TableOf(&TestModel{}))
				result, err := queryer.GetMulti(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"first_name"}).AddRow("Da").AddRow("Li")
				mock.ExpectQuery("SELECT `first_name` FROM `test_model`;").
					WillReturnRows(rows)
			},
			wantVal: func() (res []*string) {
				vals := []string{"Da", "Li"}
				for i := 0; i < len(vals); i++ {
					res = append(res, &vals[i])
				}
				return
			}(),
		},
		// struct ptr
		{
			name: "res struct ptr",
			queryRes: func(t *testing.T) any {
				queryer := NewSelector[TestModel](db).Select(C("FirstName"), C("Age")).
					From(TableOf(&TestModel{}))
				result, err := queryer.GetMulti(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"first_name", "age"}).
					AddRow("Da", 18).AddRow("Xiao", 16)
				mock.ExpectQuery("SELECT `first_name`,`age` FROM `test_model`;").
					WillReturnRows(rows)
			},
			wantVal: []*TestModel{
				{
					FirstName: "Da",
					Age:       18,
				},
				{
					FirstName: "Xiao",
					Age:       16,
				},
			},
		},
		//// sql.NullString
		{
			name: "res sql.NullString",
			queryRes: func(t *testing.T) any {
				queryer := NewSelector[sql.NullString](db).Select(C("LastName")).
					From(TableOf(&TestModel{}))
				result, err := queryer.GetMulti(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"last_name"}).
					AddRow([]byte("ming")).AddRow([]byte("gang"))
				mock.ExpectQuery("SELECT `last_name` FROM `test_model`;").
					WillReturnRows(rows)
			},
			wantVal: []*sql.NullString{
				{
					String: "ming",
					Valid:  true,
				},
				{
					String: "gang",
					Valid:  true,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockOrder(mock)
			res := tc.queryRes(t)
			assert.EqualValues(t, tc.wantVal, res)
		})
	}
}

func TestSelector_Get_baseType(t *testing.T) {
	//mockDB, mock, err := sqlmock.New()
	mockDB, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = mockDB.Close() }()
	db, err := OpenDB("mysql", mockDB)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name      string
		queryRes  func(t *testing.T) any
		mockErr   error
		mockOrder func(mock sqlmock.Sqlmock)
		wantErr   error
		wantVal   any
	}{
		//返回原生基本类型
		// int
		{
			name: "res int",
			queryRes: func(t *testing.T) any {
				queryer := NewSelector[int](db).Select(C("Age")).
					From(TableOf(&TestModel{})).Where(C("Id").EQ(1))
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"age"}).AddRow(10)
				mock.ExpectQuery("SELECT `age` FROM `test_model` WHERE `id` = ?;").
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantVal: func() *int {
				val := 10
				return &val
			}(),
		},
		// int32
		{
			name: "res int32",
			queryRes: func(t *testing.T) any {
				queryer := NewSelector[int32](db).Select(C("Age")).
					From(TableOf(&TestModel{})).Where(C("Id").EQ(1))
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"age"}).AddRow(10)
				mock.ExpectQuery("SELECT `age` FROM `test_model` WHERE `id` = ?;").
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantVal: func() *int32 {
				val := int32(10)
				return &val
			}(),
		},
		// int64
		{
			name: "res int64",
			queryRes: func(t *testing.T) any {
				queryer := NewSelector[int64](db).Select(C("Age")).
					From(TableOf(&TestModel{})).Where(C("Id").EQ(1))
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"age"}).AddRow(10)
				mock.ExpectQuery("SELECT `age` FROM `test_model` WHERE `id` = ?;").
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantVal: func() *int64 {
				val := int64(10)
				return &val
			}(),
		},
		// float32
		{
			name: "res float32",
			queryRes: func(t *testing.T) any {
				queryer := NewSelector[float32](db).Select(C("Age")).
					From(TableOf(&TestModel{})).Where(C("Id").EQ(1))
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"age"}).AddRow(10.2)
				mock.ExpectQuery("SELECT `age` FROM `test_model` WHERE `id` = ?;").
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantVal: func() *float32 {
				val := float32(10.2)
				return &val
			}(),
		},
		// float64
		{
			name: "res float64",
			queryRes: func(t *testing.T) any {
				queryer := NewSelector[float64](db).Select(C("Age")).
					From(TableOf(&TestModel{})).Where(C("Id").EQ(1))
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"age"}).AddRow(10.02)
				mock.ExpectQuery("SELECT `age` FROM `test_model` WHERE `id` = ?;").
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantVal: func() *float64 {
				val := 10.02
				return &val
			}(),
		},
		// byte
		{
			name: "res byte",
			queryRes: func(t *testing.T) any {
				queryer := NewSelector[byte](db).Select(C("FirstName")).
					From(TableOf(&TestModel{})).Where(C("Id").EQ(1))
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"first_name"}).AddRow('D')
				mock.ExpectQuery("SELECT `first_name` FROM `test_model` WHERE `id` = ?;").
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantVal: func() *byte {
				val := byte('D')
				return &val
			}(),
		},
		// bytes
		{
			name: "res bytes",
			queryRes: func(t *testing.T) any {
				queryer := NewSelector[[]byte](db).Select(C("FirstName")).
					From(TableOf(&TestModel{})).Where(C("Id").EQ(1))
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"first_name"}).AddRow([]byte("Li"))
				mock.ExpectQuery("SELECT `first_name` FROM `test_model` WHERE `id` = ?;").
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantVal: func() *[]byte {
				val := []byte("Li")
				return &val
			}(),
		},
		// string
		{
			name: "res string",
			queryRes: func(t *testing.T) any {
				queryer := NewSelector[string](db).Select(C("FirstName")).
					From(TableOf(&TestModel{})).Where(C("Id").EQ(1))
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"first_name"}).AddRow("Da")
				mock.ExpectQuery("SELECT `first_name` FROM `test_model` WHERE `id` = ?;").
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantVal: func() *string {
				val := "Da"
				return &val
			}(),
		},
		// struct ptr
		{
			name: "res struct ptr",
			queryRes: func(t *testing.T) any {
				queryer := NewSelector[TestModel](db).Select(C("FirstName"), C("Age")).
					Where(C("Id").EQ(1))
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"first_name", "age"}).AddRow("Da", 18)
				mock.ExpectQuery("SELECT `first_name`,`age` FROM `test_model` WHERE `id` = ?;").
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantVal: func() *TestModel {
				return &TestModel{
					FirstName: "Da",
					Age:       18,
				}
			}(),
		},
		{
			name: "res struct ptr all cols",
			queryRes: func(t *testing.T) any {
				queryer := NewSelector[TestModel](db).Where(C("Id").EQ(1))
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"id", "first_name", "age", "last_name"}).AddRow(1, "Da", 18, "ming")
				mock.ExpectQuery("SELECT * FROM `test_model` WHERE `id` = ?;").
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantVal: func() *TestModel {
				return &TestModel{
					Id:        1,
					FirstName: "Da",
					Age:       18,
					LastName:  &sql.NullString{String: "ming", Valid: true},
				}
			}(),
		},
		// struct sql.NullString
		{
			name: "res sql.NullString",
			queryRes: func(t *testing.T) any {
				queryer := NewSelector[sql.NullString](db).Select(C("FirstName")).
					From(TableOf(&TestModel{})).
					Where(C("Id").EQ(1))
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"first_name"}).AddRow("Da")
				mock.ExpectQuery("SELECT `first_name` FROM `test_model` WHERE `id` = ?;").
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantVal: func() *sql.NullString {
				return &sql.NullString{
					String: "Da",
					Valid:  true,
				}
			}(),
		},
		//// time
		//{
		//	name: "res time",
		//	queryRes: func(t *testing.T) any {
		//		queryer := NewSelector[time.Time](db).Select(C("CreatedAt")).
		//			From(TableOf(&TestModel{})).Where(C("Id").EQ(1))
		//		result, err := queryer.Get(context.Background())
		//		require.NoError(t, err)
		//		return result
		//	},
		//	mockOrder: func(mock sqlmock.Sqlmock) {
		//		target := "2022-07-03 22:14:02"
		//		t, _ := time.ParseInLocation("2006-01-02 15:04:05", target, time.Local)
		//		rows := mock.NewRows([]string{"created_at"}).AddRow(t)
		//		mock.ExpectQuery("SELECT `created_at` FROM `test_model` WHERE `id` = ?;").
		//			WithArgs(1).
		//			WillReturnRows(rows)
		//	},
		//	wantVal: func() *time.Time {
		//		target := "2022-07-03 22:14:02"
		//		t, _ := time.ParseInLocation("2006-01-02 15:04:05", target, time.Local)
		//		return &t
		//	}(),
		//},
		//{
		//	name: "res null time",
		//	queryRes: func(t *testing.T) any {
		//		queryer := NewSelector[sql.NullTime](db).Select(C("CreatedAt")).
		//			From(TableOf(&TestModel{})).Where(C("Id").EQ(1))
		//		result, err := queryer.Get(context.Background())
		//		require.NoError(t, err)
		//		return result
		//	},
		//	mockOrder: func(mock sqlmock.Sqlmock) {
		//		target := "2022-07-03 22:14:02"
		//		t, _ := time.ParseInLocation("2006-01-02 15:04:05", target, time.Local)
		//		rows := mock.NewRows([]string{"created_at"}).AddRow(t)
		//		mock.ExpectQuery("SELECT `created_at` FROM `test_model` WHERE `id` = ?;").
		//			WithArgs(1).
		//			WillReturnRows(rows)
		//	},
		//	wantVal: &sql.NullTime{
		//		Time: func() time.Time {
		//			tm, _ := time.ParseInLocation(
		//				"2006-01-02 15:04:05", "2022-07-03 22:14:02", time.Local)
		//			return tm
		//		}(), Valid: true},
		//},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockOrder(mock)
			res := tc.queryRes(t)
			assert.Equal(t, tc.wantVal, res)
		})
	}
}

//测试多条语句
func TestSelector_GetMulti(t *testing.T) {

	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	testCases := []struct {
		name     string
		query    string
		mockErr  error
		mockRows *sqlmock.Rows
		wantErr  error
		wantVal  []*TestModel
	}{
		{
			name:    "multi row",
			query:   "SELECT .*",
			mockErr: nil,
			mockRows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
				rows.AddRow([]byte("123"), []byte("Ming"), []byte("18"), []byte("Deng"))
				rows.AddRow([]byte("456"), []byte("Min"), []byte("19"), []byte("Da"))
				return rows
			}(),
			wantVal: []*TestModel{
				{
					Id:        123,
					FirstName: "Ming",
					Age:       18,
					LastName:  &sql.NullString{Valid: true, String: "Deng"},
				},
				{
					Id:        456,
					FirstName: "Min",
					Age:       19,
					LastName:  &sql.NullString{Valid: true, String: "Da"},
				},
			},
		},

		{
			name:    "invalid columns",
			query:   "SELECT .*",
			mockErr: nil,
			mockRows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "first_name", "gender"})
				rows.AddRow([]byte("123"), []byte("Ming"), []byte("male"))
				return rows
			}(),
			wantErr: errs.NewErrUnknownColumn("gender"),
		},

		{
			name:    "more columns",
			query:   "SELECT .*",
			mockErr: nil,
			mockRows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name", "first_name"})
				rows.AddRow([]byte("123"), []byte("Ming"), []byte("18"), []byte("Deng"), []byte("明明"))
				return rows
			}(),
			wantErr: errs.ErrTooManyReturnedColumns,
		},
	}

	for _, tc := range testCases {
		if tc.mockErr != nil {
			mock.ExpectQuery(tc.query).WillReturnError(tc.mockErr)
		} else {
			mock.ExpectQuery(tc.query).WillReturnRows(tc.mockRows)
		}
	}

	db, err := OpenDB("mysql", mockDB)
	require.NoError(t, err)
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			res, err := NewSelector[TestModel](db).GetMulti(context.Background())
			assert.Equal(t, tt.wantErr, err)
			if err != nil {
				return
			}

			assert.Equal(t, tt.wantVal, res)
		})
	}
}

func TestSelector_Get(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = mockDB.Close() }()
	db, err := OpenDB("mysql", mockDB)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name     string
		query    string
		mockErr  error
		mockRows *sqlmock.Rows
		wantErr  error
		wantVal  *TestModel
	}{
		{
			// 查询返回错误
			name:    "query error",
			mockErr: errors.New("invalid query"),
			wantErr: errors.New("invalid query"),
			query:   "SELECT .*",
		},
		{
			name:     "no row",
			wantErr:  ErrNoRows,
			query:    "SELECT .*",
			mockRows: sqlmock.NewRows([]string{"id"}),
		},
		{
			name:    "too many column",
			wantErr: errs.ErrTooManyReturnedColumns,
			query:   "SELECT .*",
			mockRows: func() *sqlmock.Rows {
				res := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name", "extra_column"})
				res.AddRow([]byte("1"), []byte("Da"), []byte("18"), []byte("Ming"), []byte("nothing"))
				return res
			}(),
		},
		{
			name:  "get data",
			query: "SELECT .*",
			mockRows: func() *sqlmock.Rows {
				res := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
				res.AddRow([]byte("1"), []byte("Da"), []byte("18"), []byte("Ming"))
				return res
			}(),
			wantVal: &TestModel{
				Id:        1,
				FirstName: "Da",
				Age:       18,
				LastName:  &sql.NullString{String: "Ming", Valid: true},
			},
		},
	}

	for _, tc := range testCases {
		exp := mock.ExpectQuery(tc.query)
		if tc.mockErr != nil {
			exp.WillReturnError(tc.mockErr)
		} else {
			exp.WillReturnRows(tc.mockRows)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := NewSelector[TestModel](db).Get(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantVal, res)
		})
	}
}

// 在 orm 目录下执行
// go test -bench=BenchmarkQuerier_Get -benchmem -benchtime=10000x
// 我的输出结果
// goos: linux
// goarch: amd64
// pkg: gitee.com/geektime-geekbang/geektime-go/orm
// cpu: Intel(R) Core(TM) i5-10400F CPU @ 2.90GHz
// BenchmarkQuerier_Get/unsafe-12             10000            453677 ns/op            3246 B/op        108 allocs/op
// BenchmarkQuerier_Get/reflect-12            10000           1173199 ns/op            3427 B/op        117 allocs/op
// PASS
// ok      gitee.com/geektime-geekbang/geektime-go/orm     16.324s
func BenchmarkQuerier_Get(b *testing.B) {
	db, err := Open("sqlite3", fmt.Sprintf("file:benchmark_get.db?cache=shared&mode=memory"))
	if err != nil {
		b.Fatal(err)
	}
	_, err = db.db.Exec(TestModel{}.CreateSQL())
	if err != nil {
		b.Fatal(err)
	}

	res, err := db.db.Exec("INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`)"+
		"VALUES (?,?,?,?)", 12, "Deng", 18, "Ming")

	if err != nil {
		b.Fatal(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		b.Fatal(err)
	}
	if affected == 0 {
		b.Fatal()
	}

	b.Run("unsafe", func(b *testing.B) {
		db.valCreator = valuer.BasicTypeCreator{
			Creator: valuer.NewUnsafeValue,
		}
		for i := 0; i < b.N; i++ {
			_, err = NewSelector[TestModel](db).Get(context.Background())
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("reflect", func(b *testing.B) {
		db.valCreator = valuer.BasicTypeCreator{
			Creator: valuer.NewReflectValue,
		}
		for i := 0; i < b.N; i++ {
			_, err = NewSelector[TestModel](db).Get(context.Background())
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
