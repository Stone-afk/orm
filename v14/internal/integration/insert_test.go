//go:build e2e

package integration

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"orm/v14"
	"orm/v14/internal/test"
	"testing"
)

type InsertTestSuite struct {
	Suite
}

// TearDownTest 清理该测试的专有数据，以及该测试产生的数据
func (i *InsertTestSuite) TearDownTest() {
	res := orm.NewDeleter[test.SimpleStruct](i.db).Exec(context.Background())
	require.NoError(i.T(), res.Err())
}

func (i *InsertTestSuite) TestGet() {
	testCases := []struct {
		name         string
		i            *orm.Inserter[test.SimpleStruct]
		wantErr      error
		rowsAffected int64
		wantData     *test.SimpleStruct
	}{
		{
			name: "id only",
			i: orm.NewInserter[test.SimpleStruct](i.db).
				Values(&test.SimpleStruct{Id: 1}),
			rowsAffected: 1,
			wantErr:      orm.ErrNoRows,
			wantData:     &test.SimpleStruct{Id: 1},
		},
		{
			name: "all field",
			i: orm.NewInserter[test.SimpleStruct](i.db).
				Values(test.NewSimpleStruct(2)),
			wantData: test.NewSimpleStruct(2),
		},
	}

	for _, tc := range testCases {
		i.T().Run(tc.name, func(t *testing.T) {
			res := tc.i.Exec(context.Background())
			affected, err := res.RowsAffected()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.rowsAffected, affected)
			data, err := orm.NewSelector[test.SimpleStruct](i.db).
				Where(orm.C("Id").EQ(tc.wantData.Id)).
				Get(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tc.wantData, data)
		})
	}
}

func TestInsertMySQL(t *testing.T) {
	suite.Run(t, &InsertTestSuite{
		Suite: Suite{
			driver: "mysql",
			dsn:    "root:123456@tcp(localhost:3306)/integration_test",
		},
	})
}
