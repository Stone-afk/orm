package orm

import (
	"context"
	"orm/v8/internal/errs"
	"orm/v8/model"
	"strings"
)

// Selector 用于构造 SELECT 语句
type Selector[T any] struct {
	sb     strings.Builder
	table  string
	args   []any
	where  *predicates
	having *predicates
	model  *model.Model
	db     *DB
	// select 查询的列
	columns []Selectable
	// as 别名映射
	aliasMap map[string]int
	groupBy  []*Column
	orderBy  []*OrderBy
	offset   int
	limit    int
}

// 定义个新的标记接口，限定传入的类型，这样我 们就可以做各种校验
// 符合的结构体有: Column、Aggregate、RawExpr

type Selectable interface {
	selectable()
}

func (s *Selector[T]) Select(cols ...Selectable) *Selector[T] {
	s.columns = cols
	return s
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		db:       db,
		aliasMap: make(map[string]int, 8),
	}
}

// GroupBy 设置 group by 子句
func (s *Selector[T]) GroupBy(cols ...*Column) *Selector[T] {
	s.groupBy = cols
	return s
}

func (s *Selector[T]) Having(ps ...*Predicate) *Selector[T] {
	s.having = &predicates{
		ps:           ps,
		useColsAlias: true,
	}
	return s
}

func (s *Selector[T]) Offset(offset int) *Selector[T] {
	s.offset = offset
	return s
}

func (s *Selector[T]) Limit(limit int) *Selector[T] {
	s.limit = limit
	return s
}

func (s *Selector[T]) OrderBy(orderBys ...*OrderBy) *Selector[T] {
	s.orderBy = orderBys
	return s
}

func (s *Selector[T]) Where(ps ...*Predicate) *Selector[T] {
	s.where = &predicates{
		ps: ps,
	}
	return s
}

// From 指定表名，如果是空字符串，那么将会使用默认表名
func (s *Selector[T]) From(tbl string) *Selector[T] {
	s.table = tbl
	return s
}

func (s *Selector[T]) addArgs(args ...any) {
	if s.args == nil {
		s.args = make([]any, 0, 8)
	}
	s.args = append(s.args, args...)
}

func (s *Selector[T]) buildAs(alias string) error {
	if alias != "" {
		_, ok := s.aliasMap[alias]
		if ok {
			return errs.NewErrDuplicateAlias(alias)
		}
		s.sb.WriteString(" AS ")
		s.sb.WriteByte('`')
		s.sb.WriteString(alias)
		s.sb.WriteByte('`')
		s.aliasMap[alias] = 1
	}
	return nil
}

func (s *Selector[T]) buildExpression(e Expression, useColsAlias, useAggreAlias bool) error {
	switch exp := e.(type) {
	case nil:
		return nil
	case *Column:
		return s.buildColumn(exp, useColsAlias)
	case *Aggregate:
		return s.buildAggregate(exp, useAggreAlias)
	case *value:
		s.sb.WriteByte('?')
		s.args = append(s.args, exp.val)
	case *RawExpr:
		s.sb.WriteString(exp.raw)
		if len(exp.args) > 0 {
			s.addArgs(exp.args...)
		}
	case *Predicate:
		_, isLp := exp.left.(*Predicate)
		if isLp {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(exp.left, useColsAlias, useAggreAlias); err != nil {
			return err
		}
		if isLp {
			s.sb.WriteByte(')')
		}

		// 可能只有左边
		if exp.op == "" {
			return nil
		}

		s.sb.WriteByte(' ')
		s.sb.WriteString(exp.op.String())
		s.sb.WriteByte(' ')
		_, isRp := exp.right.(*Predicate)
		if isRp {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(exp.right, useColsAlias, useAggreAlias); err != nil {
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

func (s *Selector[T]) buildColumn(val *Column, useAlias bool) error {
	fd, ok := s.model.FieldMap[val.name]
	if useAlias {
		if !ok {
			_, ok = s.aliasMap[val.name]
			if !ok {
				return errs.NewErrUnknownField(val.name)
			}
			s.sb.WriteByte('`')
			s.sb.WriteString(val.name)
			s.sb.WriteByte('`')
		} else {
			s.sb.WriteByte('`')
			s.sb.WriteString(fd.ColName)
			s.sb.WriteByte('`')
			err := s.buildAs(val.alias)
			if err != nil {
				return err
			}
		}
	} else {
		if !ok {
			return errs.NewErrUnknownField(val.name)
		}
		s.sb.WriteByte('`')
		s.sb.WriteString(fd.ColName)
		s.sb.WriteByte('`')
	}
	return nil
}

func (s *Selector[T]) buildAggregate(val *Aggregate, useAlias bool) error {
	fd, ok := s.model.FieldMap[val.arg]
	if !ok {
		return errs.NewErrUnknownField(val.arg)
	}
	s.sb.WriteString(val.fn)
	s.sb.WriteString("(`")
	s.sb.WriteString(fd.ColName)
	s.sb.WriteString("`)")
	if useAlias {
		err := s.buildAs(val.alias)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Selector[T]) buildColumns() error {
	if len(s.columns) == 0 {
		s.sb.WriteByte('*')
		return nil
	}
	for i, col := range s.columns {
		if i > 0 {
			s.sb.WriteByte(',')
		}
		switch val := col.(type) {
		case *Column:
			if err := s.buildColumn(val, true); err != nil {
				return err
			}
		case *Aggregate:
			if err := s.buildAggregate(val, true); err != nil {
				return err
			}
		case *RawExpr:
			s.sb.WriteString(val.raw)
			if len(val.args) > 0 {
				s.addArgs(val.args...)
			}
		default:
			return errs.NewErrUnsupportedSelectable(col)
		}
	}
	return nil
}

func (s *Selector[T]) buildPredicates(pres *predicates) error {
	ps := pres.ps[0]
	for i := 1; i < len(pres.ps); i++ {
		ps = ps.And(pres.ps[i])
	}
	return s.buildExpression(ps, pres.useColsAlias, pres.useAggreAlias)
}

func (s *Selector[T]) buildGroupBy() error {
	l := len(s.groupBy)
	for i, col := range s.groupBy {
		if i > 0 && i < l {
			s.sb.WriteByte(',')
		}
		if err := s.buildColumn(col, false); err != nil {
			return err
		}
	}
	return nil
}

func (s *Selector[T]) buildOrderBy() error {
	l := len(s.orderBy)
	for i, od := range s.orderBy {
		if i > 0 && i < l {
			s.sb.WriteByte(',')
		}
		fd, ok := s.model.FieldMap[od.col]
		if !ok {
			return errs.NewErrUnknownField(od.col)
		}
		s.sb.WriteByte('`')
		s.sb.WriteString(fd.ColName)
		s.sb.WriteByte('`')
		s.sb.WriteString(" " + od.order)
	}
	return nil
}

func (s *Selector[T]) Build() (*Query, error) {
	var (
		t   T
		err error
	)
	s.model, err = s.db.r.Get(&t)
	if err != nil {
		return nil, err
	}
	s.sb.WriteString("SELECT ")
	if err = s.buildColumns(); err != nil {
		return nil, err
	}
	s.sb.WriteString(" FROM ")
	if s.table == "" {
		s.sb.WriteByte('`')
		s.sb.WriteString(s.model.TableName)
		s.sb.WriteByte('`')
	} else {
		s.sb.WriteString(s.table)
	}
	// 构造 WHERE
	if s.where != nil && len(s.where.ps) > 0 {
		// 类似这种可有可无的部分，都要在前面加一个空格
		s.sb.WriteString(" WHERE ")
		// WHERE 是不允许用别名的
		if err = s.buildPredicates(s.where); err != nil {
			return nil, err
		}
	}
	if len(s.groupBy) > 0 {
		s.sb.WriteString(" GROUP BY ")
		// GROUP BY 理论上可以用别名，但这里不允许，用户完全可以通过简单的修改代码避免使用别名的这种用法。
		// 也不支持复杂的表达式，因为复杂的表达式和 group by 混用是非常罕见的
		if err = s.buildGroupBy(); err != nil {
			return nil, err
		}
	}
	if s.having != nil && len(s.having.ps) > 0 {
		s.sb.WriteString(" HAVING ")
		// HAVING 是可以用别名的
		if err = s.buildPredicates(s.having); err != nil {
			return nil, err
		}
	}
	if len(s.orderBy) > 0 {
		s.sb.WriteString(" ORDER BY ")
		if err = s.buildOrderBy(); err != nil {
			return nil, err
		}
	}
	if s.limit > 0 {
		s.sb.WriteString(" LIMIT ?")
		s.addArgs(s.limit)
	}
	if s.offset > 0 {
		s.sb.WriteString(" OFFSET ?")
		s.addArgs(s.offset)
	}

	s.sb.WriteString(";")
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	q, err := s.Build()
	if err != nil {
		return nil, err
	}
	// s.db 是我们定义的 DB
	// s.db.db 则是 sql.DB
	// 使用 QueryContext，从而和 GetMulti 能够复用处理结果集的代码
	rows, err := s.db.db.QueryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return nil, err
	}

	// 有 vals 了，接下来将 vals= [123, "Ming", 18, "Deng"] 反射放回去 t 里面
	t := new(T)
	// 在这里灵活切换反射或者 unsafe
	val := s.db.valCreator(t, s.model)
	err = val.SetColumns(rows)
	return t, err
}

func (s *Selector[T]) GetMulti(ctx *context.Context) (*T, error) {
	// var db *sql.DB
	// q, err := s.Build()
	// if err != nil {
	// 	return nil, err
	// }
	// rows, err := db.QueryContext(ctx, q.SQL, q.Args...)
	// if err != nil {
	// 	return nil, err
	// }
	// 想办法，把 rows 所有行转换为 []*T
	panic("`GetMulti()` must be completed")
}
