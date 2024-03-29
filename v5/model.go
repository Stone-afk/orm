//go:build v5

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
