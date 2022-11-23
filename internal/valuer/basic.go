package valuer

import (
	"database/sql"
	"orm/model"
	"reflect"
)

type supportBasicTypeValue struct {
	Value       // unsafe 或者 reflect 的实现
	val     any // 就是 *T
	valType reflect.Type
}

func (s *supportBasicTypeValue) Field(name string) (any, error) {
	return s.Value.Field(name)
}

func (s *supportBasicTypeValue) SetColumns(rows *sql.Rows) error {
	switch s.valType.Elem().Kind() {
	case reflect.Struct:
		return s.Value.SetColumns(rows)
	default:
		return rows.Scan(s.val)
	}
}

type BasicTypeCreator struct {
	Creator
}

func (s BasicTypeCreator) NewBasicTypeValue(
	t any, model *model.Model) *supportBasicTypeValue {
	return &supportBasicTypeValue{
		val:     t,
		Value:   s.Creator(t, model),
		valType: reflect.TypeOf(t),
	}
}
