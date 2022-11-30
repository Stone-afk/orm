package orm

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// TODO 待修改
func TestRawQuerier_GetMulti(t *testing.T) {
	mockDB, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = mockDB.Close() }()
	db, err := OpenDB("mysql", mockDB)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name      string
		queryRes  func(t *testing.T) any
		mockErr   error
		mockOrder func(mock sqlmock.Sqlmock)
		wantErr   error
		wantVal   any
	}{
		// 返回原生基本类型
		// int
		{
			name: "res int",
			queryRes: func(t *testing.T) any {
				queryer := RawQuery[int](db, "SELECT `age` FROM `test_model`;")
				result, err := queryer.GetMulti(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"age"}).AddRow(10).
					AddRow(18).AddRow(22)
				mock.ExpectQuery("SELECT `age` FROM `test_model`;").
					WillReturnRows(rows)
			},
			wantVal: func() (res []*int) {
				vals := []int{10, 18, 22}
				for i := 0; i < len(vals); i++ {
					res = append(res, &vals[i])
				}
				return
			}(),
		},
		// byte
		{
			name: "res byte",
			queryRes: func(t *testing.T) any {
				queryer := RawQuery[byte](db, "SELECT `first_name` FROM `test_model`;")
				result, err := queryer.GetMulti(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"first_name"}).AddRow('D').AddRow('a')
				mock.ExpectQuery("SELECT `first_name` FROM `test_model`;").
					WillReturnRows(rows)
			},
			wantVal: func() (res []*byte) {
				vals := []byte{'D', 'a'}
				for i := 0; i < len(vals); i++ {
					res = append(res, &vals[i])
				}
				return
			}(),
		},
		// bytes
		{
			name: "res bytes",
			queryRes: func(t *testing.T) any {
				queryer := RawQuery[[]byte](db, "SELECT `first_name` FROM `test_model`;")
				result, err := queryer.GetMulti(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"first_name"}).AddRow([]byte("Li")).AddRow([]byte("Liu"))
				mock.ExpectQuery("SELECT `first_name` FROM `test_model`;").
					WillReturnRows(rows)
			},
			wantVal: func() (res []*[]byte) {
				vals := [][]byte{[]byte("Li"), []byte("Liu")}
				for i := 0; i < len(vals); i++ {
					res = append(res, &vals[i])
				}
				return
			}(),
		},
		// string
		{
			name: "res string",
			queryRes: func(t *testing.T) any {
				queryer := RawQuery[string](db, "SELECT `first_name` FROM `test_model`;")
				result, err := queryer.GetMulti(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"first_name"}).AddRow("Da").AddRow("Li")
				mock.ExpectQuery("SELECT `first_name` FROM `test_model`;").
					WillReturnRows(rows)
			},
			wantVal: func() (res []*string) {
				vals := []string{"Da", "Li"}
				for i := 0; i < len(vals); i++ {
					res = append(res, &vals[i])
				}
				return
			}(),
		},
		// struct ptr
		{
			name: "res struct ptr",
			queryRes: func(t *testing.T) any {
				queryer := RawQuery[TestModel](db, "SELECT `first_name`,`age` FROM `test_model`;")
				result, err := queryer.GetMulti(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"first_name", "age"}).
					AddRow("Da", 18).AddRow("Xiao", 16)
				mock.ExpectQuery("SELECT `first_name`,`age` FROM `test_model`;").
					WillReturnRows(rows)
			},
			wantVal: []*TestModel{
				{
					FirstName: "Da",
					Age:       18,
				},
				{
					FirstName: "Xiao",
					Age:       16,
				},
			},
		},
		//// sql.NullString
		{
			name: "res sql.NullString",
			queryRes: func(t *testing.T) any {
				queryer := RawQuery[sql.NullString](db, "SELECT `last_name` FROM `test_model`;")
				result, err := queryer.GetMulti(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"last_name"}).
					AddRow([]byte("ming")).AddRow([]byte("gang"))
				mock.ExpectQuery("SELECT `last_name` FROM `test_model`;").
					WillReturnRows(rows)
			},
			wantVal: []*sql.NullString{
				{
					String: "ming",
					Valid:  true,
				},
				{
					String: "gang",
					Valid:  true,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockOrder(mock)
			res := tc.queryRes(t)
			assert.EqualValues(t, tc.wantVal, res)
		})
	}
}

func TestRawQuerier_Get(t *testing.T) {
	//mockDB, mock, err := sqlmock.New()
	mockDB, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = mockDB.Close() }()
	db, err := OpenDB("mysql", mockDB)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name      string
		queryRes  func(t *testing.T) any
		mockErr   error
		mockOrder func(mock sqlmock.Sqlmock)
		wantErr   error
		wantVal   any
	}{
		//返回原生基本类型
		{
			name: "res RawQuery int",
			queryRes: func(t *testing.T) any {
				queryer := RawQuery[int](db, "SELECT `age` FROM `test_model` LIMIT ?;", 1)
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"age"}).AddRow(10)
				mock.ExpectQuery("SELECT `age` FROM `test_model` LIMIT ?;").
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantVal: func() *int {
				val := 10
				return &val
			}(),
		},
		{
			name: "res RawQuery bytes",
			queryRes: func(t *testing.T) any {
				queryer := RawQuery[[]byte](db, "SELECT `first_name` FROM `test_model` WHERE `id`=? LIMIT ?;", 1, 1)
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"first_name"}).AddRow([]byte("Li"))
				mock.ExpectQuery("SELECT `first_name` FROM `test_model` WHERE `id`=? LIMIT ?;").
					WithArgs(1, 1).
					WillReturnRows(rows)
			},
			wantVal: func() *[]byte {
				val := []byte("Li")
				return &val
			}(),
		},
		{
			name: "res RawQuery string",
			queryRes: func(t *testing.T) any {
				queryer := RawQuery[string](db, "SELECT `first_name` FROM `test_model` WHERE `id`=? LIMIT ?;", 1, 1)
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"first_name"}).AddRow("Da")
				mock.ExpectQuery("SELECT `first_name` FROM `test_model` WHERE `id`=? LIMIT ?;").
					WithArgs(1, 1).
					WillReturnRows(rows)
			},
			wantVal: func() *string {
				val := "Da"
				return &val
			}(),
		},
		{
			name: "res RawQuery struct ptr",
			queryRes: func(t *testing.T) any {
				queryer := RawQuery[TestModel](db, "SELECT `first_name`,`age` FROM `test_model` WHERE `id`=? LIMIT ?;", 1, 1)
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"first_name", "age"}).AddRow("Da", 18)
				mock.ExpectQuery("SELECT `first_name`,`age` FROM `test_model` WHERE `id`=? LIMIT ?;").
					WithArgs(1, 1).
					WillReturnRows(rows)
			},
			wantVal: func() *TestModel {
				return &TestModel{
					FirstName: "Da",
					Age:       18,
				}
			}(),
		},
		{
			name: "res RawQuery sql.NullString",
			queryRes: func(t *testing.T) any {
				queryer := RawQuery[sql.NullString](db, "SELECT `last_name` FROM `test_model` WHERE `id`=? LIMIT ?;", 1, 1)
				result, err := queryer.Get(context.Background())
				require.NoError(t, err)
				return result
			},
			mockOrder: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"last_name"}).AddRow([]byte("ming"))
				mock.ExpectQuery("SELECT `last_name` FROM `test_model` WHERE `id`=? LIMIT ?;").
					WithArgs(1, 1).
					WillReturnRows(rows)
			},
			wantVal: func() *sql.NullString {
				return &sql.NullString{String: "ming", Valid: true}
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockOrder(mock)
			res := tc.queryRes(t)
			assert.Equal(t, tc.wantVal, res)
		})
	}
}
