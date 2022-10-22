//go:build v10 
package orm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDelerer_Build(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			// From 都不调用
			name: "no from",
			q:    NewDeleter[TestModel](db),
			wantQuery: &Query{
				SQL: "DELETE FROM `test_model`;",
			},
		},
		{
			// 调用 FROM
			name: "with from",
			q:    NewDeleter[TestModel](db).From("`test_model_t`"),
			wantQuery: &Query{
				SQL: "DELETE FROM `test_model_t`;",
			},
		},
		{
			// WHERE
			name: "where",
			q:    NewDeleter[TestModel](db).Where(C("Id").EQ(16)),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE `id` = ?;",
				Args: []any{16},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
}
