# Protobuf 和代码生成辅助方法

## Protobuf 

 在元数据里面，说过 Protobuf 这种代码生成 的，无法利用 Tag 来指定列名   
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1676013316666-c7502cf8-2e0b-42dc-98da-35ee34f2992d.png#averageHue=%232d2d2c&clientId=ud51674a4-056f-4&from=paste&height=156&id=ucbf91336&name=image.png&originHeight=212&originWidth=791&originalType=binary&ratio=1.5&rotation=0&showTitle=false&size=93221&status=done&style=none&taskId=u53dc6873-0fca-4361-9f46-539c42db426&title=&width=583.3333740234375)
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1676013284880-668023df-f62d-40d0-b723-b0108bf20bc8.png#averageHue=%232e2b2b&clientId=ud51674a4-056f-4&from=paste&height=177&id=u18063aea&name=image.png&originHeight=266&originWidth=877&originalType=binary&ratio=1.5&rotation=0&showTitle=false&size=37231&status=done&style=none&taskId=u9a9e818a-2ea8-4b2a-9760-5ad7f5bb897&title=&width=584.6666666666666)
我们希望能够达成图二这种效果，而不是图一那种。

###  Protobuf 的局限性

 Protobuf 虽然暴露了插件机制，但是插件并不能 修改生成的 Go 代码，插件只能自己额外生成一些 代码。  
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1676014358509-cec664ef-b163-498b-a7b0-9c41f2681fa2.png#averageHue=%23f9f9f9&clientId=ud51674a4-056f-4&from=paste&height=213&id=ubb2719af&name=image.png&originHeight=273&originWidth=717&originalType=binary&ratio=1.5&rotation=0&showTitle=false&size=15760&status=done&style=none&taskId=uf32ef7f9-d3ae-4165-a450-4fca4df1de7&title=&width=560)
 所以实际上不能利用 protobuf 的插件机制

###  修改 protobuf Go 插件  

-  clone 原本的 protobuf  仓库
-  修改 protobuf  仓库里的核心代码
-  安装修改后的 Go 插件  
-  执行 protoc 命令

protobuf 必然会生成 json 的标签， 所以只需要生成找到 json 标签的位置，然后插入我们的代码。  
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1676014989216-d9a57642-264c-4c1e-b6d1-1b80e3528fd0.png#averageHue=%232c2c2b&clientId=ud51674a4-056f-4&from=paste&height=363&id=uc2bf7e98&name=image.png&originHeight=758&originWidth=1309&originalType=binary&ratio=1.5&rotation=0&showTitle=false&size=102421&status=done&style=none&taskId=u0259a515-e49d-4a9c-bca5-61d817c98d6&title=&width=627.6666870117188)
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1676015015895-720987da-6c40-46b5-8054-3bb3d4693bf6.png#averageHue=%232f2e2d&clientId=ud51674a4-056f-4&from=paste&height=235&id=uab3bdf5f&name=image.png&originHeight=284&originWidth=759&originalType=binary&ratio=1.5&rotation=0&showTitle=false&size=108580&status=done&style=none&taskId=u532f13c0-3a3e-45d4-b04d-a3e4b7ac9e1&title=&width=627)
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1676015033750-c78bf61c-1f32-42f1-aa0f-f1da13bf4e2d.png#averageHue=%232f2e2c&clientId=ud51674a4-056f-4&from=paste&height=290&id=u62add16b&name=image.png&originHeight=370&originWidth=801&originalType=binary&ratio=1.5&rotation=0&showTitle=false&size=125507&status=done&style=none&taskId=ue0ac9d6a-2119-4aae-bacf-db742d849fa&title=&width=627)
 这种是侵入式的修改方案，不过我们别无选择。 如果在公司可以维护一个自己定制过的 protobuf Go 插件 仓库

###  例子

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1676015165768-16077730-598b-4e09-9086-204a8bc5961f.png#averageHue=%232c2b2b&clientId=ud51674a4-056f-4&from=paste&height=181&id=udbf5e0e6&name=image.png&originHeight=217&originWidth=631&originalType=binary&ratio=1.5&rotation=0&showTitle=false&size=15425&status=done&style=none&taskId=ufa4fc5a8-a5f1-4c37-a3e4-dc104fc66bc&title=&width=527.6666870117188)
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1676015184988-90462bbe-0b51-4504-bf3d-cab5f6a68b66.png#averageHue=%232f2c2b&clientId=ud51674a4-056f-4&from=paste&height=175&id=u45fc70e0&name=image.png&originHeight=262&originWidth=797&originalType=binary&ratio=1.5&rotation=0&showTitle=false&size=36149&status=done&style=none&taskId=ufa0500c9-7b5d-44a4-96f0-2a28a7cd33e&title=&width=531.3333333333334)
 实际上，最开始考虑过 Google 的 field option 扩展，但是依旧用不了，只能用这种方案。  

## 代码生成辅助方法  

Predicate 设计有一些缺陷，这些缺陷 是可以改进的：  

- 生成字段名的常量 
- 生成 Predicate  

```go
type User struct {
    Name     string
    Age      *int
    NickName *sql.NullString
    Picture  []byte
}

type UserDetail struct {
    Address string
}
```

 **基本上就是 AST + 模板编程  **:

-  AST 读取 Go 源文件内容，解析每个类型及其  
-  生成 import 内容，并且将 orm 依赖添加进去  
-  生成 const  内容
-  生成辅助方法  

### 具体实现

在 gen/orm_gen/ 目录下

```go
package {{ .Package }}

import (
"gitee.com/geektime-geekbang/geektime-go/orm"
{{range $idx, $import := .Imports }}
    {{$import}}
{{end -}}
)
{{- $ops := .Ops -}}
{{range $i, $type := .Types }}

    const (
    {{- range $j, $field := .Fields}}
        {{$type.Name }}{{$field.Name}} = "{{$field.Name}}"
    {{- end}}
    )

    {{range $j, $field := .Fields}}
        {{- range $k, $op := $ops}}
            func {{$type.Name }}{{$field.Name}}{{$op}}(val {{$field.Type}}) orm.Predicate {
            return orm.C("{{$field.Name}}").{{$op}}(val)
            }
        {{end}}
    {{- end}}
{{- end}}
```

```go
type SingleFileEntryVisitor struct {
	file *fileVisitor
}

func (s *SingleFileEntryVisitor) Get() File {
	if s.file != nil {
		return s.file.Get()
	}
	return File{}
}

func (s *SingleFileEntryVisitor) Visit(node ast.Node) ast.Visitor {
	n, ok := node.(*ast.File)
	if ok {
		s.file = &fileVisitor{
			pkg: n.Name.String(),
		}
		return s.file
	}
	return s
}

type fileVisitor struct {
	pkg     string
	imports []string
	types   []*typeVisitor
}

func (f *fileVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.TypeSpec:
		res := &typeVisitor{
			name:   n.Name.String(),
			fields: make([]Field, 0),
		}
		if f.types == nil {
			f.types = make([]*typeVisitor, 0)
		}
		f.types = append(f.types, res)
		return res
	case *ast.ImportSpec:
		path := n.Path.Value
		if n.Name != nil && n.Name.String() != "" {
			path = n.Name.String() + " " + path
		}
		if f.imports == nil {
			f.imports = make([]string, 0)
		}
		f.imports = append(f.imports, path)
	}
	return f
}

func (f *fileVisitor) Get() File {
	types := make([]Type, 0, len(f.types))
	for _, t := range f.types {
		types = append(types, t.Get())
	}
	return File{
		Package: f.pkg,
		Imports: f.imports,
		Types:   types,
	}
}

type typeVisitor struct {
	name   string
	fields []Field
}

func (t *typeVisitor) Visit(node ast.Node) ast.Visitor {
	fd, ok := node.(*ast.Field)
	if ok {
		// 所以实际上我们在这里并没有处理 map, channel 之类的类型
		var typName string
		switch fdType := fd.Type.(type) {
		case *ast.Ident:
			typName = fdType.String()
		case *ast.StarExpr:
			switch expr := fdType.X.(type) {
			case *ast.Ident:
				typName = fmt.Sprintf("*%s", expr.String())
			case *ast.SelectorExpr:
				x := expr.X.(*ast.Ident).String()
				name := expr.Sel.String()
				typName = fmt.Sprintf("*%s.%s", x, name)
			}
		case *ast.SelectorExpr:
			x := fdType.X.(*ast.Ident).String()
			name := fdType.Sel.String()
			typName = fmt.Sprintf("%s.%s", x, name)
		case *ast.ArrayType:
			// 其它类型我们都不能处理处理，本来在 ORM 框架里面我们也没有支持
			switch ele := fdType.Elt.(type) {
			case *ast.Ident:
				typName = fmt.Sprintf("[]%s", ele)
			}
		}
		t.fields = append(t.fields, Field{
			Type: typName,
			Name: fd.Names[0].String(),
		})
		return nil
	}
	return t
}

func (t *typeVisitor) Get() Type {
	return Type{
		Name:   t.name,
		Fields: t.fields,
	}
}

type File struct {
	Package string
	Imports []string
	Types   []Type
}

type Type struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name string
	Type string
}

```

```go
func main() {
	// 用户必须输入一个 src，限制为文件
	// 然后我们会在同目录下生成代码
	src := os.Args[1]
	// Dir返回路径的除最后一个元素之外的所有元素，通常是路径的目录。
	dstDir := filepath.Dir(src)
	// Base返回路径的最后一个元素
	fileName := filepath.Base(src)
	// LastIndexByte返回s中c的最后一个实例的索引，如果s中不存在c，则返回-1。
	idx := strings.LastIndexByte(fileName, '.')
	dst := filepath.Join(dstDir, fileName[:idx]+"_gen.go")
	f, err := os.Create(dst)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = gen(f, src)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("生成成功")

}

// Go 会读取 tpl.gohtml 里面的内容填充到变量 tpl 里面
//go:embed tpl.gohtml
var genOrm string

type OrmFile struct {
	File
	Ops []string
}

func gen(writer io.Writer, srcFile string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, srcFile, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	tv := &SingleFileEntryVisitor{}
	ast.Walk(tv, f)
	file := tv.Get()

	tpl := template.New("gen_orm")
	tpl, err = tpl.Parse(genOrm)
	if err != nil {
		return err
	}
	return tpl.Execute(writer, OrmFile{
		File: file,
		Ops:  []string{"LT", "GT", "EQ"},
	})
}

```

### 单元测试

```go
func TestFileVisitor_Get(t *testing.T) {
	testCases := []struct {
		src  string
		want File
	}{
		{
			src: `
package orm_gen
import (
	"fmt"
    "database/sql"
) 

import (
	dri "database/sql/driver"
)
type (
	StructType struct {
		// Public is a field
		// @type string
		Public string
        Ptr *sql.NullString
		Struct sql.NullInt64
		Age *int8
		Slice []byte
	}
)
`,
			want: File{
				Package: "orm_gen",
				Imports: []string{`"fmt"`, `"database/sql"`, `dri "database/sql/driver"`},
				Types: []Type{
					{
						Name: "StructType",
						Fields: []Field{
							{
								Name: "Public",
								Type: "string",
							},
							{
								Name: "Ptr",
								Type: "*sql.NullString",
							},
							{
								Name: "Struct",
								Type: "sql.NullInt64",
							},
							{
								Name: "Age",
								Type: "*int8",
							},
							{
								Name: "Slice",
								Type: "[]byte",
							},
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "src.go", tc.src, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		tv := &SingleFileEntryVisitor{}
		ast.Walk(tv, f)
		file := tv.Get()
		assert.Equal(t, tc.want, file)
	}
}
```

```go
func TestGen(t *testing.T) {
	bs := &bytes.Buffer{}
	err := gen(bs, "testdata/user.go")
	require.NoError(t, err)
	assert.Equal(t, `package testdata

import (
    "gitee.com/geektime-geekbang/geektime-go/orm"

    "database/sql"
)

const (
    UserName = "Name"
    UserAge = "Age"
    UserNickName = "NickName"
    UserPicture = "Picture"
)


func UserNameLT(val string) orm.Predicate {
    return orm.C("Name").LT(val)
}

func UserNameGT(val string) orm.Predicate {
    return orm.C("Name").GT(val)
}

func UserNameEQ(val string) orm.Predicate {
    return orm.C("Name").EQ(val)
}

func UserAgeLT(val *int) orm.Predicate {
    return orm.C("Age").LT(val)
}

func UserAgeGT(val *int) orm.Predicate {
    return orm.C("Age").GT(val)
}

func UserAgeEQ(val *int) orm.Predicate {
    return orm.C("Age").EQ(val)
}

func UserNickNameLT(val *sql.NullString) orm.Predicate {
    return orm.C("NickName").LT(val)
}

func UserNickNameGT(val *sql.NullString) orm.Predicate {
    return orm.C("NickName").GT(val)
}

func UserNickNameEQ(val *sql.NullString) orm.Predicate {
    return orm.C("NickName").EQ(val)
}

func UserPictureLT(val []byte) orm.Predicate {
    return orm.C("Picture").LT(val)
}

func UserPictureGT(val []byte) orm.Predicate {
    return orm.C("Picture").GT(val)
}

func UserPictureEQ(val []byte) orm.Predicate {
    return orm.C("Picture").EQ(val)
}


const (
    UserDetailAddress = "Address"
)


func UserDetailAddressLT(val string) orm.Predicate {
    return orm.C("Address").LT(val)
}

func UserDetailAddressGT(val string) orm.Predicate {
    return orm.C("Address").GT(val)
}

func UserDetailAddressEQ(val string) orm.Predicate {
    return orm.C("Address").EQ(val)
}
`, bs.String())
}

```

## 总结

-  **怎么修改 protobuf 生成的代码**？生成好的代码从实践上来说是不应该修改的，而很不幸的 是，protobuf 对应不同插件生成的目标代码，只能改插件的源码  ；
-  **怎么为 protobuf 的字段添加额外的属性**？可以通过 field option 来增加额外的属性，但是 这种新的属性需要你自己写代码解析  ；
-  **代码生成的常用场景**？一般来说，样板代码都可以考虑使用代码生成来替换掉，比如典型 的利用代码生成来生成数据库查询（如ENT），生成增删改查的代码，生成前端代码  ；