//go:build v12
package orm

// Creater builds CREATE Query
type Creater[T any] struct {
	builder
	db *DB
}

// Build returns CREATE Query
//func (c *Creater[T]) Build() (*Query, error) {
//	defer bytebufferpool.Put(c.buffer)
//	var (
//		t   T
//		err error
//	)
//	c.writeString("CREATE TABLE ")
//	c.model, err = c.db.r.Get(&t)
//	if err != nil {
//		return nil, err
//	}
//	c.quote(c.model.TableName)
//	c.writeString(" (")
//	for colName, fdMeta := range c.model.ColumnMap {
//		c.quote(colName)
//		c.space()
//		c.writeString(fdMeta.ColumnTyp)
//		if fdMeta.IsNull {
//			c.space()
//			c.writeString("NULL")
//		} else {
//			c.space()
//			c.writeString("NOT NULL")
//		}
//		if fdMeta.IsAutoIncrement {
//			c.space()
//			c.writeString("AUTO_INCREMENT")
//		}
//		if fdMeta.IsPrimaryKey {
//			c.space()
//			c.writeString("AUTO_INCREMENT")
//		}
//		c.comma()
//	}
//	c.writeString(" )")
//	c.end()
//	return &Query{SQL: c.buffer.String()}, nil
//}
