package orm

import (
	"context"
	"database/sql"
	"orm/v13/internal/valuer"
	"orm/v13/model"
)

// core 只是一个简单的封装，将一些 CRUD 都 需要使用的东西放到了一起。
type core struct {
	dbName     string
	r          model.Registry
	valCreator valuer.Creator
	dialect    Dialect
	ms         []Middleware
}

func getMultiHandler[T any](ctx context.Context, c core,
	sess session, qc *QueryContext) *QueryResult {
	q, err := qc.Query()
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}
	rows, err := sess.queryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}

	//if !rows.Next() {
	//	return nil, ErrNoRows
	//}
	res := make([]*T, 0, 16)
	for rows.Next() {
		tp := new(T)
		// 在这里灵活切换反射或者 unsafe
		val := c.valCreator(tp, qc.Meta)
		err = val.SetColumns(rows)
		if err != nil {
			return &QueryResult{Err: err}
		}
		res = append(res, tp)
	}
	return &QueryResult{Result: res, Err: err}
}

func getMulti[T any](ctx context.Context, c core,
	sess session, qc *QueryContext) *QueryResult {
	var handler HandleFunc = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return getMultiHandler[T](ctx, c, sess, qc)
	}
	ms := c.ms
	for i := len(ms) - 1; i >= 0; i-- {
		handler = ms[i](handler)
	}
	return handler(ctx, qc)
}

func getHandler[T any](ctx context.Context, c core,
	sess session, qc *QueryContext) *QueryResult {
	q, err := qc.Query()
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}
	// s.db 是我们定义的 DB
	// s.db.db 则是 sql.DB
	// 使用 QueryContext，从而和 GetMulti 能够复用处理结果集的代码
	rows, err := sess.queryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}

	if !rows.Next() {
		return &QueryResult{
			Err: ErrNoRows,
		}
	}

	// 有 vals 了，接下来将 vals= [123, "Ming", 18, "Deng"] 反射放回去 t 里面
	tp := new(T)

	// 在这里灵活切换反射或者 unsafe
	val := c.valCreator(tp, qc.Meta)
	err = val.SetColumns(rows)
	return &QueryResult{Result: tp, Err: err}
}

func get[T any](ctx context.Context, c core,
	sess session, qc *QueryContext) *QueryResult {
	var handler HandleFunc = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return getHandler[T](ctx, c, sess, qc)
	}
	ms := c.ms
	for i := len(ms) - 1; i >= 0; i-- {
		handler = ms[i](handler)
	}
	return handler(ctx, qc)
}

func exec[T any](ctx context.Context, c core,
	sess session, qc *QueryContext) Result {
	var handler HandleFunc = func(ctx context.Context, qc *QueryContext) *QueryResult {
		q, err := qc.Query()
		if err != nil {
			return &QueryResult{Err: err}
		}
		res, err := sess.execContext(ctx, q.SQL, q.Args...)
		return &QueryResult{Result: res, Err: err}
	}

	ms := c.ms
	for i := len(ms) - 1; i >= 0; i-- {
		handler = ms[i](handler)
	}
	qr := handler(ctx, qc)
	var res sql.Result
	if qr.Result != nil {
		res = qr.Result.(sql.Result)
	}
	return Result{err: qr.Err, res: res}
}
