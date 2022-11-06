package orm

import "context"

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
	res := get[T](ctx, r.core, r.sess, &QueryContext{
		Builder: r,
		Type:    "RAW",
	})
	if res.Err != nil {
		return nil, res.Err
	}
	return res.Result.(*T), nil
}

func (r *RawQuerier[T]) GetMulti(ctx context.Context) ([]*T, error) {
	res := getMulti[T](ctx, r.core, r.sess, &QueryContext{
		Builder: r,
		Type:    "RAW",
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
