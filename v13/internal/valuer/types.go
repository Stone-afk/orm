package valuer

import (
	"database/sql"
	"orm/v13/model"
)

// 先来一个反射和 unsafe 的抽象

// Value 是对结构体实例的内部抽象
type Value interface {
	// Field 返回字段对应的值
	Field(name string) (any, error)
	// SetColumns 设置新值
	SetColumns(rows *sql.Rows) error
}

// Creator 本质上也可以看所是 factory 模式，极其简单的 factory 模式
type Creator func(t any, model *model.Model) Value

// ResultSetHandler 这是另外一种可行的设计方案
// type ResultSetHandler interface {
// 	// SetColumns 设置新值，column 是列名
// 	SetColumns(val any, rows *sql.Rows) error
// }

type Setter interface {

	// SetColumns 设置新值
	SetColumns(val any, rows *sql.Rows) error
}

type Getter interface {
	// Field 返回字段对应的值
	Field(name string) (any, error)
}
