package orm

import (
	"context"
)

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
