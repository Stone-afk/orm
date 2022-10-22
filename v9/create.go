// Copyright 2021 gotomicro
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build v9

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
