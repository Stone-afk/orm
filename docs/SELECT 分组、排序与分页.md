## SELECT 分组、排序与分页

### 分组

常用的简单 SELECT 语句无疑包括了分组语法 GROUP BY 以及 分组后的条件过滤 HAVING，HAVING 的执行 由于在 WHERE、GROUP BY 和 聚合函数之后，所以 HAVING 也是支持聚合函数的 （**DB中， 处理顺序是 WHERE - GROUP BY - 聚合函数 - HAVING**；）。
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675156604856-2b034014-d493-43a9-bccd-3c624f7a8774.png#averageHue=%23463a2a&clientId=u3aa1eac3-b0db-4&from=paste&height=118&id=u76bd2d52&name=image.png&originHeight=221&originWidth=1620&originalType=binary&ratio=1&rotation=0&showTitle=false&size=236470&status=done&style=none&taskId=u900d211b-86bf-442f-8d36-7930e066c07&title=&width=864)

#### GROUP BY 构造

 同样限定 Column，这样就能执行校验 ；

```go
// Selector 用于构造 SELECT 语句
type Selector[T any] struct {
	sb      strings.Builder
	args    []any
	table   string
	where   []Predicate
	model   *model.Model
	db      *DB
	columns []Selectable
	groupBy []Column  // 增加 group by
}

// GroupBy 设置 group by 子句
func (s *Selector[T]) GroupBy(cols ...Column) *Selector[T] {
	s.groupBy = cols
	return s
}
```

**改造 Selector 增加对 group by 的判断**

```go
func (s *Selector[T]) Build() (*Query, error) {
	var (
		t   T
		err error
	)
	s.model, err = s.db.r.Get(&t)
	if err != nil {
		return nil, err
	}
	s.sb.WriteString("SELECT ")
	if err = s.buildColumns(); err != nil {
		return nil, err
	}
	s.sb.WriteString(" FROM ")
	if s.table == "" {
		s.sb.WriteByte('`')
		s.sb.WriteString(s.model.TableName)
		s.sb.WriteByte('`')
	} else {
		s.sb.WriteString(s.table)
	}

	// 构造 WHERE
	if len(s.where) > 0 {
		// 类似这种可有可无的部分，都要在前面加一个空格
		s.sb.WriteString(" WHERE ")
		// WHERE 是不允许用别名的
		if err = s.buildPredicates(s.where); err != nil {
			return nil, err
		}
	}

    // group by 分组
	if len(s.groupBy) > 0 {
		s.sb.WriteString(" GROUP BY ")
		for i, c := range s.groupBy {
			if i > 0 {
				s.sb.WriteByte(',')
			}
			if err = s.buildColumn(c, false); err != nil {
				return nil, err
			}
		}
	}

	s.sb.WriteString(";")
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}
```

#### HAVING 构造

HAVING 和 WHERE 其实非常像，但是也有区别： HAVING 的查询条件里面，可以使用聚合函数。 所以整体上，可以复用 WHERE 那部分代码  

```go
// Selector 用于构造 SELECT 语句
type Selector[T any] struct {
	sb      strings.Builder
	args    []any
	table   string
	where   []Predicate
	having  []Predicate // 增加 having
	model   *model.Model
	db      *DB
	columns []Selectable
	groupBy []Column
}

func (s *Selector[T]) Build() (*Query, error) {
	var (
		t   T
		err error
	)
	s.model, err = s.db.r.Get(&t)
	if err != nil {
		return nil, err
	}
	s.sb.WriteString("SELECT ")
	if err = s.buildColumns(); err != nil {
		return nil, err
	}
	s.sb.WriteString(" FROM ")
	if s.table == "" {
		s.sb.WriteByte('`')
		s.sb.WriteString(s.model.TableName)
		s.sb.WriteByte('`')
	} else {
		s.sb.WriteString(s.table)
	}

	// 构造 WHERE
	if len(s.where) > 0 {
		// 类似这种可有可无的部分，都要在前面加一个空格
		s.sb.WriteString(" WHERE ")
		p := s.where[0]
		for i := 1; i < len(s.where); i++ {
			p = p.And(s.where[i])
		}
		if err = s.buildExpression(p); err != nil {
			return nil, err
		}
	}

	if len(s.groupBy) > 0 {
		s.sb.WriteString(" GROUP BY ")
		for i, c := range s.groupBy {
			if i > 0 {
				s.sb.WriteByte(',')
			}
			if err = s.buildColumn(c, false); err != nil {
				return nil, err
			}
		}
	}

    // having 过滤
	if len(s.having) > 0 {
		s.sb.WriteString(" HAVING ")
		// HAVING 是可以用别名的
		p := s.having[0]
		for i := 1; i < len(s.having); i++ {
			p = p.And(s.having[i])
		}
		if err = s.buildExpression(p); err != nil {
			return nil, err
		}
	}

	s.sb.WriteString(";")
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}


func (s *Selector[T]) Having(ps ...Predicate) *Selector[T] {
	s.having = ps
	return s
}

```

#####  HAVING 支持聚合函  

 基本上可以参考 Column 的设计，在 Aggregate 上也设计代表 >, >= , <,  的方法

```go
// Aggregate 实现 Expression 接口
func (a Aggregate) expr() {}

// EQ 例如 C("id").Eq(12)
func (a Aggregate) EQ(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opEQ,
		right: exprOf(arg),
	}
}

func (a Aggregate) LT(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opLT,
		right: exprOf(arg),
	}
}

func (a Aggregate) GT(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opGT,
		right: exprOf(arg),
	}
}
```

** 提取 Aggregate 公共代码  **

```go
func (s *Selector[T]) buildAggregate(a Aggregate, useAlias bool) error {
	s.sb.WriteString(a.fn)
	s.sb.WriteString("(`")
	fd, ok := s.model.FieldMap[a.arg]
	if !ok {
		return errs.NewErrUnknownField(a.arg)
	}
	s.sb.WriteString(fd.ColName)
	s.sb.WriteString("`)")
	if useAlias {
		s.buildAs(a.alias)
	}
	return nil
}
```

**提取 Column 公共代码**

```go
func (s *Selector[T]) buildColumn(c Column, useAlias bool) error {
	s.sb.WriteByte('`')
	fd, ok := s.model.FieldMap[c.name]
	if !ok {
		return errs.NewErrUnknownField(c.name)
	}
	s.sb.WriteString(fd.ColName)
	s.sb.WriteByte('`')
	if useAlias {
		s.buildAs(c.alias)
	}
	return nil
}
```

 **提取 WHERE 和 HAVING 的公共部分**

```go
func (s *Selector[T]) buildPredicates(ps []Predicate) error {
	p := ps[0]
	for i := 1; i < len(ps); i++ {
		p = p.And(ps[i])
	}
	return s.buildExpression(p)
}
```

```go
func (s *Selector[T]) Build() (*Query, error) {
	var (
		t   T
		err error
	)
	s.model, err = s.db.r.Get(&t)
	if err != nil {
		return nil, err
	}
	s.sb.WriteString("SELECT ")
	if err = s.buildColumns(); err != nil {
		return nil, err
	}
	s.sb.WriteString(" FROM ")
	if s.table == "" {
		s.sb.WriteByte('`')
		s.sb.WriteString(s.model.TableName)
		s.sb.WriteByte('`')
	} else {
		s.sb.WriteString(s.table)
	}

	// 构造 WHERE
	if len(s.where) > 0 {
		// 类似这种可有可无的部分，都要在前面加一个空格
		s.sb.WriteString(" WHERE ")
		// WHERE 是不允许用别名的
		if err = s.buildPredicates(s.where); err != nil {
			return nil, err
		}
	}

	if len(s.groupBy) > 0 {
		s.sb.WriteString(" GROUP BY ")
		for i, c := range s.groupBy {
			if i > 0 {
				s.sb.WriteByte(',')
			}
			if err = s.buildColumn(c, false); err != nil {
				return nil, err
			}
		}
	}
	// 构造 having
	if len(s.having) > 0 {
		s.sb.WriteString(" HAVING ")
		// HAVING 是可以用别名的
		if err = s.buildPredicates(s.having); err != nil {
			return nil, err
		}
	}

	s.sb.WriteString(";")
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}
```

 **HAVING 中要不要支持直接用别名做查询条件？  **
现在还有一个问题要处理，在 HAVING 中使用 SELECT 中出现的别名 ，这里不会支持这种写法，因为它带来的收益 很有限，但是会极大的破坏代码的美感。  

### 排序

#### ORDER BY

ORDER BY 语法也是在日常开发中常用的语法，本文的排序方法可以参考上面的分组方法的方式构建

```go
// Selector 用于构造 SELECT 语句
type Selector[T any] struct {
	sb      strings.Builder
	args    []any
	table   string
	where   []Predicate
	having  []Predicate
	model   *model.Model
	db      *DB
	columns []Selectable
	groupBy []Column
	orderBy []OrderBy
	offset  int
	limit   int
}

func (s *Selector[T]) OrderBy(orderBys...OrderBy) *Selector[T] {
	s.orderBy = orderBys
	return s
}

func (s *Selector[T]) buildOrderBy() error {
	for idx, ob := range s.orderBy {
		if idx > 0 {
			s.sb.WriteByte(',')
		}
		err := s.buildColumn(ob.col, "")
		if err != nil {
			return err
		}
		s.sb.WriteByte(' ')
		s.sb.WriteString(ob.order)
	}
	return nil

```

### 分页

#### Offset 和 Limi  

 在 MySQL 中，典型的分页查询是 Limit x, y。x 是 偏移量，y 是目标数量。 例如 Limit 10, 20 是指从偏移量 10 开始往后取 20 条数据。 而实际上标准 SQL 的写法是 LIMIT y OFFSET  
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675159590352-23a6a334-82e9-426d-aef4-caf58057aaae.png#averageHue=%23f7f6f5&clientId=u9a9f8e63-5288-4&from=paste&height=507&id=zGIdw&name=image.png&originHeight=634&originWidth=1070&originalType=binary&ratio=1&rotation=0&showTitle=false&size=103187&status=done&style=none&taskId=u040a8de5-6eb1-42cb-84d9-6c347b0580b&title=&width=856)

```go
// Selector 用于构造 SELECT 语句
type Selector[T any] struct {
	sb      strings.Builder
	args    []any
	table   string
	where   []Predicate
	having  []Predicate
	model   *model.Model
	db      *DB
	columns []Selectable
	groupBy []Column
    orderBy []OrderBy
	offset  int
	limit   int
}

func (s *Selector[T]) Build() (*Query, error) {
	var (
		t   T
		err error
	)
	s.model, err = s.db.r.Get(&t)
	if err != nil {
		return nil, err
	}
	s.sb.WriteString("SELECT ")
	if err = s.buildColumns(); err != nil {
		return nil, err
	}
	s.sb.WriteString(" FROM ")
	if s.table == "" {
		s.sb.WriteByte('`')
		s.sb.WriteString(s.model.TableName)
		s.sb.WriteByte('`')
	} else {
		s.sb.WriteString(s.table)
	}

	// 构造 WHERE
	if len(s.where) > 0 {
		// 类似这种可有可无的部分，都要在前面加一个空格
		s.sb.WriteString(" WHERE ")
		// WHERE 是不允许用别名的
		if err = s.buildPredicates(s.where); err != nil {
			return nil, err
		}
	}

	if len(s.groupBy) > 0 {
		s.sb.WriteString(" GROUP BY ")
		for i, c := range s.groupBy {
			if i > 0 {
				s.sb.WriteByte(',')
			}
			if err = s.buildColumn(c, false); err != nil {
				return nil, err
			}
		}
	}

	if len(s.having) > 0 {
		s.sb.WriteString(" HAVING ")
		// HAVING 是可以用别名的
		if err = s.buildPredicates(s.having); err != nil {
			return nil, err
		}
	}
	// 构造 orderby
    if len(s.orderBy) > 0 {
		s.sb.WriteString(" ORDER BY ")
		if err = s.buildOrderBy(); err != nil {
			return nil, err
		}
	}

	if s.limit > 0 {
		s.sb.WriteString(" LIMIT ?")
		s.addArgs(s.limit)
	}

	if s.offset > 0 {
		s.sb.WriteString(" OFFSET ?")
		s.addArgs(s.offset)
	}

	s.sb.WriteString(";")
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) Offset(offset int) *Selector[T] {
	s.offset = offset
	return s
}

func (s *Selector[T]) Limit(limit int) *Selector[T] {
	s.limit = limit
	return s
}
```

