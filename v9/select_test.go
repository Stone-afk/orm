//go:build v9 
package orm

import (
"context"
"database/sql"
"github.com/DATA-DOG/go-sqlmock"
"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
"orm/v9/internal/errs"
"testing"
)

func TestSelector_OrderBy(t *testing.T) {
db := memoryDB(t)
testCases := []struct {
name      string
q         QueryBuilder
wantQuery *Query
wantErr   error
}{
{
name: "column",
q:    NewSelector[TestModel](db).OrderBy(Asc("Age")),
wantQuery: &Query{
SQL: "SELECT * FROM `test_model` ORDER BY `age` ASC;",
},
},
{
name: "columns",
q:    NewSelector[TestModel](db).OrderBy(Asc("Age"), Desc("Id")),
wantQuery: &Query{
SQL: "SELECT * FROM `test_model` ORDER BY `age` ASC,`id` DESC;",
},
},
{
name:    "invalid column",
q:       NewSelector[TestModel](db).OrderBy(Asc("Invalid")),
wantErr: errs.NewErrUnknownField("Invalid"),
},
}

for _, tc := range testCases {
t.Run(tc.name, func (t *testing.T) {
query, err := tc.q.Build()
assert.Equal(t, tc.wantErr, err)
if err != nil {
return
}
assert.Equal(t, tc.wantQuery, query)
})
}
}

func TestSelector_OffsetLimit(t *testing.T) {
db := memoryDB(t)
testCases := []struct {
name      string
q         QueryBuilder
wantQuery *Query
wantErr   error
}{
{
name: "offset only",
q:    NewSelector[TestModel](db).Offset(10),
wantQuery: &Query{
SQL:  "SELECT * FROM `test_model` OFFSET ?;",
Args: []any{10},
},
},
{
name: "limit only",
q:    NewSelector[TestModel](db).Limit(10),
wantQuery: &Query{
SQL:  "SELECT * FROM `test_model` LIMIT ?;",
Args: []any{10},
},
},
{
name: "limit offset",
q:    NewSelector[TestModel](db).Limit(20).Offset(10),
wantQuery: &Query{
SQL:  "SELECT * FROM `test_model` LIMIT ? OFFSET ?;",
Args: []any{20, 10},
},
},
}

for _, tc := range testCases {
t.Run(tc.name, func (t *testing.T) {
query, err := tc.q.Build()
assert.Equal(t, tc.wantErr, err)
if err != nil {
return
}
assert.Equal(t, tc.wantQuery, query)
})
}
}

func TestSelector_Having(t *testing.T) {
db := memoryDB(t)
testCases := []struct {
name      string
q         QueryBuilder
wantQuery *Query
wantErr   error
}{
{
// 调用了，但是啥也没传
name: "none",
q:    NewSelector[TestModel](db).GroupBy(C("Age")).Having(),
wantQuery: &Query{
SQL: "SELECT * FROM `test_model` GROUP BY `age`;",
},
},
{
// 单个条件
name: "single",
q: NewSelector[TestModel](db).GroupBy(C("Age")).
Having(C("FirstName").EQ("Deng")),
wantQuery: &Query{
SQL:  "SELECT * FROM `test_model` GROUP BY `age` HAVING `first_name` = ?;",
Args: []any{"Deng"},
},
},
{
// 多个条件
name: "multiple",
q: NewSelector[TestModel](db).GroupBy(C("Age")).
Having(C("FirstName").EQ("Deng"), C("LastName").EQ("Ming")),
wantQuery: &Query{
SQL:  "SELECT * FROM `test_model` GROUP BY `age` HAVING (`first_name` = ?) AND (`last_name` = ?);",
Args: []any{"Deng", "Ming"},
},
},
{
// 聚合函数
name: "avg",
q: NewSelector[TestModel](db).GroupBy(C("Age"), C("FirstName")).
Having(Avg("Age").EQ(18)),
wantQuery: &Query{
SQL:  "SELECT * FROM `test_model` GROUP BY `age`,`first_name` HAVING AVG(`age`) = ?;",
Args: []any{18},
},
},
// having 使用别名
{
name: "use as name",
q: NewSelector[TestModel](db).
Select(Sum("Age").As("sum_age"), C("FirstName")).
GroupBy(C("FirstName")).
Having(C("sum_age").GT(18)),
wantQuery: &Query{
SQL:  "SELECT SUM(`age`) AS `sum_age`,`first_name` FROM `test_model` GROUP BY `first_name` HAVING `sum_age` > ?;",
Args: []any{18},
},
},
}
for _, tc := range testCases {
t.Run(tc.name, func (t *testing.T) {
query, err := tc.q.Build()
assert.Equal(t, tc.wantErr, err)
if err != nil {
return
}
assert.Equal(t, tc.wantQuery, query)
})
}
}

func TestSelector_GroupBy(t *testing.T) {
db := memoryDB(t)
testCases := []struct {
name      string
q         QueryBuilder
wantQuery *Query
wantErr   error
}{
{
// 调用了，但是啥也没传
name: "none",
q:    NewSelector[TestModel](db).GroupBy(),
wantQuery: &Query{
SQL: "SELECT * FROM `test_model`;",
},
},
{
// 单个
name: "single",
q:    NewSelector[TestModel](db).GroupBy(C("Age")),
wantQuery: &Query{
SQL: "SELECT * FROM `test_model` GROUP BY `age`;",
},
},
{
// 多个
name: "multiple",
q:    NewSelector[TestModel](db).GroupBy(C("Age"), C("FirstName")),
wantQuery: &Query{
SQL: "SELECT * FROM `test_model` GROUP BY `age`,`first_name`;",
},
},
{
// 不存在
name:    "invalid column",
q:       NewSelector[TestModel](db).GroupBy(C("Invalid")),
wantErr: errs.NewErrUnknownField("Invalid"),
},
}
for _, tc := range testCases {
t.Run(tc.name, func (t *testing.T) {
query, err := tc.q.Build()
assert.Equal(t, tc.wantErr, err)
if err != nil {
return
}
assert.Equal(t, tc.wantQuery, query)
})
}
}

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
t.Run(tc.name, func (t *testing.T) {
query, err := tc.q.Build()
assert.Equal(t, tc.wantErr, err)
if err != nil {
return
}
assert.Equal(t, tc.wantQuery, query)
})
}
}

func TestSelector_Build(t *testing.T) {
db := memoryDB(t)
testCases := []struct {
name      string
q         QueryBuilder
wantQuery *Query
wantErr   error
}{
{
// From 都不调用
name: "no from",
q:    NewSelector[TestModel](db),
wantQuery: &Query{
SQL: "SELECT * FROM `test_model`;",
},
},
{
// 调用 FROM
name: "with from",
q:    NewSelector[TestModel](db).From("`test_model_t`"),
wantQuery: &Query{
SQL: "SELECT * FROM `test_model_t`;",
},
},
{
// 调用 FROM，但是传入空字符串
name: "empty from",
q:    NewSelector[TestModel](db).From(""),
wantQuery: &Query{
SQL: "SELECT * FROM `test_model`;",
},
},
{
// 调用 FROM，同时出入看了 DB
name: "with db",
q:    NewSelector[TestModel](db).From("`test_db`.`test_model`"),
wantQuery: &Query{
SQL: "SELECT * FROM `test_db`.`test_model`;",
},
},
{
// 单一简单条件
name: "single and simple predicate",
q: NewSelector[TestModel](db).From("`test_model_t`").
Where(C("Id").EQ(1)),
wantQuery: &Query{
SQL:  "SELECT * FROM `test_model_t` WHERE `id` = ?;",
Args: []any{1},
},
},
{
// 多个 predicate
name: "multiple predicates",
q: NewSelector[TestModel](db).
Where(C("Age").GT(18), C("Age").LT(35)),
wantQuery: &Query{
// TestModel -> test_model
SQL:  "SELECT * FROM `test_model` WHERE (`age` > ?) AND (`age` < ?);",
Args: []any{18, 35},
},
},
{
// 使用 AND
name: "and",
q: NewSelector[TestModel](db).
Where(C("Age").GT(18).And(C("Age").LT(35))),
wantQuery: &Query{
SQL:  "SELECT * FROM `test_model` WHERE (`age` > ?) AND (`age` < ?);",
Args: []any{18, 35},
},
},
{
// 使用 OR
name: "or",
q: NewSelector[TestModel](db).
Where(C("Age").GT(18).Or(C("Age").LT(35))),
wantQuery: &Query{
SQL:  "SELECT * FROM `test_model` WHERE (`age` > ?) OR (`age` < ?);",
Args: []any{18, 35},
},
},
{
// 使用 NOT
name: "not",
q:    NewSelector[TestModel](db).Where(Not(C("Age").GT(18))),
wantQuery: &Query{
// NOT 前面有两个空格，因为我们没有对 NOT 进行特殊处理
SQL:  "SELECT * FROM `test_model` WHERE  NOT (`age` > ?);",
Args: []any{18},
},
},
}

for _, tc := range testCases {
t.Run(tc.name, func (t *testing.T) {
query, err := tc.q.Build()
assert.Equal(t, tc.wantErr, err)
if err != nil {
return
}
assert.Equal(t, tc.wantQuery, query)
})
}
}

//测试多条语句
func TestSelector_GetMulti(t *testing.T) {

mockDB, mock, err := sqlmock.New()
require.NoError(t, err)
testCases := []struct {
name     string
query    string
mockErr  error
mockRows *sqlmock.Rows
wantErr  error
wantVal  []*TestModel
}{
{
name:    "multi row",
query:   "SELECT .*",
mockErr: nil,
mockRows: func () *sqlmock.Rows {
rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
rows.AddRow([]byte("123"), []byte("Ming"), []byte("18"), []byte("Deng"))
rows.AddRow([]byte("456"), []byte("Min"), []byte("19"), []byte("Da"))
return rows
}(),
wantVal: []*TestModel{
{
Id:        123,
FirstName: "Ming",
Age:       18,
LastName:  &sql.NullString{Valid: true, String: "Deng"},
},
{
Id:        456,
FirstName: "Min",
Age:       19,
LastName:  &sql.NullString{Valid: true, String: "Da"},
},
},
},

{
name:    "invalid columns",
query:   "SELECT .*",
mockErr: nil,
mockRows: func () *sqlmock.Rows {
rows := sqlmock.NewRows([]string{"id", "first_name", "gender"})
rows.AddRow([]byte("123"), []byte("Ming"), []byte("male"))
return rows
}(),
wantErr: errs.NewErrUnknownColumn("gender"),
},

{
name:    "more columns",
query:   "SELECT .*",
mockErr: nil,
mockRows: func () *sqlmock.Rows {
rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name", "first_name"})
rows.AddRow([]byte("123"), []byte("Ming"), []byte("18"), []byte("Deng"), []byte("明明"))
return rows
}(),
wantErr: errs.ErrTooManyReturnedColumns,
},
}

for _, tc := range testCases {
if tc.mockErr != nil {
mock.ExpectQuery(tc.query).WillReturnError(tc.mockErr)
} else {
mock.ExpectQuery(tc.query).WillReturnRows(tc.mockRows)
}
}

db, err := OpenDB(mockDB)
require.NoError(t, err)
for _, tt := range testCases {
t.Run(tt.name, func (t *testing.T) {
res, err := NewSelector[TestModel](db).GetMulti(context.Background())
assert.Equal(t, tt.wantErr, err)
if err != nil {
return
}

assert.Equal(t, tt.wantVal, res)
})
}
}

func TestSelector_Get(t *testing.T) {

mockDB, mock, err := sqlmock.New()
require.NoError(t, err)

testCases := []struct {
name     string
query    string
mockErr  error
mockRows *sqlmock.Rows
wantErr  error
wantVal  *TestModel
}{
{
name:    "single row",
query:   "SELECT .*",
mockErr: nil,
mockRows: func () *sqlmock.Rows {
rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
rows.AddRow([]byte("123"), []byte("Ming"), []byte("18"), []byte("Deng"))
return rows
}(),
wantVal: &TestModel{
Id:        123,
FirstName: "Ming",
Age:       18,
LastName:  &sql.NullString{Valid: true, String: "Deng"},
},
},

{
// SELECT 出来的行数小于你结构体的行数
name:    "less columns",
query:   "SELECT .*",
mockErr: nil,
mockRows: func () *sqlmock.Rows {
rows := sqlmock.NewRows([]string{"id", "first_name"})
rows.AddRow([]byte("123"), []byte("Ming"))
return rows
}(),
wantVal: &TestModel{
Id:        123,
FirstName: "Ming",
},
},

{
name:    "invalid columns",
query:   "SELECT .*",
mockErr: nil,
mockRows: func () *sqlmock.Rows {
rows := sqlmock.NewRows([]string{"id", "first_name", "gender"})
rows.AddRow([]byte("123"), []byte("Ming"), []byte("male"))
return rows
}(),
wantErr: errs.NewErrUnknownColumn("gender"),
},

{
name:    "more columns",
query:   "SELECT .*",
mockErr: nil,
mockRows: func () *sqlmock.Rows {
rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name", "first_name"})
rows.AddRow([]byte("123"), []byte("Ming"), []byte("18"), []byte("Deng"), []byte("明明"))
return rows
}(),
wantErr: errs.ErrTooManyReturnedColumns,
},
}

for _, tc := range testCases {
if tc.mockErr != nil {
mock.ExpectQuery(tc.query).WillReturnError(tc.mockErr)
} else {
mock.ExpectQuery(tc.query).WillReturnRows(tc.mockRows)
}
}

db, err := OpenDB(mockDB)
require.NoError(t, err)
for _, tt := range testCases {
t.Run(tt.name, func (t *testing.T) {
res, err := NewSelector[TestModel](db).Get(context.Background())
assert.Equal(t, tt.wantErr, err)
if err != nil {
return
}
assert.Equal(t, tt.wantVal, res)
})
}
}

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

//func BenchmarkQuerier_Get(b *testing.B) {
//	db, err := Open("sqlite3", fmt.Sprintf("file:benchmark_get.db?cache=shared&mode=memory"))
//	if err != nil {
//		b.Fatal(err)
//	}
//	_, err = db.db.Exec(TestModel{}.CreateSQL())
//	if err != nil {
//		b.Fatal(err)
//	}
//
//	res, err := db.db.Exec("INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`)"+
//		"VALUES (?,?,?,?)", 12, "Deng", 18, "Ming")
//
//	if err != nil {
//		b.Fatal(err)
//	}
//	affected, err := res.RowsAffected()
//	if err != nil {
//		b.Fatal(err)
//	}
//	if affected == 0 {
//		b.Fatal()
//	}
//
//	b.Run("unsafe", func(b *testing.B) {
//		db.valCreator = valuer.NewUnsafeValue
//		for i := 0; i < b.N; i++ {
//			_, err = NewSelector[TestModel](db).Get(context.Background())
//			if err != nil {
//				b.Fatal(err)
//			}
//		}
//	})
//
//	b.Run("reflect", func(b *testing.B) {
//		db.valCreator = valuer.NewReflectValue
//		for i := 0; i < b.N; i++ {
//			_, err = NewSelector[TestModel](db).Get(context.Background())
//			if err != nil {
//				b.Fatal(err)
//			}
//		}
//	})
//}
