## SELECT 结果集处理



### 创建 sql.D

很显然，sql.DB 应该和我们 ORM 层面上的 DB 概念绑定再一起。可以将我们的 DB 看作 是 sql.DB 的一个封装。

```go
type DBOption func(*DB)

type DB struct {
	r  *registry
    db         *sql.DB  、// 封装 sql.DB
}

func NewDB(opts ...DBOption) (*DB, error) {
	db := &DB{
		r: &registry{},
	}
	for _, opt := range opts {
		opt(db)
	}
	return db, nil
}
```

如果将 DB 看作启动的引擎，那么就要设计启动 DB 的方法 Open 和 OpenD

#### Open 和 OpenD的设计

```go
type DBOption func(*DB)

type DB struct {
	r  *registry
    db         *sql.DB  、// 封装 sql.DB
}

func NewDB(opts ...DBOption) (*DB, error) {
	db := &DB{
		r: &registry{},
	}
	for _, opt := range opts {
		opt(db)
	}
	return db, nil
}

func Open(driver string, dsn string, opts ...DBOption) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	return OpenDB(db, opts...)
}

func OpenDB(db *sql.DB, opts ...DBOption) (*DB, error) {
	res := &DB{
		r:          NewRegistry(),
		db:         db,
		valCreator: valuer.NewUnsafeValue,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}
```

从直觉上来说，可能只需要一个 Open 方法，它会创建一个我们的 DB 实例；实际上，**因为用户可能自己创建了 sql.DB 实例**，所以**要允许用户直接用 sql.DB 来创建我们的 DB**

> OpenDB 常用于测试，以及集成别的数据库 中间件。

### 处理结果集

#### 准备处理结果集

在单元测试里，我们不希望依赖于真实的数据 库，因为数据难以模拟，error 更加难以模拟， 所以我们采用 sqlmock 来做单元测试

```go
mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = mockDB.Close() }()
	db, err := OpenDB(mockDB)
	if err != nil {
		t.Fatal(err)
	}
```

**sqlmock 使用步骤:**

- **初始化**：返回一个 mockDB，类型是 *sql.DB，还有 mock 用于构造模拟的场景
- **设置 mock**：基本上是 ExpectXXX WillXXX，严格依赖于顺序

#### 接口设计

对于查询数据的接口，我们需要定义 Get 和 GetMulti 来让用户调用，获取查询的结果集。

```go
type Querier[T any] interface {
	Get(ctx context.Context) (*T, error)
	(ctx context.Context) ([]*T, error)
}
```

设计 Value 是对结构体实例的内部抽象，暴露  SetColumns 方法，目的是把要返回 的结构体，包装成一个 Value 对象

Creator 则是创建 reflect 对象的方法

```go
// orm\internal\valuer\value.go

package valuer

import (
	"database/sql"
	"gitee.com/geektime-geekbang/geektime-go/orm/v7/model"
)

type Value interface {
	// SetColumns 设置新值
	SetColumns(rows *sql.Rows) error
}

type Creator func(val interface{}, meta *model.Model) Value

// ResultSetHandler 这是另外一种可行的设计方案
// type ResultSetHandler interface {
// 	// SetColumns 设置新值，column 是列名
// 	SetColumns(val any, rows *sql.Rows) error
// }

```

#### reflect 方案

##### 构造基于反射的 Value （反射类）

 **NewReflectValue 返回一个封装好的**，基于反射实现的 Value，输入 val 必须是一个指向结构体实例的指针，而不能是任何其它类型

```go
// reflectValue 基于反射的 Value
type reflectValue struct {
	val  reflect.Value
	meta *model.Model
}

var _ Creator = NewReflectValue

// NewReflectValue 返回一个封装好的，基于反射实现的 Value
// 输入 val 必须是一个指向结构体实例的指针，而不能是任何其它类型
func NewReflectValue(val interface{}, meta *model.Model) Value {
	return reflectValue{
		val:  reflect.ValueOf(val).Elem(),
		meta: meta,
	}
}
```

##### 构造结构体；实现 SetColumns 方法

```go
func (r reflectValue) SetColumns(rows *sql.Rows) error {
	cs, err := rows.Columns()
	if err != nil {
		return err
	}
	if len(cs) > len(r.meta.FieldMap) {
		return errs.ErrTooManyReturnedColumns
	}

	// colValues 和 colEleValues 实质上最终都指向同一个对象
	colValues := make([]interface{}, len(cs))
	colEleValues := make([]reflect.Value, len(cs))
	for i, c := range cs {
		cm, ok := r.meta.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		val := reflect.New(cm.Type)
		colValues[i] = val.Interface()
		colEleValues[i] = val.Elem()
	}
	if err = rows.Scan(colValues...); err != nil {
		return err
	}
	for i, c := range cs {
		cm := r.meta.ColumnMap[c]
		fd := r.val.FieldByName(cm.GoName)
		fd.Set(colEleValues[i])
	}
	return nil
}

```

![image-20230125212233758](/docs/images/image-20230125212233758.png)

之后，需要做一个小的调整， 因为在 reflectValue 里面维持住 Model， 而 model 是直接在 orm 中维护的，  于是你得到了一个依赖循环引用。

![image-20230125214109997](/docs/images/image-20230125214109997.png)

要解决循环引用的问题，就得抽取 model。

**方案一**

- 整个包放到 internal 里面 
- Model 的字段都是公有的，这样本项目的其它包都 可以用

**internal 包特性**：只有本项目能用，外部项目不能用。 即用户用不了我们的 registry，也就调用不了 Register 的方法。

**方案二**

- 整个包放到顶级目录下 
- Model 的字段都是公有的

缺点就是用户就可以随意访问我们的 TableName、 FieldMap、ColumnMap。不过大多数时候这也还算可以接受，毕竟正经人也不会 去用这些东西。

**这里我们采用方案二**，接下来要修改 Option 命名，

因为我们的包名就叫做 model，所以不能再加 Model 前缀了，Go 里面，是不建议使用包名来作为结构体名，或者 方法名的前缀的。例如，如果你有一个 user 包，那么里面的服务应该 叫做 Service，而不能是 UserService。 对应的测试我们也挪动了位置。

```go
type Option func(m *Model) error

func WithTableName(tableName string) Option {
	return func(model *Model) error {
		model.TableName = tableName
		return nil
	}
}

func WithColumnName(field string, columnName string) Option {
	return func(model *Model) error {
		fd, ok := model.FieldMap[field]
		if !ok {
			return errs.NewErrUnknownField(field)
		}
		// 注意，这里我们根本没有检测 ColName 会不会是空字符串
		// 因为正常情况下，用户都不会写错
		// 即便写错了，也很容易在测试中发现
		fd.ColName = columnName
		return nil
	}
}
```

##### 设计 DBWithRegistry

到这一步，因为 Model 也被暴露出去了，所以用户 完全可以实现自己的 Registry，也可以使用默认 的。这里暴露一个 DBWithRegistry 这样的选项，允许用 户进一步指定在 DB 中使用的 Registr

```go
func DBWithRegistry(r model.Registry) DBOption {
	return func(db *DB) {
		db.r = r
	}
}
```

改造 DB，增加 valCreator 的定义

```go
type DB struct {
	r          model.Registry
	db         *sql.DB
	valCreator valuer.Creator
}

func Open(driver string, dsn string, opts ...DBOption) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	return OpenDB(db, opts...)
}

func OpenDB(db *sql.DB, opts ...DBOption) (*DB, error) {
	res := &DB{
		r:          model.NewRegistry(),
		db:         db,
		valCreator: valuer.NewUnsafeValue,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}
```

#### unsafe 方案

除了使用反射，**还可以使用 unsafe 来构造结构体**。主要步骤如下：

- 计算字段偏移量 
- 计算对象起始地址 
- 字段真实地址=对象起始地址 + 字段偏移量 
- reflect.NewAt 在特定地址创建对象 
- 调用 Scan

![image-20230125213107371](/docs/images/image-20230125213107371.png)



##### 计算字段地址偏移量

们要用 unsafe 操作字段，所以需要首先 计算出来字段的地址。在Go 里面，运行时刻我们才能知道对象的地址，但是每个字段相对于对象起始地址的偏移 量是固定的。

所以元数据字段结构体必须有字段偏移量才行；增加一个新的字段 Offset，定义成 uintptr 类 型。 在 Go 里面，表示地址一般都是用 uintptr。 这个偏移量，可以通过反射直接拿到。

```go
// Field 字段
type Field struct {
	ColName string
	GoName string
	Type   reflect.Type
	// Offset 相对于对象起始地址的字段偏移量
	Offset uintptr
}
```

在 parseModel 方法中得到偏移量

```go
// parseModel 支持从标签中提取自定义设置
// 标签形式 orm:"key1=value1,key2=value2"
func (r *registry) parseModel(val any) (*Model, error) {
	typ := reflect.TypeOf(val)
	if typ.Kind() != reflect.Ptr ||
		typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}
	typ = typ.Elem()

	// 获得字段的数量
	numField := typ.NumField()
	fds := make(map[string]*Field, numField)
	colMap := make(map[string]*Field, numField)
	for i := 0; i < numField; i++ {
		fdType := typ.Field(i)
		tags, err := r.parseTag(fdType.Tag)
		if err != nil {
			return nil, err
		}
		colName := tags[tagKeyColumn]
		if colName == "" {
			colName = underscoreName(fdType.Name)
		}
		f := &Field{
			ColName: colName,
			Type:    fdType.Type,
			GoName:  fdType.Name,
			Offset:  fdType.Offset,  // 得到偏移量
		}
		fds[fdType.Name] = f
		colMap[colName] = f
	}
	var tableName string
	if tn, ok := val.(TableName); ok {
		tableName = tn.TableName()
	}

	if tableName == "" {
		tableName = underscoreName(typ.Name())
	}

	return &Model{
		TableName: tableName,
		FieldMap:  fds,
		ColumnMap: colMap,
	}, nil
}
```

##### 计算基准地址

```go
type unsafeValue struct {
	addr unsafe.Pointer
	meta *model.Model
}

var _ Creator = NewUnsafeValue

func NewUnsafeValue(val interface{}, meta *model.Model) Value {
	return unsafeValue{
		addr: unsafe.Pointer(reflect.ValueOf(val).Pointer()),
		meta: meta,
	}
}
```

#### 处理结果集

就是在 ptr 这个位置，创建了一个字段对 应的对象。 然后要求 Scan 方法将内容填充到 ptr 那个位置。

```go
func (u unsafeValue) SetColumns(rows *sql.Rows) error {
	cs, err := rows.Columns()
	if err != nil {
		return err
	}
	if len(cs) > len(u.meta.ColumnMap) {
		return errs.ErrTooManyReturnedColumns
	}

	colValues := make([]interface{}, len(cs))
	for i, c := range cs {
		cm, ok := u.meta.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		ptr := unsafe.Pointer(uintptr(u.addr) + cm.Offset)
		val := reflect.NewAt(cm.Type, ptr)
		colValues[i] = val.Interface()
	}
	return rows.Scan(colValues...)
}

```

#### 改造 Selector

目标是让用户可以自由选择是使用反射 实现还是使用unsafe 实现。 

方案一：在 Selector 里面维持一个标记位，比 如 isUnsafe 这种。 

方案二：在 Selector 里面维持一个 Creator

![image-20230125224938405](/docs/images/image-20230125224938405.png)

这里采用方案二，代码如下：

```go
type Selector[T any] struct {
	sb    strings.Builder
	args  []any
	table string
	where []Predicate
	model *model.Model
	db    *DB
}
```

定义一个使用反射或 unsafe 的方法 

```go
func DBUseReflectValuer() DBOption {
   return func(db *DB) {
      db.valCreator = valuer.NewReflectValue
   }
}
```

##### 增加 Get 与 GetMulti 的方法实现

让 Selector 实现 Querier[T any]  接口

```go
func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	q, err := s.Build()
	if err != nil {
		return nil, err
	}
	// s.db 是我们定义的 DB
	// s.db.db 则是 sql.DB
	// 使用 QueryContext，从而和 GetMulti 能够复用处理结果集的代码
	rows, err := s.db.db.QueryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, ErrNoRows
	}

	tp := new(T)
	meta, err := s.db.r.Get(tp)
	if err != nil {
		return nil, err
	}
	val := s.db.valCreator(tp, meta)
	err = val.SetColumns(rows)
	return tp, err
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	var db sql.DB
	q, err := s.Build()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		// 在这里构造 []*T
	}

	panic("implement me")
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		db: db,
	}
}
```

#### 单元测试

reflect 测试

```go
func Test_reflectValue_SetColumn(t *testing.T) {
	testCases := []struct {
		name    string
		cs      map[string][]byte
		val     *test.SimpleStruct
		wantVal *test.SimpleStruct
		wantErr error
	}{
		{
			name: "normal value",
			cs: map[string][]byte{
				"id":               []byte("1"),
				"bool":             []byte("true"),
				"bool_ptr":         []byte("false"),
				"int":              []byte("12"),
				"int_ptr":          []byte("13"),
				"int8":             []byte("8"),
				"int8_ptr":         []byte("-8"),
				"int16":            []byte("16"),
				"int16_ptr":        []byte("-16"),
				"int32":            []byte("32"),
				"int32_ptr":        []byte("-32"),
				"int64":            []byte("64"),
				"int64_ptr":        []byte("-64"),
				"uint":             []byte("14"),
				"uint_ptr":         []byte("15"),
				"uint8":            []byte("8"),
				"uint8_ptr":        []byte("18"),
				"uint16":           []byte("16"),
				"uint16_ptr":       []byte("116"),
				"uint32":           []byte("32"),
				"uint32_ptr":       []byte("132"),
				"uint64":           []byte("64"),
				"uint64_ptr":       []byte("164"),
				"float32":          []byte("3.2"),
				"float32_ptr":      []byte("-3.2"),
				"float64":          []byte("6.4"),
				"float64_ptr":      []byte("-6.4"),
				"byte":             []byte("8"),
				"byte_ptr":         []byte("18"),
				"byte_array":       []byte("hello"),
				"string":           []byte("world"),
				"null_string_ptr":  []byte("null string"),
				"null_int16_ptr":   []byte("16"),
				"null_int32_ptr":   []byte("32"),
				"null_int64_ptr":   []byte("64"),
				"null_bool_ptr":    []byte("true"),
				"null_float64_ptr": []byte("6.4"),
				"json_column":      []byte(`{"name": "Tom"}`),
			},
			val:     &test.SimpleStruct{},
			wantVal: test.NewSimpleStruct(1),
		},
		{
			name: "invalid field",
			cs: map[string][]byte{
				"invalid_column": nil,
			},
			wantErr: errs.NewErrUnknownColumn("invalid_column"),
		},
	}

	r := model.NewRegistry()
	meta, err := r.Get(&test.SimpleStruct{})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = db.Close() }()
			val := NewReflectValue(tc.val, meta)
			cols := make([]string, 0, len(tc.cs))
			colVals := make([]driver.Value, 0, len(tc.cs))
			for k, v := range tc.cs {
				cols = append(cols, k)
				colVals = append(colVals, v)
			}
			mock.ExpectQuery("SELECT *").
				WillReturnRows(sqlmock.NewRows(cols).
					AddRow(colVals...))
			rows, _ := db.Query("SELECT *")
			rows.Next()
			err = val.SetColumns(rows)
			if err != nil {
				assert.Equal(t, tc.wantErr, err)
				return
			}
			if tc.wantErr != nil {
				t.Fatalf("期望得到错误，但是并没有得到 %v", tc.wantErr)
			}
			assert.Equal(t, tc.wantVal, tc.val)
		})
	}

}
```

unsafe 测试

```go
func Test_unsafeValue_SetColumn(t *testing.T) {
	testCases := []struct {
		name    string
		cs      map[string][]byte
		val     *test.SimpleStruct
		wantVal *test.SimpleStruct
		wantErr error
	}{
		{
			name: "normal value",
			cs: map[string][]byte{
				"id":               []byte("1"),
				"bool":             []byte("true"),
				"bool_ptr":         []byte("false"),
				"int":              []byte("12"),
				"int_ptr":          []byte("13"),
				"int8":             []byte("8"),
				"int8_ptr":         []byte("-8"),
				"int16":            []byte("16"),
				"int16_ptr":        []byte("-16"),
				"int32":            []byte("32"),
				"int32_ptr":        []byte("-32"),
				"int64":            []byte("64"),
				"int64_ptr":        []byte("-64"),
				"uint":             []byte("14"),
				"uint_ptr":         []byte("15"),
				"uint8":            []byte("8"),
				"uint8_ptr":        []byte("18"),
				"uint16":           []byte("16"),
				"uint16_ptr":       []byte("116"),
				"uint32":           []byte("32"),
				"uint32_ptr":       []byte("132"),
				"uint64":           []byte("64"),
				"uint64_ptr":       []byte("164"),
				"float32":          []byte("3.2"),
				"float32_ptr":      []byte("-3.2"),
				"float64":          []byte("6.4"),
				"float64_ptr":      []byte("-6.4"),
				"byte":             []byte("8"),
				"byte_ptr":         []byte("18"),
				"byte_array":       []byte("hello"),
				"string":           []byte("world"),
				"null_string_ptr":  []byte("null string"),
				"null_int16_ptr":   []byte("16"),
				"null_int32_ptr":   []byte("32"),
				"null_int64_ptr":   []byte("64"),
				"null_bool_ptr":    []byte("true"),
				"null_float64_ptr": []byte("6.4"),
				"json_column":      []byte(`{"name": "Tom"}`),
			},
			val:     &test.SimpleStruct{},
			wantVal: test.NewSimpleStruct(1),
		},
		{
			name: "invalid field",
			cs: map[string][]byte{
				"invalid_column": nil,
			},
			wantErr: errs.NewErrUnknownColumn("invalid_column"),
		},
	}
	r := model.NewRegistry()
	meta, err := r.Get(&test.SimpleStruct{})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = db.Close() }()
			val := NewUnsafeValue(tc.val, meta)
			cols := make([]string, 0, len(tc.cs))
			colVals := make([]driver.Value, 0, len(tc.cs))
			for k, v := range tc.cs {
				cols = append(cols, k)
				colVals = append(colVals, v)
			}
			mock.ExpectQuery("SELECT *").
				WillReturnRows(sqlmock.NewRows(cols).
					AddRow(colVals...))
			rows, _ := db.Query("SELECT *")
			rows.Next()
			err = val.SetColumns(rows)
			if err != nil {
				assert.Equal(t, tc.wantErr, err)
				return
			}
			if tc.wantErr != nil {
				t.Fatalf("期望得到错误，但是并没有得到 %v", tc.wantErr)
			}
			assert.Equal(t, tc.wantVal, tc.val)
		})
	}

}
```

Selector Get 测试

```go
func TestSelector_Get(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = mockDB.Close() }()
	db, err := OpenDB(mockDB)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name     string
		query    string
		mockErr  error
		mockRows *sqlmock.Rows
		wantErr  error
		wantVal  *TestModel
	}{
		{
			// 查询返回错误
			name:    "query error",
			mockErr: errors.New("invalid query"),
			wantErr: errors.New("invalid query"),
			query:   "SELECT .*",
		},
		{
			name:     "no row",
			wantErr:  ErrNoRows,
			query:    "SELECT .*",
			mockRows: sqlmock.NewRows([]string{"id"}),
		},
		{
			name:    "too many column",
			wantErr: errs.ErrTooManyReturnedColumns,
			query:   "SELECT .*",
			mockRows: func() *sqlmock.Rows {
				res := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name", "extra_column"})
				res.AddRow([]byte("1"), []byte("Da"), []byte("18"), []byte("Ming"), []byte("nothing"))
				return res
			}(),
		},
		{
			name:  "get data",
			query: "SELECT .*",
			mockRows: func() *sqlmock.Rows {
				res := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
				res.AddRow([]byte("1"), []byte("Da"), []byte("18"), []byte("Ming"))
				return res
			}(),
			wantVal: &TestModel{
				Id:        1,
				FirstName: "Da",
				Age:       18,
				LastName:  &sql.NullString{String: "Ming", Valid: true},
			},
		},
	}

	for _, tc := range testCases {
		exp := mock.ExpectQuery(tc.query)
		if tc.mockErr != nil {
			exp.WillReturnError(tc.mockErr)
		} else {
			exp.WillReturnRows(tc.mockRows)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := NewSelector[TestModel](db).Get(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantVal, res)
		})
	}
}
```

#### 性能测试

```go
// 在 orm 目录下执行
// go test -bench=BenchmarkQuerier_Get -benchmem -benchtime=10000x
// 我的输出结果
// goos: linux
// goarch: amd64
// pkg: gitee.com/geektime-geekbang/geektime-go/orm
// cpu: Intel(R) Core(TM) i5-10400F CPU @ 2.90GHz
// BenchmarkQuerier_Get/unsafe-12             10000            453677 ns/op            3246 B/op        108 allocs/op
// BenchmarkQuerier_Get/reflect-12            10000           1173199 ns/op            3427 B/op        117 allocs/op
// PASS
// ok      gitee.com/geektime-geekbang/geektime-go/orm     16.324s
func BenchmarkQuerier_Get(b *testing.B) {
	db, err := Open("sqlite3", fmt.Sprintf("file:benchmark_get.db?cache=shared&mode=memory"))
	if err != nil {
		b.Fatal(err)
	}
	_, err = db.db.Exec(TestModel{}.CreateSQL())
	if err != nil {
		b.Fatal(err)
	}

	res, err := db.db.Exec("INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`)"+
		"VALUES (?,?,?,?)", 12, "Deng", 18, "Ming")

	if err != nil {
		b.Fatal(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		b.Fatal(err)
	}
	if affected == 0 {
		b.Fatal()
	}

	b.Run("unsafe", func(b *testing.B) {
		db.valCreator = valuer.NewUnsafeValue
		for i := 0; i < b.N; i++ {
			_, err = NewSelector[TestModel](db).Get(context.Background())
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("reflect", func(b *testing.B) {
		db.valCreator = valuer.NewReflectValue
		for i := 0; i < b.N; i++ {
			_, err = NewSelector[TestModel](db).Get(context.Background())
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

```

**unsafe 和反射对比**

![image-20230125230332527](/docs/images/image-20230125230332527.png)

### 总结

- ORM 框架怎么处理数据库返回的数据？要点在于，将列映射过去字段（借助于元数据），然后将每一列 的数据解析为字段的类型（这个过程在 Go 里面是由 sql 包完成的），利用反射将转化后的数据塞到结构 体里。 unsafe 的过程则类似，不同在于，unsafe 是先在目标地址创建了字段的零值，后面 sql 包把数 据注入到这些地方
- 使用 unsafe 有什么优点？性能更好
- ORM 的性能瓶颈在哪里，以及怎么解决？两个：构造 SQL 的过程，和处理结果集。前者主要通过 buffer pool 能够极大缓解；后者处理结果集则可以使用 unsafe 来加速
- 为什么 unsafe 要比反射更快？因为反射可以看做是对 unsafe 的封装，因此直接使用 unsafe 就相当于 绕开了中间商。绕开整个中间商之后，同时减少了 CPU 消耗和内存消
