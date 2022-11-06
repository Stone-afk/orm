package model

import (
	"orm/v14/internal/errs"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

// Registry 元数据注册中心的抽象
type Registry interface {
	// Get 查找元数据
	Get(val any) (*Model, error)
	// Register 注册一个模型
	Register(val any, opts ...Option) (*Model, error)
}

type registry struct {
	lock   sync.RWMutex
	models map[reflect.Type]*Model
}

func NewRegistry() Registry {
	return &registry{models: make(map[reflect.Type]*Model, 8)}
}

func (r *registry) Register(val any, opts ...Option) (*Model, error) {
	m, err := r.parseModel(val)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		err = opt(m)
		if err != nil {
			return nil, err
		}
	}

	typ := reflect.TypeOf(val)
	r.models[typ] = m
	return m, nil
}

// Get 查找元数据模型
// 使用读写锁的并发安全解决思路
func (r *registry) Get(val any) (*Model, error) {
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
	return r.Register(val)
}

//// Get 查找元数据模型
//func (r *registry) Get(val any) (*Model, error) {
//	typ := reflect.TypeOf(val)
//	m, ok := r.model.Load(typ)
//	if ok {
//		return m.(*Model), nil
//	}
//	return r.Register(val)
//}
//
//func (r *registry) Register(val any, opts ...ModelOpt) (*Model, error) {
//	m, err := r.parseModel(val)
//	if err != nil {
//		return nil, err
//	}
//
//	for _, opt := range opts {
//		err = opt(m)
//		if err != nil {
//			return nil, err
//		}
//	}
//
//	typ := reflect.TypeOf(val)
//	r.model.Store(typ, m)
//	return m, nil
//}

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

	typ := reflect.TypeOf(val)
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}

	typ = typ.Elem()

	// 获得字段的数量
	numField := typ.NumField()
	fields := make([]*Field, 0, numField)
	fds := make(map[string]*Field, numField)
	colMap := make(map[string]*Field, numField)
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
		fdName := fdType.Name
		fdMeta := &Field{
			ColName: colName,
			Type:    fdType.Type,
			GoName:  fdName,
			Offset:  fdType.Offset,
			Index:   i,
		}
		fds[fdName] = fdMeta
		colMap[colName] = fdMeta
		fields = append(fields, fdMeta)
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
		Type:      typ,
		TableName: tableName,
		FieldMap:  fds,
		ColumnMap: colMap,
		Fields:    fields,
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
