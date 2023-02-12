# AOP 方案

## AOP是什么？

**面向切面编程（Aspect Oriented Programming）一种编程思想**
** OOP（面向对象编程）与AOP区别:**

-  OOP针对业务处理过程的实体及其属性和行为进行抽象封装，以获得更加清晰高效的逻辑单元划分管理。
-  AOP则是针对业务处理过程中的切面进行提取，它所面对的是处理过程中的某个步骤或阶段，以获得逻辑过程中各部分之间低耦合性的隔离效果
-  OOP负责抽象和管理，AOP负责解耦和复用
-  OOP面向名词领域，AOP面向动词领域
-  OOP面向纵向，AOP面向横向

## 为啥要使用AOP？

- 降低业务耦合度;
- 提高程序可复用性;
- 提高代码可读性，易维护性;
- 提高开发效率

实际上，基本上任何框架都需要提供类似的接口，因 为大家都需要解决一些共性问题，例如日志、追踪、 性能监控等。  

## 开源实例

### Beego  ORM

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1676196248698-b1399706-4b14-4fc8-a612-26ac3f844633.png#averageHue=%232f2e2d&clientId=u12a86523-e425-4&from=paste&height=402&id=u1eaf8146&name=image.png&originHeight=603&originWidth=1245&originalType=binary&ratio=1.5&rotation=0&showTitle=false&size=104903&status=done&style=none&taskId=uc7a3f048-7f31-42f4-8c6c-5518b709486&title=&width=830)
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1676196418458-33f021f0-baf7-4bb1-9cbd-56172a07ef8e.png#averageHue=%232d2c2b&clientId=u12a86523-e425-4&from=paste&height=376&id=u8359f99e&name=image.png&originHeight=564&originWidth=1380&originalType=binary&ratio=1.5&rotation=0&showTitle=false&size=57941&status=done&style=none&taskId=uae3bfddd-33a7-45d1-a48a-6c76fd3d342&title=&width=920)
可以说是没有。 Beego 在 ORM 层面上，类似的需求都 是通过侵入式的方案解决的，所以看不到一个显式的类 似于 Middleware ；
后来加一个 FilterChain，但看起来效果不是很好， 根源在于 ORM 没有一个统一的出口（即和数据库交互的 统一的出口）。 用户可以通过装饰器模式封装 Beego ORM 的接口来间 接实现类似的需求。  

### GORM

 在 GORM 里面这个东西叫做 Hook，它是一个和时机有关的概念 ：

- Create：对应于插入 ；
- Update: 对应更新 ；
- Delete：对应删除 ；
- Query：对应于查找  ；

所以用户需要根据自己的需求，选择不同的 Hook。 当然，其实 GORM 还提供了一些额外的接口，实现这些接口也能达成类似的效 果，比如说 driver ；

#### GORM Create Hook

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1676198133363-8fd27b54-5a35-472c-94fb-4c1f2aab725c.png#averageHue=%23302d2c&clientId=u12a86523-e425-4&from=paste&height=494&id=uc8b5d29c&name=image.png&originHeight=741&originWidth=1257&originalType=binary&ratio=1.5&rotation=0&showTitle=false&size=96942&status=done&style=none&taskId=u17e925a1-5a24-4867-992c-b7a0904f149&title=&width=838)
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1676198444137-e385c6e9-9a2a-465c-889d-43e3606cc88a.png#averageHue=%23faf0ef&clientId=u12a86523-e425-4&from=paste&height=191&id=ub3ec96e9&name=image.png&originHeight=287&originWidth=1239&originalType=binary&ratio=1.5&rotation=0&showTitle=false&size=97684&status=done&style=none&taskId=u2b466769-52d1-4dc7-b5b0-6ac874ebb5e&title=&width=826)
**Create 有四个，分成两对**： 

- BeforeSave 和 AfterSave 
- BeforeCreate 和 AfterCreate  

#### GORM Update Hook

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1676198546762-47e081de-d1bf-43b8-8c6f-6576078df182.png#averageHue=%232d2c2b&clientId=u12a86523-e425-4&from=paste&height=331&id=u06549947&name=image.png&originHeight=497&originWidth=1117&originalType=binary&ratio=1.5&rotation=0&showTitle=false&size=55094&status=done&style=none&taskId=u358eb9ac-ef5e-436f-bfac-93ab8772ae5&title=&width=744.6666666666666)
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1676198359707-a326fc26-b715-45d6-955a-b1cfaa237090.png#averageHue=%23f9f0ef&clientId=u12a86523-e425-4&from=paste&height=189&id=u80476141&name=image.png&originHeight=283&originWidth=1215&originalType=binary&ratio=1.5&rotation=0&showTitle=false&size=96267&status=done&style=none&taskId=u14872c02-0393-4510-ad6a-d02d052cf2d&title=&width=810)
**Update 也是四个 Hook，分成两对**： 

- BeforeSave 和 AfterSave 
- BeforeUpdate 和 AfterUpdate  

#### GORM Delete Hook    

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1676198802929-ff1bc418-161e-4e60-a83a-295a08155a37.png#averageHue=%232d2c2b&clientId=u12a86523-e425-4&from=paste&height=307&id=udeca44dc&name=image.png&originHeight=461&originWidth=1157&originalType=binary&ratio=1.5&rotation=0&showTitle=false&size=49547&status=done&style=none&taskId=u52e7b013-6255-4e04-a870-e7eae7b151f&title=&width=771.3333333333334)

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1676198690931-c87421c2-d20d-4e57-b747-4001f01d753c.png#averageHue=%23faf1f0&clientId=u12a86523-e425-4&from=paste&height=188&id=u0abfb7b9&name=image.png&originHeight=282&originWidth=1220&originalType=binary&ratio=1.5&rotation=0&showTitle=false&size=71776&status=done&style=none&taskId=ufc6f5dd3-7619-4586-aac2-31737d5afb2&title=&width=813.3333333333334)
**Delete 有两个 Hook，它们构成了一对**：

- BeforeDelete 和 AfterDelete  

#### GORM Query Hook  

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1676198862486-d39c4284-e51e-4c5c-b746-11f34c56f8c3.png#averageHue=%232d2c2b&clientId=u12a86523-e425-4&from=paste&height=172&id=ub83fa6ba&name=image.png&originHeight=232&originWidth=1004&originalType=binary&ratio=1.5&rotation=0&showTitle=false&size=23999&status=done&style=none&taskId=u0f3b2541-cb3b-4021-96c0-d42f8607ced&title=&width=745.3333740234375)
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1676198834892-081223ae-b337-47d4-829d-124f5c3fd8b5.png#averageHue=%23fbfbfb&clientId=u12a86523-e425-4&from=paste&height=204&id=uec758dfa&name=image.png&originHeight=306&originWidth=1207&originalType=binary&ratio=1.5&rotation=0&showTitle=false&size=57453&status=done&style=none&taskId=u72a28232-ec1f-404e-b452-be84b7a591c&title=&width=804.6666666666666)
 Query 只有一个 Hook，就是 AfterFind  

#### GORM 设计总结  

**优点**：

- **分查询类型**：对增删改查有不同的 Hook ;
- **分时机**：在查询执行前，或者在查询执行后。这种顺序是预定义好的 ; 
- **修改上下文**：每一个 Hook 内部都是可以修改执行上下文的。例如可以利用 这个特性实现一个简单的分库分表中间件  ;
- 用户用起来还是比较简单的，例如使用 AfterUpdate 的时 候，可以很清楚确定这个会在 Update  语句的时候被调用。  

**缺点也很明显**：  

- 缺乏扩展性，用户指定不了顺序 
- BeforeSave 和 AfterSave 有点令人困惑 
- 如果 GORM 要扩展支持别的接入点，例如 BeforeFind，需要修改  

## API 设计

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1676202365971-14847182-8c5d-448b-856d-0fb8f1adb486.png#averageHue=%23fdfcfc&clientId=u12a86523-e425-4&from=paste&height=375&id=u9ecbc6a6&name=image.png&originHeight=562&originWidth=1510&originalType=binary&ratio=1.5&rotation=0&showTitle=false&size=156740&status=done&style=none&taskId=u2a6e2139-4b16-4481-98d8-afa9bf23927&title=&width=1006.6666666666666)

- 抽象出来一个 QueryContext，代表查询上下文
- 抽象出来一个 QueryResult，代表查询结果
- 抽象出来 Handler，代表在这个上下文里面做点什么事情 
- 抽象出来 Middleware，连接不同的 Handler  

```go
type QueryContext struct {
    Type string
    builder QueryBuilder
    Model   *model.Model
    q *Query
}

func (qc *QueryContext) Query() (*Query, error) {
    if qc.q != nil {
        return qc.q, nil
    }
    var err error
    qc.q, err = qc.builder.Build()
    return qc.q, err
}

type QueryResult struct {
    Result any
    Err error
}

type Middleware func(next HandleFunc) HandleFunc

type HandleFunc func(ctx context.Context, qc *QueryContext) *QueryResult
```

这种设计的缺陷就是用户实现 Middleware 的时候，可能存在大量的类型 断言之类的东西，或者需要自己判断是什么查询。  

## 具体实现

###  查询日志  

```go
type MiddlewareBuilder struct {
	logFunc func(query string, args []any)
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{
		logFunc: func(query string, args []any) {
			log.Printf("sql: %s, args: %v", query, args)
		},
	}
}

func (m *MiddlewareBuilder) LogFunc(fn func(query string, args []any)) *MiddlewareBuilder {
	m.logFunc = fn
	return m
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	return func(next orm.HandleFunc) orm.HandleFunc {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			q, err := qc.Builder.Build()
			if err != nil {
				return &orm.QueryResult{
					Err: err,
				}
			}
			m.logFunc(q.SQL, q.Args)
			res := next(ctx, qc)
			return res
		}
	}
}
```

现构造 SQL 都失败了，就可以直接返回 了。也可以选择继续执行下去，因为后面的 Middleware 可能还需要继续处理。
大多数的 ORM 框架都喜欢引入一个 DEBUG 的标记 位，这种 DEBUG 标记位的缺点是侵入式的方案，需要我们修改 Get、GetMulti 和 Exec 这几个方法。    
相比之下，这种做法无侵入，用户的可控性更强。 另外，这里并没有处理敏感信息，也就是 Args 里面可能有密码之类的信息，logFunc 的提供者要 处理这种问题。  
另外一种所谓的 dry run，其实也就是在这里记录了 SQL 之后就直接返回，根本不会发起真实调用。  

###  opentelemetry  

```go
const defaultInstrumentationName = "middleware/opentelemetry"

type MiddlewareBuilder struct {
	Tracer trace.Tracer
}

func (b *MiddlewareBuilder) Build() orm.Middleware {
	if b.Tracer == nil {
		b.Tracer = otel.GetTracerProvider().Tracer(defaultInstrumentationName)
	}
	return func(next orm.HandleFunc) orm.HandleFunc {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			tbl := qc.Meta.TableName
			reqCtx, span := b.Tracer.Start(ctx, qc.Type+"-"+tbl, trace.WithAttributes())
			defer span.End()
			span.SetAttributes(attribute.String("component", "orm"))
			q, err := qc.Builder.Build()
			if err != nil {
				span.RecordError(err)
			}
			span.SetAttributes(attribute.String("table", tbl))
			if q != nil {
				span.SetAttributes(attribute.String("sql", q.SQL))
			}
			return next(reqCtx, qc)
		}
	}
}
```

 简单记录了一下表名和 SQL。 但是没有记录参数，如果要记录参数， 同样要处理加密的问题。  

###  prometheus  

```go
type MiddlewareBuilder struct {
	Name        string
	Subsystem   string
	ConstLabels map[string]string
	Help        string
}

func (m *MiddlewareBuilder) Build() orm.Middleware {
	summaryVec := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:        m.Name,
		Subsystem:   m.Subsystem,
		ConstLabels: m.ConstLabels,
		Help:        m.Help,
	}, []string{"type", "table"})
	prometheus.MustRegister(summaryVec)
	return func(next orm.HandleFunc) orm.HandleFunc {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			startTime := time.Now()
			defer func() {
				endTime := time.Now()
				summaryVec.WithLabelValues(qc.Type, qc.Meta.TableName).
					Observe(float64(endTime.Sub(startTime).Milliseconds()))
			}()
			return next(ctx, qc)
		}
	}
}
```

prometheus 也就是简单记录了一下操 作，以及对应的表。 在这几个 Middleware 里面可以看到，其 实没有办法拿到 IP 之类的信息。因为 ORM 层面并 不知道 Go sql 包内部的信息。 对于分库分表的数据库，这种监控过于弱了，因为在 分库分表之下，会希望能够单独监控每一个库。  

## 单元测试

```go
func Test_Middleware(t *testing.T) {
	testCases := []struct {
		name    string
		wantErr error
		mdls    []Middleware
	}{
		{
			name: "one middleware",
			mdls: func() []Middleware {
				var mdl Middleware = func(next HandleFunc) HandleFunc {
					return func(ctx context.Context, queryContext *QueryContext) *QueryResult {
						return &QueryResult{}
					}
				}
				return []Middleware{mdl}
			}(),
		},
		{
			name: "many middleware",
			mdls: func() []Middleware {
				mdl1 := func(next HandleFunc) HandleFunc {
					return func(ctx context.Context, queryContext *QueryContext) *QueryResult {
						return &QueryResult{Result: "mdl1"}
					}
				}
				mdl2 := func(next HandleFunc) HandleFunc {
					return func(ctx context.Context, queryContext *QueryContext) *QueryResult {
						return &QueryResult{Result: "mdl2"}
					}
				}
				return []Middleware{mdl1, mdl2}
			}(),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			orm, err := Open("sqlite3", "file:test.db?cache=shared&mode=memory",
				DBWithMiddlewares(tc.mdls...))
			if err != nil {
				t.Error(err)
			}
			defer func() {
				_ = orm.Close()
			}()
			assert.EqualValues(t, tc.mdls, orm.ms)
		})
	}
}

```

```go
func TestNewMiddlewareBuilder(t *testing.T) {
	var query string
	var args []any
	m := (&MiddlewareBuilder{}).LogFunc(func(q string, as []any) {
		query=q
		args =as
	})

	db, err := orm.Open("sqlite3",
		"file:test.db?cache=shared&mode=memory",
		orm.DBWithMiddlewares(m.Build()))
	require.NoError(t, err)
	_, _ = orm.NewSelector[TestModel](db).Where(orm.C("Id").EQ(10)).Get(context.Background())
	assert.Equal(t, "SELECT * FROM `test_model` WHERE `id` = ?;", query)
	assert.Equal(t, []any{10}, args)

	orm.NewInserter[TestModel](db).Values(&TestModel{Id: 18}).Exec(context.Background())
	assert.Equal(t, "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?);", query)
	assert.Equal(t, []any{int64(18), "", int8(0), (*sql.NullString)(nil)}, args)
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}

```

## 总结

- **GORM 的 Hook 设计原理**：GORM 的 Hook 按照 SQL 类型划分，例如 BeforeCreate 之类的。本质上 只是 GORM 的研发者在内部找准地方（其实就是指执行语句前后）调用用户注册的 Hook ;  
- **怎么监控慢查询**？就是可以利用 AOP 方案，写一个 AOP 的实现，里面计算 SQL 执行时间，当 SQL 执 行时间超过阈值的时候就可以告警或者打印出来。但是所有 SQL 监控都要注意不要把敏感数据打印出来  ;