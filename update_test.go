package orm

import (
	"github.com/stretchr/testify/assert"
	"orm/internal/errs"
	"testing"
)

func TestUpdater_Build(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct {
		name    string
		u       QueryBuilder
		want    *Query
		wantErr error
	}{
		{
			name:    "no columns",
			u:       NewUpdater[TestModel](db),
			wantErr: errs.ErrNoUpdatedColumns,
		},
		{
			name: "single column",
			u: NewUpdater[TestModel](db).Update(&TestModel{
				Age: 18,
			}).Set(C("Age")),
			want: &Query{
				SQL:  "UPDATE `test_model` SET `age` = ?;",
				Args: []any{int8(18)},
			},
		},
		{
			name: "assignment",
			u: NewUpdater[TestModel](db).Update(&TestModel{
				Age:       18,
				FirstName: "Tom",
			}).Set(C("Age"), Assign("FirstName", "DaMing")),
			want: &Query{
				SQL:  "UPDATE `test_model` SET `age` = ?,`first_name` = ?;",
				Args: []any{int8(18), "DaMing"},
			},
		},
		{
			name: "where",
			u: NewUpdater[TestModel](db).Update(&TestModel{
				Age:       18,
				FirstName: "Tom",
			}).Set(C("Age"), Assign("FirstName", "DaMing")).
				Where(C("Id").EQ(1)),
			want: &Query{
				SQL:  "UPDATE `test_model` SET `age` = ?,`first_name` = ? WHERE `id` = ?;",
				Args: []any{int8(18), "DaMing", 1},
			},
		},
		{
			name: "incremental",
			u: NewUpdater[TestModel](db).Update(&TestModel{
				Age:       18,
				FirstName: "Tom",
			}).Set(Assign("Age", C("Age").Add(1))),
			want: &Query{
				SQL:  "UPDATE `test_model` SET `age` = `age` + ?;",
				Args: []any{1},
			},
		},
		{
			name: "incremental-raw",
			u: NewUpdater[TestModel](db).Update(&TestModel{
				Age:       18,
				FirstName: "Tom",
			}).Set(Assign("Age", Raw("`age`+?", 1))),
			want: &Query{
				SQL:  "UPDATE `test_model` SET `age` = `age`+?;",
				Args: []any{1},
			},
		},
		//{
		//	name: "set age=id+(age*100)",
		//	u: NewUpdater[TestModel](db).Update(&TestModel{
		//		Id:        12,
		//		FirstName: "Tom",
		//		Age:       18,
		//		LastName:  &sql.NullString{String: "Jerry", Valid: true},
		//	}).Set(C("FirstName"), Assign("Age", C("Id").Add(C("Age").Multi(100)))),
		//	want: &Query{
		//		// &orm.Query{SQL:"UPDATE `test_model` SET `first_name` = ?,`age` = `id` + (`age` * ?);"
		//		SQL:  "UPDATE `test_model` SET `first_name` = ?,`age` = (`id` + (`age` * ?));",
		//		Args: []interface{}{"Tom", 100},
		//	},
		//},
		//{
		//	name: "set age=(id+(age*100))*110",
		//	u: NewUpdater[TestModel](db).Update(&TestModel{
		//		Id:        12,
		//		FirstName: "Tom",
		//		Age:       18,
		//		LastName:  &sql.NullString{String: "Jerry", Valid: true},
		//	}).Set(C("FirstName"), Assign("Age", C("Id").Add(C("Age").Multi(100)).Multi(110))),
		//	want: &Query{
		//		SQL:  "UPDATE `test_model` SET `first_name` = ?,`age` = ((`id` + (`age` * ?)) * ?);",
		//		Args: []interface{}{"Tom", 100, 110},
		//	},
		//},
		{
			name: "not nil columns",
			u:    NewUpdater[TestModel](db).Update(&TestModel{}).Set(AssignNotNilColumns(&TestModel{Id: 13})...),
			want: &Query{
				SQL:  "UPDATE `test_model` SET `id` = ?,`first_name` = ?,`age` = ?;",
				Args: []interface{}{int64(13), "", int8(0)},
			},
		},
		{
			name: "not zero columns",
			u:    NewUpdater[TestModel](db).Update(&TestModel{}).Set(AssignNotZeroColumns(&TestModel{Id: 13})...),
			want: &Query{
				SQL:  "UPDATE `test_model` SET `id` = ?;",
				Args: []interface{}{int64(13)},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.u.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.want, q)
		})
	}
}
