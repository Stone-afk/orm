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
	data *test.SimpleStruct
}

func (s *SelectTestSuite) SetupSuite() {
	s.Suite.SetupSuite()
	s.data = test.NewSimpleStruct(1)
	res := orm.NewInserter[test.SimpleStruct](s.db).
		Values(s.data, test.NewSimpleStruct(2), test.NewSimpleStruct(3)).
		Exec(context.Background())
	require.NoError(s.T(), res.Err())

}

// TearDownSuite 关闭环境，清理全部数据
func (s *SelectTestSuite) TearDownSuite() {
	res := orm.NewDeleter[test.SimpleStruct](s.db).Exec(context.Background())
	require.NoError(s.T(), res.Err())
}

func (s *SelectTestSuite) TestSelectorGet() {
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
			wantRes: s.data,
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

func (s *SelectTestSuite) TestSelectorGetMulti() {
	testCases := []struct {
		name    string
		s       *orm.Selector[test.SimpleStruct]
		wantErr error
		wantRes []*test.SimpleStruct
	}{
		{
			name: "not found",
			s: orm.NewSelector[test.SimpleStruct](s.db).
				Where(orm.C("Id").EQ(9)),
			wantRes: []*test.SimpleStruct{},
		},
		{
			name: "found",
			s:    orm.NewSelector[test.SimpleStruct](s.db),
			wantRes: []*test.SimpleStruct{
				test.NewSimpleStruct(1),
				test.NewSimpleStruct(2),
				test.NewSimpleStruct(3),
			},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			res, err := tc.s.GetMulti(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func (s *SelectTestSuite) TestSelectorGet_baseType() {
	testCases := []struct {
		name     string
		queryRes func() (any, error)
		wantErr  error
		wantRes  any
	}{
		{
			name: "res int",
			queryRes: func() (any, error) {
				queryer := orm.NewSelector[int](s.db).Select(orm.C("Id")).
					From(orm.TableOf(&test.SimpleStruct{})).
					Where(orm.C("Id").EQ(1))
				return queryer.Get(context.Background())
			},
			wantRes: func() *int {
				res := 1
				return &res
			}(),
		},
		{
			name: "res string",
			queryRes: func() (any, error) {
				queryer := orm.NewSelector[string](s.db).Select(orm.C("String")).
					From(orm.TableOf(&test.SimpleStruct{})).Where(orm.C("Id").EQ(1))
				return queryer.Get(context.Background())
			},
			wantRes: func() *string {
				res := "world"
				return &res
			}(),
		},
		{
			name: "res bytes",
			queryRes: func() (any, error) {
				queryer := orm.NewSelector[[]byte](s.db).Select(orm.C("ByteArray")).
					From(orm.TableOf(&test.SimpleStruct{})).Where(orm.C("Id").EQ(1))
				return queryer.Get(context.Background())
			},
			wantRes: func() *[]byte {
				res := []byte("hello")
				return &res
			}(),
		},
		{
			name: "res bool",
			queryRes: func() (any, error) {
				queryer := orm.NewSelector[bool](s.db).Select(orm.C("Bool")).
					From(orm.TableOf(&test.SimpleStruct{})).Where(orm.C("Id").EQ(1))
				return queryer.Get(context.Background())
			},
			wantRes: func() *bool {
				res := true
				return &res
			}(),
		},
		{
			name: "res null string ptr",
			queryRes: func() (any, error) {
				queryer := orm.NewSelector[sql.NullString](s.db).Select(orm.C("NullStringPtr")).
					From(orm.TableOf(&test.SimpleStruct{})).Where(orm.C("Id").EQ(1))
				return queryer.Get(context.Background())
			},
			wantRes: func() *sql.NullString {
				res := sql.NullString{String: "null string", Valid: true}
				return &res
			}(),
		},
		{
			name: "res null int32 ptr",
			queryRes: func() (any, error) {
				queryer := orm.NewSelector[sql.NullInt32](s.db).Select(orm.C("NullInt32Ptr")).
					From(orm.TableOf(&test.SimpleStruct{})).Where(orm.C("Id").EQ(1))
				return queryer.Get(context.Background())
			},
			wantRes: func() *sql.NullInt32 {
				res := sql.NullInt32{Int32: 32, Valid: true}
				return &res
			}(),
		},
		{
			name: "res null bool ptr",
			queryRes: func() (any, error) {
				queryer := orm.NewSelector[sql.NullBool](s.db).Select(orm.C("NullBoolPtr")).
					From(orm.TableOf(&test.SimpleStruct{})).Where(orm.C("Id").EQ(1))
				return queryer.Get(context.Background())
			},
			wantRes: func() *sql.NullBool {
				res := sql.NullBool{Bool: true, Valid: true}
				return &res
			}(),
		},
		{
			name: "res null float64 ptr",
			queryRes: func() (any, error) {
				queryer := orm.NewSelector[sql.NullFloat64](s.db).Select(orm.C("NullFloat64Ptr")).
					From(orm.TableOf(&test.SimpleStruct{})).Where(orm.C("Id").EQ(1))
				return queryer.Get(context.Background())
			},
			wantRes: func() *sql.NullFloat64 {
				res := sql.NullFloat64{Float64: 6.4, Valid: true}
				return &res
			}(),
		},
		{
			name: "res null float64 ptr",
			queryRes: func() (any, error) {
				queryer := orm.NewSelector[sql.NullFloat64](s.db).Select(orm.C("NullFloat64Ptr")).
					From(orm.TableOf(&test.SimpleStruct{})).Where(orm.C("Id").EQ(1))
				return queryer.Get(context.Background())
			},
			wantRes: func() *sql.NullFloat64 {
				res := sql.NullFloat64{Float64: 6.4, Valid: true}
				return &res
			}(),
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			res, err := tc.queryRes()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func (s *SelectTestSuite) TestSelectorGetMulti_baseType() {
	testCases := []struct {
		name     string
		queryRes func() (any, error)
		wantErr  error
		wantRes  any
	}{
		{
			name: "res int",
			queryRes: func() (any, error) {
				queryer := orm.NewSelector[int](s.db).Select(orm.C("Id")).
					From(orm.TableOf(&test.SimpleStruct{}))
				return queryer.GetMulti(context.Background())
			},
			wantRes: func() (res []*int) {
				vals := []int{1, 2, 3}
				for i := 0; i < len(vals); i++ {
					res = append(res, &vals[i])
				}
				return
			}(),
		},
		{
			name: "res string",
			queryRes: func() (any, error) {
				queryer := orm.NewSelector[string](s.db).Select(orm.C("String")).
					From(orm.TableOf(&test.SimpleStruct{}))
				return queryer.GetMulti(context.Background())
			},
			wantRes: func() (res []*string) {
				vals := []string{"world", "world", "world"}
				for i := 0; i < len(vals); i++ {
					res = append(res, &vals[i])
				}
				return
			}(),
		},
		{
			name: "res bytes",
			queryRes: func() (any, error) {
				queryer := orm.NewSelector[[]byte](s.db).Select(orm.C("ByteArray")).
					From(orm.TableOf(&test.SimpleStruct{}))
				return queryer.GetMulti(context.Background())
			},
			wantRes: func() (res []*[]byte) {
				vals := [][]byte{[]byte("hello"), []byte("hello"), []byte("hello")}
				for i := 0; i < len(vals); i++ {
					res = append(res, &vals[i])
				}
				return
			}(),
		},
		{
			name: "res bool",
			queryRes: func() (any, error) {
				queryer := orm.NewSelector[bool](s.db).Select(orm.C("Bool")).
					From(orm.TableOf(&test.SimpleStruct{}))
				return queryer.GetMulti(context.Background())
			},
			wantRes: func() (res []*bool) {
				vals := []bool{true, true, true}
				for i := 0; i < len(vals); i++ {
					res = append(res, &vals[i])
				}
				return
			}(),
		},
		{
			name: "res null string ptr",
			queryRes: func() (any, error) {
				queryer := orm.NewSelector[sql.NullString](s.db).Select(orm.C("NullStringPtr")).
					From(orm.TableOf(&test.SimpleStruct{}))
				return queryer.GetMulti(context.Background())
			},
			wantRes: []*sql.NullString{
				{
					String: "null string",
					Valid:  true,
				},
				{
					String: "null string",
					Valid:  true,
				},
				{
					String: "null string",
					Valid:  true,
				},
			},
		},
		{
			name: "res null int32 ptr",
			queryRes: func() (any, error) {
				queryer := orm.NewSelector[sql.NullInt32](s.db).Select(orm.C("NullInt32Ptr")).
					From(orm.TableOf(&test.SimpleStruct{}))
				return queryer.GetMulti(context.Background())
			},
			wantRes: []*sql.NullInt32{
				{
					Int32: 32,
					Valid: true,
				},
				{
					Int32: 32,
					Valid: true,
				},
				{
					Int32: 32,
					Valid: true,
				},
			},
		},
		{
			name: "res null bool ptr",
			queryRes: func() (any, error) {
				queryer := orm.NewSelector[sql.NullBool](s.db).Select(orm.C("NullBoolPtr")).
					From(orm.TableOf(&test.SimpleStruct{}))
				return queryer.GetMulti(context.Background())
			},
			wantRes: []*sql.NullBool{
				{
					Bool:  true,
					Valid: true,
				},
				{
					Bool:  true,
					Valid: true,
				},
				{
					Bool:  true,
					Valid: true,
				},
			},
		},
		{
			name: "res null float64 ptr",
			queryRes: func() (any, error) {
				queryer := orm.NewSelector[sql.NullFloat64](s.db).Select(orm.C("NullFloat64Ptr")).
					From(orm.TableOf(&test.SimpleStruct{}))
				return queryer.GetMulti(context.Background())
			},
			wantRes: []*sql.NullFloat64{
				{
					Float64: 6.4,
					Valid:   true,
				},
				{
					Float64: 6.4,
					Valid:   true,
				},
				{
					Float64: 6.4,
					Valid:   true,
				},
			},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			res, err := tc.queryRes()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.EqualValues(t, tc.wantRes, res)
		})
	}
}

func (s *SelectTestSuite) TestRawQueryGet() {
	testCases := []struct {
		name    string
		s       *orm.RawQuerier[test.SimpleStruct]
		wantErr error
		wantRes *test.SimpleStruct
	}{
		{
			name:    "not found",
			s:       orm.RawQuery[test.SimpleStruct](s.db, "SELECT `id` FROM `simple_struct` WHERE `id` = ?;", 9),
			wantErr: orm.ErrNoRows,
		},
		{
			name:    "found",
			s:       orm.RawQuery[test.SimpleStruct](s.db, "SELECT * FROM `simple_struct` WHERE `id` = ?;", 1),
			wantRes: s.data,
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

func (s *SelectTestSuite) TestRawQueryGet_baseType() {
	testCases := []struct {
		name     string
		queryRes func() (any, error)
		wantErr  error
		wantRes  any
	}{
		{
			name: "res int",
			queryRes: func() (any, error) {
				queryer := orm.RawQuery[int](s.db, "SELECT `id` FROM `simple_struct` WHERE `id` = ?;", 1)
				return queryer.Get(context.Background())
			},
			wantRes: func() *int {
				res := 1
				return &res
			}(),
		},
		{
			name: "res int convert string",
			queryRes: func() (any, error) {
				queryer := orm.RawQuery[string](s.db, "SELECT `id` FROM `simple_struct` WHERE `id` = ?;", 1)
				return queryer.Get(context.Background())
			},
			wantRes: func() *string {
				res := "1"
				return &res
			}(),
		},
		{
			name: "res int convert bytes",
			queryRes: func() (any, error) {
				queryer := orm.RawQuery[[]byte](s.db, "SELECT `id` FROM `simple_struct` WHERE `id` = ?;", 1)
				return queryer.Get(context.Background())
			},
			wantRes: func() *[]byte {
				res := []byte("1")
				return &res
			}(),
		},
		{
			name: "res string",
			queryRes: func() (any, error) {
				queryer := orm.RawQuery[string](s.db, "SELECT `string` FROM `simple_struct` WHERE `id` = ?;", 1)
				return queryer.Get(context.Background())
			},
			wantRes: func() *string {
				res := "world"
				return &res
			}(),
		},
		{
			name: "res string  convert bytes",
			queryRes: func() (any, error) {
				queryer := orm.RawQuery[[]byte](s.db, "SELECT `string` FROM `simple_struct` WHERE `id` = ?;", 1)
				return queryer.Get(context.Background())
			},
			wantRes: func() *[]byte {
				res := []byte("world")
				return &res
			}(),
		},
		{
			name: "res bytes",
			queryRes: func() (any, error) {
				queryer := orm.RawQuery[[]byte](s.db, "SELECT `byte_array` FROM `simple_struct` WHERE `id` = ?;", 1)
				return queryer.Get(context.Background())
			},
			wantRes: func() *[]byte {
				res := []byte("hello")
				return &res
			}(),
		},
		{
			name: "res bytes convert string",
			queryRes: func() (any, error) {
				queryer := orm.RawQuery[string](s.db, "SELECT `byte_array` FROM `simple_struct` WHERE `id` = ?;", 1)
				return queryer.Get(context.Background())
			},
			wantRes: func() *string {
				res := "hello"
				return &res
			}(),
		},
		{
			name: "res bool",
			queryRes: func() (any, error) {
				queryer := orm.RawQuery[bool](s.db, "SELECT `bool` FROM `simple_struct` WHERE `id` = ?;", 1)
				return queryer.Get(context.Background())
			},
			wantRes: func() *bool {
				res := true
				return &res
			}(),
		},
		{
			name: "res bool convert string",
			queryRes: func() (any, error) {
				queryer := orm.RawQuery[string](s.db, "SELECT `bool` FROM `simple_struct` WHERE `id` = ?;", 1)
				return queryer.Get(context.Background())
			},
			wantRes: func() *string {
				res := "1"
				return &res
			}(),
		},
		{
			name: "res bool convert in",
			queryRes: func() (any, error) {
				queryer := orm.RawQuery[int](s.db, "SELECT `bool` FROM `simple_struct` WHERE `id` = ?;", 1)
				return queryer.Get(context.Background())
			},
			wantRes: func() *int {
				res := 1
				return &res
			}(),
		},
		{
			name: "res null string ptr",
			queryRes: func() (any, error) {
				queryer := orm.RawQuery[sql.NullString](s.db, "SELECT `null_string_ptr` FROM `simple_struct` WHERE `id` = ?;", 1)
				return queryer.Get(context.Background())
			},
			wantRes: func() *sql.NullString {
				res := sql.NullString{String: "null string", Valid: true}
				return &res
			}(),
		},
		{
			name: "res sring convert null string ptr",
			queryRes: func() (any, error) {
				queryer := orm.RawQuery[sql.NullString](s.db, "SELECT `string` FROM `simple_struct` WHERE `id` = ?;", 1)
				return queryer.Get(context.Background())
			},
			wantRes: func() *sql.NullString {
				res := sql.NullString{String: "world", Valid: true}
				return &res
			}(),
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			res, err := tc.queryRes()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func (s *SelectTestSuite) TestRawQueryGetMulti_baseType() {
	testCases := []struct {
		name     string
		queryRes func() (any, error)
		wantErr  error
		wantRes  any
	}{
		{
			name: "res int",
			queryRes: func() (any, error) {
				queryer := orm.RawQuery[int](s.db, "SELECT `id` FROM `simple_struct`;")
				return queryer.GetMulti(context.Background())
			},
			wantRes: func() (res []*int) {
				vals := []int{1, 2, 3}
				for i := 0; i < len(vals); i++ {
					res = append(res, &vals[i])
				}
				return
			}(),
		},
		{
			name: "res string",
			queryRes: func() (any, error) {
				queryer := orm.RawQuery[string](s.db, "SELECT `string` FROM `simple_struct`;")
				return queryer.GetMulti(context.Background())
			},
			wantRes: func() (res []*string) {
				vals := []string{"world", "world", "world"}
				for i := 0; i < len(vals); i++ {
					res = append(res, &vals[i])
				}
				return
			}(),
		},
		{
			name: "res bytes",
			queryRes: func() (any, error) {
				queryer := orm.RawQuery[[]byte](s.db, "SELECT `byte_array` FROM `simple_struct`;")
				return queryer.GetMulti(context.Background())
			},
			wantRes: func() (res []*[]byte) {
				vals := [][]byte{[]byte("hello"), []byte("hello"), []byte("hello")}
				for i := 0; i < len(vals); i++ {
					res = append(res, &vals[i])
				}
				return
			}(),
		},
		{
			name: "res bool",
			queryRes: func() (any, error) {
				queryer := orm.RawQuery[bool](s.db, "SELECT `bool` FROM `simple_struct`;")
				return queryer.GetMulti(context.Background())
			},
			wantRes: func() (res []*bool) {
				vals := []bool{true, true, true}
				for i := 0; i < len(vals); i++ {
					res = append(res, &vals[i])
				}
				return
			}(),
		},
		{
			name: "res null string ptr",
			queryRes: func() (any, error) {
				queryer := orm.RawQuery[sql.NullString](s.db, "SELECT `null_string_ptr` FROM `simple_struct`;")
				return queryer.GetMulti(context.Background())
			},
			wantRes: []*sql.NullString{
				{
					String: "null string",
					Valid:  true,
				},
				{
					String: "null string",
					Valid:  true,
				},
				{
					String: "null string",
					Valid:  true,
				},
			},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			res, err := tc.queryRes()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.EqualValues(t, tc.wantRes, res)
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
