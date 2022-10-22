//go:build v9 
package orm

const (
asc = "ASC"
desc = "DESC"
)

type OrderBy struct {
col   string
order string
}

func Asc(col string) *OrderBy {
return &OrderBy{
col:   col,
order: asc,
}
}

func Desc(col string) *OrderBy {
return &OrderBy{
col:   col,
order: desc,
}
}

type Column struct {
name  string
alias string
}

func (c *Column) assign() {}

func (c *Column) expr() {}

func (c *Column) selectable() {}

func C(name string) *Column {
return &Column{name: name}
}

type value struct {
val any
}

func (v *value) expr() {}

func valueOf(val any) *value {
return &value{
val: val,
}
}

func (c *Column) As(alias string) *Column {
return &Column{
name:  c.name,
alias: alias,
}
}

// EQ 例如 C("id").Eq(12)
func (c *Column) EQ(arg any) *Predicate {
return &Predicate{
left:  c,
op:    opEQ,
right: exprOf(arg),
}
}

func (c *Column) LT(arg any) *Predicate {
return &Predicate{
left:  c,
op:    opLT,
right: exprOf(arg),
}
}

func (c *Column) GT(arg any) *Predicate {
return &Predicate{
left:  c,
op:    opGT,
right: exprOf(arg),
	}
}
