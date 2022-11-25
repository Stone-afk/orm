//go:build e2e

package integration

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"orm"
	"orm/internal/test"
	"testing"
)

type SelectTestSuite struct {
	Suite
}

func (s *SelectTestSuite) SetupSuite() {
	s.Suite.SetupSuite()
	res := orm.NewInserter[test.SimpleStruct](s.db).
		Values(test.NewSimpleStruct(1), test.NewSimpleStruct(2), test.NewSimpleStruct(3)).
		Exec(context.Background())
	require.NoError(s.T(), res.Err())

}

// TearDownSuite 关闭环境，清理全部数据
func (s *SelectTestSuite) TearDownSuite() {
	res := orm.NewDeleter[test.SimpleStruct](s.db).Exec(context.Background())
	require.NoError(s.T(), res.Err())
}

func (s *SelectTestSuite) TestGet() {
	testCases := []struct {
		name    string
		s       *orm.Selector[test.SimpleStruct]
		wantErr error
		wantRes *test.SimpleStruct
	}{
		{
			name: "not found",
			s: orm.NewSelector[test.SimpleStruct](s.db).
				Where(orm.C("Id").EQ(9)),
			wantErr: orm.ErrNoRows,
		},
		{
			name: "found",
			s: orm.NewSelector[test.SimpleStruct](s.db).
				Where(orm.C("Id").EQ(1)),
			wantRes: test.NewSimpleStruct(1),
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			res, err := tc.s.Get(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func (s *SelectTestSuite) TestGet_baseType() {
	testCases := []struct {
		name     string
		queryRes func(t *testing.T) any
		wantErr  error
		wantRes  any
	}{
		{
			name: "not found",
			queryRes: func(t *testing.T) any {
				queryer := orm.NewSelector[test.SimpleStruct](s.db).
					Where(orm.C("Id").EQ(9))
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			wantErr: orm.ErrNoRows,
		},
		{
			name: "res struct ptr",
			queryRes: func(t *testing.T) any {
				queryer := orm.NewSelector[test.SimpleStruct](s.db).
					Where(orm.C("Id").EQ(1))
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			wantRes: test.NewSimpleStruct(1),
		},
		{
			name: "res int",
			queryRes: func(t *testing.T) any {
				queryer := orm.NewSelector[int](s.db).Select(orm.Avg("Id")).
					From(orm.TableOf(&test.SimpleStruct{}))
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			wantRes: func() *int {
				res := 2
				return &res
			}(),
		},
		{
			name: "res string",
			queryRes: func(t *testing.T) any {
				queryer := orm.NewSelector[string](s.db).Select(orm.C("String")).
					From(orm.TableOf(&test.SimpleStruct{})).Where(orm.C("Id").EQ(1))
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			wantRes: func() *string {
				res := "word"
				return &res
			}(),
		},
		{
			name: "res bytes",
			queryRes: func(t *testing.T) any {
				queryer := orm.NewSelector[[]byte](s.db).Select(orm.C("ByteArray")).
					From(orm.TableOf(&test.SimpleStruct{})).Where(orm.C("Id").EQ(1))
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			wantRes: func() *[]byte {
				res := []byte("hello")
				return &res
			}(),
		},
		{
			name: "res bool",
			queryRes: func(t *testing.T) any {
				queryer := orm.NewSelector[bool](s.db).Select(orm.C("Bool")).
					From(orm.TableOf(&test.SimpleStruct{})).Where(orm.C("Id").EQ(1))
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			wantRes: func() *bool {
				res := true
				return &res
			}(),
		},
		{
			name: "res null string ptr",
			queryRes: func(t *testing.T) any {
				queryer := orm.NewSelector[sql.NullString](s.db).Select(orm.C("NullStringPtr")).
					From(orm.TableOf(&test.SimpleStruct{})).Where(orm.C("Id").EQ(1))
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			wantRes: func() *sql.NullString {
				res := sql.NullString{String: "null string", Valid: true}
				return &res
			}(),
		},
		{
			name: "res null int32 ptr",
			queryRes: func(t *testing.T) any {
				queryer := orm.NewSelector[sql.NullInt32](s.db).Select(orm.C("NullInt32Ptr")).
					From(orm.TableOf(&test.SimpleStruct{})).Where(orm.C("Id").EQ(1))
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			wantRes: func() *sql.NullInt32 {
				res := sql.NullInt32{Int32: 32, Valid: true}
				return &res
			}(),
		},
		{
			name: "res null bool ptr",
			queryRes: func(t *testing.T) any {
				queryer := orm.NewSelector[sql.NullBool](s.db).Select(orm.C("NullBoolPtr")).
					From(orm.TableOf(&test.SimpleStruct{})).Where(orm.C("Id").EQ(1))
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			wantRes: func() *sql.NullBool {
				res := sql.NullBool{Bool: true, Valid: true}
				return &res
			}(),
		},
		{
			name: "res null float64 ptr",
			queryRes: func(t *testing.T) any {
				queryer := orm.NewSelector[sql.NullFloat64](s.db).Select(orm.C("NullFloat64Ptr")).
					From(orm.TableOf(&test.SimpleStruct{})).Where(orm.C("Id").EQ(1))
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			wantRes: func() *sql.NullFloat64 {
				res := sql.NullFloat64{Float64: 6.4, Valid: true}
				return &res
			}(),
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			res := tc.queryRes(t)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func TestSelectMySQL(t *testing.T) {
	suite.Run(t, &SelectTestSuite{
		Suite: Suite{
			driver: "mysql",
			dsn:    "root:123456@tcp(localhost:3306)/integration_test",
		},
	})
}
