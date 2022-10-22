//go:build v12
package orm

// 后面可以每次支持新的操作符就加一个
const (
	avg   = "AVG"
	sum   = "SUM"
	max   = "MAX"
	min   = "MIN"
	count = "COUNT"
)

type Aggregate struct {
	fn    string
	arg   string
	alias string
}

func (a Aggregate) expr() {}

func (a Aggregate) selectable() {}

func (a Aggregate) As(alias string) Aggregate {
	return Aggregate{
		fn:    a.fn,
		arg:   a.arg,
		alias: alias,
	}
}

// EQ 例如 C("id").Eq(12)
func (a Aggregate) EQ(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opEQ,
		right: exprOf(arg),
	}
}

func (a Aggregate) LT(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opLT,
		right: exprOf(arg),
	}
}

func (a Aggregate) GT(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opGT,
		right: exprOf(arg),
	}
}

func Avg(col string) Aggregate {
	return Aggregate{
		arg: col,
		fn:  avg,
	}
}

func Min(col string) Aggregate {
	return Aggregate{
		arg: col,
		fn:  min,
	}
}

func Max(col string) Aggregate {
	return Aggregate{
		arg: col,
		fn:  max,
	}
}

func Count(col string) Aggregate {
	return Aggregate{
		arg: col,
		fn:  count,
	}
}

func Sum(col string) Aggregate {
	return Aggregate{
		arg: col,
		fn:  sum,
	}
}
