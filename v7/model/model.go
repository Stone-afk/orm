package model

import (
	"orm/internal/errs"
	"reflect"
)

// field 字段
type Field struct {
	ColName string
	GoName  string
	Type    reflect.Type

	// Offset 相对于对象起始地址的字段偏移量
	Offset uintptr
}

type Model struct {
	// tableName 结构体对应的表名
	TableName string
	// 字段名到字段的元数据
	FieldMap  map[string]*Field
	ColumnMap map[string]*Field
}

type Option func(model *Model) error

func WithTableName(tableName string) Option {
	return func(model *Model) error {
		model.TableName = tableName
		return nil
	}
}

func WithColumnName(field string, columnName string) Option {
	return func(model *Model) error {
		fd, ok := model.FieldMap[field]
		if !ok {
			return errs.NewErrUnknownField(field)
		}
		// 注意，这里根本没有检测 colName 会不会是空字符串
		// 因为正常情况下，用户都不会写错
		// 即便写错了，也很容易在测试中发现
		fd.ColName = columnName
		return nil
	}
}

// 我们支持的全部标签上的 key 都放在这里
// 方便用户查找，和我们后期维护
const (
	tagKeyColumn = "column"
)

// 用户自定义一些模型信息的接口，集中放在这里
// 方便用户查找和我们后期维护

// TableName 用户实现这个接口来返回自定义的表名
type TableName interface {
	TableName() string
}
