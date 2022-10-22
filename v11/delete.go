//go:build v11
package orm

import "github.com/valyala/bytebufferpool"

type Deleter[T any] struct {
	builder
	core
	sess  session
	table string
	where *predicates
}

func NewDeleter[T any](sess session) *Deleter[T] {
	c := sess.getCore()
	return &Deleter[T]{
		core: c,
		sess: sess,
		builder: builder{
			buffer:   bytebufferpool.Get(),
			aliasMap: make(map[string]int, 8),
			quoter:   c.dialect.quoter(),
		},
	}
}

func (d *Deleter[T]) Where(ps ...*Predicate) *Deleter[T] {
	d.where = &predicates{
		ps: ps,
	}
	return d
}

// From 指定表名，如果是空字符串，那么将会使用默认表名
func (d *Deleter[T]) From(tbl string) *Deleter[T] {
	d.table = tbl
	return d
}

func (d *Deleter[T]) Build() (*Query, error) {
	defer bytebufferpool.Put(d.buffer)
	var (
		t   T
		err error
	)
	d.model, err = d.r.Get(&t)
	if err != nil {
		return nil, err
	}
	d.writeString("DELETE FROM ")
	if d.table == "" {
		d.writeByte('`')
		d.writeString(d.model.TableName)
		d.writeByte('`')
	} else {
		d.writeString(d.table)
	}

	// 构造 WHERE
	if d.where != nil && len(d.where.ps) > 0 {
		// 类似这种可有可无的部分，都要在前面加一个空格
		d.writeString(" WHERE ")
		// WHERE 是不允许用别名的
		if err = d.buildPredicates(d.where); err != nil {
			return nil, err
		}
	}
	d.writeString(";")
	return &Query{
		SQL:  d.buffer.String(),
		Args: d.args,
	}, nil
}
