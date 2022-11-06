//go:build e2e

package integration

import (
	"context"
	"orm"
	"orm/internal/test"
	"testing"

	_ "github.com/go-sql-driver/mysql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type DeleteTestSuite struct {
	Suite
}

func (d *DeleteTestSuite) SetupSuite() {
	d.Suite.SetupSuite()
	data1 := test.NewSimpleStruct(1)
	data2 := test.NewSimpleStruct(2)
	data3 := test.NewSimpleStruct(3)
	res := orm.NewInserter[test.SimpleStruct](d.db).Values(data1, data2, data3).Exec(context.Background())
	require.NoError(d.T(), res.Err())
}

func (d *DeleteTestSuite) TearDownTest() {
	res := orm.NewDeleter[test.SimpleStruct](d.db).Exec(context.Background())
	require.NoError(d.T(), res.Err())
}

func (d *DeleteTestSuite) TestDeleter() {
	testCases := []struct {
		name         string
		d            *orm.Deleter[test.SimpleStruct]
		rowsAffected int64
		wantErr      error
	}{
		{
			name:         "id only",
			d:            orm.NewDeleter[test.SimpleStruct](d.db).Where(orm.C("Id").EQ("1")),
			rowsAffected: 1,
		},
		{
			name:         "delete all",
			d:            orm.NewDeleter[test.SimpleStruct](d.db),
			rowsAffected: 2,
		},
	}
	for _, tc := range testCases {
		d.T().Run(tc.name, func(t *testing.T) {
			res := tc.d.Exec(context.Background())
			require.Equal(t, tc.wantErr, res.Err())
			affected, err := res.RowsAffected()
			require.Nil(t, err)
			assert.Equal(t, tc.rowsAffected, affected)
		})
	}
}

func TestMySQL8tDelete(t *testing.T) {
	suite.Run(t, &DeleteTestSuite{
		Suite{
			driver: "mysql",
			dsn:    "root:123456@tcp(localhost:3306)/integration_test",
		},
	})
}
