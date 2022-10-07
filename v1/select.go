//go:build v1

package orm

import (
	"context"
	"reflect"
	"strings"
)

func NewSelector[T any]() *Selector[T] {
	return &Selector[T]{}
}

// Selector 用于构造 SELECT 语句
type Selector[T any] struct {
	table string
}

// From 指定表名，如果是空字符串，那么将会使用默认表名
func (s *Selector[T]) From(tbl string) *Selector[T] {
	s.table = tbl
	return s
}

func (s *Selector[T]) Build() (*Query, error) {
	var sb strings.Builder
	sb.WriteString("SELECT * FROM ")
	if s.table == "" {
		var t T
		tbl := reflect.TypeOf(t).Name()
		sb.WriteByte('`')
		sb.WriteString(tbl)
		sb.WriteByte('`')
	} else {
		sb.WriteString(s.table)
	}
	sb.WriteString(";")
	return &Query{
		SQL: sb.String(),
	}, nil
}

func (s *Selector[T]) Get(ctx *context.Context) (*T, error) {
	panic("`Get()` must be completed")
}

func (s *Selector[T]) GetMulti(ctx *context.Context) (*T, error) {
	panic("`GetMulti()` must be completed")
}
