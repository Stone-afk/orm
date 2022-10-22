//go:build v9 
package orm

import (
"context"
"github.com/valyala/bytebufferpool"
"orm/v9/internal/errs"
"orm/v9/model"
)

type OnConflictBuilder[T any] struct {
i               *Inserter[T]
conflictColumns []string
}

type OnConflict struct {
assigns         []Assignable
conflictColumns []string
}

type Inserter[T any] struct {
builder
db      *DB
values  []*T
columns []string
// 方案二
onConflict *OnConflict

// 方案一
// onDuplicate []Assignable
}

func NewInserter[T any](db *DB) *Inserter[T] {
return &Inserter[T]{
db: db,
builder: builder{
buffer:   bytebufferpool.Get(),
aliasMap: make(map[string]int, 8),
dialect:  db.dialect,
quoter:   db.dialect.quoter(),
},
}
}

func (i *Inserter[T]) OnConflictKey() *OnConflictBuilder[T] {
return &OnConflictBuilder[T]{
i: i,
}

}

func (i *Inserter[T]) Columns(cols ...string) *Inserter[T] {
i.columns = cols
return i
}

func (i *Inserter[T]) Values(vals ...*T) *Inserter[T] {
i.values = vals
return i
}

func (o *OnConflictBuilder[T]) ConflictColumns(cols ...string) *OnConflictBuilder[T] {
o.conflictColumns = cols
return o
}

func (o *OnConflictBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
o.i.onConflict = &OnConflict{
assigns:         assigns,
conflictColumns: o.conflictColumns,
}
return o.i
}

func (i *Inserter[T]) Build() (*Query, error) {
if len(i.values) == 0 {
return nil, errs.ErrInsertZeroRow
}
defer bytebufferpool.Put(i.buffer)
var (
t   T
err error
)
if i.model == nil {
i.model, err = i.db.r.Get(&t)
if err != nil {
return nil, err
}
}
i.writeString("INSERT INTO ")
i.quote(i.model.TableName)

fields := i.model.Fields
if len(i.columns) > 0 {
fields = make([]*model.Field, 0, len(i.columns))
for _, col := range i.columns {
fd, ok := i.model.FieldMap[col]
if !ok {
return nil, errs.NewErrUnknownField(col)
}
fields = append(fields, fd)
}
}

i.writeLeftParenthesis()
for idx, fd := range fields {
if idx > 0 {
i.writeComma()
}
i.quote(fd.ColName)
}
i.writeRightParenthesis()

i.writeSpace()
i.writeString("VALUES")
i.args = make([]any, 0, len(fields)*len(i.values)+1)
for vIdx, val := range i.values {
if vIdx > 0 {
i.writeComma()
}
refVal := i.db.valCreator(val, i.model)
i.writeLeftParenthesis()
for fIdx, fd := range fields {
if fIdx > 0 {
i.writeComma()
}
i.writePlaceholder()
fdVal, err := refVal.Field(fd.GoName)
if err != nil {
return nil, err
}
i.addArgs(fdVal)
}
i.writeRightParenthesis()
}
if i.onConflict != nil {
err = i.dialect.buildOnConflict(&i.builder, i.onConflict)
if err != nil {
return nil, err
}
}
i.end()
return &Query{
SQL:  i.buffer.String(),
Args: i.args,
}, nil
}

func (i *Inserter[T]) Exec(ctx context.Context) Result {
q, err := i.Build()
if err != nil {
return Result{err: err}
}
res, err := i.db.db.ExecContext(ctx, q.SQL, q.SQL)
return Result{res: res, err: err}
}
