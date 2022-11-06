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

type UpdateTestSuite struct {
	Suite
}

func (u *UpdateTestSuite) SetupSuite() {
	u.Suite.SetupSuite()
	data1 := test.NewSimpleStruct(1)
	res := orm.NewInserter[test.SimpleStruct](u.db).Values(data1).Exec(context.Background())
	require.NoError(u.T(), res.Err())
}

func (u *UpdateTestSuite) TearDownTest() {
	res := orm.NewDeleter[test.SimpleStruct](u.db).Exec(context.Background())
	require.NoError(u.T(), res.Err())
}

func (u *UpdateTestSuite) TestUpdate() {
	testCases := []struct {
		name         string
		u            *orm.Updater[test.SimpleStruct]
		rowsAffected int64
		wantErr      error
	}{
		{
			name: "update columns",
			u: orm.NewUpdater[test.SimpleStruct](u.db).Update(&test.SimpleStruct{Int: 18}).
				Set(orm.C("Int")).Where(orm.C("Id").EQ(1)),
			rowsAffected: 1,
		},
	}
	for _, tc := range testCases {
		u.T().Run(tc.name, func(t *testing.T) {
			res := tc.u.Exec(context.Background())
			assert.Equal(t, tc.wantErr, res.Err())
			if res.Err() != nil {
				return
			}
			affected, err := res.RowsAffected()
			require.Nil(t, err)
			assert.Equal(t, tc.rowsAffected, affected)
		})
	}
}

func TestMySQL8Update(t *testing.T) {
	suite.Run(t, &UpdateTestSuite{
		Suite{
			driver: "mysql",
			dsn:    "root:123456@tcp(localhost:3306)/integration_test",
		},
	})
}
