## 简单 SELECT

### sql 语句构造分析

要从orm转变为一个sql查询，那么首先第一步是先解决如何构造一个sql；这可以以参考不同orm的设计：

#### Beego ORM

Beego 通过鲜明的语义的构造形式来告诉用户，这个接口是用来满足某些方法的

- **优点**：对于用户来说，这些 API 极其简单，懂sql语句的用户来说看语义就基本明白这个接口的功能

- **缺点**：代码耦合性强，可扩展性差

- **耦合性强**：SQL 的构造和执行得到结果集处理混在一起，职责不清晰

  ![image-20230115161216562](/docs/images/image-20230115161216562.png)
  ![image-20230115161521682](/docs/images/image-20230115161521682.png)

- **可扩展性差**：SELECT 的语法形式可以说非常复杂，这种处理方式难以支持完整的语法

  ![image-20230115160349817](/docs/images/image-20230115160349817.png)

**Beego 的另一种设计**，将构造 sql 语句的功能统一一个抽象 QueryBuilder， 这种设计解耦了 sql 语句的构造与执行结果集功能，使得扩展性大大提高

  ![image-20230115161911506](/docs/images/image-20230115161911506.png)
  
但深入这个 QueryBuilder 的实现可以发现，具体的实现sql构造的功能是根据 sql 语义的顺序来调用的（**可以观察到使用字符串不断的往后追究拼接的参数就很好的证明了这点**）

![image-20230115162235094](/docs/images/image-20230115162235094.png)

![image-20230115162256253](/docs/images/image-20230115162256253.png)

- **优点**: 操作语义简单易懂，用户看一眼就懂；
- **缺点**：约束性太强，难以灵活构建sql；

#### Gorm

Gorm主要分为以下**四个抽象:**

- **Builder**: sql 语句构造的抽象

  ![image-20230115170607021](/docs/images/image-20230115170607021.png)

- **Expression**：sql 语句表达式，表达式和表达式可以组合成复合表达式

  ![image-20230115170655117](/docs/images/image-20230115170655117.png)

- **Clause**：按照特定需要组合而成的 sql 的一个部分

  ![image-20230115170625057](/docs/images/image-20230115170625057.png)

- **Interface**：构造它自身，以及和其它部分 Clause 组合

  ![image-20230115170739363](/docs/images/image-20230115170739363.png)

可以看到 Gorm 的 sql 构造的基本思想是，sql 语句不同的部分分开构造，最后再组合在一起，这就意味着不再像 beego orm 那样需要按照 sql 语义的顺序来调用构造的方法了。

例如 ： **这是一个 Expression 的实现**，对应到 SELECT XXX 这个单一部分。

![image-20230115164210852](/docs/images/image-20230115164210852.png)

#### Ent

Ent 可以称为经典 Builder 模式。与 Beego orm 相比不要求调用顺序，与 Gorm 相比没有复杂的接口机制，这就导致灵活性不如 Gorm。

在设计业务系统、中间件的时候都要 平衡扩展性和系统复杂度的关系。 往往高扩展性带来的就是复杂的接口机 制。任何非功能特性都是有代价的。

![image-20230115170858727](/docs/images/image-20230115170858727.png)

![image-20230115171517458](/docs/images/image-20230115171517458.png)

### 定义核心接口

1. 使用 **Builder 模式**，不同的语句的具体实现不同；

   **QueryBuilder** 作为构建 SQL 这一个单独 步骤的顶级抽象

   ```go
   type QueryBuilder interface {
      Build() (*Query, error)
   }
   ```

2. 使用**泛型做约束**，这样这样约束目标模型的类型， 以免参数直接定义为接口后，传入未知的类型而报错；

   Querier 泛型约束接口，发起结果集查询的抽象

   ```go
   type Querier[T any] interface {
   	Get(ctx *context.Context) (*T, error)
   	GetMulti(ctx *context.Context) (*T, error)
   }
   ```

3. **Executor** 对于 Update、Insect、Dlete 语句的抽象

```go
type Executor interface {
	Exec(ctx *context.Context) (sql.Result, error)
}
```

### Selector 定义

#### SELECT 语句规范

**以MySQL 语法规范为例**： 

- **SELECT**：代表这是一个查询语句；
- **FROM**：普通表、子查询、JOIN 查询 ；
- **WHERE**：各种查询条件，以及由 AND、OR、NOT 混合在一起的复杂查询条件 ；
- **HAVING**：可以使用 WHERE 里面的条件，以及使用 聚合函数的条件 ；
- **ORDER BY** ：对表中字段排序；
- **GROUP BY** ：对表中字段分组；
- **LIMIT 和 OFFSE**：对查询数据做分页；

##### 首先先定义 Selector 结构体

```go
func NewSelector[T any]() *Selector[T] {
	return &Selector[T]{}
}

// Selector 用于构造 SELECT 语句
type Selector[T any] struct {
	table string
}

func (s *Selector[T]) Build() (*Query, error) {
    retuen nil, err
}
```

##### From 方法的的定义

需要 From 方法的原因是因为：

- 第一为了有更强的语义；
- 第二使用户可以自定义传入表名，不依赖泛型的约束的模型结构体，如果不使用 From 方 法，默认就用泛型类型作为表名；

```go
// From 指定表名，如果是空字符串，那么将会使用默认表名
func (s *Selector[T]) From(tbl string) *Selector[T] {
	s.table = tbl
	return s
}
```

**WHERE 方法**

如果参考 From 的设计，Where 方法可以直接用接口作为参数接收者接收字符串和参数作为输入

例如 Grom：

![image-20230115182450396](/docs/images/image-20230115182450396.png)

**优点:**

- 实现简单
- 灵活

**缺点：**

- 缺少参数校验
- args 作为不定参数，容易误用切片

为了支持更加复杂的 Where，Where 方法不再接收一个字符串，而是接收结构化的 Predicate 作为输入；那么 Predicate 是什么？ 如何定义？

这里又有要参考一下 Gorm 了：

**Gorm** 首先定义了一个对表达式的抽象 **Expression**

![image-20230115183218332](/docs/images/image-20230115183218332.png)

![image-20230115183406923](/docs/images/image-20230115183406923.png)

定义对 **Expression** 的实现，例如 **AndConditions**、 **OrConditions**；以此来实现复杂的条语义的支持。

![image-20230115183541942](/docs/images/image-20230115183541942.png)

![image-20230115183911027](/docs/images/image-20230115183911027.png)

**总的来说 Gorm 的设计特点如下:**

- Expression 抽象 

- 各种表达式都有一个实现，例如 Eq、IN 、Not、And 和 Or 

- 被认为是一个 Expression 的集合

  ![image-20230115184121843](/docs/images/image-20230115184121843.png)



**再来参考 Ent 的 Where 设计:**

ent 的 where 是一个结构体指针 *Predicate

![image-20230115184301431](/docs/images/image-20230115184301431.png)

也采用 **Builder 模式**

![image-20230115184649437](/docs/images/image-20230115184649437.png)

![image-20230115184748911](/docs/images/image-20230115184748911.png)

![image-20230115184906454](/docs/images/image-20230115184906454.png)

这里将所有的条件拼接，组合到 *Predicate 里

![image-20230115185239028](/docs/images/image-20230115185239028.png)

![image-20230115190452709](/docs/images/image-20230115190452709.png)

将 *Predicate 里 的条件语句解析并拼接到 sql 中

![image-20230115185727802](/docs/images/image-20230115185727802.png)

![image-20230115185757364](/docs/images/image-20230115185757364.png)

**ent 的 特点是：**

- 函数式设计，任何一个 Predicate 都被认为是对 Selector 本身的修

这里 Gorm 的设计无疑是更灵活的，所以应该首先考虑与 Gorm 类似的设计；

定义 Expression 抽象，在定义 Predicate;

**查询条件可以看做是 Left Op Right 的模式**： 

- **基本比较符**：Left 是列名，Op 是各个比较符号，右边是表达式，常见的是一个值 
- **Not**：左边缺省，只剩下 Op Right，如 NOT (id = ?) 
- **And、Or**：左边右边都是一个 Predicat

```go
// Expression 代表语句，或者语句的部分
type Expression interface {
	expr()
}

// Predicate 代表一个查询条件
// Predicate 可以通过和 Predicate 组合构成复杂的查询条件
type Predicate struct {
	left  Expression
	op    op
	right Expression
}


func (Predicate) expr() {}

func exprOf(e any) Expression {
	switch exp := e.(type) {
	case Expression:
		return exp
	default:
		return valueOf(exp)
	}
}

```

```go
type Column struct {
	name string
}

func (c Column) expr() {}

type value struct {
	val any
}

func (c value) expr() {}

func valueOf(val any) value {
	return value{
		val: val,
	}
}

func C(name string) Column {
	return Column{name: name}
}

// EQ 例如 C("id").Eq(12)
func (c Column) EQ(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opEQ,
		right: exprOf(arg),
	}
}

func (c Column) LT(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opLT,
		right: exprOf(arg),
	}
}

func (c Column) GT(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opGT,
		right: exprOf(arg),
	}
}
```

**Predicate 是 op(比较符号) 之后构成一颗二叉树（只有 left right)**；这与 Gorm 设计的区别是 Gorm 实际上可以看作是多叉树 (切片代表了多叉树)

![image-20230115205139429](/docs/images/image-20230115205139429.png)

接下来直接在 再设计 buildExpression 方法， 理论上目前 Column、value、 Predicate 都属于 Expression 抽象的实现；

```go
type Selector[T any] struct {
	sb strings.Builder
	args []any
	table string
	where []Predicate
}

// Where 用于构造 WHERE 查询条件。如果 ps 长度为 0，那么不会构造 WHERE 部分
func (s *Selector[T]) Where(ps ...Predicate) *Selector[T] {
	s.where = ps
	return s
}

func (s *Selector[T]) Build() (*Query, error) {
	s.sb.WriteString("SELECT * FROM ")
	if s.table == "" {
		var t T
		s.sb.WriteByte('`')
		s.sb.WriteString(reflect.TypeOf(t).Name())
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
		if err := s.buildExpression(p); err != nil {
			return nil, err
		}
	}
	s.sb.WriteString(";")
	return &Query{
		SQL: s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) buildExpression(e Expression) error {
	if e == nil {
		return nil
	}
	switch exp := e.(type) {
	case Column:
		s.sb.WriteByte('`')
		s.sb.WriteString(exp.name)
		s.sb.WriteByte('`')
	case value:
		s.sb.WriteByte('?')
		s.args = append(s.args, exp.val)
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
		return fmt.Errorf("orm: 不支持的表达式 %v", exp)
	}
	return nil
}

```

**buildExpression** 其实逻辑很简单： 

- Column 代表是列名，直接拼接列名 
- value 代表参数，加入参数列表 
- Predicate 代表一个查询条件： 
  - 如果左边是一个 Predicate，那么加上括号 
  - 递归构造左边 
  - 构造操作符 
  - 如果右边是一个 Predicate，
  - 那么加上括号 
  - 递归构造右边

#####  简单的测试用例：

```go
func TestSelector_Build(t *testing.T) {
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			// From 都不调用
			name: "no from",
			q:    NewSelector[TestModel](),
			wantQuery: &Query{
				SQL: "SELECT * FROM `TestModel`;",
			},
		},
		{
			// 调用 FROM
			name: "with from",
			q:    NewSelector[TestModel]().From("`test_model_t`"),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model_t`;",
			},
		},
		{
			// 调用 FROM，但是传入空字符串
			name: "empty from",
			q:    NewSelector[TestModel]().From(""),
			wantQuery: &Query{
				SQL: "SELECT * FROM `TestModel`;",
			},
		},
		{
			// 调用 FROM，同时出入看了 DB
			name: "with db",
			q:    NewSelector[TestModel]().From("`test_db`.`test_model`"),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_db`.`test_model`;",
			},
		},
		{
			// 单一简单条件
			name: "single and simple predicate",
			q:    NewSelector[TestModel]().From("`test_model_t`").
				Where(C("Id").EQ(1)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model_t` WHERE `Id` = ?;",
				Args: []any{1},
			},
		},
		{
			// 多个 predicate
			name: "multiple predicates",
			q: NewSelector[TestModel]().
				Where(C("Age").GT(18), C("Age").LT(35)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE (`Age` > ?) AND (`Age` < ?);",
				Args: []any{18, 35},
			},
		},
		{
			// 使用 AND
			name: "and",
			q: NewSelector[TestModel]().
				Where(C("Age").GT(18).And(C("Age").LT(35))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE (`Age` > ?) AND (`Age` < ?);",
				Args: []any{18, 35},
			},
		},
		{
			// 使用 OR
			name: "or",
			q:    NewSelector[TestModel]().
				Where(C("Age").GT(18).Or(C("Age").LT(35))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE (`Age` > ?) OR (`Age` < ?);",
				Args: []any{18, 35},
			},
		},
		{
			// 使用 NOT
			name: "not",
			q:    NewSelector[TestModel]().Where(Not(C("Age").GT(18))),
			wantQuery: &Query{
				// NOT 前面有两个空格，因为我们没有对 NOT 进行特殊处理
				SQL:  "SELECT * FROM `TestModel` WHERE  NOT (`Age` > ?);",
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
```

### 简单 SELECT 总结

- GORM 是如何构造 SQL 的？**在 GORM 里面主要有四个抽象**：**Builder、Expression、Clause 和 Interface**。简单一句话概括 **GORM 的设计思路就是 SQL 的不同部分分开构造，最后再拼接在一起**。 
- **什么是 Builder 模式**？能**用来干什么**？用我们的 ORM 的例子就可以，**Builder 模式尤其适合用于构造复杂多变的对象**。 
- **在 ORM 框架使用泛型有什么优点**？能用来**约束用户传入的参数**或者**用户希望得到的返回值**，**加强类 型安全。**
