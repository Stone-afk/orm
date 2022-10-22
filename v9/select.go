//go:build v9 
package orm

import (
"context"
"github.com/valyala/bytebufferpool"
"orm/v9/internal/errs"
)

// Selector 用于构造 SELECT 语句
type Selector[T any] struct {
builder
table  string
where  *predicates
having *predicates
db     *DB
// select 查询的列
columns []Selectable
groupBy []*Column
orderBy []*OrderBy
offset  int
limit   int
}

// 定义个新的标记接口，限定传入的类型，这样我 们就可以做各种校验
// 符合的结构体有: Column、Aggregate、RawExpr

type Selectable interface {
selectable()
}

func (s *Selector[T]) Select(cols ...Selectable) *Selector[T] {
s.columns = cols
return s
}

func NewSelector[T any](db *DB) *Selector[T] {
return &Selector[T]{
db: db,
builder: builder{
buffer:   bytebufferpool.Get(),
aliasMap: make(map[string]int, 8),
quoter:   db.dialect.quoter(),
},
}
}

// GroupBy 设置 group by 子句
func (s *Selector[T]) GroupBy(cols ...*Column) *Selector[T] {
s.groupBy = cols
return s
}

func (s *Selector[T]) Having(ps ...*Predicate) *Selector[T] {
s.having = &predicates{
ps:           ps,
useColsAlias: true,
}
return s
}

func (s *Selector[T]) Offset(offset int) *Selector[T] {
s.offset = offset
return s
}

func (s *Selector[T]) Limit(limit int) *Selector[T] {
s.limit = limit
return s
}

func (s *Selector[T]) OrderBy(orderBys ...*OrderBy) *Selector[T] {
s.orderBy = orderBys
return s
}

func (s *Selector[T]) Where(ps ...*Predicate) *Selector[T] {
s.where = &predicates{
ps: ps,
}
return s
}

// From 指定表名，如果是空字符串，那么将会使用默认表名
func (s *Selector[T]) From(tbl string) *Selector[T] {
s.table = tbl
return s
}

func (s *Selector[T]) buildColumns() error {
if len(s.columns) == 0 {
s.writeByte('*')
return nil
}
for i, col := range s.columns {
if i > 0 {
s.writeByte(',')
}
switch val := col.(type) {
case *Column:
if err := s.buildColumn(val, true); err != nil {
return err
}
case *Aggregate:
if err := s.buildAggregate(val, true); err != nil {
return err
}
case *RawExpr:
s.writeString(val.raw)
if len(val.args) > 0 {
s.addArgs(val.args...)
}
default:
return errs.NewErrUnsupportedSelectable(col)
}
}
return nil
}

func (s *Selector[T]) buildGroupBy() error {
for i, col := range s.groupBy {
if i > 0 {
s.writeByte(',')
}
if err := s.buildColumn(col, false); err != nil {
return err
}
}
return nil
}

func (s *Selector[T]) buildOrderBy() error {
for i, od := range s.orderBy {
if i > 0 {
s.writeByte(',')
}
fd, ok := s.model.FieldMap[od.col]
if !ok {
return errs.NewErrUnknownField(od.col)
}
s.writeByte('`')
s.writeString(fd.ColName)
s.writeByte('`')
s.writeString(" " + od.order)
}
return nil
}

func (s *Selector[T]) Build() (*Query, error) {
defer bytebufferpool.Put(s.buffer)
var (
t   T
err error
)
s.model, err = s.db.r.Get(&t)
if err != nil {
return nil, err
}
s.writeString("SELECT ")
if err = s.buildColumns(); err != nil {
return nil, err
}
s.writeString(" FROM ")
if s.table == "" {
s.writeByte('`')
s.writeString(s.model.TableName)
s.writeByte('`')
} else {
s.writeString(s.table)
}
// 构造 WHERE
if s.where != nil && len(s.where.ps) > 0 {
// 类似这种可有可无的部分，都要在前面加一个空格
s.writeString(" WHERE ")
// WHERE 是不允许用别名的
if err = s.buildPredicates(s.where); err != nil {
return nil, err
}
}
if len(s.groupBy) > 0 {
s.writeString(" GROUP BY ")
// GROUP BY 理论上可以用别名，但这里不允许，用户完全可以通过简单的修改代码避免使用别名的这种用法。
// 也不支持复杂的表达式，因为复杂的表达式和 group by 混用是非常罕见的
if err = s.buildGroupBy(); err != nil {
return nil, err
}
}
if s.having != nil && len(s.having.ps) > 0 {
s.writeString(" HAVING ")
// HAVING 是可以用别名的
if err = s.buildPredicates(s.having); err != nil {
return nil, err
}
}
if len(s.orderBy) > 0 {
s.writeString(" ORDER BY ")
if err = s.buildOrderBy(); err != nil {
return nil, err
}
}
if s.limit > 0 {
s.writeString(" LIMIT ?")
s.addArgs(s.limit)
}
if s.offset > 0 {
s.writeString(" OFFSET ?")
s.addArgs(s.offset)
}

s.writeString(";")
return &Query{
SQL:  s.buffer.String(),
Args: s.args,
}, nil
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
q, err := s.Build()
if err != nil {
return nil, err
}
// s.db 是我们定义的 DB
// s.db.db 则是 sql.DB
// 使用 QueryContext，从而和 GetMulti 能够复用处理结果集的代码
rows, err := s.db.db.QueryContext(ctx, q.SQL, q.Args...)
if err != nil {
return nil, err
}

if !rows.Next() {
return nil, ErrNoRows
}

// 有 vals 了，接下来将 vals= [123, "Ming", 18, "Deng"] 反射放回去 t 里面
t := new(T)
// 在这里灵活切换反射或者 unsafe
val := s.db.valCreator(t, s.model)
err = val.SetColumns(rows)
return t, err
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
q, err := s.Build()
if err != nil {
return nil, err
}
rows, err := s.db.db.QueryContext(ctx, q.SQL, q.Args...)
if err != nil {
return nil, err
}

//if !rows.Next() {
//	return nil, ErrNoRows
//}
res := make([]*T, 0, 16)
for rows.Next() {
t := new(T)
// 在这里灵活切换反射或者 unsafe
val := s.db.valCreator(t, s.model)
err = val.SetColumns(rows)
if err != nil {
return nil, err
}
res = append(res, t)
}
return res, nil
}
