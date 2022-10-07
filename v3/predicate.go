//go:build v3

package orm

// 后面可以每次支持新的操作符就加一个
const (
	opEQ  = "="
	opLT  = "<"
	opGT  = ">"
	opIN  = "IN"
	opAND = "AND"
	opOR  = "OR"
	opNOT = "NOT"
)

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

// Expression 代表语句，或者语句的部分
// 暂时没想好怎么设计方法，所以直接做成标记接口
type Expression interface {
	expr()
}

func exprOf(e any) Expression {
	switch exp := e.(type) {
	case Expression:
		return exp
	default:
		return valueOf(exp)
	}
}

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
