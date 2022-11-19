package orm

import (
	"fmt"
	"orm/internal/errs"
	"reflect"
	"time"
)

var (
	MySQL   Dialect = &mysqlDialect{}
	SQLite3 Dialect = &sqlite3Dialect{}
)

type Dialect interface {
	// quoter 返回一个引号，引用列名，表名的引号
	quoter() byte
	ColTypeOf(typ reflect.Value) string
	// buildOnConflict 构造插入冲突部分
	buildOnConflict(b *Builder, odk *OnConflict) error
}

type standardSQL struct{}

func (d *standardSQL) quoter() byte {
	// TODO implement me
	panic("implement me")
}

func (d *standardSQL) buildOnConflict(b *Builder, odk *OnConflict) error {
	// TODO implement me
	panic("implement me")
}

func (d *standardSQL) ColTypeOf(typ reflect.Value) string {
	// TODO implement me
	panic("implement me")
}

type mysqlDialect struct {
	standardSQL
}

func (d *mysqlDialect) quoter() byte {
	return '`'
}

func (d *mysqlDialect) buildOnConflict(b *Builder, odk *OnConflict) error {
	b.writeString(" ON DUPLICATE KEY UPDATE ")
	for pos, assign := range odk.assigns {
		if pos > 0 {
			b.writeComma()
		}
		switch a := assign.(type) {
		case Assignment:
			if err := b.buildAssignment(a); err != nil {
				return err
			}
		case Column:
			if err := d.buildConflictColumn(b, a); err != nil {
				return err
			}
		default:
			return errs.NewErrUnsupportedAssignableType(assign)
		}
	}
	return nil
}

func (d *mysqlDialect) buildConflictColumn(b *Builder, c Column) error {
	fd, ok := b.model.FieldMap[c.name]
	if !ok {
		return errs.NewErrUnknownField(c.name)
	}
	b.quote(fd.ColName)
	b.writeString(" = ")
	b.writeString("VALUES")
	b.writeLeftParenthesis()
	b.quote(fd.ColName)
	b.writeRightParenthesis()
	return nil
}

func (d *mysqlDialect) ColTypeOf(typ reflect.Value) string {
	switch typ.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uintptr:
		return "int(11)"
	case reflect.Int64, reflect.Uint64:
		return "bigint(11)"
	case reflect.Float32, reflect.Float64:
		return "float(11)"
	case reflect.String:
		return "longtext"
	case reflect.Array, reflect.Slice:
		return "blob"
	case reflect.Struct:
		if _, ok := typ.Interface().(time.Time); ok {
			return "datetime"
		}
	}
	panic(fmt.Sprintf("invalid sql type %s (%s)", typ.Type().Name(), typ.Kind()))
}

type sqlite3Dialect struct {
	standardSQL
}

func (d *sqlite3Dialect) quoter() byte {
	return '`'
}

func (d *sqlite3Dialect) ColTypeOf(typ reflect.Value) string {
	switch typ.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uintptr:
		return "integer"
	case reflect.Int64, reflect.Uint64:
		return "bigint"
	case reflect.Float32, reflect.Float64:
		return "real"
	case reflect.String:
		return "text"
	case reflect.Array, reflect.Slice:
		return "blob"
	case reflect.Struct:
		if _, ok := typ.Interface().(time.Time); ok {
			return "datetime"
		}
	}
	panic(fmt.Sprintf("invalid sql type %s (%s)", typ.Type().Name(), typ.Kind()))

}

func (d *sqlite3Dialect) buildOnConflict(b *Builder, odk *OnConflict) error {
	b.writeString(" ON CONFLICT")
	if len(odk.conflictColumns) > 0 {
		b.writeLeftParenthesis()
		for i, col := range odk.conflictColumns {
			if i > 0 {
				b.writeComma()
			}
			fd, ok := b.model.FieldMap[col]
			if !ok {
				return errs.NewErrUnknownField(col)
			}
			b.quote(fd.ColName)
		}
		b.writeRightParenthesis()
	}
	b.writeString(" DO UPDATE SET ")
	for pos, assign := range odk.assigns {
		if pos > 0 {
			b.writeComma()
		}
		switch a := assign.(type) {
		case Assignment:
			if err := b.buildAssignment(a); err != nil {
				return err
			}
		case Column:
			if err := d.buildConflictColumn(b, a); err != nil {
				return err
			}
		default:
			return errs.NewErrUnsupportedAssignableType(assign)
		}
	}
	return nil
}

func (d *sqlite3Dialect) buildConflictColumn(b *Builder, c Column) error {
	fd, ok := b.model.FieldMap[c.name]
	if !ok {
		return errs.NewErrUnknownField(c.name)
	}
	b.quote(fd.ColName)
	b.writeString(" = ")
	b.writeString("excluded.")
	b.quote(fd.ColName)
	return nil
}
