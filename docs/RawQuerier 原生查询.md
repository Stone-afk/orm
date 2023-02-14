## 语法分析

 就 MySQL SELECT 语句来说，ORM 框架能支持全部 语法吗？   显然不能，也不愿意支持全部 ， 也不仅仅是 ORM 框架，大多数框架设计的时候，都要考 虑提供兜底的措施，或者提供绕开你的框架的机制。 在 ORM 这里，就是要允许用户手写 SQL，直接绕开 ORM 的各种机制。  
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675911362630-fa99cbeb-71ec-495f-a801-411847f563e7.png#averageHue=%23f5f5f4&clientId=uab4ab7df-3bf5-4&from=paste&height=502&id=u318040b3&name=image.png&originHeight=753&originWidth=631&originalType=binary&ratio=1&rotation=0&showTitle=false&size=312075&status=done&style=none&taskId=u5a2162b1-fbb5-4d0f-916b-594eb0622db&title=&width=420.6666666666667)

## 开源实例

### Beego ORM

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675912014053-40ce8751-fdd8-4945-bc0a-be10e4e8665e.png#averageHue=%23302d2b&clientId=uab4ab7df-3bf5-4&from=paste&height=479&id=u83e76bbb&name=image.png&originHeight=718&originWidth=1284&originalType=binary&ratio=1&rotation=0&showTitle=false&size=136986&status=done&style=none&taskId=ud89e1672-10ad-4b12-9561-cbfd8d4f266&title=&width=856)
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675912079080-dfa5c568-441a-4e21-a02b-2aa68aa13f2c.png#averageHue=%232d2c2b&clientId=uab4ab7df-3bf5-4&from=paste&height=489&id=u26a80c26&name=image.png&originHeight=734&originWidth=1112&originalType=binary&ratio=1&rotation=0&showTitle=false&size=98419&status=done&style=none&taskId=u4b2540e8-47e4-4464-8679-77402c81de7&title=&width=741.3333333333334)
Beego ORM 的原生查询接口是，直接让用户传入 sql 语句与对应的参数，然后返回一个原生查询的抽象 RawSeter，该抽象提供了众多的方法支持。

### GORM

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675914599557-72be3db1-5deb-4b2f-b674-cf59fb6d777f.png#averageHue=%232d2c2c&clientId=uab4ab7df-3bf5-4&from=paste&height=235&id=u038134ca&name=image.png&originHeight=369&originWidth=857&originalType=binary&ratio=1&rotation=0&showTitle=false&size=49108&status=done&style=none&taskId=u62679a02-c82b-43ed-9291-e05f7922aa8&title=&width=545.3333740234375)
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675913553016-aa5fa65c-7775-4764-81ef-b47b3ec92ee6.png#averageHue=%232d2c2c&clientId=uab4ab7df-3bf5-4&from=paste&height=295&id=u46f57f2d&name=image.png&originHeight=443&originWidth=821&originalType=binary&ratio=1&rotation=0&showTitle=false&size=58122&status=done&style=none&taskId=ue90df1ff-091f-49ec-8d51-159e4d6e272&title=&width=547.3333333333334)
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675914961944-860bf537-f760-437f-b3da-8724c36d4bac.png#averageHue=%232d2c2c&clientId=uab4ab7df-3bf5-4&from=paste&height=383&id=u29fcbb08&name=image.png&originHeight=611&originWidth=869&originalType=binary&ratio=1&rotation=0&showTitle=false&size=81227&status=done&style=none&taskId=u0b50eba6-3ef8-48a1-8efa-9f3932e4650&title=&width=545.3333740234375)
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675915007032-c2fb0879-e78e-421d-bff4-a928013bd1a6.png#averageHue=%232c2b2b&clientId=uab4ab7df-3bf5-4&from=paste&height=333&id=u6f3d1227&name=image.png&originHeight=742&originWidth=1216&originalType=binary&ratio=1&rotation=0&showTitle=false&size=78912&status=done&style=none&taskId=ucb8ac3fc-6b49-4ab2-aa5c-79260137fd5&title=&width=545.6666870117188)
GORM 也是依赖于用户传 sql 语句与参数，但是返回给用户的是 DB，用户负责调用目标获取结果的接口，例如: Row()、Rows()、Scan() 等方法。

```go
type Result struct {
  ID   int
  Name string
  Age  int
}

var result Result
db.Raw("SELECT id, name, age FROM users WHERE name = ?", 3).Scan(&result)

db.Raw("SELECT id, name, age FROM users WHERE name = ?", 3).Scan(&result)

var age int
db.Raw("SELECT SUM(age) FROM users WHERE role = ?", "admin").Scan(&age)

var users []User
db.Raw("UPDATE users SET name = ? WHERE age = ? RETURNING id, name", "jinzhu", 20).Scan(&users)

// 使用原生 SQL
row := db.Raw("select name, age, email from users where name = ?", "jinzhu").Row()
row.Scan(&name, &age, &email)

db.Exec("DROP TABLE users")
db.Exec("UPDATE orders SET shipped_at = ? WHERE id IN ?", time.Now(), []int64{1, 2, 3})

// Exec with SQL Expression
db.Exec("UPDATE users SET money = ? WHERE name = ?", gorm.Expr("money * ? + ?", 10000, 1), "jinzhu")

```

## API 设计

 本文主要支持中间这种。 第三种用户可以直接使用 sql.DB，都用不着 ORM 框架  
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675915944464-dd569af6-e873-4299-8ed6-17dc06d0fc4b.png#averageHue=%23fafafa&clientId=uab4ab7df-3bf5-4&from=paste&height=400&id=ucf6ce428&name=image.png&originHeight=600&originWidth=997&originalType=binary&ratio=1&rotation=0&showTitle=false&size=126324&status=done&style=none&taskId=u254d5a38-0c32-4b10-9cf7-ba7cab36b96&title=&width=664.6666666666666)

```go
var _ Querier[any] = &RawQuerier[any]{}

// RawQuerier 原生查询器
type RawQuerier[T any] struct {
	core
	sess session
	sql string
	args []any
}

func (r *RawQuerier[T]) Exec(ctx context.Context) Result {
	// TODO implement me
	panic("implement me")
}

func (r *RawQuerier[T]) Get(ctx context.Context) (*T, error) {
	// TODO implement me
	panic("implement me")
}

func (r *RawQuerier[T]) GetMulti(ctx context.Context) ([]*T, error) {
	// TODO implement me
	panic("implement me")
}
```

## 具体实现

 其中exec 和 get 两个方法是从我们原本的 Selector 与 Updateor 等实现里面抽取出来的。  目的是为了实现
Executor、Querier 等接口，另外为了兼容 Seletor 等构造模式的构造方式，也必须实现 QueryBuilder 接口。

```go
var _ Querier[any] = &RawQuerier[any]{}

// RawQuerier 原生查询器
type RawQuerier[T any] struct {
	core
	sess session
	sql  string
	args []any
}

// RawQuery 创建一个 RawQuerier 实例
// 泛型参数 T 是目标类型。
// 例如，如果查询 User 的数据，那么 T 就是 User
func RawQuery[T any](sess session, sql string, args ...any) *RawQuerier[T] {
	return &RawQuerier[T]{
		core: sess.getCore(),
		sess: sess,
		sql:  sql,
		args: args,
	}
}

func (r *RawQuerier[T]) Build() (*Query, error) {
	return &Query{
		SQL:  r.sql,
		Args: r.args,
	}, nil
}

func (r *RawQuerier[T]) Get(ctx context.Context) (*T, error) {
	//  当通过 RawQuery 方法调用 Get ,如果 T 是 time.Time, sql.Scanner 的实现，
	//  内置类型或者基本类型时， 在这里都会报错，但是这种情况我们认为是可以接受的
	//  所以在此将报错忽略，因为基本类型取值用不到 meta 里的数据
	model, _ := r.r.Get(new(T))
	res := get[T](ctx, r.core, r.sess, &QueryContext{
		Builder: r,
		Type:    "RAW",
		Meta:    model,
	})
	if res.Err != nil {
		return nil, res.Err
	}
	return res.Result.(*T), nil
}

func (r *RawQuerier[T]) GetMulti(ctx context.Context) ([]*T, error) {
	//  当通过 RawQuery 方法调用 Get ,如果 T 是 time.Time, sql.Scanner 的实现，
	//  内置类型或者基本类型时， 在这里都会报错，但是这种情况我们认为是可以接受的
	//  所以在此将报错忽略，因为基本类型取值用不到 meta 里的数据
	model, _ := r.r.Get(new(T))
	res := getMulti[T](ctx, r.core, r.sess, &QueryContext{
		Builder: r,
		Type:    "RAW",
		Meta:    model,
	})
	if res.Err != nil {
		return nil, res.Err
	}
	return res.Result.([]*T), nil
}

func (r *RawQuerier[T]) Exec(ctx context.Context) Result {
	return exec[T](ctx, r.core, r.sess, &QueryContext{
		Type:    "RAW",
		Builder: r,
	})
}

```

## 单元测试

```go
func TestRawQuerier_Get(t *testing.T) {
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
		{
			name: "res RawQuery int",
			queryRes: func(t *testing.T) any {
				queryer := RawQuery[int](db, "SELECT `age` FROM `test_model` LIMIT ?;", 1)
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"age"}).AddRow(10)
				mock.ExpectQuery("SELECT `age` FROM `test_model` LIMIT ?;").
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantVal: func() *int {
				val := 10
				return &val
			}(),
		},
		{
			name: "res RawQuery bytes",
			queryRes: func(t *testing.T) any {
				queryer := RawQuery[[]byte](db, "SELECT `first_name` FROM `test_model` WHERE `id`=? LIMIT ?;", 1, 1)
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"first_name"}).AddRow([]byte("Li"))
				mock.ExpectQuery("SELECT `first_name` FROM `test_model` WHERE `id`=? LIMIT ?;").
					WithArgs(1, 1).
					WillReturnRows(rows)
			},
			wantVal: func() *[]byte {
				val := []byte("Li")
				return &val
			}(),
		},
		{
			name: "res RawQuery string",
			queryRes: func(t *testing.T) any {
				queryer := RawQuery[string](db, "SELECT `first_name` FROM `test_model` WHERE `id`=? LIMIT ?;", 1, 1)
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"first_name"}).AddRow("Da")
				mock.ExpectQuery("SELECT `first_name` FROM `test_model` WHERE `id`=? LIMIT ?;").
					WithArgs(1, 1).
					WillReturnRows(rows)
			},
			wantVal: func() *string {
				val := "Da"
				return &val
			}(),
		},
		{
			name: "res RawQuery struct ptr",
			queryRes: func(t *testing.T) any {
				queryer := RawQuery[TestModel](db, "SELECT `first_name`,`age` FROM `test_model` WHERE `id`=? LIMIT ?;", 1, 1)
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"first_name", "age"}).AddRow("Da", 18)
				mock.ExpectQuery("SELECT `first_name`,`age` FROM `test_model` WHERE `id`=? LIMIT ?;").
					WithArgs(1, 1).
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
			name: "res RawQuery sql.NullString",
			queryRes: func(t *testing.T) any {
				queryer := RawQuery[sql.NullString](db, "SELECT `last_name` FROM `test_model` WHERE `id`=? LIMIT ?;", 1, 1)
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"last_name"}).AddRow([]byte("ming"))
				mock.ExpectQuery("SELECT `last_name` FROM `test_model` WHERE `id`=? LIMIT ?;").
					WithArgs(1, 1).
					WillReturnRows(rows)
			},
			wantVal: func() *sql.NullString {
				return &sql.NullString{String: "ming", Valid: true}
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockOrder(mock)
			res := tc.queryRes(t)
			assert.Equal(t, tc.wantVal, res)
		})
	}
}

```
