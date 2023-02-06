# SELECT JOIN

## SQL 语法分析

 JOIN 查询有点像我们的 Expression，就是可以 查询套查询无限套下去。  

### MySQL 

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675253105690-516c68e6-038b-4a1d-97fb-12441ecba0ea.png#averageHue=%23f7f6f6&clientId=ub764e046-43c5-4&from=paste&height=572&id=uf50e4f3f&name=image.png&originHeight=629&originWidth=1082&originalType=binary&ratio=1&rotation=0&showTitle=false&size=102452&status=done&style=none&taskId=u4a2994c9-2f2e-4fef-8c02-9b30ec4d86b&title=&width=983.6363423166201)
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675253193286-831e479e-1ed6-47b7-8002-852c3e052250.png#averageHue=%23f7f6f6&clientId=ub764e046-43c5-4&from=paste&height=595&id=u29a00a92&name=image.png&originHeight=654&originWidth=1079&originalType=binary&ratio=1&rotation=0&showTitle=false&size=88737&status=done&style=none&taskId=u2f848843-7474-4075-a38d-6e725c5a19d&title=&width=980.9090696484594)

###  SQLite  

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675253461931-e4b862b8-d9b9-479e-8987-63fa89920c58.png#averageHue=%23efeded&clientId=ub764e046-43c5-4&from=paste&height=455&id=ue5915de8&name=image.png&originHeight=501&originWidth=894&originalType=binary&ratio=1&rotation=0&showTitle=false&size=121199&status=done&style=none&taskId=ud6b80aa7-2e3e-4f46-89b8-f04fa943f9f&title=&width=812.7272551118839)
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675253487222-467c8c2a-8dd1-43d9-9aba-27ab144b37d7.png#averageHue=%23ece9e9&clientId=ub764e046-43c5-4&from=paste&height=593&id=u64242d32&name=image.png&originHeight=652&originWidth=956&originalType=binary&ratio=1&rotation=0&showTitle=false&size=259882&status=done&style=none&taskId=u8bc271e1-34d0-4116-a2a4-ad59c7fac64&title=&width=869.0908902538713)
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675253503818-300e5f1a-b641-4b2f-ba89-c71e573cb95c.png#averageHue=%23eeeeee&clientId=ub764e046-43c5-4&from=paste&height=205&id=u7b3e2329&name=image.png&originHeight=225&originWidth=1150&originalType=binary&ratio=1&rotation=0&showTitle=false&size=63842&status=done&style=none&taskId=u0824d146-3bd8-4e38-a890-801bd6be5eb&title=&width=1045.4545227949288)

###  PostgreSQL  

 和 MySQL、SQLite 也差不多
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675253722138-70f90639-f0da-492c-9520-9dffb5310272.png#averageHue=%23eff2f4&clientId=ub764e046-43c5-4&from=paste&height=192&id=u54e368e4&name=image.png&originHeight=211&originWidth=1424&originalType=binary&ratio=1&rotation=0&showTitle=false&size=175237&status=done&style=none&taskId=u1c8925ef-7ce5-4af4-ae65-eea7ce29655&title=&width=1294.5454264869381)

###  JOIN 语法总结  

** JOIN 语法有两种形态  **
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675253880032-0146736e-795f-48e5-a199-a01ae192ebf2.png#averageHue=%2347392a&clientId=ub764e046-43c5-4&from=paste&height=123&id=u3d7788e2&name=image.png&originHeight=135&originWidth=1056&originalType=binary&ratio=1&rotation=0&showTitle=false&size=90737&status=done&style=none&taskId=u5ae853a6-aed9-499f-8046-a58522b0ba6&title=&width=959.9999791925608)

- JOIN ... ON 
- JOIN ... USING：USING 后面使用的是列  

** JOIN 本身有：**

- INNER JOIN， JOIN 
- LEFT JOIN，RIGHT JOIN

也就是说 JOIN 的结构大概可以描述成下面这几种，而且还可以嵌套。

- 表 JOIN 表
- （表 JOIN 表） JOIN 表
- 表 JOIN 子查询
- 子查询 JOIN 子查询

## 开源实例

###  Beego JOIN 查询  

** Beego 的 JOIN 查询主要出现在两个地方 ：**
 QueryBuilder 接口设计了 InnerJoin、LeftJoin 和 RightJoin 三个方法
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675254621997-1d412a17-a609-4b4b-a4d7-d8a77f017672.png#averageHue=%23322e2c&clientId=ub764e046-43c5-4&from=paste&height=336&id=u15d072e9&name=image.png&originHeight=370&originWidth=870&originalType=binary&ratio=1&rotation=0&showTitle=false&size=67976&status=done&style=none&taskId=uc323842a-409e-46cc-8685-79cbcaede02&title=&width=790.9090737665985)
 Beego 本身支持一对一、一对多和多对多的关联关 系，所以如果设置了正确的关联关系，那么 Beego 在部分情况下会生成 JOIN 查  

###  GROM JOIN 查询  

 GORM中 JOIN 查询主要是为了所谓的 Preload 而服务的  
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675254825403-447d9333-82f5-4d41-918b-38bf6dc2fbf2.png#averageHue=%232d2c2b&clientId=ub764e046-43c5-4&from=paste&height=253&id=u52f22e9f&name=image.png&originHeight=278&originWidth=1324&originalType=binary&ratio=1&rotation=0&showTitle=false&size=42430&status=done&style=none&taskId=u98703bca-1a66-4a54-ad3f-0c1d7324caa&title=&width=1203.6363375482485)

## 相关接口与方法的设计

实现简单的 JOIN 其实不怎么复杂，主要功能包括起别名、选择列、复杂一点的是 JOIN 可以嵌套。
常用的 JOIN 结构大概就是下面这几个：

- 表 A JOIN 表 B ON …
- 表 A AS 新名字 JOIN 表 B AS 新名字 ON …
- 表 A JOIN (表 B JOIN 表 C ON …) ON …

之前处理 FROM 后面那个位置的时候是直接用的数据库表名，但是那个位置其实可以放的玩意有表名、JOIN、子查询，这明摆了是要有一个抽象的，官方文档也告诉你了，叫 table_references（这里叫表表达式）。但是这三种对象的处理方式肯定是不一样的，语句形态差的都很远，基本没有共同点。

###  TableReference 抽象

- Table： 代表普通的表 
- Join：代表 JOIN 查询 
- Subquery：子查询  

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675255414165-d64dad14-0893-46f8-aee7-0f026ffcf99a.png#averageHue=%23fdfdfd&clientId=ub764e046-43c5-4&from=paste&height=495&id=uf89182c3&name=image.png&originHeight=742&originWidth=1228&originalType=binary&ratio=1&rotation=0&showTitle=false&size=61096&status=done&style=none&taskId=uef401d01-b759-47ea-b136-32570479e0a&title=&width=818.6666666666666)

```go
type TableReference interface {
	tableAlias() string
}
```

>  TableReference 可以在将来有需要的时候 不断增加方法。  

###  JoinBuilder 和 Join 定义  

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675255665271-ceed3459-0846-4b73-8f4d-46d534ce6cac.png#averageHue=%23f7f6f6&clientId=ub764e046-43c5-4&from=paste&height=180&id=ue3c468e9&name=image.png&originHeight=270&originWidth=1348&originalType=binary&ratio=1&rotation=0&showTitle=false&size=81237&status=done&style=none&taskId=u18533494-54c0-4a5e-ae33-c870bab3643&title=&width=898.6666666666666)

```go
var _ TableReference = Join{}

type JoinBuilder struct {
    left TableReference
    right TableReference
    typ string
}

type Join struct {
    left TableReference
    right TableReference
    typ string
    on []Predicate
    using []string
}

func (j Join) tableAlias() string {
	return ""
}

```

 JoinBuilder 里面的 On 和 Using 是终结方法，也就是 直接返回了 Join  ,   这种设计可以避免用户同时调用 On  或者 Using。

```go
func (j *JoinBuilder) On(ps...Predicate) Join {
	return Join {
		left: j.left,
		right: j.right,
		on: ps,
		typ: j.typ,
	}
}

func (j *JoinBuilder) Using(cs...string) Join {
	return Join {
		left: j.left,
		right: j.right,
		using: cs,
		typ: j.typ,
	}
}
```

 JOIN 本身也可以进一步 JOIN，所以我们 同样需要在 Join 上面定义类似的方法。 同样地，子查询也可以用来构造 JOIN 查 询。  

```go
func (j Join) Join(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: target,
		typ:   "JOIN",
	}
}

func (j Join) LeftJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: target,
		typ:   "LEFT JOIN",
	}
}

func (j Join) RightJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: target,
		typ:   "RIGHT JOIN",
	}
}
```

###  Table  ( 普通表 )

 Table 代表一个普通的表，它也是 JOIN 查询的起点

```go
type Table struct {
	entity any
	alias string
}

func TableOf(entity any) Table {
	return Table{
		entity: entity,
	}
}

func (t Table) tableAlias() string {
	return t.alias
}

func (t Table) As(alias string) Table {
	return Table {
		entity: t.entity,
		alias: alias,
	}
}

func (t Table) Join(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left: t,
		right: target,
		typ: "JOIN",
	}
}

func (t Table) LeftJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left: t,
		right: target,
		typ: "LEFT JOIN",
	}
}

func (t Table) RightJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left: t,
		right: target,
		typ: "RIGHT JOIN",
	}
}
```

##  重构 Selector

 将 From 改为接收一个 TableReference 作为 输入，后续 subquery 同样可以复用这个方法。  

```go
func (s *Selector[T]) From(tbl TableReference) *Selector[T] {
	s.table = tbl
	return s
}
```

- case nil : 如果用户没有调用 From 方法 
- case Table：用户传入了一个普通的表
- case Join：是一个 Join 查询 

```go
func (s *Selector[T]) buildTable(table TableReference) error {
	switch tab := table.(type) {
	case nil:
		s.quote(s.model.TableName)
	case Table:
		m, err := s.r.Get(tab.entity)
		if err != nil {
			return err
		}
		s.quote(m.TableName)
		if tab.alias != "" {
			s.sb.WriteString(" AS ")
			s.quote(tab.alias)
		}
	case Join:
		return s.buildJoin(tab)
	default:
		return errs.NewErrUnsupportedExpressionType(tab)
	}
	return nil
}

func (s *Selector[T]) buildJoin(tab Join) error {
	s.sb.WriteByte('(')
	if err := s.buildTable(tab.left); err != nil {
		return err
	}
	s.sb.WriteString(" ")
	s.sb.WriteString(tab.typ)
	s.sb.WriteString(" ")
	if err := s.buildTable(tab.right); err != nil {
		return err
	}
	if len(tab.using) > 0 {
		s.sb.WriteString(" USING (")
		for i, col := range tab.using {
			if i > 0 {
				s.sb.WriteByte(',')
			}
			err := s.buildColumn(Column{name: col}, false)
			if err != nil {
				return err
			}
		}
		s.sb.WriteString(")")
	}
	if len(tab.on) > 0 {
		s.sb.WriteString(" ON ")
		err := s.buildPredicates(tab.on)
		if err != nil {
			return err
		}
	}
	s.sb.WriteByte(')')
	return nil
}

```

###  重构列校验逻辑  

在原本不支持 JOIN 查询的时候，只需要 看一下操作的元数据里面有没有这个列。在支持了 JOIN 之后，那么所有的列、聚合函 数都可能有一个拥有者（owner）。 例如** t1.col1 其中 t1 就是 col1 的拥有者**。   那么校验逻辑就是：  

-  **如果用户指定了表，那么就检查指定的表上 面有没有这个列  **
-  **如果没有指定表，就走老逻辑，也就是右图 case nil 的分支** 

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
            // buildColumn 方法
			if err := s.buildColumn(val, true); err != nil {
				return err
			}
		case Aggregate:
			if err := s.buildAggregate(val, true); err != nil {
				return err
			}
		case RawExpr:
			s.raw(val)
		default:
			return errs.NewErrUnsupportedSelectable(c)
		}
	}
	return nil
}
```

buildColumn 方法

```go
func (s *Selector[T]) buildColumn(c Column, useAlias bool) error {
	err := s.builder.buildColumn(c.table, c.name)
	if err != nil {
		return err
	}
	if useAlias {
		s.buildAs(c.alias)
	}
	return nil
}
```

```go
// buildColumn 构造列
// 如果 table 没有指定，我们就用 model 来判断列是否存在
func (b *builder) buildColumn(table TableReference, fd string) error {
	var alias string
	if table != nil {
		alias = table.tableAlias()
	}
	if alias != "" {
		b.quote(alias)
		b.sb.WriteByte('.')
	}
	colName, err := b.colName(table, fd)
	if err != nil {
		return err
	}
	b.quote(colName)
	return nil
}

func (b *builder) colName(table TableReference, fd string) (string, error) {
	switch tab := table.(type) {
	case nil:
		fdMeta, ok := b.model.FieldMap[fd]
		if !ok {
			return "", errs.NewErrUnknownField(fd)
		}
		return fdMeta.ColName, nil
	case Table:
		m, err := b.r.Get(tab.entity)
		if err != nil {
			return "", err
		}
		fdMeta, ok := m.FieldMap[fd]
		if !ok {
			return "", errs.NewErrUnknownField(fd)
		}
		return fdMeta.ColName, nil
	case Join:
		colName, err := b.colName(tab.left, fd)
		if err != nil {
			return colName, nil
		}
		return b.colName(tab.right, fd)
	default:
		return "", errs.NewErrUnsupportedExpressionType(tab)
	}
}
```

## 总结

- GORM 的 Preload 是什么？本质上就是一个 JOIN 查询，并且严格来说，在 Go 语言里面是很难实现 lazy load的。GORM 的 Preload 就是通过 Join 把相关的数据都查询出来，并且组装成结构体 ； 
- WHERE、ON 和 HAVING 的区别：在 JOIN 查询里面，一般的建议 都是尽量把条件放到 ON 上面，这样 JOIN 生成的中间数据要少很多；
- JOIN 的执行原理：可以近似理解为一个双重循环 。 