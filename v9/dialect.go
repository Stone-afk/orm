//go:build v9 
package orm

import (
"fmt"
"orm/v9/internal/errs"
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
// buildOnConflict 构造插入冲突部分
buildOnConflict(b *builder, odk *OnConflict) error
ColTypeOf(typ reflect.Value) string
}

type standardSQL struct{}

func (d *standardSQL) quoter() byte {
// TODO implement me
panic("implement me")
}

type mysqlDialect struct {
standardSQL
}

func (d *mysqlDialect) quoter() byte {
return '`'
}

func (d *mysqlDialect) buildOnConflict(b *builder, odk *OnConflict) error {
b.writeString(" ON DUPLICATE KEY UPDATE ")
for pos, assign := range odk.assigns {
if pos > 0 {
b.writeComma()
}
switch a := assign.(type) {
case *Assignment:
fd, ok := b.model.FieldMap[a.column]
if !ok {
return errs.NewErrUnknownField(a.column)
}
b.quote(fd.ColName)
b.writeString(" = ")
b.writePlaceholder()
b.addArgs(a.val)
case *Column:
fd, ok := b.model.FieldMap[a.name]
if !ok {
return errs.NewErrUnknownField(a.name)
}
b.quote(fd.ColName)
b.writeString(" = ")
b.writeString("VALUES")
b.writeLeftParenthesis()
b.quote(fd.ColName)
b.writeRightParenthesis()
default:
return errs.NewErrUnsupportedAssignableType(assign)
}
}
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

func (d *standardSQL) buildOnConflict(b *builder, odk *OnConflict) error {
	return nil
}
