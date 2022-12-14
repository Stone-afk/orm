package orm

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Middleware(t *testing.T) {
	testCases := []struct {
		name    string
		wantErr error
		mdls    []Middleware
	}{
		{
			name: "one middleware",
			mdls: func() []Middleware {
				var mdl Middleware = func(next HandleFunc) HandleFunc {
					return func(ctx context.Context, queryContext *QueryContext) *QueryResult {
						return &QueryResult{}
					}
				}
				return []Middleware{mdl}
			}(),
		},
		{
			name: "many middleware",
			mdls: func() []Middleware {
				mdl1 := func(next HandleFunc) HandleFunc {
					return func(ctx context.Context, queryContext *QueryContext) *QueryResult {
						return &QueryResult{Result: "mdl1"}
					}
				}
				mdl2 := func(next HandleFunc) HandleFunc {
					return func(ctx context.Context, queryContext *QueryContext) *QueryResult {
						return &QueryResult{Result: "mdl2"}
					}
				}
				return []Middleware{mdl1, mdl2}
			}(),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			orm, err := Open("sqlite3", "file:test.db?cache=shared&mode=memory",
				DBWithMiddlewares(tc.mdls...))
			if err != nil {
				t.Error(err)
			}
			defer func() {
				_ = orm.Close()
			}()
			assert.EqualValues(t, tc.mdls, orm.ms)
		})
	}
}
