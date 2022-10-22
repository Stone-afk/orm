package valuer

import (
	"database/sql"
	"orm/v10/internal/errs"
	"orm/v10/model"
	"reflect"
)

//type reflectValue struct {
//	t    any
//	meta *model.Model
//}

//func NewReflectValue(t any, model *model.Model) Value {
//	return &reflectValue{
//		t:    t,
//		meta: model,
//	}
//}

type reflectValue struct {
	val  reflect.Value
	meta *model.Model
}

func NewReflectValue(t any, model *model.Model) Value {
	return &reflectValue{
		val:  reflect.ValueOf(t).Elem(),
		meta: model,
	}
}

func (u *reflectValue) SetColumns(rows *sql.Rows) error {

	// 先看一下你返回了哪些列
	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	if len(cols) > len(u.meta.ColumnMap) {
		return errs.ErrTooManyReturnedColumns
	}

	colValues := make([]any, 0, len(cols))
	colElemVals := make([]reflect.Value, 0, len(cols))
	for _, col := range cols {
		fd, ok := u.meta.ColumnMap[col]
		if !ok {
			return errs.NewErrUnknownColumn(col)
		}
		// fd.Type 是 int，那么  reflect.New(fd.typ) 是 *int
		fdVal := reflect.New(fd.Type)
		colElemVals = append(colElemVals, fdVal.Elem())

		// 因为 Scan 要指针，所以在这里，不需要调用 Elem
		colValues = append(colValues, fdVal.Interface())
	}

	// 要把 cols 映射过去字段
	err = rows.Scan(colValues...)
	if err != nil {
		return err
	}

	// 有 vals 了，接下来将 vals= [123, "Ming", 18, "Deng"] 反射放回去 t 里面
	//t := u.t
	// 由于传递给reflect.ValueOf的 t 是一个指针，所以得到的 T 的类型是Ptr, 而FieldByName方法需要调用者类型为Struct
	//tVal := reflect.ValueOf(t).Elem()
	//for i, col := range cols {
	//	fd := u.meta.ColumnMap[col]
	//	tVal.FieldByName(fd.GoName).Set(colElemVals[i])
	//}

	for i, col := range cols {
		fd := u.meta.ColumnMap[col]
		u.val.FieldByName(fd.GoName).Set(colElemVals[i])
	}

	return nil
}

func (u *reflectValue) Field(name string) (any, error) {
	fd, ok := u.meta.FieldMap[name]
	if !ok {
		return nil, errs.NewErrUnknownField(name)
	}
	fdVal := u.val.Field(fd.Index)
	return fdVal.Interface(), nil
}
