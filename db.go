package orm

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"log"
	"orm/internal/errs"
	"orm/internal/valuer"
	"orm/model"
	"time"
)

type DBOption func(db *DB)

func DBWithMiddlewares(ms ...Middleware) DBOption {
	return func(db *DB) {
		db.ms = ms
	}
}

func DBWithDialect(d Dialect) DBOption {
	return func(db *DB) {
		db.dialect = d
	}
}

func DBWithMysqlDialect() DBOption {
	return func(db *DB) {
		db.dialect = MySQL
	}
}

func DBWithRegistry(r model.Registry) DBOption {
	return func(db *DB) {
		db.r = r
	}
}

func DBWithValCreator(c valuer.Creator) DBOption {
	return func(db *DB) {
		db.valCreator = valuer.BasicTypeCreator{Creator: c}
	}
}

//func DBUseUnsafeValuer() DBOption {
//	return func(db *DB) {
//		db.valCreator = valuer.NewUnsafeValue
//	}
//}

// Wait 会等待数据库连接
// 注意只能用于测试
func (db *DB) Wait() error {
	err := db.db.Ping()
	for err == driver.ErrBadConn {
		log.Printf("等待数据库启动...")
		err = db.db.Ping()
		time.Sleep(time.Second)
	}
	return err
}

type DB struct {
	core
	db *sql.DB
}

// type txKey struct {
//
// }

// BeginTxAndDiff 事务扩散
//func (db *DB) BeginTxAndDiff(ctx context.Context,
//	opts *sql.TxOptions) (context.Context, *Tx, error) {
//	val := ctx.Value(txKey{})
//	if val != nil {
//		tx := val.(*Tx)
//		if !tx.done {
//			return ctx, tx, nil
//		}
//	}
//	tx, err := db.BeginTx(ctx, opts)
//	if err != nil {
//		return ctx, nil, err
//	}
//	ctx = context.WithValue(ctx, txKey{}, tx)
//	return ctx, tx, nil
//}

// DoTx 将会开启事务执行 fn。如果 fn 返回错误或者发生 panic，事务将会回滚，
// 否则提交事务
func (db *DB) DoTx(ctx context.Context,
	fn func(ctx context.Context, tx *Tx) error,
	opts *sql.TxOptions) (err error) {
	var tx *Tx
	tx, err = db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	panicked := true
	defer func() {
		if panicked || err != nil {
			exc := tx.Rollback()
			if exc != nil {
				err = errs.NewErrFailToRollbackTx(err, exc, panicked)
			}
		} else {
			err = tx.Commit()
		}
	}()
	err = fn(ctx, tx)
	panicked = false
	return err
}

func (db *DB) getCore() core {
	return db.core
}

func (db *DB) queryContext(ctx context.Context, sql string, args ...any) (*sql.Rows, error) {
	return db.db.QueryContext(ctx, sql, args...)
}

func (db *DB) execContext(ctx context.Context, sql string, args ...any) (sql.Result, error) {
	return db.db.ExecContext(ctx, sql, args...)
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx, db: db}, nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

func Open(driver string, dsn string, opts ...DBOption) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	return OpenDB(driver, db, opts...)
}

func OpenDB(driver string, db *sql.DB, opts ...DBOption) (*DB, error) {
	dl, err := dialectOf(driver)
	if err != nil {
		return nil, err
	}

	res := &DB{
		core: core{
			dialect: dl,
			r:       model.NewRegistry(),
			// 可以设为默认，因为原本这里也有默认
			valCreator: valuer.BasicTypeCreator{
				Creator: valuer.NewReflectValue,
			},
		},
		db: db,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

// MustNewDB 创建一个 DB，如果失败则会 panic
// 个人不太喜欢这种
func MustNewDB(driver string, dsn string, opts ...DBOption) *DB {
	db, err := Open(driver, dsn, opts...)
	if err != nil {
		panic(err)
	}
	return db
}
