//go:build e2e

package integration

import (
	"github.com/stretchr/testify/suite"
	"orm"
)

type Suite struct {
	suite.Suite
	db     *orm.DB
	driver string
	dsn    string
}

func (s *Suite) SetupSuite() {
	t := s.T()
	db, err := orm.Open(s.driver, s.dsn)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Wait(); err != nil {
		t.Fatal(err)
	}
	s.db = db
}

//func (i *Suite) SetupSuite() {
//	db, err := orm.Open(i.driver, i.dsn)
//	require.NoError(i.T(), err)
//	err = db.Wait()
//	require.NoError(i.T(), err)
//	i.db = db
//}
