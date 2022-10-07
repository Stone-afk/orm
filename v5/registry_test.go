//go:build v5

package orm

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"orm/internal/errs"
	"reflect"
	"testing"
)

func TestRegistry_get(t *testing.T) {
	testCases := []struct {
		name      string
		val       any
		wantModel *Model
		wantErr   error
	}{
		{
			name: "test model",
			val:  &TestModel{},
			wantModel: &Model{
				tableName: "test_model",
				fieldMap: map[string]*Field{
					"Id": {
						colName: "id",
					},
					"FirstName": {
						colName: "first_name",
					},
					"Age": {
						colName: "age",
					},
					"LastName": {
						colName: "last_name",
					},
				},
			},
		},
		{
			// 多级指针
			name: "multiple pointer",
			// 因为 Go 编译器的原因，所以我们写成这样
			val: func() any {
				val := &TestModel{}
				return &val
			}(),
			wantErr: errors.New("orm: 只支持一级指针作为输入，例如 &User{}"),
		},
		{
			name:    "map",
			val:     map[string]string{},
			wantErr: errors.New("orm: 只支持一级指针作为输入，例如 &User{}"),
		},
		{
			name:    "slice",
			val:     []int{},
			wantErr: errors.New("orm: 只支持一级指针作为输入，例如 &User{}"),
		},
		{
			name:    "basic type",
			val:     0,
			wantErr: errors.New("orm: 只支持一级指针作为输入，例如 &User{}"),
		},
		// 标签相关测试用例
		{
			name: "column tag",
			val: func() any {
				// 把测试结构体定义在方法内部，防止被其它用例访问
				type ColumnTag struct {
					ID uint64 `orm:"column=id"`
				}
				return &ColumnTag{}
			}(),
			wantModel: &Model{
				tableName: "column_tag",
				fieldMap: map[string]*Field{
					"ID": {
						colName: "id",
					},
				},
			},
		},
		{
			// 如果用户设置了 column，但是传入一个空字符串，那么会用默认的名字
			name: "empty column",
			val: func() any {
				// 把测试结构体定义在方法内部，防止被其它用例访问
				type EmptyColumn struct {
					FirstName uint64 `orm:"column="`
				}
				return &EmptyColumn{}
			}(),
			wantModel: &Model{
				tableName: "empty_column",
				fieldMap: map[string]*Field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
		},
		{
			// 如果用户设置了 column，但是没有赋值
			name: "invalid tag",
			val: func() any {
				// 把测试结构体定义在方法内部，防止被其它用例访问
				type InvalidTag struct {
					FirstName uint64 `orm:"column"`
				}
				return &InvalidTag{}
			}(),
			wantErr: errs.NewErrInvalidTagContent("column"),
		},
		{
			// 如果用户设置了一些奇奇怪怪的内容，这部分内容会忽略掉
			name: "ignore tag",
			val: func() any {
				// 把测试结构体定义在方法内部，防止被其它用例访问
				type IgnoreTag struct {
					FirstName uint64 `orm:"abc=abc"`
				}
				return &IgnoreTag{}
			}(),
			wantModel: &Model{
				tableName: "ignore_tag",
				fieldMap: map[string]*Field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
		},
		// 利用接口自定义模型信息
		{
			name: "table name",
			val:  &CustomTableName{},
			wantModel: &Model{
				tableName: "custom_table_name_t",
				fieldMap: map[string]*Field{
					"Name": {
						colName: "name",
					},
				},
			},
		},
		{
			name: "table name ptr",
			val:  &CustomTableNamePtr{},
			wantModel: &Model{
				tableName: "custom_table_name_ptr_t",
				fieldMap: map[string]*Field{
					"Name": {
						colName: "name",
					},
				},
			},
		},
		{
			name: "empty table name",
			val:  &EmptyTableName{},
			wantModel: &Model{
				tableName: "empty_table_name",
				fieldMap: map[string]*Field{
					"Name": {
						colName: "name",
					},
				},
			},
		},
	}

	r := &registry{
		models: make(map[reflect.Type]*Model, 16),
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.get(tc.val)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantModel, m)
		})
	}

}

type CustomTableName struct {
	Name string
}

func (c CustomTableName) TableName() string {
	return "custom_table_name_t"
}

type CustomTableNamePtr struct {
	Name string
}

func (c *CustomTableNamePtr) TableName() string {
	return "custom_table_name_ptr_t"
}

type EmptyTableName struct {
	Name string
}

func (c *EmptyTableName) TableName() string {
	return ""
}

//func Test_underscoreName(t *testing.T) {
//	testCases := []struct {
//		name    string
//		srcStr  string
//		wantStr string
//	}{
//		// 我们这些用例就是为了确保
//		// 在忘记 underscoreName 的行为特性之后
//		// 可以从这里找回来
//		// 比如说过了一段时间之后
//		// 忘记了 ID 不能转化为 id
//		// 那么这个测试能帮我们确定 ID 只能转化为 i_d
//		{
//			name:    "upper cases",
//			srcStr:  "ID",
//			wantStr: "i_d",
//		},
//		{
//			name:    "use number",
//			srcStr:  "Table1Name",
//			wantStr: "table1_name",
//		},
//	}
//	for _, tc := range testCases {
//		t.Run(tc.name, func(t *testing.T) {
//			res := underscoreName(tc.srcStr)
//			assert.Equal(t, tc.wantStr, res)
//		})
//	}
//}
