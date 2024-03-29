package main

import (
	_ "embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

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
