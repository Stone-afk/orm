//go:build v5

package orm

import (
	"orm/internal/errs"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

// 这种包变量对测试不友好，缺乏隔离
// var defaultRegistry = &registry{
// 	model: make(map[reflect.Type]*model, 16),
// }

type registry struct {
	// model key 是类型名
	// 这种定义方式是不行的
	// 1. 类型名冲突，例如都是 User，但是一个映射过去 buyer_t
	// 一个映射过去 seller_t
	// 2. 并发不安全
	// model map[string]*model

	lock   sync.RWMutex
	models map[reflect.Type]*Model

	//// 使用 sync.Map
	//model sync.Map
}

// 使用 sync.Map
//func (r *registry) get(val any) (*Model, error) {
//	typ := reflect.TypeOf(val)
//	m, ok := r.model.Load(typ)
//	if !ok {
//		var err error
//		if m, err = r.parseModel(typ); err != nil {
//			return nil, err
//		}
//	}
//	r.model.Store(typ, m)
//	return m.(*Model), nil
//}

// 直接 map
// func (r *registry) get(val any) (*model, error) {
// 	typ := reflect.TypeOf(val)
// 	m, ok := r.model[typ]
// 	if !ok {
// 		var err error
// 		if m, err = r.parseModel(typ); err != nil {
// 			return nil, err
// 		}
// 	}
// 	r.model[typ] = m
// 	return m, nil
// }

// 使用读写锁的并发安全解决思路
func (r *registry) get(val any) (*Model, error) {
	if val == nil {
		return nil, errs.ErrInputNil
	}
	r.lock.RLock()
	typ := reflect.TypeOf(val)
	m, ok := r.models[typ]
	r.lock.RUnlock()
	if ok {
		return m, nil
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	m, ok = r.models[typ]
	if ok {
		return m, nil
	}
	var err error
	if m, err = r.parseModel(val); err != nil {
		return nil, err
	}
	r.models[typ] = m
	return m, nil
}

func (r *registry) parseTag(tag reflect.StructTag) (map[string]string, error) {
	ormTag := tag.Get("orm")
	if ormTag == "" {
		// 返回一个空的 map，这样调用者就不需要判断 nil 了
		return map[string]string{}, nil
	}
	// 这个初始化容量就是支持的 key 的数量，
	// 现在只有一个，所以我们初始化为 1
	res := make(map[string]string, 1)

	// 接下来就是字符串处理了
	pairs := strings.Split(ormTag, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return nil, errs.NewErrInvalidTagContent(pair)
		}
		res[kv[0]] = kv[1]
	}
	return res, nil
}

func (r *registry) parseModel(val any) (*Model, error) {

	//ptrTyp := reflect.TypeOf(val)
	//typ := ptrTyp.Elem()
	//if ptrTyp.Kind() != reflect.Ptr || typ.Kind() != reflect.Struct {
	//	return nil, errs.ErrPointerOnly
	//}

	//ptrTyp := reflect.TypeOf(val)
	//if ptrTyp.Kind() != reflect.Ptr || ptrTyp.Elem().Kind() != reflect.Struct {
	//	return nil, errs.ErrPointerOnly
	//}
	//typ := ptrTyp.Elem()

	typ := reflect.TypeOf(val)

	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}

	typ = typ.Elem()

	// 获得字段的数量
	numField := typ.NumField()
	fds := make(map[string]*Field, numField)
	for i := 0; i < numField; i++ {
		fdType := typ.Field(i)
		tags, err := r.parseTag(fdType.Tag)
		if err != nil {
			return nil, err
		}
		colName := tags[tagKeyColumn]
		if colName == "" {
			colName = underscoreName(fdType.Name)
		}
		fds[fdType.Name] = &Field{
			colName: colName,
		}
	}
	var tableName string
	if tn, ok := val.(TableName); ok {
		tableName = tn.TableName()
	}
	// 注意: 极端情况在 tn.TableName() 中表名有可能被自定义为 "", 这时将使用默认格式的表名
	if tableName == "" {
		tableName = underscoreName(typ.Name())
	}

	return &Model{
		tableName: tableName,
		fieldMap:  fds,
	}, nil
}

// underscoreName 驼峰转字符串命名
func underscoreName(tableName string) string {
	var buf []byte
	for i, v := range tableName {
		if unicode.IsUpper(v) {
			if i != 0 {
				buf = append(buf, '_')
			}
			buf = append(buf, byte(unicode.ToLower(v)))
		} else {
			buf = append(buf, byte(v))
		}
	}
	return string(buf)
}
