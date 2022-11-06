package orm

// 后面可以每次支持新的操作符就加一个
const (
	opEQ     = "="
	opLT     = "<"
	opGT     = ">"
	opAND    = "AND"
	opOR     = "OR"
	opNOT    = "NOT"
	opAdd    = "+"
	opMulti  = "*"
	opIN     = "IN"
	opExists = "EXIST"
	preALL   = "ALL"
	preAny   = "ANY"
	preSome  = "SOME"
)

type MathExpr binaryExpr

func (m MathExpr) expr() {}

func (m MathExpr) Add(val any) MathExpr {
	return MathExpr{
		left:  m,
		op:    opAdd,
		right: valueOf(val),
	}
}

func (m MathExpr) Multi(val any) MathExpr {
	return MathExpr{
		left:  m,
		op:    opMulti,
		right: valueOf(val),
	}
}

// RawExpr 代表一个原生表达式
// 意味着 ORM 不会对它进行任何处理
type RawExpr struct {
	raw  string
	args []any
}

//func (r RawExpr) selectable() {}

func (r RawExpr) selectedAlias() string {
	return ""
}

func (r RawExpr) fieldName() string {
	return ""
}

func (r RawExpr) target() TableReference {
	return nil
}

func (r RawExpr) expr() {}

func (r RawExpr) AsPredicate() Predicate {
	return Predicate{
		left: r,
	}
}

// Raw 创建一个 RawExpr
func Raw(expr string, args ...any) RawExpr {
	return RawExpr{
		raw:  expr,
		args: args,
	}
}

// Expression 代表语句，或者语句的部分
// 暂时没想好怎么设计方法，所以直接做成标记接口
//目前该接口的结构体有 Column、value、 Predicate、RawExpr
type Expression interface {
	expr()
}

type binaryExpr struct {
	left  Expression
	op    op
	right Expression
}

func (b binaryExpr) expr() {}

func exprOf(e any) Expression {
	switch exp := e.(type) {
	case Expression:
		return exp
	default:
		return valueOf(exp)
	}
}

// SubqueryExpr 注意，这个谓词这种不是在所有的数据库里面都支持的
// 这里采取的是和 Upsert 不同的做法
// Upsert 里面我们是属于用 dialect 来区别不同的实现
// 这里我们采用另外一种方案，就是直接生成，依赖于数据库来报错
// 实际中两种方案你可以自由替换
type SubqueryExpr struct {
	s Subquery
	// 谓词，ALL，ANY 或者 SOME
	pred string
}

func (s SubqueryExpr) expr() {}

func All(sub Subquery) SubqueryExpr {
	return SubqueryExpr{
		s:    sub,
		pred: preALL,
	}
}

func Any(sub Subquery) SubqueryExpr {
	return SubqueryExpr{
		s:    sub,
		pred: preAny,
	}
}

func Some(sub Subquery) SubqueryExpr {
	return SubqueryExpr{
		s:    sub,
		pred: preSome,
	}
}
