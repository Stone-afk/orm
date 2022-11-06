//go:build e2e

package integration

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"orm"
)

type Suite struct {
	suite.Suite
	db     *orm.DB
	driver string
	dsn    string
}

func (i *Suite) SetupSuite() {
	db, err := orm.Open(i.driver, i.dsn)
	require.NoError(i.T(), err)
	err = db.Wait()
	require.NoError(i.T(), err)
	i.db = db
}
