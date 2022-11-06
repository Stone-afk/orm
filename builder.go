package orm

import (
	"github.com/valyala/bytebufferpool"
	"orm/internal/errs"
	"orm/model"
)

type Builder struct {
	core
	// sb     stringb.Builder // 普通字符串 Builder
	// 使用 bytebufferpool 以减少内存分配
	// 每次调用 Get 之后不要忘记再调用 Put
	buffer *bytebufferpool.ByteBuffer
	// 元数据模型
	model *model.Model
	args  []any
	// as 别名映射
	aliasMap map[string]int

	//dialect Dialect
	quoter byte
}

func (b *Builder) writeSpace() {
	_ = b.buffer.WriteByte(' ')
}

func (b *Builder) writeComma() {
	b.writeByte(',')
}

func (b *Builder) writePlaceholder() {
	b.writeByte('?')
}

func (b *Builder) writeLeftParenthesis() {
	b.writeString("(")
}

func (b *Builder) writeRightParenthesis() {
	b.writeString(")")
}

func (b *Builder) quote(name string) {
	b.writeByte(b.quoter)
	b.writeString(name)
	b.writeByte(b.quoter)

}

func (b *Builder) end() {
	b.writeString(";")
}

func (b *Builder) writeString(val string) {
	_, _ = b.buffer.WriteString(val)
}

func (b *Builder) writeByte(c byte) {
	_ = b.buffer.WriteByte(c)
}

func (b *Builder) buildAssignment(a Assignment) error {
	fd, ok := b.model.FieldMap[a.column]
	if !ok {
		return errs.NewErrUnknownField(a.column)
	}
	b.quote(fd.ColName)
	b.writeString(" = ")
	return b.buildExpression(a.val, false, false)
}

func (b *Builder) buildPredicates(pres *predicates) error {
	ps := pres.ps[0]
	for i := 1; i < len(pres.ps); i++ {
		ps = ps.And(pres.ps[i])
	}
	return b.buildExpression(ps, pres.useColsAlias, pres.useAggreAlias)

}

func (b *Builder) addArgs(args ...any) {
	if b.args == nil {
		b.args = make([]any, 0, 8)
	}
	b.args = append(b.args, args...)
}

func (b *Builder) buildAs(alias string) error {
	if alias != "" {
		_, ok := b.aliasMap[alias]
		if ok {
			return errs.NewErrDuplicateAlias(alias)
		}
		b.writeString(" AS ")
		b.quote(alias)
		b.aliasMap[alias] = 1
	}
	return nil

}

func (b *Builder) buildRaw(r RawExpr) error {
	b.writeString(r.raw)
	if len(r.args) > 0 {
		b.addArgs(r.args...)
	}
	return nil
}

func (b *Builder) buildValue(v value) error {
	b.writePlaceholder()
	b.addArgs(v.val)
	return nil
}

func (b *Builder) colName(table TableReference, fdName string, useAlias bool) (string, error) {
	switch tab := table.(type) {
	case nil:
		_, ok := b.aliasMap[fdName]
		if useAlias && ok {
			return fdName, nil
		}
		fd, ok := b.model.FieldMap[fdName]
		if !ok {
			return "", errs.NewErrUnknownField(fdName)
		}
		return fd.ColName, nil
	case Table:
		m, err := b.r.Get(tab.entity)
		if err != nil {
			return "", err
		}
		fd, ok := m.FieldMap[fdName]
		if !ok {
			return "", errs.NewErrUnknownField(fdName)
		}
		return fd.ColName, nil
	case Subquery:
		if len(tab.columns) > 0 {
			for _, col := range tab.columns {
				if col.selectedAlias() == fdName {
					return fdName, nil
				}
				if col.fieldName() == fdName {
					return b.colName(col.target(), fdName, useAlias)
				}
				return "", errs.NewErrUnknownField(fdName)
			}
		}
		return b.colName(tab.table, fdName, useAlias)
	default:
		return "", errs.NewErrUnsupportedExpressionType(tab)
	}
}

func (b *Builder) buildColumn(val Column, useAlias bool) error {
	var alias string
	if val.table != nil {
		alias = val.table.tableAlias()
	}
	if alias != "" {
		b.quote(alias)
		b.writeByte('.')
	}
	colName, err := b.colName(val.table, val.name, useAlias)
	if err != nil {
		return err
	}
	b.quote(colName)
	return nil
}

func (b *Builder) buildAggregate(val Aggregate, useAlias bool) error {
	b.writeString(val.fn)
	b.writeLeftParenthesis()
	err := b.buildColumn(
		Column{table: val.table, name: val.arg}, useAlias)
	if err != nil {
		return err
	}
	b.writeRightParenthesis()
	if useAlias {
		err := b.buildAs(val.alias)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) buildSubquery(sub Subquery, useAlias bool) error {
	q, err := sub.s.Build()
	if err != nil {
		return err
	}
	b.writeLeftParenthesis()
	b.writeString(q.SQL[:len(q.SQL)-1])
	b.writeRightParenthesis()
	if len(q.Args) > 0 {
		b.addArgs(q.Args...)
	}
	if useAlias {
		if err = b.buildAs(sub.alias); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) buildBinaryExpr(
	exp binaryExpr, colsAlias, aggreAlias bool) error {
	err := b.buildSubExpr(
		exp.left, colsAlias, aggreAlias)
	if err != nil {
		return err
	}
	if exp.op != "" {
		b.writeSpace()
		b.writeString(exp.op.String())
		b.writeSpace()
	}

	return b.buildSubExpr(
		exp.right, colsAlias, aggreAlias)
}

func (b *Builder) buildSubExpr(expr Expression, colsAlias, aggreAlias bool) error {
	switch e := expr.(type) {
	case MathExpr:
		b.writeLeftParenthesis()
		err := b.buildBinaryExpr(
			binaryExpr(e), colsAlias, aggreAlias)
		if err != nil {
			return err
		}
		b.writeRightParenthesis()
	case binaryExpr:
		b.writeLeftParenthesis()
		err := b.buildBinaryExpr(
			e, colsAlias, aggreAlias)
		if err != nil {
			return err
		}
		b.writeRightParenthesis()
	case Predicate:
		b.writeLeftParenthesis()
		err := b.buildBinaryExpr(
			binaryExpr(e), colsAlias, aggreAlias)
		if err != nil {
			return err
		}
		b.writeRightParenthesis()
	default:
		err := b.buildExpression(
			e, colsAlias, aggreAlias)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) buildExpression(
	e Expression, colsAlias, aggreAlias bool) error {
	switch exp := e.(type) {
	case nil:
		return nil
	case Column:
		return b.buildColumn(exp, colsAlias)
	case Aggregate:
		return b.buildAggregate(exp, aggreAlias)
	case value:
		return b.buildValue(exp)
	case RawExpr:
		return b.buildRaw(exp)
	case MathExpr:
		return b.buildBinaryExpr(
			binaryExpr(exp), colsAlias, aggreAlias)
	case Predicate:
		return b.buildBinaryExpr(
			binaryExpr(exp), colsAlias, aggreAlias)
	case Subquery:
		return b.buildSubquery(exp, false)
	case SubqueryExpr:
		b.writeString(exp.pred)
		b.writeSpace()
		return b.buildSubquery(exp.s, false)
	default:
		return errs.NewErrUnsupportedExpressionType(exp)
	}
}
