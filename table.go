package orm

type TableReference interface {
	tableAlias() string
}

var _ TableReference = Table{}
var _ TableReference = Join{}
var _ TableReference = Subquery{}
var _ TableReference = &Union{}

type Table struct {
	entity any
	alias  string
}

func (t Table) tableAlias() string {
	return t.alias
}

func TableOf(entity any) Table {
	return Table{
		entity: entity,
	}
}

func (t Table) C(name string) Column {
	return Column{
		name:  name,
		table: t,
	}
}

func (t Table) As(alias string) Table {
	return Table{
		entity: t.entity,
		alias:  alias,
	}
}

func (t Table) Join(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: target,
		typ:   "JOIN",
	}
}

func (t Table) LeftJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: target,
		typ:   "LEFT JOIN",
	}
}

func (t Table) RightJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: target,
		typ:   "RIGHT JOIN",
	}
}

type JoinBuilder struct {
	left  TableReference
	right TableReference
	typ   string
}

func (j *JoinBuilder) On(ps ...Predicate) Join {
	pres := &predicates{ps: ps}
	return Join{
		typ:   j.typ,
		on:    pres,
		left:  j.left,
		right: j.right,
	}
}

func (j *JoinBuilder) Using(cs ...string) Join {
	return Join{
		typ:   j.typ,
		using: cs,
		left:  j.left,
		right: j.right,
	}
}

type Join struct {
	typ   string
	on    *predicates
	left  TableReference
	right TableReference
	using []string
}

func (j Join) tableAlias() string {
	return ""
}

func (j Join) Join(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: target,
		typ:   "JOIN",
	}
}

func (j Join) LeftJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: target,
		typ:   "LEFT JOIN",
	}
}

func (j Join) RightJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: target,
		typ:   "RIGHT JOIN",
	}
}

type Subquery struct {
	// 使用 QueryBuilder 仅仅是为了让 Subquery 可以是非泛型的。
	s     QueryBuilder
	table TableReference
	// select 查询的列
	columns []Selectable
	alias   string
}

func (s Subquery) expr() {}

func (s Subquery) tableAlias() string {
	return s.alias
}

func (s Subquery) C(name string) Column {
	return Column{table: s, name: name}
}

func (s Subquery) Join(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  s,
		right: target,
		typ:   "JOIN",
	}
}

func (s Subquery) LeftJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  s,
		right: target,
		typ:   "LEFT JOIN",
	}
}

func (s Subquery) RightJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  s,
		right: target,
		typ:   "RIGHT JOIN",
	}
}
