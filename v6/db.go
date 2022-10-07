//go:build v6

package orm

import "reflect"

type DBOption func(db *DB)

type DB struct {
	r *registry
}

func NewDB(opts ...DBOption) (*DB, error) {
	res := &DB{
		r: &registry{
			models: make(map[reflect.Type]*Model, 16),
		},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

// MustNewDB 创建一个 DB，如果失败则会 panic
// 我个人不太喜欢这种
func MustNewDB(opts ...DBOption) *DB {
	db, err := NewDB(opts...)
	if err != nil {
		panic(err)
	}
	return db
}
