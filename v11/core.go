//go:build v11
package orm

import (
	"orm/v11/internal/valuer"
	"orm/v11/model"
)

// core 只是一个简单的封装，将一些 CRUD 都 需要使用的东西放到了一起。
type core struct {
	r          model.Registry
	valCreator valuer.Creator
	dialect    Dialect
}
