# DELETE Statement

DELETE 应该是增删改查里最简单的语句了

## 语法分析

由于不同类型 db 包括 SQLite3、 MySQL、PostgreSQL 的 DELETE 最简单形态的语法是一样的，本文只以实现最简单的形态为目标，所以，这里只拿 MySQL 举例
MYSQL 的 DELETE 语句也有两种形态:

- 删除单表的：  额外支持了 ORDER BY 和 LIMIT

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675844492707-389a0c6b-15f1-4f26-be47-c21139046e00.png#averageHue=%23f6f6f5&clientId=ude9730c4-2ae3-4&from=paste&height=128&id=u1ccaeb0a&name=image.png&originHeight=192&originWidth=1072&originalType=binary&ratio=1&rotation=0&showTitle=false&size=33785&status=done&style=none&taskId=uf3dd4f08-2115-4dd5-9ea4-779b477c067&title=&width=714.6666666666666)

- 删除多表的：只支持 WHERE 条件  

  ![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675844574692-63d1f379-1e60-47af-8c65-6b66638bbe59.png#averageHue=%23f6f6f5&clientId=ude9730c4-2ae3-4&from=paste&height=194&id=u0eb6f349&name=image.png&originHeight=291&originWidth=1083&originalType=binary&ratio=1&rotation=0&showTitle=false&size=50445&status=done&style=none&taskId=ub191939d-50d2-477e-a5b3-b1349a56543&title=&width=722)

## 开源实例

### Beego ORM

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675844813298-3d5db3d4-be6d-4317-b476-fcf7979d2cb2.png#averageHue=%23312d2c&clientId=ude9730c4-2ae3-4&from=paste&height=393&id=ubacf6f52&name=image.png&originHeight=590&originWidth=1208&originalType=binary&ratio=1&rotation=0&showTitle=false&size=112828&status=done&style=none&taskId=u135e881e-d8ed-4a11-9cde-6c5137a8a22&title=&width=805.3333333333334)
Beego 的 DELETE API 定义和 UPDATE 一样，如果 cols 没有传，默认是根据主键进行删除。

### GORM

![](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675845363413-491ceb95-7c9d-46e8-a1c8-5294f0e503fe.png#averageHue=%232d2c2c&from=url&id=GO8Wc&originHeight=396&originWidth=1457&originalType=binary&ratio=1&rotation=0&showTitle=false&status=done&style=none&title=)
DELETE 相关方法只有一个， 具体实现思路是，删除一条记录时，删除对象需要指定主键，否则会触发批量Delete，例如：

```go
// Email 的 ID 是 `10`
db.Delete(&email)
// DELETE from emails where id = 10;

// 带额外条件的删除
db.Where("name = ?", "jinzhu").Delete(&email)
// DELETE from emails where id = 10 AND name = "jinzhu";

db.Delete(&users, []int{1,2,3})
// DELETE FROM users WHERE id IN (1,2,3);

```

## API 设计

与 Update 类似， 需要结构体 Deletor 结构体去实现 QueryBuilder 与 Executor 接口。以及需要包含条件语句的拼接。

```go
type Deleter[T any] struct {
}

func (d *Deleter[T]) Build() (*Query, error) {
	panic("implement me")
}

// From accepts model definition
func (d *Deleter[T]) From(table string) *Deleter[T] {
	panic("implement me ")
}

// Where accepts predicates
func (d *Deleter[T]) Where(predicates ...Predicate) *Deleter[T] {
	panic("implement me")
}
```

## 具体实现

Build 方法

```go
// Deleter builds DELETE query
type Deleter[T any] struct {
	builder
	table string
	where []Predicate
}

// Build returns DELETE query
func (d *Deleter[T]) Build() (*Query, error) {
	_, _ = d.sb.WriteString("DELETE FROM ")

	if d.table == "" {
		var t T
		d.sb.WriteByte('`')
		d.sb.WriteString(reflect.TypeOf(t).Name())
		d.sb.WriteByte('`')
	} else {
		d.sb.WriteString(d.table)
	}
	if len(d.where) > 0 {
		d.sb.WriteString(" WHERE ")
		err := d.buildPredicates(d.where)
		if err != nil {
			return nil, err
		}
	}
	d.sb.WriteByte(';')
	return &Query{SQL: d.sb.String(), Args: d.args}, nil
}

// From accepts model definition
func (d *Deleter[T]) From(table string) *Deleter[T] {
	d.table = table
	return d
}

// Where accepts predicates
func (d *Deleter[T]) Where(predicates ...Predicate) *Deleter[T] {
	d.where = predicates
	return d
}

```

与 Update 一样需要实现一个 Excute方法

```go
func (d *Deleter[T]) Exec(ctx context.Context) Result {
	q, err := d.Build()
	if err != nil {
		return Result{err: err}
	}
	res, err := d.db.db.ExecContext(ctx, q.SQL, q.Args...)
	return Result{err: err, res: res}
}

```

## 单元测试

```go
func TestDeleter_Build(t *testing.T) {
	testCases := []struct {
		name      string
		builder   QueryBuilder
		wantErr   error
		wantQuery *Query
	}{
		{
			name:    "no where",
			builder: (&Deleter[TestModel]{}).From("`test_model`"),
			wantQuery: &Query{
				SQL: "DELETE FROM `test_model`;",
			},
		},
		{
			name:    "where",
			builder: (&Deleter[TestModel]{}).Where(C("Id").EQ(16)),
			wantQuery: &Query{
				SQL: "DELETE FROM `TestModel` WHERE `Id` = ?;",
				Args: []any{16},
			},
		},
		{
			name:    "from",
			builder: (&Deleter[TestModel]{}).From("`test_model`").Where(C("Id").EQ(16)),
			wantQuery: &Query{
				SQL: "DELETE FROM `test_model` WHERE `Id` = ?;",
				Args: []any{16},
			},
		},
	}

	for _, tc := range testCases {
		c := tc
		t.Run(c.name, func(t *testing.T) {
			query, err := c.builder.Build()
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
}
```

只要了解前面几个模块的设计，就会发现最简单的 delete 的构造基本没有难点，也没啥好总结的。