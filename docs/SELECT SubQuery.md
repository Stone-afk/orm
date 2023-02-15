# SELECT SubQuery （子查询）

子查询可以用在很多地方，需要考虑的语法特性也非常多，这是一个比较复杂的功能。

| 语法                                                         | 说明                                                         |
| ------------------------------------------------------------ | ------------------------------------------------------------ |
| SELECT(SELECT s2 FROM t1);                                   | 这是用作 SELECT 部分。需要注意的是，这里面 s2 是一个列，所以最终外层里面返回的也是一个列。 |
| 这种可以考虑不支持，因为本身站在工程的角度来看，它并不是一种很好的实践方式。 |                                                              |
| 在 WHERE 条件里面和操作符进行比较，例如：                    |                                                              |
| WHERE'a'=(SELECT column1 FROM t1)                            | 除了 =，其它操作符 >, >=, <, <= 都是可以的。                 |
| 一般来说，可能会配合聚合函数一起使用：                       |                                                              |
| SELECT*FROM t1                                               |                                                              |
| WHERE column1 =(SELECTMAX(column2)FROM t2);                  |                                                              |
| IN 查询：                                                    |                                                              |
| SELECT s1 FROM t1 WHERE s1 IN(SELECT s1 FROM t2);            | 这也是我们在日常使用中最常使用的形态了                       |
| 和 ANY, ALL, SOME 谓词配合使用：                             |                                                              |
| SELECT s1 FROM t1 WHERE s1 >ANY(SELECT s1 FROM t2);          | ANY，ALL 一般会和比较操作符一起使用，例如 < ANY, <= ALL 等   |
| 多个列混合使用：                                             |                                                              |
| SELECT*FROM t1                                               |                                                              |
| WHERE(col1,col2)=(SELECT col3, col4 FROM t2 WHERE id =10);   | 这种用法比较罕见，稍微比较常见的是 IN 和多个列混合使用，例如： |
| SELECT*FROM t1                                               |                                                              |
| WHERE(col1,col2)IN(SELECT col3, col4 FROM t2 WHERE id =10);  |                                                              |
| Exist 和 Not Exist：                                         |                                                              |
| WHEREEXISTS(SELECT*FROM cities_stores                        |                                                              |
| WHERE cities_stores.store_type = stores.store_type);         | 用于表达存在或者不存在。这种用法也是比较常见                 |
| 在子查询内部使用外部表：                                     |                                                              |
| SELECT*FROM t1                                               |                                                              |
| WHERE column1 =ANY(SELECT column1 FROM t2                    |                                                              |
| WHERE t2.column2 = t1.column2);                              | 虽然从理论上来说这种写法是可以的，但是我们并不建议在实践中使用，因为可读性并不高，对于部分用户来说可能搞不清楚这种语句最终会返回什么值 |
| 用作 SELECT 的表：                                           |                                                              |
| SELECT...FROM(_subquery_)[AS]_tbl_name_...                   | 这个类似于 JOIN 查询。类似地，子查询也可以作为 JOIN 的左边或者右边。 |
| 更进一步地，用户可以使用子查询的别名来指定列，例如           |                                                              |
| SELECT sub.name FROM (SELECT * FROM XXX) AS sub。            |                                                              |
| 类似地，在 WHERE 部分也可以使用子查询的列：                  |                                                              |
| SELECT sub.name FROM (SELECT * FROM XXX) AS sub WHERE sub.id > 10 |                                                              |
| 同样地，在 HAVING 和 ON 里面应该都可以使用子查询的别名       |                                                              |
|                                                              |                                                              |

## 语句分析

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675580711875-84b9a4f2-0b69-4260-90f7-fde9e1189bba.png#averageHue=%23f8f8f7&clientId=u8796a7b3-ae08-4&from=paste&height=496&id=uc7b9a0d8&name=image.png&originHeight=992&originWidth=2184&originalType=binary&ratio=1&rotation=0&showTitle=false&size=541582&status=done&style=none&taskId=u3144066b-0ef4-4808-9679-ad1daf2d504&title=&width=1092)
子查询看着很复杂，但是构成子查询的元素，上面都已经处理过了。前面说过，子查询就是一个临时表，把别名当做表名，左右两个 table_references 的列就是临时表的列，如果子查询设置了查询表达式，那么临时表的列就是查询表达式里面写的那些列。
整体思路和 JOIN 的思路是差不多的，区别的地方在于对列的处理方式不同。处理 JOIN 的时候，查询表达式里的列可能来自不同的表。但是在子查询这里，所有的列都来自子查询的临时表。

## **场景分析**

### 将子查询用在 WHERE 语句中

- 使用 IN 和 NOT IN 查询，可以只支持单个列的查询，而不必支持多个列的 IN 查询
- 支持操作符，例如 > (SELECT XXX) 这种形态
- 可以支持 ANY，SOME 和 ALL 三个谓词，但是要注意并不是所有的方言都支持这三个谓词
- 支持 EXIST 和 NOT EXIST 两个查询条件

### 将子查询用在 FROM 里面

- 子查询独立作为 FROM 的部分
- 子查询组合成 JOIN 查询。子查询、JOIN 查询和普通的表之间可以层层嵌套组成复杂的结构
  - 可以是子查询和单独的物理表
  - 子查询和子查询
  - 子查询和 JOIN 查询
- 使用子查询的别名指定列
  - 指定列用在 SELECT 部分，也可以和聚合函数一起使用
  - 指定列用在 WHERE 部分
  - 指定列用在 HAVING 部分
  - 指定列用在 ON 部分

## 开源实例

### Beego ORM

Beego 本身对子查询的支持非常有限，核心是在 dbQuerier 里面有一个方法：
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675568642554-8e3c56d5-56e1-455d-b08a-d1e321267a12.png#averageHue=%2332342f&clientId=u631d5832-1b8c-4&from=paste&height=268&id=u74b08a5d&name=image.png&originHeight=536&originWidth=1558&originalType=binary&ratio=1&rotation=0&showTitle=false&size=458674&status=done&style=none&taskId=uddd01737-9969-434a-bf7d-94890380548&title=&width=779)
这个设计基本上只能依赖于用户自己去手写 SQL。
Beego 在 dbQuerier 里面的方法大部分都是接收一个 string 参数，例如 From，那么也就是说我们天然就可以直接使用这些方法。类似地，Beego 并没有显式地定义 Exist，ALL 之类的方法，但是 Where 方法本身是接收一个 string 类型的，所以用户可以自己输入类似于 EXIST (SELECT XXX) 这样的字符串。

### GORM

#### where 子查询

 GORM 允许在使用 *gorm.DB 对象作为参数时生成子查询 

```go
db.Where("amount > (?)", db.Table("orders").Select("AVG(amount)")).Find(&orders)
// SELECT * FROM "orders" WHERE amount > (SELECT AVG(amount) FROM "orders");
subQuery := db.Select("AVG(age)").Where("name LIKE ?", "name%").Table("users")
db.Select("AVG(age) as avgage").Group("name").Having("AVG(age) > (?)",
subQuery).Find(&results)
// SELECT AVG(age) as avgage FROM `users` GROUP BY `name` HAVING AVG(age) >
(SELECT AVG(age) FROM `users` WHERE name LIKE "name%")

```

GORM 将 WHERE 语句后的参数定义为接口类型，表示该参数接收任何类型同时也就包括了子查询 
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675568045768-b3ceeb90-f741-4376-98f6-e6542e3b333a.png#averageHue=%232e2c2b&clientId=u631d5832-1b8c-4&from=paste&height=231&id=uc6e9a5b3&name=image.png&originHeight=462&originWidth=1544&originalType=binary&ratio=1&rotation=0&showTitle=false&size=194610&status=done&style=none&taskId=u79072ad1-d4bc-448f-bc06-1ac0eca30ed&title=&width=772)

#### From 子查询

GORM 允许您在 Table 方法中通过 FROM 子句使用子查询，例如: 

```go
db.Table("(?) as u", db.Model(&User{}).Select("name", "age")).Where("age = ?",
18).Find(&User{})
// SELECT * FROM (SELECT `name`,`age` FROM `users`) as u WHERE `age` = 18
subQuery1 := db.Model(&User{}).Select("name")
subQuery2 := db.Model(&Pet{}).Select("name")
db.Table("(?) as u, (?) as p", subQuery1, subQuery2).Find(&User{})
// SELECT * FROM (SELECT `name` FROM `users`) as u, (SELECT `name` FROM `pets`)
as p
```

Table 接口和 WHERE 接口一样，参数定义为接口类型 
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675568344751-07709a93-b59e-4ca7-a0ea-3483d4823acd.png#averageHue=%232e2d2c&clientId=u631d5832-1b8c-4&from=paste&height=380&id=ud1ffcbbb&name=image.png&originHeight=760&originWidth=1548&originalType=binary&ratio=1&rotation=0&showTitle=false&size=380131&status=done&style=none&taskId=u5319020f-41fa-4369-a110-13ce310dab3&title=&width=774)
**总的来说也就是查询本身就可以被用在 Table 方法调用里面**。
GORM 的 API 也是大量接收了 interface 或者 string 作为输入，所以用户可以自己手写这种，例如在 WHERE 部分写 WHERE a IN (SELECT XXX)。

## API 设计

总体上可以复用 JOIN 查询里面的 TableReference 抽象，提供一个代表子查询的抽象：

```go
type Subquery struct {
	// 使用 QueryBuilder 仅仅是为了让 Subquery 可以是非泛型的。
	s QueryBuilder
	columns []Selectable
	alias string
	table TableReference
}

func (s Subquery) expr() {}

func (s Subquery) tableAlias() string {
	return s.alias
}
```

特殊支持在于，Subquery 里面理论上来说应该使用 Selector 的。但是因为本身 Selector 是泛型的，而我们并不希望在 Subquery 中引入类型参数，所以我们就直接使用 QueryBuilder 本身。
**Subquery 会实现以下接口**：

- TableReference，这确保了子查询可以用在 FROM 部分
- Expression，确保了子查询可以用在 WHERE 部分

但依赖于 Selector 构建

```go
func (s *Selector[T]) AsSubquery(alias string) Subquery {
	tbl := s.table
	if tbl == nil {
		tbl = TableOf(new(T))
	}
	return Subquery {
		s: s,
		alias: alias,
		table: tbl,
		columns: s.columns,
	}
}
```

### 指定列

类似于 JOIN 查询，可以在 Subquery 结构体上定义一个指定列的方法：

```go
func (s Subquery) C(name string) Column {
	return Column{
		table: s,
		name:  name,
	}
}
```

那么这个返回的 Colum 就可以被用在 SELECT 部分，或者 WHERE 部分

### JOIN 查询

类似于 Table 结构体，我们在 Subquery 上定义 Join, LeftJoin 和 RightJoin 方法：

```go
func (s Subquery) Join(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  s,
		right: target,
		typ:   "JOIN",
	}
}

func (s Subquery) LeftJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  s,
		right: target,
		typ:   "LEFT JOIN",
	}
}

func (s Subquery) RightJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  s,
		right: target,
		typ:   "RIGHT JOIN",
	}
}

```

### Any，All 和 Some

```go
type SubqueryExpr struct {
	s Subquery
	// 谓词，ALL，ANY 或者 SOME
	pred string
}

func (s SubqueryExpr) expr() {}

func All(sub Subquery) SubqueryExpr {
	return SubqueryExpr{
		s:    sub,
		pred: preALL,
	}
}

func Any(sub Subquery) SubqueryExpr {
	return SubqueryExpr{
		s:    sub,
		pred: preAny,
	}
}

func Some(sub Subquery) SubqueryExpr {
	return SubqueryExpr{
		s:    sub,
		pred: preSome,
	}
}
```

注意，这里的设计并没有把谓词直接定义在 Subquery 上，而是额外定义了一个结构体。是因为我们希望用户不能将这个 SubqueryExpr 用在 FROM 部分，避免类似于 Aggregate 别名处理的尴尬问题。
SubqueryExpr 本身会实现 Expression 接口，这也确保了 SubqueryExpr 可以被用在 WHERE 部分，以构建复杂的查询条件。

### EXIST 和 NOT EXIST

类似于 EQ 之类的方法，可以定义新的方法：

```go
func Exist(sub Subquery) Predicate {
	return Predicate{
		op: opExist,
		right: sub,
	}
}

```

和 Eq 之类的方法比起来，不同之处就是 Exist 是 left 是没有取值的，并且 op 定义了一个新的 opExist。
显然，Not Exist 可以复用已有的 Not 方法，而不需要我们额外定义一个新的方法。

## 重构 Selector 与 builder 的部分方法

### Selector

增加子查询构建逻辑 buildSubquery

```go
func (b *Builder) buildSubquery(sub Subquery, useAlias bool) error {
	q, err := sub.s.Build()
	if err != nil {
		return err
	}
	b.writeLeftParenthesis()
	b.writeString(q.SQL[:len(q.SQL)-1])
	b.writeRightParenthesis()
	if len(q.Args) > 0 {
		b.addArgs(q.Args...)
	}
	if useAlias {
		if err = b.buildAs(sub.alias); err != nil {
			return err
		}
	}
	return nil
}

```

```go
func (s *Selector[T]) buildTable(table TableReference) error {
	switch tab := table.(type) {
	case nil:
		s.quote(s.model.TableName)
	case Table:
		meta, err := s.r.Get(tab.entity)
		if err != nil {
			return err
		}
		s.quote(meta.TableName)
		return s.buildAs(tab.alias)
	case Join:
		return s.buildJoin(tab)
    // Subquery
	case Subquery:
		return s.buildSubquery(tab, true)
	default:
		return errs.NewErrUnsupportedExpressionType(tab)
	}
	return nil
}
```

```go
func (s *Selector[T]) buildColumn(val Column, useAlias bool) error {
	err := s.Builder.buildColumn(val, useAlias)
	if err != nil {
		return err
	}
	if useAlias {
		if err = s.buildAs(val.alias); err != nil {
			return err
		}
	}
	return nil
}
```

### builder

```go
func (b *Builder) buildColumn(val Column, useAlias bool) error {
	var alias string
	if val.table != nil {
		alias = val.table.tableAlias()
	}
	if alias != "" {
		b.quote(alias)
		b.writeByte('.')
	}
	colName, err := b.colName(val.table, val.name, useAlias)
	if err != nil {
		return err
	}
	b.quote(colName)
	return nil
}
```

要考虑列的构造，是否依赖于子查询

```go
func (b *Builder) colName(table TableReference, fdName string, useAlias bool) (string, error) {
	switch tab := table.(type) {
	case nil:
		_, ok := b.aliasMap[fdName]
		if useAlias && ok {
			return fdName, nil
		}
		fd, ok := b.model.FieldMap[fdName]
		if !ok {
			return "", errs.NewErrUnknownField(fdName)
		}
		return fd.ColName, nil
	case Table:
		m, err := b.r.Get(tab.entity)
		if err != nil {
			return "", err
		}
		fd, ok := m.FieldMap[fdName]
		if !ok {
			return "", errs.NewErrUnknownField(fdName)
		}
		return fd.ColName, nil
    // 子查询
	case Subquery:
		if len(tab.columns) > 0 {
			for _, col := range tab.columns {
				if col.selectedAlias() == fdName {
					return fdName, nil
				}
				if col.fieldName() == fdName {
					return b.colName(col.target(), fdName, useAlias)
				}
				return "", errs.NewErrUnknownField(fdName)
			}
		}
		return b.colName(tab.table, fdName, useAlias)
	default:
		return "", errs.NewErrUnsupportedExpressionType(tab)
	}
}
```

加入了子查询后考虑，重构 buildExpression

```go
func (b *Builder) buildBinaryExpr(
	exp binaryExpr, colsAlias, aggreAlias bool) error {
	err := b.buildSubExpr(
		exp.left, colsAlias, aggreAlias)
	if err != nil {
		return err
	}
	if exp.op != "" {
		b.writeSpace()
		b.writeString(exp.op.String())
		b.writeSpace()
	}

	return b.buildSubExpr(
		exp.right, colsAlias, aggreAlias)
}

func (b *Builder) buildSubExpr(expr Expression, colsAlias, aggreAlias bool) error {
	switch e := expr.(type) {
	case MathExpr:
		b.writeLeftParenthesis()
		err := b.buildBinaryExpr(
			binaryExpr(e), colsAlias, aggreAlias)
		if err != nil {
			return err
		}
		b.writeRightParenthesis()
	case binaryExpr:
		b.writeLeftParenthesis()
		err := b.buildBinaryExpr(
			e, colsAlias, aggreAlias)
		if err != nil {
			return err
		}
		b.writeRightParenthesis()
	case Predicate:
		b.writeLeftParenthesis()
		err := b.buildBinaryExpr(
			binaryExpr(e), colsAlias, aggreAlias)
		if err != nil {
			return err
		}
		b.writeRightParenthesis()
	default:
		err := b.buildExpression(
			e, colsAlias, aggreAlias)
		if err != nil {
			return err
		}
	}
	return nil
}


func (b *Builder) buildExpression(
	e Expression, colsAlias, aggreAlias bool) error {
	switch exp := e.(type) {
	case nil:
		return nil
	case Column:
		return b.buildColumn(exp, colsAlias)
	case Aggregate:
		return b.buildAggregate(exp, aggreAlias)
	case value:
		return b.buildValue(exp)
	case RawExpr:
		return b.buildRaw(exp)
	case MathExpr:
		return b.buildBinaryExpr(
			binaryExpr(exp), colsAlias, aggreAlias)
	case Predicate:
		return b.buildBinaryExpr(
			binaryExpr(exp), colsAlias, aggreAlias)
    // 增加 子查询 与 谓语条件包含子查询 的分支 
	case Subquery:
		return b.buildSubquery(exp, false)
	case SubqueryExpr:
		b.writeString(exp.pred)
		b.writeSpace()
		return b.buildSubquery(exp.s, false)
	default:
		return errs.NewErrUnsupportedExpressionType(exp)
	}
}
```

## 单元测试

```go
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
```