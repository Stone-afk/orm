### 指定查询列

 在 SELECT 语句中，我们可以指定列，严格来说，可以指定：

- 普通列 
- 聚合函数 
- 子查询

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675136496864-951bbe31-4d90-407c-81b8-9ff89ea6f7c8.png#averageHue=%232e2d2c&clientId=u8bae189c-24d9-4&from=paste&height=480&id=ua9d97385&name=image.png&originHeight=600&originWidth=1519&originalType=binary&ratio=1&rotation=0&showTitle=false&size=325817&status=done&style=none&taskId=uee5ad8b3-c6ee-4bc6-a918-2d428146db9&title=&width=1215.2)
**方案一**: 直接传入字符串指定列
 也就是说，我们需要一个指定列的接口。 最简单的情况下，我们就让用户传入字符串。  

```go
func (s *Selector[T]) Select(cols ...string) *Selector[T] {
	s.columns = cols
	return s
}
```

 **接下来构造查询列逻辑很简单**： 

- **没有指定列**：那就是搜索全部列 
- **指定列**：则只搜索指定的列  

**优点:**

- 简单明了

**缺点**：

- 缺乏校验，手一抖写错了都发现不了 
- 用户直接写的是列名，而不是我们希望的字段名  

**方案二**： Selectable 抽象  
 定义个新的标记接口，限定传入的类型，这样我 们就可以做各种校验。 **这是一种严苛的设计方案**，而不是宽容的设计方案。

```go
type Selectable interface {
	selectable()
}

func (s *Selector[T]) Select(cols ...Selectable) *Selector[T] {
	s.columns = cols
	return s
}
```

对应这个抽象，可以对应两个实现里，分别是 Column与Aggregate

```go
func (c Column) selectable() {}
```

```go
// Aggregate 代表聚合函数，例如 AVG, MAX, MIN 等
type Aggregate struct {
	fn    string
	arg   string
}

func (a Aggregate) selectable() {}

func Avg(c string) Aggregate {
	return Aggregate{
		fn:  "AVG",
		arg: c,
	}
}
```

####  别名 As

 在 SQL 里面，我们可以使用 As 这种关键字来指定返回的字段的别名，比较常见 于为聚合函数设置别名，少数情况下在 JOIN 或者子查询的情况下也会为列设置 别名。 AS 也是可以省略的  
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675143157547-b0b9b139-0e1f-4ce2-b8c8-61aadd572a63.png#averageHue=%23483a2b&clientId=uf67b7122-2a63-4&from=paste&height=246&id=ub559535c&name=image.png&originHeight=307&originWidth=1589&originalType=binary&ratio=1&rotation=0&showTitle=false&size=306950&status=done&style=none&taskId=u0ff21e97-f297-4e45-986a-eed59fe4bd6&title=&width=1271.2)
目前 column 与 aggregate 模块都需要考虑支持别名的场景，所以只需要在 Column 和 Aggregate 两个类型上新加一个方法 As 即可

```go
type Column struct {
	name  string
	alias string
}

func (c Column) As(alias string) Column {
	return Column {
		name:  c.name,
		alias: alias,
	}
}
```

```go
// Aggregate 代表聚合函数，例如 AVG, MAX, MIN 等
type Aggregate struct {
	fn    string
	arg   string
	alias string  // 别名
}


func (a Aggregate) As(alias string) Aggregate {
	return Aggregate{
		fn:    a.fn,
		arg:   a.arg,
		alias: alias,
	}
}

```

 虽然这种设计是不可变对象的设计思路，即每次都返回一个新对 象，但是本质上并不是为了并发安全，而是为了链式调用。  

####  怎么支持其它乱七八糟的查询呢？  

例如下图的sql语句:
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675142417966-b6ba748a-6a7c-43de-90a0-c13cc16b0a69.png#averageHue=%232e2d2b&clientId=uf67b7122-2a63-4&from=paste&height=337&id=u6428060c&name=image.png&originHeight=421&originWidth=1391&originalType=binary&ratio=1&rotation=0&showTitle=false&size=212243&status=done&style=none&taskId=u6aea2868-d538-4ccb-8c90-fc182e3a5fa&title=&width=1112.8)
它体现的系统设计的时候不得不处理的一个问 题，就是用户的脑洞总会比你的大。 所以一定要设计一个兜底的方案，它可能不太好用， 用户容易犯错，但是得有。  这就是  RawExpr  (原生查询模块) 的支持

####   RawExpr  

 RawExpr 就是让用户自己编写sql语法，然后让用户保证正确性。  

```go
// RawExpr 代表一个原生表达式
// 意味着 ORM 不会对它进行任何处理
type RawExpr struct {
	raw  string
	args []interface{}
}

func (r RawExpr) selectable() {}


// Raw 创建一个 RawExpr
func Raw(expr string, args ...interface{}) RawExpr {
	return RawExpr{
		raw:  expr,
		args: args,
	}
}
```

RawExpr 也是要支持 Where 语法的，所以也要让 RawExpr 实现  Expression 接口，那么就可 以用于构造 WHERE  语句

```go
func (r RawExpr) expr() {}

func (r RawExpr) AsPredicate() Predicate {
	return Predicate{
		left: r,
	}
}
```

#### 改造 Selector 的相关方法

```go
func (s *Selector[T]) buildColumns() error {
	if len(s.columns) == 0 {
		s.sb.WriteByte('*')
		return nil
	}
	for i, c := range s.columns {
		if i > 0 {
			s.sb.WriteByte(',')
		}
		switch val := c.(type) {
		case Column:
			s.sb.WriteByte('`')
			fd, ok := s.model.FieldMap[val.name]
			if !ok {
				return errs.NewErrUnknownField(val.name)
			}
			s.sb.WriteString(fd.ColName)
			s.sb.WriteByte('`')
			s.buildAs(val.alias)  // 只在构建列的时候使用别名
		case Aggregate:
			s.sb.WriteString(val.fn)
			s.sb.WriteString("(`")
			fd, ok := s.model.FieldMap[val.arg]
			if !ok {
				return errs.NewErrUnknownField(val.arg)
			}
			s.sb.WriteString(fd.ColName)
			s.sb.WriteString("`)")
			s.buildAs(val.alias)
		case RawExpr:
			s.sb.WriteString(val.raw)
			if len(val.args) != 0 {
				s.addArgs(val.args...)
			}
		default:
			return errs.NewErrUnsupportedSelectable(c)
		}
	}
	return nil
}
```

注意： 在拼接 SQL 的时候，要注意，这两个用在 WHERE （或者 HAVING） 都是要忽略掉别名的  

```go
func (s *Selector[T]) buildExpression(e Expression) error {
	if e == nil {
		return nil
	}
	switch exp := e.(type) {
	case Column:
		fd, ok := s.model.FieldMap[exp.name]
		if !ok {
			return errs.NewErrUnknownField(exp.name)
		}
		s.sb.WriteByte('`')
		s.sb.WriteString(fd.ColName)
		s.sb.WriteByte('`')
	case value:
		s.sb.WriteByte('?')
		s.addArgs(exp.val)
	case RawExpr:  //  RawExpr
		s.sb.WriteString(exp.raw)
		if len(exp.args) != 0 {
			s.addArgs(exp.args...)
		}
	case Predicate:
		_, lp := exp.left.(Predicate)
		if lp {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(exp.left); err != nil {
			return err
		}
		if lp {
			s.sb.WriteByte(')')
		}

		// 可能只有左边
		if exp.op == "" {
			return nil
		}

		s.sb.WriteByte(' ')
		s.sb.WriteString(exp.op.String())
		s.sb.WriteByte(' ')

		_, rp := exp.right.(Predicate)
		if rp {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(exp.right); err != nil {
			return err
		}
		if rp {
			s.sb.WriteByte(')')
		}
	default:
		return errs.NewErrUnsupportedExpressionType(exp)
	}
	return nil
}
```

最后加上单元测试

```go
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
```

### 总结

-  **为什么 WHERE 里面不能使用聚合函数**？从 DB 实现的角度来说，是因为聚合函数必须要在数据筛选出来 之后才能计算，也因此，**HAVING 是可以使用聚合函数的**。简单概括就是，在**DB中， 处理顺序是 WHERE - GROUP BY - 聚合函数 - HAVING**；
-  **WHERE 和 HAVING 的区别**：最重要的就是能不能使用聚合函数作为查询条件，以及两者的执行顺序  ；
-  **聚合函数有哪些**？常用的就是 Max、Min、Count、Sum、Avg。另外一个类似的问题是，DISTINCT 是不 是聚合函数？显然不是，DISTINCT 是去重的关键字  ；
-  ** 当使用聚合函数的时候，SELECT 后面有什么限制**？当使用聚合函数之后，在 SELECT 后面只能是常数， 或者聚合函数，或者出现在 GROUP BY 中的列。理论上我们在 ORM 框架里面是可以进行这种校验的，只 是在我们这个实现里面并没有执行校验。另外要小心面试官要求写 SQL  ；