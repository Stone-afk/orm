# INSERT Statement

## MySQL 规范

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675580141680-88c5b9d3-2a7b-4b48-af5d-18d390fc55d6.png#averageHue=%23ebe7e0&clientId=u28c090ba-c046-4&from=paste&height=901&id=uc421b305&name=image.png&originHeight=1802&originWidth=1886&originalType=binary&ratio=1&rotation=0&showTitle=false&size=2523759&status=done&style=none&taskId=uf36f86ff-6e2d-48e4-b918-902604f5727&title=&width=943)
**MySQL 支持的 INSERT 语句有三种风格:**

- 最普通的
- 带 SET 赋值的
- 带 SELECT 子句的

同时这三种都支持 ON DUPLICATE KEY UPDATE 

## 开源实例

### Beego ORM

**api 设计**

- Insert 和 InsertWithCtx：普通的插入 
- InsertOrUpdate：也就是 UPSERT 语义 
- InsertMulti：也就是批量插入，但是 Beego 进行了封装，允许将一大批拆分成 几个小批次分批插入到数据库，例如你传入 1000 个，然后设定每一批只能有 100 个， 那么 Beego 会帮你拆成 10  批

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675581348455-f2fa520d-4eba-4186-9fb2-7d36ccadf688.png#averageHue=%23342f2d&clientId=u61ef4c7d-f833-4&from=paste&height=363&id=u997da104&name=image.png&originHeight=544&originWidth=1273&originalType=binary&ratio=1&rotation=0&showTitle=false&size=108511&status=done&style=none&taskId=ucaf0db8a-99d1-403b-9602-34a4ba2fe04&title=&width=848.6666666666666)
**具体实现**
这里根据 driver 的类型来实现语句中不同的方言
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675581466638-2febf094-3ce8-4d25-abca-5a3c84bee3be.png#averageHue=%23322f2e&clientId=u61ef4c7d-f833-4&from=paste&height=457&id=u1364fa65&name=image.png&originHeight=685&originWidth=1227&originalType=binary&ratio=1&rotation=0&showTitle=false&size=114799&status=done&style=none&taskId=u332b3b3d-84f1-421b-9dac-80ff25e5890&title=&width=818)
准备好了要插入的列、插入的值， 以及 INSERT or UPDATE 的处理，然后拼接 SQL  
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675581582635-176e80c7-f7de-42db-8263-bc5673dbd277.png#averageHue=%232e2c2c&clientId=u61ef4c7d-f833-4&from=paste&height=481&id=u58e184bb&name=image.png&originHeight=721&originWidth=1200&originalType=binary&ratio=1&rotation=0&showTitle=false&size=89036&status=done&style=none&taskId=u2f0bdcce-3136-4d16-b6af-59d584ee0e0&title=&width=800)

### GORM

**api 设计**
![](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675581711996-473a4980-3be0-4b4f-811c-8d6836c8735c.png#averageHue=%23302e2d&from=url&id=nGu6M&originHeight=410&originWidth=1265&originalType=binary&ratio=1&rotation=0&showTitle=false&status=done&style=none&title=)

-  Create：支持单个、批量，或者分批次插入  
-  CreateInBatches：分批次插入  
-  Save：更新，如果没有主键就是插入。里面会 进一步判断用户有没有设置 OnConflict 的子句  

**具体实现**
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675582412944-9271d655-496d-433e-b8be-630b3291fc23.png#averageHue=%232d2d2c&clientId=u61ef4c7d-f833-4&from=paste&height=451&id=u45797832&name=image.png&originHeight=677&originWidth=952&originalType=binary&ratio=1&rotation=0&showTitle=false&size=80125&status=done&style=none&taskId=u2e80e402-e6da-4ad4-a0b5-d06eb9aaaa5&title=&width=634.6666666666666)
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675582540347-98260e01-f66d-46f7-be25-67416cba20d3.png#averageHue=%232d2d2c&clientId=u61ef4c7d-f833-4&from=paste&height=410&id=ud600072a&name=image.png&originHeight=615&originWidth=956&originalType=binary&ratio=1&rotation=0&showTitle=false&size=82247&status=done&style=none&taskId=u25bebf2d-3c67-414d-8b40-4ef22dc8236&title=&width=637.3333333333334)
GORM 的实现非常分散： 

- INSERT 语句的拼接是通过 Insert、Values 等 几个 Clause 来实现的 
- 最终是在 processor 的 Execute 中执行  

## API设计

依旧延续 Builder 模式。从 Inserter 起步，在前面， SELECT 语句有相当一部分设计和代码是可以复用的。  

```go
type Inserter[T any] struct {
    values  []*T
    db      *DB
}

func NewInserter[T any](db *DB) *Inserter[T] {
    return &Inserter[T]{
        db: db,
    }
}

func (i *Inserter[T]) Values(vals ...*T) *Inserter[T] {
    i.values = vals
    return i
}

func (i *Inserter[T]) Build() (*Query, error) {

}
```

###  指定列  

```go
// Fields 指定要插入的列
// TODO 目前我们只支持指定具体的列，但是不支持复杂的表达式
// 例如不支持 VALUES(..., now(), now()) 这种在 MySQL 里面常用的
func (i *Inserter[T]) Columns(cols ...string) *Inserter[T] {
    i.columns = cols
    return i
}
```

INSERT 在某些情况下，可以指定要插入的列，比如在 TestModel 里面，可以指定只插入 Age、FirstName 和 LastName 三个列。 实际上，插入的列可以是 

- 普通列 
- 函数，例如 MySQL 上的 now 
- 复合表达式 

但是这里只支持普通列。  

###  UPSERT  

####  MySQL UPSERT  

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675584132291-bd129ce4-7975-4cb2-bd78-ccc3ea8d3fbd.png#averageHue=%23f6f2f2&clientId=u61ef4c7d-f833-4&from=paste&height=197&id=u25a00bea&name=image.png&originHeight=296&originWidth=860&originalType=binary&ratio=1&rotation=0&showTitle=false&size=46406&status=done&style=none&taskId=ub08c3f1f-8f40-4996-b959-1591be48443&title=&width=573.3333333333334)
现实中有时会遇到一种 INSERT or Update 的场 景，也就是所谓的 UPSERT  

- 如果数据不存在，那么就插入 
- 如果数据存在，那么就更新  

而数据存不存在，也就是判断冲突的标准，就是依靠主键，或者唯一索引冲突。 在 MySQL 里面这种特性是 ON DUPLICATE KEY。  
![](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675584242164-113aad0e-4bcb-4e96-88a8-6c7a0b48984c.png#averageHue=%23f6f5f5&from=url&id=ThjFp&originHeight=303&originWidth=833&originalType=binary&ratio=1&rotation=0&showTitle=false&status=done&style=none&title=)
进一步看 assignment 的右边，发现它有几种类型：  

1. value：纯粹的值
2. [row_alias.]col_name：也就是指定使用某行 的某个列的值 
3. [tbl_name.]col_name：也就是另一个列
4. [row_alias.]col_alias：指定使用某行的某个 列的值 

 只支持1、3 两种情况，因为在 INSERT 部分就没有支持 Row 的写法  
 从前面的分析来看，要支持 MySQL 的 UPSERT，至少需要两个东西：  

- ON CONFLICT KEY UPDAT  
- assignment 也就是赋值语句

```go
// Assignable 标记接口，
// 实现该接口意味着可以用于赋值语句，
// 用于在 UPDATE 和 UPSERT 中
type Assignable interface {
    assign()
}
```

**方案一：直接在 Inserter 里面维持一个 onConflict 的字 段。  **

```go
type Inserter[T any] struct {
    values  []*T
    db      *DB
    columns []string
    sb      strings.Builder
    args    []any
    model   *model.Model

    onDuplicate []Assignable
}


func (i *Inserter[T]) OnDuplicateKeyBuilder(assigns...Assignable) *Inserter[T] {
    i.onDuplicate = assigns
    return i
}

```

**方案二：使用一个 Upsert 结构体，从而允 许将来扩展更加复杂的行为（后面考虑不同数据库 的时候就能看到效果了）**

```go
type UpsertBuilder[T any] struct {
	i               *Inserter[T]
}

type Upsert struct {
	assigns         []Assignable
}


// Update 也可以看做是一个终结方法，重新回到 Inserter 里面
func (o *UpsertBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.onDuplicate = &Upsert{
		assigns:         assigns,
	}
	return o.i
}
```

这里采用在方案二的设计， 方案二利用了 OnDuplicateKeyBudiler 这种中间结构  
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675673220825-d84f74f9-482f-4264-bdb1-bf252277f2d3.png#averageHue=%23f9f9f9&clientId=u3030462a-8a1a-4&from=paste&height=242&id=ua6d10030&name=image.png&originHeight=363&originWidth=1091&originalType=binary&ratio=1&rotation=0&showTitle=false&size=27190&status=done&style=none&taskId=u63cf923d-fdb7-4dc8-867e-35544c2d763&title=&width=727.3333333333334)

####  更新为特定的值  

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675586464227-5f654a0b-bd25-4586-ba9e-d4a34c443947.png#averageHue=%2342382b&clientId=u61ef4c7d-f833-4&from=paste&height=165&id=u3ea7a0f4&name=image.png&originHeight=247&originWidth=1252&originalType=binary&ratio=1&rotation=0&showTitle=false&size=151277&status=done&style=none&taskId=u84b2d03a-86c0-4b42-b69a-6401a9cd930&title=&width=834.6666666666666)
 在这种情况下，考虑引入一种新的结构体 Assignment，表达更新为特定值的语义。  

```go
type Assignment struct {
	column string
	val Expression
}

func Assign(column string, val any) Assignment {
	v, ok := val.(Expression)
	if !ok {
		v = value{val: val}
	}
	return Assignment{
		column: column,
		val: v,
	}
}

func (a Assignment) assign() {}

```

#### 更新为插入的值  

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675586846211-7b6f1ff4-ea1d-4706-b10b-4e42f34fe265.png#averageHue=%233f392c&clientId=u61ef4c7d-f833-4&from=paste&height=112&id=uc5e972a7&name=image.png&originHeight=168&originWidth=876&originalType=binary&ratio=1&rotation=0&showTitle=false&size=104570&status=done&style=none&taskId=u9032dfaf-c0f0-4c91-8616-e367e0096ca&title=&width=584)
第二种情况，更新为插入的值。 更新为插入的值，需要用到 MySQL 里面的语法: 
**UPDATE col1 = VALUES(col1) **
这里我们希望改造一下原本的 Column 结构，这样可 以直接使用。  

```go
type Column struct {
	name  string
	alias string
}

func (c Column) assign() {}
```

#### 方言抽象  Dialect   

**UPSERT 中的方言问题  **
SQL 作为一个不太强的规范，有些数据库会有一些自己定义 的语法，**这种不同数据库支持的 SQL 称作方言** 。 在前面处理 SELECT 的时候，根本没有考虑方言的问题， 而到了 UPSERT 这里，就不得不考虑了  
例如 UPSERT 语句，在标准 SQL 里面是 ON CONFLICT(col1, col2...) DO UPDATE SET，SQLite 和 PostgreSQL 都遵循这 种风格 。
其实，还有一个也没考虑方言，就是 ` 。用于引用列名或 者表名，在不同的方言里面也是不同的，这一次一并解 决。  
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675660951246-261c74d9-cac3-4a1e-846c-db9d80b07eba.png#averageHue=%23e2ddd4&clientId=u7b20470e-8cf7-4&from=paste&height=270&id=u3f274e60&name=image.png&originHeight=405&originWidth=936&originalType=binary&ratio=1&rotation=0&showTitle=false&size=311045&status=done&style=none&taskId=uf01c3335-8f91-41f5-89c2-eaafdfb6ed1&title=&width=624)
需要一个抽象，来帮我们解决不同方言之间 SQL 语句不 同的问题。这个抽象就是 Dialect。   在这个抽象的基础上，我们 ORM 框架的 SQL 构造部分就分 成两个部分：  

-  **公共部分**：大家语法都一样。这部分我们主要参考 SQL 标 准  
-  **个性部分**：如果一个方言的做法和 SQL 标准不一样，那么就要求该方言的对应的实现负责解决这种差异  
-  **Dialect 是一个接口**：每当我们发现有一个 SQL 部分不同 方言写法不一样，就加一个方法  
-  实现了 Dialect 的 **standardSQL**  
-  **其它方言继承自 standardSQL**。当然， ”继承”在 Go 语 境下，指的是组合用法。

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675661122085-b5621f09-c1a9-4650-a75f-cdfb9a5e2391.png#averageHue=%23fcfcfc&clientId=u7b20470e-8cf7-4&from=paste&height=255&id=u03050b22&name=image.png&originHeight=383&originWidth=952&originalType=binary&ratio=1&rotation=0&showTitle=false&size=72183&status=done&style=none&taskId=u911cf0cf-84ed-4d8b-926f-ec5e3a8dc5b&title=&width=634.6666666666666)
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675661291100-909b672c-e49c-4a97-8a06-f3e4bb765595.png#averageHue=%23fdfdfd&clientId=u7b20470e-8cf7-4&from=paste&height=393&id=uadd2b606&name=image.png&originHeight=590&originWidth=893&originalType=binary&ratio=1&rotation=0&showTitle=false&size=48272&status=done&style=none&taskId=u4f11f97e-cb60-42c8-967d-33f435c64c0&title=&width=595.3333333333334)

##### 接口定义  

Dialect 多定义了一个 quoter 方法，是因为希望能够一并解决掉引号的问题。 MySQL 引号是 ` ，而 Oracle 是双引 "

```go
type Dialect interface {
	// quoter 返回一个引号，引用列名，表名的引号
	quoter() byte
	// buildUpsert 构造插入冲突部分
	buildUpsert(b *builder, odk *Upsert) error
}

type standardSQL struct {
}

func (s *standardSQL) quoter() byte {
	// TODO implement me
	panic("implement me")
}

func (s *standardSQL) buildUpsert(b *builder,
	odk *Upsert) error {
	panic("implement me")
}
 
```

##### 局限性

-  Dialect 本身及其容易膨胀。每一点不同都会导致 Dialect 添加方法
-  一些方言独有的特性，加入到 Dialect 里面就不是很合适。例如对 JSON 的支持，只有 PostgreSQL 支持
-  Dialect 抽象无法挪到 internal 包里面    

 可选的其它方案： 可以有不同方言的 Builder，例如 MySQLInserter、PostgreSQLInserter 等  

## 具体实现

### mysqlDialect  与  sqlite3Dialect

 mysqlDialect 与 sqlite3Dialect 组合了 standardSQL，那么它就只需要实现和 标准 SQL 不一样的部分了  
 到实际上在这种场景下，同样需要三个东西： 

- sb：这里设计为 strings.Builder 
- i.model：实际上是元数据 
- i.addArgs：添加执行参数  

mysqlDialect 遇到的困境，以及省视 Selector 和 Inserter, 就能发现需要引入一个公共的父类 builder，

-  它可以封装一些轻量级的操作，简化代码  
-  它持有一些公共字段 —— Selector、Inserter 以及将来支持的 Deleter 和 Updater 大概率都会使用的字 段 

```go
type builder struct {
	sb      strings.Builder
	args    []any
	model   *model.Model
	dialect Dialect
	quoter  byte
}

// buildColumn 构造列
func (b *builder) buildColumn(fd string) error {
	meta, ok := b.model.FieldMap[fd]
	if !ok {
		return errs.NewErrUnknownField(fd)
	}
	b.quote(meta.ColName)
	return nil
}

func (b *builder) quote(name string) {
	b.sb.WriteByte(b.quoter)
	b.sb.WriteString(name)
	b.sb.WriteByte(b.quoter)
}

func (b *builder) addArgs(args ...any) {
	if b.args == nil {
		// 很少有查询能够超过八个参数
		// INSERT 除外
		b.args = make([]any, 0, 8)
	}
	b.args = append(b.args, args...)
}

```

 引入 builder 之后要改造：  

- DB 要支持 Dialect 

```go
type DB struct {
	dialect    Dialect
	r          model.Registry
	db         *sql.DB
	valCreator valuer.Creator
}

// Open 创建一个 DB 实例。
// 默认情况下，该 DB 将使用 MySQL 作为方言
// 如果你使用了其它数据库，可以使用 DBWithDialect 指定
func Open(driver string, dsn string, opts ...DBOption) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	return OpenDB(db, opts...)
}

func OpenDB(db *sql.DB, opts ...DBOption) (*DB, error) {
	res := &DB{
		dialect:    MySQL,
		r:          model.NewRegistry(),
		db:         db,
		valCreator: valuer.NewUnsafeValue,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func DBWithDialect(dialect Dialect) DBOption {
	return func(db *DB) {
		db.dialect = dialect
	}
}
```

- 创建 Selector 和 Inserter 的地方 

```go
func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		builder: builder{
			dialect: db.dialect,
			quoter:  db.dialect.quoter(),
		},
		db: db,
	}
}
```

```go
func NewInserter[T any](db *DB) *Inserter[T] {
	return &Inserter[T]{
		db: db,
		builder: builder{
			dialect: db.dialect,
			quoter:  db.dialect.quoter(),
		},
	}
}

```

- Selector 和 Inserter 的 build 方法

#### mysqlDialect buildUpsert 方法的实现

```go
func (m *mysqlDialect) buildUpsert(b *builder,
	odk *Upsert) error {
	b.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
	for idx, a := range odk.assigns {
		if idx > 0 {
			b.sb.WriteByte(',')
		}
		switch assign := a.(type) {
		case Column:
			fd, ok := b.model.FieldMap[assign.name]
			if !ok {
				return errs.NewErrUnknownField(assign.name)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=VALUES(")
			b.quote(fd.ColName)
			b.sb.WriteByte(')')
		case Assignment:
			err := b.buildColumn(assign.column)
			if err != nil {
				return err
			}
			b.sb.WriteString("=?")
			b.addArgs(assign.val)
		default:
			return errs.NewErrUnsupportedAssignableType(a)
		}
	}
	return nil
}
```

#### sqlite3Dialect buildUpsert 方法的实现

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675671890635-c7ea1341-2c54-4106-a2f7-f14d865b939d.png#averageHue=%23e0d8d0&clientId=u3030462a-8a1a-4&from=paste&height=297&id=u528a8f6f&name=image.png&originHeight=446&originWidth=1083&originalType=binary&ratio=1&rotation=0&showTitle=false&size=359563&status=done&style=none&taskId=u5d0c59d7-8ed2-4999-a6a0-ada91a57bed&title=&width=722)
通过了解 SQLite3 的语法特征：

-  可以指定哪些列冲突  
-  和 MySQL 不同的是，SQLite3 使用的是 excluded.col1 的语法  

然后相应的，修改 Upsert 和 UpsertBuilder：

```go
type UpsertBuilder[T any] struct {
	i               *Inserter[T]
	conflictColumns []string
}

type Upsert struct {
	conflictColumns []string
	assigns         []Assignable
}

func (o *UpsertBuilder[T]) ConflictColumns(cols ...string) *UpsertBuilder[T] {
	o.conflictColumns = cols
	return o
}

// Update 也可以看做是一个终结方法，重新回到 Inserter 里面
func (o *UpsertBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.onDuplicate = &Upsert{
		conflictColumns: o.conflictColumns,
		assigns:         assigns,
	}
	return o.i
}
```

实现 sqlite3Dialect buildUpsert

```go
type sqlite3Dialect struct {
	standardSQL
}

func (s *sqlite3Dialect) quoter() byte {
	return '`'
}

func (s *sqlite3Dialect) buildUpsert(b *builder,
	odk *Upsert) error {
	b.sb.WriteString(" ON CONFLICT")
	if len(odk.conflictColumns) > 0 {
		b.sb.WriteByte('(')
		for i, col := range odk.conflictColumns {
			if i > 0 {
				b.sb.WriteByte(',')
			}
			err := b.buildColumn(col)
			if err != nil {
				return err
			}
		}
		b.sb.WriteByte(')')
	}
	b.sb.WriteString(" DO UPDATE SET ")

	for idx, a := range odk.assigns {
		if idx > 0 {
			b.sb.WriteByte(',')
		}
		switch assign := a.(type) {
		case Column:
			fd, ok := b.model.FieldMap[assign.name]
			if !ok {
				return errs.NewErrUnknownField(assign.name)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=excluded.")
			b.quote(fd.ColName)
		case Assignment:
			err := b.buildColumn(assign.column)
			if err != nil {
				return err
			}
			b.sb.WriteString("=?")
			b.addArgs(assign.val)
		default:
			return errs.NewErrUnsupportedAssignableType(a)
		}
	}
	return nil
}
```

### INSERT 执行  

执行功能一半只关注 sql 执行影响的行数，考虑到 DELETE、UPDATE 都要实现该功能，所以这里统一设计一个接口：

```go
type Executor interface {
	Exec(ctx context.Context) (sql.Result, error)
}
```

然后 Inserter 则实现该接口

```go
func (i *Inserter[T]) Exec(ctx context.Context) (sql.Result, err) {
	q, err := i.Build()
	if err != nil {
		return nil, err
	}
	return i.db.db.ExecContext(ctx, q.SQL, q.Args...)
}

```

该实现的缺点是在想要获得 Id 或者受影响行数的 时候，需要两次处理 err， 例如：

```go
    res, err := NewInserter[TestModel](db).Values(
        &TestModel{
            Id:        1,
            FirstName: "Deng",
            Age:       18,
            LastName:  &sql.NullString{String: "Ming", Valid: true},
        }).Exec(context.Background())
    if err != nil {
        t.Fatal(err)
    }
    id, err := res.LastInsertId()
    if err != nil {
        t.Fatal(err)
    }  
```

### Result 

 引入一个 Result 来简化错误处理，  这种做法类似于 sql.Row 的设计 

```go
type Result struct {
	err error
	res sql.Result
}

func (r Result) Err() error {
	return r.err
}

func (r Result) LastInsertId() (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return r.res.LastInsertId()
}

func (r Result) RowsAffected() (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return r.res.RowsAffected()
}
```

### unsafe 与 reflect 读取字段  

而且前面 Selector 处理结果集的时候我们已经使用 到了 reflect 与 unsafe 来操作对象，这一次我们可以考虑借 助 valuer 抽象 。

```go
// Value 是对结构体实例的内部抽象
type Value interface {
	// Field 返回字段对应的值
	Field(name string) (any, error)
	// SetColumns 设置新值
	SetColumns(rows *sql.Rows) error
}
```

增加一个新的方法 Field，同时修改 reflect 和 unsafe 的实现。  

```go
func (r reflectValue) Field(name string) (any, error) {
	res := r.val.FieldByName(name)
	if res == (reflect.Value{}) {
		return nil, errs.NewErrUnknownField(name)
	}
	return res.Interface(), nil
}

```

```go
func (u unsafeValue) Field(name string) (interface{}, error) {
	fd, ok := u.meta.FieldMap[name]
	if !ok {
		return nil, errs.NewErrUnknownField(name)
	}
	ptr := unsafe.Pointer(uintptr(u.addr) + fd.Offset)
	val := reflect.NewAt(fd.Type, ptr).Elem()
	return val.Interface(), nil
}

```

### Insert Build

还是和 Seletor 一样，使用 Build 构造整个 sql 与 参数

```go
func (i *Inserter[T]) Build() (*Query, error) {
	if len(i.values) == 0 {
		return nil, errs.ErrInsertZeroRow
	}
	m, err := i.db.r.Get(i.values[0])
	if err != nil {
		return nil, err
	}
	i.model = m

	i.sb.WriteString("INSERT INTO ")
	i.quote(m.TableName)
	i.sb.WriteString("(")

	fields := m.Fields
	if len(i.columns) != 0 {
		fields = make([]*model.Field, 0, len(i.columns))
		for _, c := range i.columns {
			field, ok := m.FieldMap[c]
			if !ok {
				return nil, errs.NewErrUnknownField(c)
			}
			fields = append(fields, field)
		}
	}

	// (len(i.values) + 1) 中 +1 是考虑到 UPSERT 语句会传递额外的参数
	i.args = make([]any, 0, len(fields)*(len(i.values)+1))
	for idx, fd := range fields {
		if idx > 0 {
			i.sb.WriteByte(',')
		}
		i.quote(fd.ColName)
	}

	i.sb.WriteString(") VALUES")
	for vIdx, val := range i.values {
		if vIdx > 0 {
			i.sb.WriteByte(',')
		}
		refVal := i.db.valCreator(val, i.model)
		i.sb.WriteByte('(')
		for fIdx, field := range fields {
			if fIdx > 0 {
				i.sb.WriteByte(',')
			}
			i.sb.WriteByte('?')
			fdVal, err := refVal.Field(field.GoName)
			if err != nil {
				return nil, err
			}
			i.addArgs(fdVal)
		}
		i.sb.WriteByte(')')
	}

	if i.upsert != nil {
		err = i.dialect.buildUpsert(&i.builder, i.upsert)
		if err != nil {
			return nil, err
		}
	}

	i.sb.WriteString(";")
	return &Query{
		SQL:  i.sb.String(),
		Args: i.args,
	}, nil
}

```

## 单元测试

```go
func TestInserter_Build(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			// 一个都不插入
			name:    "no value",
			q:       NewInserter[TestModel](db).Values(),
			wantErr: errs.ErrInsertZeroRow,
		},
		{
			name: "single values",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?);",
				Args: []any{int64(1), "Deng", int8(18), &sql.NullString{String: "Ming", Valid: true}},
			},
		},
		{
			name: "multiple values",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				},
				&TestModel{
					Id:        2,
					FirstName: "Da",
					Age:       19,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?),(?,?,?,?);",
				Args: []any{int64(1), "Deng", int8(18), &sql.NullString{String: "Ming", Valid: true},
					int64(2), "Da", int8(19), &sql.NullString{String: "Ming", Valid: true}},
			},
		},
		{
			// 指定列
			name: "specify columns",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}).Columns("FirstName", "LastName"),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model`(`first_name`,`last_name`) VALUES(?,?);",
				Args: []any{"Deng", &sql.NullString{String: "Ming", Valid: true}},
			},
		},
		{
			// 指定列
			name: "invalid columns",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}).Columns("FirstName", "Invalid"),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},

		{
			// upsert
			name: "upsert",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}).OnDuplicateKey().Update(Assign("FirstName", "Da")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?) " +
					"ON DUPLICATE KEY UPDATE `first_name`=?;",
				Args: []any{int64(1), "Deng", int8(18), &sql.NullString{String: "Ming", Valid: true}, "Da"},
			},
		},
		{
			// upsert invalid column
			name: "upsert invalid column",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}).OnDuplicateKey().Update(Assign("Invalid", "Da")),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			// 使用原本插入的值
			name: "upsert use insert value",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				},
				&TestModel{
					Id:        2,
					FirstName: "Da",
					Age:       19,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}).OnDuplicateKey().Update(C("FirstName"), C("LastName")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?),(?,?,?,?) " +
					"ON DUPLICATE KEY UPDATE `first_name`=VALUES(`first_name`),`last_name`=VALUES(`last_name`);",
				Args: []any{int64(1), "Deng", int8(18), &sql.NullString{String: "Ming", Valid: true},
					int64(2), "Da", int8(19), &sql.NullString{String: "Ming", Valid: true}},
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

func TestUpsert_SQLite3_Build(t *testing.T) {
	db := memoryDB(t, DBWithDialect(SQLite3))
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			// upsert
			name: "upsert",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}).OnDuplicateKey().ConflictColumns("Id").
				Update(Assign("FirstName", "Da")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?) " +
					"ON CONFLICT(`id`) DO UPDATE SET `first_name`=?;",
				Args: []any{int64(1), "Deng", int8(18), &sql.NullString{String: "Ming", Valid: true}, "Da"},
			},
		},
		{
			// upsert invalid column
			name: "upsert invalid column",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}).OnDuplicateKey().ConflictColumns("Id").
				Update(Assign("Invalid", "Da")),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			// conflict invalid column
			name: "conflict invalid column",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}).OnDuplicateKey().ConflictColumns("Invalid").
				Update(Assign("FirstName", "Da")),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			// 使用原本插入的值
			name: "upsert use insert value",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				},
				&TestModel{
					Id:        2,
					FirstName: "Da",
					Age:       19,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}).OnDuplicateKey().ConflictColumns("Id").
				Update(C("FirstName"), C("LastName")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?),(?,?,?,?) " +
					"ON CONFLICT(`id`) DO UPDATE SET `first_name`=excluded.`first_name`,`last_name`=excluded.`last_name`;",
				Args: []any{int64(1), "Deng", int8(18), &sql.NullString{String: "Ming", Valid: true},
					int64(2), "Da", int8(19), &sql.NullString{String: "Ming", Valid: true}},
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

## 总结

**ORM 框架是如何支持不同的数据库的**？一般是引入了 Dialect（方言） 抽象，通过设计一个公共的接口，为不同的数据库提供不同的方言实现。这些实现核心都是构造 SQL，因为数据库驱动已经屏蔽掉了不同数据库返回结果集的差异了 ；
**ORM 框架在插入的时候如何处理主键**？ORM 如果知道主键是一个自增主键（或者 TIDB）的随机主键， 并且这个主键是零值，那么插入的时候就会忽略主键列 ；
**unsafe 读取字段，如何计算偏移量**？可以直接使用反射。但是在组合的情况下，一个组合结构体字段的偏移量等于组合结构体的起始偏移量 +  该字段的偏移量  ；
**INSERT 语句能不能插入复杂的表达式**？能，而且这种表达式可以非常复杂。有一种比较特殊的情况，就是 INSERT xxx VALUES(a, a+1) 这种形态，要注意后面的列可以用前面的列来组成表达式，反过来则不行； 