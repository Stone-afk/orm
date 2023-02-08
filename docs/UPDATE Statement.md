# UPDATE Statement

## 语法分析

### MySQL 语法  

 MySQL 的 UPDATE 语句有两种形态 ：

-  更新单表的： 额外支持了 ORDER BY 和 LIMIT

![](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675762233714-3a2cbb15-59be-4284-8eab-ea818ce18fe1.png#averageHue=%23f7f3f3&from=url&id=SXQLI&originHeight=442&originWidth=1119&originalType=binary&ratio=1&rotation=0&showTitle=false&status=done&style=none&title=)

-  更新多表的：只支持 WHERE 条件  

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675762365595-2575cf83-b1a7-45fd-9b31-165acc5c5244.png#averageHue=%23f6f6f5&clientId=ufe9e2719-1725-4&from=paste&height=84&id=u77f6b86e&name=image.png&originHeight=126&originWidth=1080&originalType=binary&ratio=1&rotation=0&showTitle=false&size=20463&status=done&style=none&taskId=u18c017d7-8a8f-40ca-b91f-3b8ff47a241&title=&width=720)

### SQLite 语法  

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675762541333-c96539e2-593a-4268-8edb-9d7eba135aaa.png#averageHue=%23eeeeed&clientId=ufe9e2719-1725-4&from=paste&height=461&id=u88b7cb2a&name=image.png&originHeight=692&originWidth=714&originalType=binary&ratio=1&rotation=0&showTitle=false&size=164527&status=done&style=none&taskId=ue9886764-38f4-4a69-b9be-77fab92d1ae&title=&width=476)
从图里面看，SQLite 的语法形态总结为： UPDATE xxx SET xxxx WHERE xx ； 除此以外，它还额外支持了 FROM，但是不支持 ORDER BY 和 LIMIT  

### PostgreSQL 语法

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675762672065-c1fbbbdd-4869-4783-b132-bcb061e88a2c.png#averageHue=%23f3f4f7&clientId=ufe9e2719-1725-4&from=paste&height=195&id=ucc0a8f56&name=image.png&originHeight=293&originWidth=837&originalType=binary&ratio=1&rotation=0&showTitle=false&size=166212&status=done&style=none&taskId=u4cbe8f8d-4364-4161-bd3d-e5b6fd79f7d&title=&width=558)
从右图中来看，PostgreSQL 和 SQLite 的语法很像： UPDATE xxx SET xxxx , 它也支持 FROM，但是也不支持 ORDER BY 和 LIMIT。

## 开源实例

### Beego ORM  

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675763277687-417a209d-7e24-41dc-82dd-b19457cd554e.png#averageHue=%23312d2c&clientId=ufe9e2719-1725-4&from=paste&height=398&id=u08e8b5de&name=image.png&originHeight=597&originWidth=1292&originalType=binary&ratio=1&rotation=0&showTitle=false&size=115453&status=done&style=none&taskId=u71a087be-15df-4f0e-8c72-ee7075f9d5b&title=&width=861.3333333333334)
Beego 的 Update API 定义还是很简单的。 它默认是根据主键进行更新，如果 cols 没有传，那 么就默认更新所有的字段。   它构造 SQL 的部分非常复杂，不具备参考价值。  

### GORM ORM  

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675763668201-bfb30969-7712-4a2b-ae6f-66edc9e8b4df.png#averageHue=%232e2d2c&clientId=ufe9e2719-1725-4&from=paste&height=213&id=ude372ac4&name=image.png&originHeight=319&originWidth=1271&originalType=binary&ratio=1&rotation=0&showTitle=false&size=58720&status=done&style=none&taskId=uae3a07f5-89d6-4db4-8a3e-d200d945b4e&title=&width=847.3333333333334)
GORM 的 UPDATE 相关的方法有四 个：

- ** Update 和 Updates**：区别在于更新单个字段还是 更新多个字段。Updates 支持传入 map  ；
- **UpdateColumn 和 UpdateColumns**：和 Update、Updates 比起来，这两个方法支持复杂的表达式。  

```go
	db.Model(&product).Update("Price", 200)
	// Update - update multiple fields
	db.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // non-zero fields
	db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})
	db.Model(&product).UpdateColumn("Price", 200)
	db.Model(&product).UpdateColumns(Product{Price: 200, Code: "F42"})
```

## API 设计

只需要考虑最简单的 UPDATE xxx SET xxxx WHERE xxx 的形态。 同时考虑支持复杂的表达式，例如自增。 到这一步，几乎就能断定需要 一个新的结构体 Updater  , 并且同时需要实现 QueryBuilder 与 Executor 接口。

```go
type Updater[T any] struct {
}

func (u *Updater[T]) Build() (*Query, error) {
}

func (u *Updater[T]) Exec(ctx context.Context) Result {
}
```

## 具体实现

###  复用 Assignable  

 在支持 INSERT 语句中的 upsert 写法的时候，我们 定义了一个 Assignable 接口，它代表的是一个 a=b 的抽象。  

```go
func (u *Updater[T]) Set(assigns ...Assignable) *Updater[T] {
	u.assigns = assigns
	return u
}
```

 **支持已有的 Assignable 实现  **
 现在 Assignable 就只有 Column 和 Assignment 两个实现，支持起来很简单。  
 逻辑： 

- 如果用户指定的是 Column，那么我们就从传入 进来的实体里面读取字段的值，作为更新的值 ；
- 如果传进来的是 Assignment，那么我们就直接 使用 Assignment 的值  ；

```go
func (u *Updater[T]) buildAssignment(assign Assignment) error {
	if err := u.buildColumn(assign.column); err != nil {
		return err
	}
	u.sb.WriteByte('=')
	return u.buildExpression(assign.val)
}
```

###  集成 WHERE  

 在 SELECT 语句里面，已经支持了 WHERE。现在我们需要在 UPDATE 里面也支持 WHERE 了。 方法也很简单，把 Selector 里面的和 Predicate 相 关的代码提升到 builder  下

```go
func (u *Updater[T]) Where(ps ...Predicate) *Updater[T] {
	u.where = ps
	return u
}
```

### 复杂表达式的支持

在日常中，比较常使用的 UPDATE 语句，如 UPDATE xxx SET age=age+1。 这种是常见的自增。 另外是更新时间： UPDATE xxx SET update_time = now()

- 方案一：使用 RawExpr 
- 方案二：在 Column 上定义新的方法 
- 方案三：全都要  

 RawExpr 肯定要支持，因为不确定用户是不是 会写出来各种奇奇怪怪的表达式。  除了 RawExpr  如果需要引入新方法，那就必须定义新的抽象。

####  MathExpr  支持

MathExpr 实现了 Expression 接口，所以基本上它可以非常灵活地用于 UPDATE 语句。 

```go
type MathExpr struct {
    left Expression
	op    op
    right Expression
}
```

** MathExpr 与 Predicate**  
 MathExpr 和 Predicate 非常像，所以可以为这 两种结构体引入一个公共的抽象：binaryExpr。   它代表的是二元操作符，也就是 a op b 的形态 ,  这也是 Go 里面一种独有的设计技巧。即本质上 内核就是一个，但是对外部用户而言有不同的形式 （实现了不同接口）。  

```go
type binaryExpr struct {
	left  Expression
	op    op
	right Expression
}

func (binaryExpr) expr() {}
```

```go
type Predicate binaryExpr
```

>  本质都是 binaryExpr，但是对外表现成了 MathExpr 和 Predicate 两种。  

```go
type MathExpr binaryExpr

func (m MathExpr) Add(val interface{}) MathExpr {
	return MathExpr{
		left:  m,
		op:    opAdd,
		right: valueOf(val),
	}
}

func (m MathExpr) Multi(val interface{}) MathExpr {
	return MathExpr{
		left:  m,
		op:    opMulti,
		right: valueOf(val),
	}
}

func (m MathExpr) expr() {}
```

而 Column 本身就变成了构建 MathExpr 的起点

```go
func (c Column) Add(delta int) MathExpr {
	return MathExpr{
		left: c,
		op: opAdd,
		right: value{val: delta},
	}
}

func (c Column) Multi(delta int) MathExpr {
	return MathExpr{
		left: c,
		op: opAdd,
		right: value{val: delta},
	}
}
```

 最终大家都是 Expression，所以都调到 buildExpression 上 , 所以这里需要重构 buildExpression：

```go
func (b *builder) buildExpression(e Expression) error {
	if e == nil {
		return nil
	}
	switch exp := e.(type) {
	case Column:
		return b.buildColumn(exp.name)
	case Aggregate:
		return b.buildAggregate(exp, false)
	case value:
		b.sb.WriteByte('?')
		b.addArgs(exp.val)
	case RawExpr:
		b.sb.WriteString(exp.raw)
		if len(exp.args) != 0 {
			b.addArgs(exp.args...)
		}
	case MathExpr:
		return b.buildBinaryExpr(binaryExpr(exp))
	case Predicate:
		return b.buildBinaryExpr(binaryExpr(exp))
	case binaryExpr:
		return b.buildBinaryExpr(exp)
	default:
		return errs.NewErrUnsupportedExpressionType(exp)
	}
	return nil
}

func (b *builder) buildBinaryExpr(e binaryExpr) error {
	err := b.buildSubExpr(e.left)
	if err != nil {
		return err
	}
	if e.op != "" {
		b.sb.WriteByte(' ')
		b.sb.WriteString(e.op.String())
	}
	if e.right != nil {
		b.sb.WriteByte(' ')
		return b.buildSubExpr(e.right)
	}
	return nil
}

func (b *builder) buildSubExpr(subExpr Expression) error {
	switch sub := subExpr.(type) {
	case MathExpr:
		_ = b.sb.WriteByte('(')
		if err := b.buildBinaryExpr(binaryExpr(sub)); err != nil {
			return err
		}
		_ = b.sb.WriteByte(')')
	case binaryExpr:
		_ = b.sb.WriteByte('(')
		if err := b.buildBinaryExpr(sub); err != nil {
			return err
		}
		_ = b.sb.WriteByte(')')
	case Predicate:
		_ = b.sb.WriteByte('(')
		if err := b.buildBinaryExpr(binaryExpr(sub)); err != nil {
			return err
		}
		_ = b.sb.WriteByte(')')
	default:
		if err := b.buildExpression(sub); err != nil {
			return err
		}
	}
	return nil
}

```

 **buildExpression 和 GORM 的设计对比**
 可以看到，本文设计是采用了 switch case，平摊下 来，构造 SQL 的过程聚在一起。 GORM 则是不同，利用 Expression 和 Build 抽象分 散在各种实现里面，而后拼凑在一起。  

### Build 构造

```go
type Updater[T any] struct {
	builder
	assigns []Assignable
	val     *T
	where   []Predicate

	sess session
	core
}

func NewUpdater[T any](sess session) *Updater[T] {
	c := sess.getCore()
	return &Updater[T]{
		builder: builder{
			dialect: c.dialect,
			quoter:  c.dialect.quoter(),
		},
		sess: sess,
		core: c,
	}
}

func (u *Updater[T]) Build() (*Query, error) {
	if len(u.assigns) == 0 {
		return nil, errs.ErrNoUpdatedColumns
	}
	var (
		err error
		t   T
	)
	u.model, err = u.r.Get(&t)
	if err != nil {
		return nil, err
	}
	u.sb.WriteString("UPDATE ")
	u.quote(u.model.TableName)
	u.sb.WriteString(" SET ")
	val := u.valCreator(u.val, u.model)
	for i, a := range u.assigns {
		if i > 0 {
			u.sb.WriteByte(',')
		}
		switch assign := a.(type) {
		case Column:
			if err = u.buildColumn(assign.name); err != nil {
				return nil, err
			}
			u.sb.WriteString("=?")
			arg, err := val.Field(assign.name)
			if err != nil {
				return nil, err
			}
			u.addArgs(arg)
		case Assignment:
			if err = u.buildAssignment(assign); err != nil {
				return nil, err
			}
		default:
			return nil, errs.NewErrUnsupportedAssignableType(a)
		}
	}
	if len(u.where) > 0 {
		u.sb.WriteString(" WHERE ")
		if err = u.buildPredicates(u.where); err != nil {
			return nil, err
		}
	}
	u.sb.WriteByte(';')
	return &Query{
		SQL:  u.sb.String(),
		Args: u.args,
	}, nil
}

func (u *Updater[T]) Exec(ctx context.Context) Result {
	q, err := u.Build()
	if err != nil {
		return Result{err: err}
	}
	res, err := u.db.db.ExecContext(ctx, q.SQL, q.Args...)
	return Result{err: err, res: res}
}


```

## 单元测试

```go
func TestUpdater_Build(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct {
		name    string
		u       QueryBuilder
		want    *Query
		wantErr error
	}{
		{
			name:    "no columns",
			u:       NewUpdater[TestModel](db),
			wantErr: errs.ErrNoUpdatedColumns,
		},
		{
			name: "single column",
			u: NewUpdater[TestModel](db).Update(&TestModel{
				Age: 18,
			}).Set(C("Age")),
			want: &Query{
				SQL:  "UPDATE `test_model` SET `age`=?;",
				Args: []any{int8(18)},
			},
		},
		{
			name: "assignment",
			u: NewUpdater[TestModel](db).Update(&TestModel{
				Age:       18,
				FirstName: "Tom",
			}).Set(C("Age"), Assign("FirstName", "DaMing")),
			want: &Query{
				SQL:  "UPDATE `test_model` SET `age`=?,`first_name`=?;",
				Args: []any{int8(18), "DaMing"},
			},
		},
		{
			name: "where",
			u: NewUpdater[TestModel](db).Update(&TestModel{
				Age:       18,
				FirstName: "Tom",
			}).Set(C("Age"), Assign("FirstName", "DaMing")).
				Where(C("Id").EQ(1)),
			want: &Query{
				SQL:  "UPDATE `test_model` SET `age`=?,`first_name`=? WHERE `id` = ?;",
				Args: []any{int8(18), "DaMing", 1},
			},
		},
		{
			name: "incremental",
			u: NewUpdater[TestModel](db).Update(&TestModel{
				Age:       18,
				FirstName: "Tom",
			}).Set(Assign("Age", C("Age").Add(1))),
			want: &Query{
				SQL:  "UPDATE `test_model` SET `age`=`age` + ?;",
				Args: []any{1},
			},
		},
		{
			name: "incremental-raw",
			u: NewUpdater[TestModel](db).Update(&TestModel{
				Age:       18,
				FirstName: "Tom",
			}).Set(Assign("Age", Raw("`age`+?", 1))),
			want: &Query{
				SQL:  "UPDATE `test_model` SET `age`=`age`+?;",
				Args: []any{1},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.u.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.want, q)
		})
	}
}

```

## 总结

- **UPDATE 如实现复杂的表达式?** 首先需要一个 二元表达式，其实现了 Expression 接口，与 WHERE 条件类似，然后利用 column 为起点，构造方法；
- **UPDATE 如何实现 SET 方法**？与 Insert 一样，待修改的列沿用 Assignable 抽象，然后利用 relfect 或者 unsafe 将目标值取出。