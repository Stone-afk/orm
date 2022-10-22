package orm

type Column struct {
	name  string
	alias string
}

func (c Column) assign() {}

func (c Column) expr() {}

func (c Column) selectable() {}

func C(name string) Column {
	return Column{name: name}
}

func (c Column) Add(val any) MathExpr {
	return MathExpr{
		left:  c,
		op:    opAdd,
		right: valueOf(val),
	}
}

func (c Column) Multi(val any) MathExpr {
	return MathExpr{
		left:  c,
		op:    opMulti,
		right: valueOf(val),
	}
}

type value struct {
	val any
}

func (v value) expr() {}

func valueOf(val any) Expression {
	switch v := val.(type) {
	case Expression:
		return v
	default:
		return value{val: val}
	}
}

func (c Column) As(alias string) Column {
	return Column{
		name:  c.name,
		alias: alias,
	}
}

// EQ 例如 C("id").Eq(12)
func (c Column) EQ(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opEQ,
		right: exprOf(arg),
	}
}

func (c Column) LT(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opLT,
		right: exprOf(arg),
	}
}

func (c Column) GT(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opGT,
		right: exprOf(arg),
	}
}
