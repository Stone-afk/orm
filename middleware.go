package orm

import (
	"context"
	"orm/model"
)

type QueryContext struct {
	// 用在 UPDATE，DELETE，SELECT，以及 INSERT 语句上的
	// Type 声明查询类型。即 SELECT, UPDATE, DELETE 和 INSERT
	Type string
	// builder 使用的时候，大多数情况下你需要转换到具体的类型
	// 才能篡改查询
	Builder QueryBuilder
	Meta    *model.Model
	q       *Query
}

func (qc *QueryContext) Query() (*Query, error) {
	if qc.q != nil {
		return qc.q, nil
	}
	return qc.Builder.Build()
}

type QueryResult struct {
	// SELECT 语句，你的返回值是 T 或者 []T
	// UPDATE, DELETE, INSERT 返回值是 Result
	Result any
	Err    error
}

type Middleware func(next HandleFunc) HandleFunc

type HandleFunc func(ctx context.Context, qc *QueryContext) *QueryResult
