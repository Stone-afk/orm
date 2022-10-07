//go:build v4

package orm

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
