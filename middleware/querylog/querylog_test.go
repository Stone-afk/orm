package querylog

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"orm"
	"testing"
	"time"
)

func TestMiddlewareBuilder_Build(t *testing.T) {
	var query string
	var as []any
	builder := &MiddlewareBuilder{}
	builder.LogFunc(func(sql string, args ...any) {
		fmt.Println(sql)
		query = sql
		as = args
	}).SlowQueryThreshold(100) // 100 ms 就是慢查询

	slowFuncMdl := func(next orm.HandleFunc) orm.HandleFunc {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			time.Sleep(time.Millisecond)
			return next(ctx, qc)
		}
	}

	db, err := orm.Open("sqlite3",
		"file:test.db?cache=shared&mode=memory", orm.DBWithMiddlewares(
			builder.Build(), slowFuncMdl))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = orm.NewSelector[TestModel](db).Where(orm.C("Id").EQ(10)).Get(context.Background())
	assert.Equal(t, "SELECT * FROM `test_model` WHERE `id` = ?;", query)
	assert.Equal(t, []any{10}, as)

	orm.NewInserter[TestModel](db).Values(&TestModel{Id: 18}).Exec(context.Background())
	assert.Equal(t, "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?);", query)
	assert.Equal(t, []any{int64(18), "", int8(0), (*sql.NullString)(nil)}, as)
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
