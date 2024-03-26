package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

var (
	src           = flag.String("src", "./", "source dir or file")
	filePatternRE = regexp.MustCompile(`\[[A-Za-z0-9_\-\.\\\/]+\.go\:\d+\]`)
)

func readDir(d string) {
	dirs, err := os.ReadDir(d)
	if err != nil {
		log.Printf("read dir %s error, err=%s", d, err)
	}
	b := path.Base(d)
	for _, item := range dirs {
		if item.IsDir() {
			readDir(path.Join(b, item.Name()))
			continue
		}
		if strings.HasSuffix(item.Name(), ".go") {
			f := path.Join(b, item.Name())
			readFile(f)
		}
	}
}

func readFile(f string) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, f, nil, parser.ParseComments)
	if err != nil {
		log.Printf("parse go file %s error, err=%s", f, err)
		return
	}
	isModify := false
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			for idx, arg := range x.Args {
				if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					if len(lit.Value) >= 7 &&
						lit.Value[1] == '[' && lit.Value[len(lit.Value)-2] == ']' &&
						filePatternRE.MatchString(lit.Value) {
						pos := fset.Position(n.Pos())
						replaceTo := fmt.Sprintf("\"[%s:%d]\"", pos.Filename, pos.Line)
						if lit.Value == replaceTo {
							continue
						}
						isModify = true
						log.Printf("Found a constant string argument:%+v, replace to \"[%s:%d]\"\n", lit.Value, pos.Filename, pos.Line)
						newArg := &ast.BasicLit{
							Kind:  token.STRING,
							Value: replaceTo,
						}
						x.Args[idx] = newArg
					}
				}
			}
		}
		return true
	})
	if !isModify {
		return
	}
	var buf bytes.Buffer
	err = printer.Fprint(&buf, fset, file)
	if err != nil {
		log.Printf("format go file %s error, err=%s", f, err)
	}
	if err := os.WriteFile(f, buf.Bytes(), os.ModePerm); err != nil {
		log.Printf("write go file %s error, err=%s", f, err)
	}
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	flag.Parse()

	if len(*src) == 0 {
		*src = "./"
	}
	info, err := os.Stat(*src)
	if err != nil {
		log.Println(err)
	}
	if info.IsDir() {
		readDir(*src)
		return
	}
	if strings.HasSuffix(*src, ".go") {
		readFile(*src)
	}
}
