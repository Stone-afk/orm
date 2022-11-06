package main

import (
	"fmt"
	"go/ast"
)

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
