package orm

// 后面可以每次支持新的操作符就加一个
const (
	opEQ  = "="
	opLT  = "<"
	opGT  = ">"
	opAND = "AND"
	opOR  = "OR"
	opNOT = "NOT"
)

type predicates struct {
	ps            []*Predicate
	useColsAlias  bool
	useAggreAlias bool
}

// op 代表操作符
type op string

func (o op) String() string {
	return string(o)
}

type Predicate struct {
	left  Expression
	op    op
	right Expression
}

func (p *Predicate) expr() {}

// Not(C("id").Eq(12))
// NOT (id = ?), 12

func Not(p *Predicate) *Predicate {
	return &Predicate{
		op:    opNOT,
		right: p,
	}
}

// C("id").Eq(12).And(C("name").Eq("Tom"))

func (p *Predicate) And(r *Predicate) *Predicate {
	return &Predicate{
		left:  p,
		op:    opAND,
		right: r,
	}
}

func (p *Predicate) Or(r *Predicate) *Predicate {
	return &Predicate{
		left:  p,
		op:    opOR,
		right: r,
	}
}
