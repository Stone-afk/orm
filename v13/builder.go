package orm

import (
	"github.com/valyala/bytebufferpool"
	"orm/v13/internal/errs"
	"orm/v13/model"
)

type Builder struct {
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
		b.writeByte('`')
		b.writeString(alias)
		b.writeByte('`')
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

func (b *Builder) buildColumn(val Column, useAlias bool) error {
	fd, ok := b.model.FieldMap[val.name]
	if useAlias {
		if !ok {
			_, ok = b.aliasMap[val.name]
			if !ok {
				return errs.NewErrUnknownField(val.name)
			}
			b.quote(val.name)
		} else {
			b.quote(fd.ColName)
			err := b.buildAs(val.alias)
			if err != nil {
				return err
			}
		}
	} else {
		if !ok {
			return errs.NewErrUnknownField(val.name)
		}
		b.quote(fd.ColName)
	}
	return nil
}

func (b *Builder) buildAggregate(val Aggregate, useAlias bool) error {
	fd, ok := b.model.FieldMap[val.arg]
	if !ok {
		return errs.NewErrUnknownField(val.arg)
	}
	b.writeString(val.fn)
	b.writeLeftParenthesis()
	b.quote(fd.ColName)
	b.writeRightParenthesis()
	if useAlias {
		err := b.buildAs(val.alias)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) buildBinaryExpr(exp binaryExpr, useColsAlias, useAggreAlias bool) error {
	err := b.buildSubExpr(exp.left, useColsAlias, useAggreAlias)
	if err != nil {
		return err
	}
	if exp.op != "" {
		b.writeSpace()
		b.writeString(exp.op.String())
		b.writeSpace()
	}

	return b.buildSubExpr(exp.right, useColsAlias, useAggreAlias)
}

func (b *Builder) buildSubExpr(expr Expression, useColsAlias, useAggreAlias bool) error {
	switch e := expr.(type) {
	case MathExpr:
		b.writeLeftParenthesis()
		err := b.buildBinaryExpr(binaryExpr(e), useColsAlias, useAggreAlias)
		if err != nil {
			return err
		}
		b.writeRightParenthesis()
	case binaryExpr:
		b.writeLeftParenthesis()
		err := b.buildBinaryExpr(e, useColsAlias, useAggreAlias)
		if err != nil {
			return err
		}
		b.writeRightParenthesis()
	case Predicate:
		b.writeLeftParenthesis()
		err := b.buildBinaryExpr(binaryExpr(e), useColsAlias, useAggreAlias)
		if err != nil {
			return err
		}
		b.writeRightParenthesis()
	default:
		err := b.buildExpression(e, useColsAlias, useAggreAlias)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) buildExpression(e Expression, useColsAlias, useAggreAlias bool) error {
	switch exp := e.(type) {
	case nil:
		return nil
	case Column:
		return b.buildColumn(exp, useColsAlias)
	case Aggregate:
		return b.buildAggregate(exp, useAggreAlias)
	case value:
		return b.buildValue(exp)
	case RawExpr:
		return b.buildRaw(exp)
	case MathExpr:
		return b.buildBinaryExpr(binaryExpr(exp), useColsAlias, useAggreAlias)
	case Predicate:
		return b.buildBinaryExpr(binaryExpr(exp), useColsAlias, useAggreAlias)
	default:
		return errs.NewErrUnsupportedExpressionType(exp)
	}
}
