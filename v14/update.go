package orm

import (
	"context"
	"github.com/valyala/bytebufferpool"
	"orm/v14/internal/errs"
	"reflect"
)

type Updater[T any] struct {
	Builder
	core
	sess    session
	where   *predicates
	assigns []Assignable
	table   *T
}

func NewUpdater[T any](sess session) *Updater[T] {
	c := sess.getCore()
	return &Updater[T]{
		core: c,
		sess: sess,
		Builder: Builder{
			buffer:   bytebufferpool.Get(),
			aliasMap: make(map[string]int, 8),
			quoter:   c.dialect.quoter(),
		},
	}
}

func (u *Updater[T]) Set(assigns ...Assignable) *Updater[T] {
	u.assigns = assigns
	return u
}

func (u *Updater[T]) Where(ps ...Predicate) *Updater[T] {
	u.where = &predicates{ps: ps}
	return u
}

func (u *Updater[T]) Update(val *T) *Updater[T] {
	u.table = val
	return u
}

func (u *Updater[T]) Build() (*Query, error) {
	defer bytebufferpool.Put(u.buffer)
	var err error
	if u.model == nil {
		u.model, err = u.r.Get(u.table)
		if err != nil {
			return nil, err
		}
	}
	u.writeString("UPDATE ")
	u.quote(u.model.TableName)
	if len(u.assigns) == 0 {
		return nil, errs.ErrNoUpdatedColumns
	}
	refVal := u.valCreator(u.table, u.model)
	u.writeSpace()
	u.writeString("SET")
	u.writeSpace()
	for i, assign := range u.assigns {
		if i > 0 {
			u.writeComma()
		}
		switch a := assign.(type) {
		case Assignment:
			err = u.buildAssignment(a)
			if err != nil {
				return nil, err
			}
		case Column:
			fd, ok := u.model.FieldMap[a.name]
			if !ok {
				return nil, errs.NewErrUnknownField(a.name)
			}
			u.quote(fd.ColName)
			u.writeString(" = ")
			u.writePlaceholder()

			fdVal, err := refVal.Field(fd.GoName)
			if err != nil {
				return nil, err
			}
			u.addArgs(fdVal)
		default:
			return nil, errs.NewErrUnsupportedAssignableType(assign)
		}
	}
	if u.where != nil && len(u.where.ps) > 0 {
		u.writeString(" WHERE ")
		err = u.buildPredicates(u.where)
		if err != nil {
			return nil, err
		}
	}
	u.end()
	return &Query{SQL: u.buffer.String(), Args: u.args}, nil
}

func AssignNotNilColumns(entity any) []Assignable {
	return AssignColumns(entity,
		func(typ reflect.StructField, val reflect.Value) bool {
			switch val.Kind() {
			case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
				return !val.IsNil()
			}
			return true
		})
}

func AssignNotZeroColumns(entity any) []Assignable {
	return AssignColumns(entity,
		func(typ reflect.StructField, val reflect.Value) bool {
			return !val.IsZero()
		})
}

func AssignColumns(entity any,
	filter func(typ reflect.StructField, val reflect.Value) bool) []Assignable {
	val := reflect.ValueOf(entity).Elem()
	typ := reflect.TypeOf(entity).Elem()
	numField := val.NumField()
	res := make([]Assignable, 0, numField)
	for i := 0; i < numField; i++ {
		fieldVal := val.Field(i)
		fieldTyp := typ.Field(i)
		if filter(fieldTyp, fieldVal) {
			res = append(res, Assign(fieldTyp.Name, fieldVal.Interface()))
		}
	}
	return res
}

func (u *Updater[T]) Exec(ctx context.Context) Result {
	if u.model == nil {
		m, err := u.r.Get(new(T))
		if err != nil {
			return Result{err: err}
		}
		u.model = m
	}
	return exec[T](ctx, u.core, u.sess, &QueryContext{
		Type:    "UPDATE",
		Builder: u,
		Meta:    u.model,
	})
}
