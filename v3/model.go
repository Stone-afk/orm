//go:build v3

package orm

import (
	"orm/internal/errs"
	"reflect"
	"unicode"
)

// field 字段
type Field struct {
	colName string
}

type Model struct {
	// tableName 结构体对应的表名
	tableName string
	// 字段名到字段的元数据
	fieldMap map[string]*Field
}

func parseModel(val any) (*Model, error) {
	if val == nil {
		return nil, errs.ErrInputNil
	}
	ptrTyp := reflect.TypeOf(val)
	typ := ptrTyp.Elem()
	if ptrTyp.Kind() != reflect.Ptr || typ.Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}

	// 冷酷无情版
	// if typ.Kind() != reflect.Struct {
	// 	return nil, errors.New("只支持结构体作为输入")
	// }
	// 温情版 只支持一级指针

	// 三级烧伤版
	// for typ.Kind() == reflect.Ptr {
	// 	typ = typ.Elem()
	// }
	//
	// if typ.Kind() != reflect.Struct {
	// 	return nil, errors.New("只支持结构体或者结构体的（多级）指针作为输入")
	// }

	// 获得字段的数量
	numField := typ.NumField()
	fds := make(map[string]*Field, numField)
	for i := 0; i < numField; i++ {
		fdType := typ.Field(i)
		fds[fdType.Name] = &Field{
			colName: underscoreName(fdType.Name),
		}
	}
	return &Model{
		tableName: underscoreName(typ.Name()),
		fieldMap:  fds,
	}, nil
}

// underscoreName 驼峰转字符串命名
func underscoreName(tableName string) string {
	var buf []byte
	for i, v := range tableName {
		if unicode.IsUpper(v) {
			if i != 0 {
				buf = append(buf, '_')
			}
			buf = append(buf, byte(unicode.ToLower(v)))
		} else {
			buf = append(buf, byte(v))
		}
	}
	return string(buf)
}
