//go:build v11
package orm

import (
	"context"
	"database/sql"
)

// session 代表一个抽象的概念，即会话
// 暂时做成私有的，后面考虑重构，因为这个东西用户可能有点难以理解
type session interface {
	getCore() core
	queryContext(ctx context.Context, sql string, args ...any) (*sql.Rows, error)
	execContext(ctx context.Context, sql string, args ...any) (sql.Result, error)
}

func (t *Tx) getCore() core {
	return t.db.core
}

func (t *Tx) queryContext(ctx context.Context, sql string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, sql, args)
}

func (t *Tx) execContext(ctx context.Context, sql string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, sql, args)
}

func (t *Tx) Commit() error {
	return t.tx.Commit()
}

func (t *Tx) Rollback() error {
	return t.tx.Rollback()
}

// RollbackIfNotCommit 只需要尝试回滚，如果此时事务已经被提交，或者 被回滚掉了，
// 那么就会得到 sql.ErrTxDone 错误， 这时候我们忽略这个错误就可以
func (t *Tx) RollbackIfNotCommit() error {
	err := t.Rollback()
	if err != sql.ErrTxDone {
		return err
	}
	return nil
}

type Tx struct {
	tx *sql.Tx
	db *DB
	// 事务扩散方案里面，
	// 这个要在 commit 或者 rollback 的时候修改为 true
	// done bool
}
