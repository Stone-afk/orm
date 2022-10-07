//go:build v7

package orm

import (
	"database/sql"
	"orm/v7/internal/valuer"
	"orm/v7/model"
)

type DBOption func(db *DB)

type DB struct {
	r          model.Registry
	db         *sql.DB
	valCreator valuer.Creator
}

func DBWithRegistry(r model.Registry) DBOption {
	return func(db *DB) {
		db.r = r
	}
}

func DBUseUnsafeValuer() DBOption {
	return func(db *DB) {
		db.valCreator = valuer.NewUnsafeValue
	}
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
		valCreator: valuer.NewReflectValue,
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
