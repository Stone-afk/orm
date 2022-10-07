//go:build v4

package orm

import (
	"context"
	"orm/internal/errs"
	"strings"
)

// cols 是用于 WHERE 的列，难以解决 And Or 和 Not 等问题
// func (s *Selector[T]) Where(cols []string, args...any) *Selector[T] {
// 	s.whereCols = cols
// 	s.args = append(s.args, args...)
// }

// 最为灵活的设计
// func (s *Selector[T]) Where(where string, args...any) *Selector[T] {
// 	s.where = where
// 	s.args = append(s.args, args...)
// }

// Selector 用于构造 SELECT 语句
type Selector[T any] struct {
	sb    strings.Builder
	table string
	args  []any
	where []*Predicate
	model *Model
	db    *DB
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		db: db,
	}
}

func (s *Selector[T]) Where(ps ...*Predicate) *Selector[T] {
	s.where = ps
	return s
}

// From 指定表名，如果是空字符串，那么将会使用默认表名
func (s *Selector[T]) From(tbl string) *Selector[T] {
	s.table = tbl
	return s
}

func (s *Selector[T]) buildExpression(e Expression) error {
	switch exp := e.(type) {
	case nil:
		return nil
	case *Column:
		fd, ok := s.model.fieldMap[exp.name]
		if !ok {
			return errs.NewErrUnknownField(exp.name)
		}
		s.sb.WriteByte('`')
		s.sb.WriteString(fd.colName)
		s.sb.WriteByte('`')
	case *value:
		s.sb.WriteByte('?')
		s.args = append(s.args, exp.val)
	case *Predicate:
		_, isLp := exp.left.(*Predicate)
		if isLp {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(exp.left); err != nil {
			return err
		}
		if isLp {
			s.sb.WriteByte(')')
		}
		s.sb.WriteByte(' ')
		s.sb.WriteString(exp.op.String())
		s.sb.WriteByte(' ')
		_, isRp := exp.right.(*Predicate)
		if isRp {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(exp.right); err != nil {
			return err
		}
		if isRp {
			s.sb.WriteByte(')')
		}
	default:
		return errs.NewErrUnsupportedExpressionType(exp)
	}
	return nil
}

func (s *Selector[T]) Build() (*Query, error) {
	var (
		t   T
		err error
	)
	s.model, err = s.db.r.get(&t)
	if err != nil {
		return nil, err
	}
	s.sb.WriteString("SELECT * FROM ")
	if s.table == "" {
		s.sb.WriteByte('`')
		s.sb.WriteString(s.model.tableName)
		s.sb.WriteByte('`')
	} else {
		s.sb.WriteString(s.table)
	}
	// 构造 WHERE
	if len(s.where) > 0 {
		// 类似这种可有可无的部分，都要在前面加一个空格
		s.sb.WriteString(" WHERE ")
		pre := s.where[0]
		for i := 1; i < len(s.where); i++ {
			pre = pre.And(s.where[i])
		}
		if err := s.buildExpression(pre); err != nil {
			return nil, err
		}
	}
	s.sb.WriteString(";")
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) Get(ctx *context.Context) (*T, error) {
	panic("`Get()` must be completed")
}

func (s *Selector[T]) GetMulti(ctx *context.Context) (*T, error) {
	panic("`GetMulti()` must be completed")
}
