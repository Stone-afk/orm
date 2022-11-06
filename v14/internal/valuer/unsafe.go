package valuer

import (
	"database/sql"
	"orm/v14/internal/errs"
	"orm/v14/model"
	"reflect"
	"unsafe"
)

type unsafeValue struct {
	meta *model.Model
	addr unsafe.Pointer
}

func NewUnsafeValue(t any, model *model.Model) Value {
	// t 的起始地址， 用来支持通过 t 里的字段的偏移量来计算 t 里字段的地址
	addr := unsafe.Pointer(reflect.ValueOf(t).Pointer())
	return &unsafeValue{
		meta: model,
		addr: addr,
	}
}

func (u *unsafeValue) SetColumns(rows *sql.Rows) error {

	// 先看一下你返回了哪些列
	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	if len(cols) > len(u.meta.ColumnMap) {
		return errs.ErrTooManyReturnedColumns
	}

	colValues := make([]any, len(cols))
	for i, col := range cols {
		fd, ok := u.meta.ColumnMap[col]
		if !ok {
			return errs.NewErrUnknownColumn(col)
		}
		// 要计算 字段 的真实地址：对象起始地址 + 字段偏移量
		uPtr := unsafe.Pointer(uintptr(u.addr) + fd.Offset)
		fdVal := reflect.NewAt(fd.Type, uPtr)
		// 通过索引约束所得到的 cols，如果列超过了长度，那么就会超过索引然后报错
		colValues[i] = fdVal.Interface()
	}
	return rows.Scan(colValues...)
}

func (u *unsafeValue) Field(name string) (any, error) {
	fd, ok := u.meta.FieldMap[name]
	if !ok {
		return nil, errs.NewErrUnknownField(name)
	}
	ptr := unsafe.Pointer(uintptr(u.addr) + fd.Offset)
	fdVal := reflect.NewAt(fd.Type, ptr).Elem()
	return fdVal.Interface(), nil
}
