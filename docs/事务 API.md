# 事务 API

到目前为止，已经解决了增删改查的问题，是时候步入到事务阶段了，对于事务来说，核心就是要允许用户创建事务，然后 在事务内部执行增删改查。  

## 开源实例

### Beego ORM

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675847574843-0357c1ca-39d2-43c5-90dd-329eda942545.png#averageHue=%232d2c2b&clientId=u03176505-d321-4&from=paste&height=295&id=uc00f0cfe&name=image.png&originHeight=443&originWidth=1433&originalType=binary&ratio=1&rotation=0&showTitle=false&size=92140&status=done&style=none&taskId=ud168c692-27f2-4578-bdb4-689f55da7ac&title=&width=955.3333333333334)
Beego ORM 这里提供了两种类型的事务接口：

-  一种是用户自己控制的， 
-  另一种是框架控制的（将执行语句打包传进框架提供的指定的事务方法， 这里指的是 Do 开头的事务方法）

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675847619662-3c959bb5-97d1-4941-a64b-ae137b7e230d.png#averageHue=%232c2b2b&clientId=u03176505-d321-4&from=paste&height=202&id=u3f582fdd&name=image.png&originHeight=303&originWidth=1123&originalType=binary&ratio=1&rotation=0&showTitle=false&size=20867&status=done&style=none&taskId=uc0dbbe4b-0672-4e50-a0f5-23e380aa7c0&title=&width=748.6666666666666)
对外暴露的操作事务的提交或回滚的接口，可以理解成这是与用户自己控制的 Begin 开头的事务接口配合使用的。

### GORM

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675848188937-16b03fa0-1b33-459f-ba0f-5e6a0cba175b.png#averageHue=%232d2c2c&clientId=u03176505-d321-4&from=paste&height=339&id=ud9346d35&name=image.png&originHeight=508&originWidth=1183&originalType=binary&ratio=1&rotation=0&showTitle=false&size=66581&status=done&style=none&taskId=u16a923a6-18f8-4dca-a0d4-30cf5c32e7f&title=&width=788.6666666666666)

- DB 本身也可以被看做是事务 
- 普通的事务开启、提交和回滚功能 
- 额外实现了一个 SavePoint 的功能 
- 事务闭包 API  Transaction

## API 设计

目标：

1. 开启事务
2. 回滚或者提交事务
3. 闭包 API 
4. 不准备支持 SavePoint 的功能  

###  Tx 定义  

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675848775172-8085106f-cbb5-407b-911a-665cf389c424.png#averageHue=%23faf9f9&clientId=u03176505-d321-4&from=paste&height=465&id=u3fc4144f&name=image.png&originHeight=698&originWidth=1150&originalType=binary&ratio=1&rotation=0&showTitle=false&size=34219&status=done&style=none&taskId=ub8e289c4-123d-4ec1-8697-c6582fb06fc&title=&width=766.6666666666666)

**事务的核心 API**： 

- Begin：开始一个事务 
- Commit：提交一个事务
- Rollback：回滚一个事务  

需要定义一个新的结构体来表达事务的含义 ,  这里本文引入全新的 Tx 来表达事务，和 GORM 的设计是很不一样的。这意味着 DB 在创建好之后，就是一个不可变的对象。  

```go
type Tx struct {
	tx *sql.Tx
	db *DB
}

func (t *Tx) Commit() error {
	return t.tx.Commit()
}

func (t *Tx) Rollback() error {
	return t.tx.Rollback()
}

func (t *Tx) RollbackIfNotCommit() error {
	err := t.tx.Rollback()
	if err != sql.ErrTxDone {
		return err
	}
	return nil
}
```

 这种设计也暗含了一个限制，即一个事务无法开启另 外一个事务，也就是我们的事务都是单独一个个的。  
 如何使用 Tx 呢？ 原本的 Selector 接收的是 DB 作为参数，现在需要利用 Tx 来创建 Selector，怎么办 ？

###  Session 抽象

 需要一个 DB 和 Tx 的公共抽象
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675849339101-eb02e2f9-7265-49d4-b189-2d4ec4417995.png#averageHue=%23fafafa&clientId=u03176505-d321-4&from=paste&height=299&id=ucc3e61f1&name=image.png&originHeight=449&originWidth=1047&originalType=binary&ratio=1&rotation=0&showTitle=false&size=11526&status=done&style=none&taskId=u4381f6c6-fc8b-4e71-a93f-139c4fb8eac&title=&width=698)
Session 在 Web 里面有比较特殊的含义。 在 ORM 的语境下，一般代表一个上下文； 也可以理解为一种分组机制，**在这个分组内所有的查询会共享一些基本的配置**。  

```go
type Session interface {
	getCore() core
	queryContext(ctx context.Context, query string, args...any) (*sql.Rows, error)
	execContext(ctx context.Context, query string, args...any) (sql.Result, error)
}
```

### 事务闭包 API

在 Beego ORM 和 GORM 里面都看到了事 务闭包 API 的设计。 所谓事务闭包 API ，即用户传入一个方法，ORM 框架会创建事务，利用事务执行该方法，然后根 据该方法的执行情况来判断需要提交还是回滚。  
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675850864010-911155ca-9b02-42af-9b40-478eb8f54a9a.png#averageHue=%23fafafa&clientId=u03176505-d321-4&from=paste&height=425&id=ud33a6a96&name=image.png&originHeight=637&originWidth=1452&originalType=binary&ratio=1&rotation=0&showTitle=false&size=36448&status=done&style=none&taskId=ub5619a37-3cff-4604-aa1d-ee288544fd5&title=&width=968)
**Beego ORM 实例  **
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675851050172-148781dd-241e-4be4-b08a-41eafab76559.png#averageHue=%232c2c2b&clientId=u03176505-d321-4&from=paste&height=482&id=u19b76aac&name=image.png&originHeight=723&originWidth=968&originalType=binary&ratio=1&rotation=0&showTitle=false&size=84502&status=done&style=none&taskId=u4e84399e-7dc2-4066-a654-e70d2f70261&title=&width=645.3333333333334)
核心点： 

- 要判断事务内部有没有发生 panic，也就是 panicked 变量的作用 
- 要判断业务代码有没有返回 error 
- 发生了 panic 或者返回了 error，则回滚，否则提交  

**GORM 实例**
GORM 和 Beego ORM 的处理逻辑都是类似 的，要判断有没有 panic，以及业务代码有没有 返回 error，两者决定是否提交  
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675851176040-9d30127a-b61b-4218-9cc2-7c0c8863fd5e.png#averageHue=%232e2b2b&clientId=u03176505-d321-4&from=paste&height=515&id=u6090f742&name=image.png&originHeight=773&originWidth=1116&originalType=binary&ratio=1&rotation=0&showTitle=false&size=71794&status=done&style=none&taskId=u08331554-3e8c-4f19-be9c-3226d367c98&title=&width=744)
那么本文实现的事务 API 也是类似的逻辑

```go
// DoTx 将会开启事务执行 fn。如果 fn 返回错误或者发生 panic，事务将会回滚，
// 否则提交事务
func (db *DB) DoTx(ctx context.Context,
	fn func(ctx context.Context, tx *Tx) error,
	opts *sql.TxOptions) (err error) {
	var tx *Tx
	tx, err = db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	panicked := true
	defer func() {
		if panicked || err != nil {
			e := tx.Rollback()
			if e != nil {
				err = errs.NewErrFailToRollbackTx(err, e, panicked)
			}
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(ctx, tx)
	panicked = false
	return err
}

```

## 具体实现

### DB 开启一个 Tx 

```go
// BeginTx 开启事务
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx, db: db}, nil
} 
```

### 实现 Session 抽象

```go
func (t *Tx) getCore() core {
	return t.db.core
}

func (t *Tx) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (t *Tx) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}
```

```go
func (db *DB) getCore() core {
	return db.core
}

func (db *DB) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.db.QueryContext(ctx, query, args...)
}

func (db *DB) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.db.ExecContext(ctx, query, args...)
}

```

### ORM Session 重构 Selector、Updator、Insertor 与 deletor

添加 core 模块， core 只是一个简单的封装，将一些 CRUD 都 需要使用的东西放到了一起。    

```go
type core struct {
	r          model.Registry
	dialect    Dialect
	valCreator valuer.Creator
}

```

```go
type Selector[T any] struct {
	builder
	table   string
	where   []Predicate
	having  []Predicate
	columns []Selectable
	groupBy []Column
	offset  int
	limit   int

	core
	sess session
}

func NewSelector[T any](sess session) *Selector[T] {
	c := sess.getCore()
	return &Selector[T]{
		core: c,
		sess: sess,
		builder: builder{
			dialect: c.dialect,
			quoter:  c.dialect.quoter(),
		},
	}
}
```

```go
type Inserter[T any] struct {
	builder
	values  []*T
	columns []string
	upsert  *Upsert

	sess session
	core
}

func NewInserter[T any](sess session) *Inserter[T] {
	c := sess.getCore()
	return &Inserter[T]{
		core: c,
		sess: sess,
		builder: builder{
			dialect: c.dialect,
			quoter:  c.dialect.quoter(),
		},
	}
}
```

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
```

```go
type Deleter[T any] struct {
	builder
	sess  session
	table string
	where []Predicate
}

func NewDeleter[T any](sess session) *Deleter[T] {
	c := sess.getCore()
	return &Deleter[T]{
		sess: sess,
		builder: builder{
			dialect: c.dialect,
			quoter:  c.dialect.quoter(),
		},
	}
}

```

### RollbackIfNotCommit  

Go 因为没有类似于 Java、Python 的异常捕获机制，所以经常会写出呆板代码。 前面的 DoTx 能够解决很大一部分问题，但是有些时候还是要自己控制事务。 因此我们会希望有一个方法，如果事务没有提交， 那么该方法就回滚。  

```go
func (t *Tx) RollbackIfNotCommit() error {
	err := t.tx.Rollback()
	if err != sql.ErrTxDone {
		return err
	}
	return nil
}
```

只需要尝试回滚，如果此时事务已经被提交，或者 被回滚掉了，那么就会得到 sql.ErrTxDone 错误， 这时候我们忽略这个错误就可以。  

### 事务扩散方案  

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675852126305-b8f5c191-9f00-4d3c-862f-663619db118c.png#averageHue=%23fafafa&clientId=u03176505-d321-4&from=paste&height=457&id=ucf69791b&name=image.png&originHeight=686&originWidth=1009&originalType=binary&ratio=1&rotation=0&showTitle=false&size=43526&status=done&style=none&taskId=u24f75cf1-b988-4010-988f-6511c5ef087&title=&width=672.6666666666666)
所谓事务扩散方案，也就是在调用链里面，如果上游的方法开启了事务，那么下游的所有方法也会使用这个事务，否则 ：

-  下游可以开一个新事务 
-  也可以无事务运行 
-  还可以报错  

**context 传递事务**
但凡别的语言用 thread-local 的，在 Go 里面都是 用 context.Context。 核心就是创建事务的时候要检查一下 context 里面 存不存在还没有完成的事务，有就直接返回，没有 就创建一个新的。 Tx 也需要在提交或者回滚的时候将 done 设置为 true。  

```go
func (db *DB) BeginTxV2(ctx context.Context,
	opts *sql.TxOptions) (context.Context, *Tx, error) {
	val := ctx.Value(txKey{})
	if val != nil {
		tx := val.(*Tx)
		if !tx.done {
			return ctx, tx, nil
		}
	}
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return ctx, nil, err
	}
	ctx = context.WithValue(ctx, txKey{}, tx)
	return ctx, tx, nil
}
```

## 单元测试

```go
func TestTx_Commit(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = mockDB.Close() }()

	db, err := OpenDB(mockDB)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		mock.ExpectClose()
		_ = db.Close()
	}()

	// 事务正常提交
	mock.ExpectBegin()
	mock.ExpectCommit()

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	assert.Nil(t, err)
	err = tx.Commit()
	assert.Nil(t, err)


}

func TestTx_Rollback(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = mockDB.Close() }()

	db, err := OpenDB(mockDB)
	if err != nil {
		t.Fatal(err)
	}

	// 事务回滚
	mock.ExpectBegin()
	mock.ExpectRollback()
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	assert.Nil(t, err)
	err = tx.Rollback()
	assert.Nil(t, err)
}
```

## 总结

- **什么是事务扩散**？在 Go 里面怎么解决？其实本质就是上下文里面有事务就用事务，没有事务就开新事 务。Go 里面要解决的话只能依赖于 context.Context，基本上在别的语言里面用 thread-local 解决 的，到 Go 里面都是用 context.Context ; 
- **事务扩散中，如果没有开启事务应该怎么办**？看业务，可以选择报错，可以选择开启新事务，也可以无事务运行 ;
- **事务重复提交会怎样**？在 ORM 层面上，有些 ORM 会维护一个标记位，标记一个事务有没有被提交。 即便没有这个标记位，数据库也会返回错误 ;
- **Go 里面实现一个事务闭包要考虑一些什么问题**？如何实现？主要是考虑 panic 的问题，而后要在 panic 的时候，以及业务代码返回 error 的时候，回滚事务;