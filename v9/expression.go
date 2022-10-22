//go:build v9 
package orm

// RawExpr 代表一个原生表达式
// 意味着 ORM 不会对它进行任何处理
type RawExpr struct {
raw  string
args []any
}

func (r *RawExpr) selectable() {}

func (r *RawExpr) expr() {}

func (r *RawExpr) AsPredicate() *Predicate {
return &Predicate{
left: r,
}
}

// Raw 创建一个 RawExpr
func Raw(expr string, args ...any) *RawExpr {
return &RawExpr{
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

func exprOf(e any) Expression {
switch exp := e.(type) {
case Expression:
return exp
default:
return valueOf(exp)
	}
}
