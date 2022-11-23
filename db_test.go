package orm

import (
	"orm/internal/valuer"
	"testing"
)

func TestDBUseReflectValuer(t *testing.T) {
	Open("sqlite3", "file:test.db?cache=shared&mode=memory",
		DBWithValCreator(valuer.NewUnsafeValue))
}
