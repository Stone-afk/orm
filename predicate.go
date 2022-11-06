package orm

type predicates struct {
	ps            []Predicate
	useColsAlias  bool
	useAggreAlias bool
}

// op 代表操作符
type op string

func (o op) String() string {
	return string(o)
}

type Predicate binaryExpr

func (p Predicate) expr() {}

// Not(C("id").Eq(12))
// NOT (id = ?), 12

func Not(p Predicate) Predicate {
	return Predicate{
		op:    opNOT,
		right: p,
	}
}

func Exists(sub Subquery) Predicate {
	return Predicate{
		op:    opExists,
		right: sub,
	}
}

// C("id").Eq(12).And(C("name").Eq("Tom"))

func (p Predicate) And(r Predicate) Predicate {
	return Predicate{
		left:  p,
		op:    opAND,
		right: r,
	}
}

func (p Predicate) Or(r Predicate) Predicate {
	return Predicate{
		left:  p,
		op:    opOR,
		right: r,
	}
}
