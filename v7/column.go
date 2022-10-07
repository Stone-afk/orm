//go:build v7

package orm

type Column struct {
	name string
}

func (c *Column) expr() {}

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
