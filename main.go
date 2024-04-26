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
	"path/filepath"
	"regexp"
	"strings"
)

var (
	src           = flag.String("src", "./", "source dir or file")
	exclude       = flag.String("exclude", "", "exclude path names, separated by commas")
	filePatternRE = regexp.MustCompile(`\[[A-Za-z0-9_\-\.\\\/]+\.go\:\d+\]`)
)

func readDir(d string, prefixLen int) {
	if len(d) == 0 {
		log.Printf("empty dir name")
		return
	}
	if !strings.HasPrefix(d, "..") && d[0] == '.' {
		log.Printf("skip hidden dir: %s", d)
		return
	}
	dirs, err := os.ReadDir(d)
	if err != nil {
		cur, _ := os.Getwd()
		log.Printf("read dir %s error, err=%s\n\t%s", d, err, cur)
	}
	//b := path.Base(d)
	for _, item := range dirs {
		if item.IsDir() {
			if _, has := excludePaths[item.Name()]; has {
				continue
			}
			newDir := filepath.Join(d, item.Name())
			//fmt.Printf("  read: %s\n", newDir)
			readDir(newDir, prefixLen)
			continue
		}
		if strings.HasSuffix(item.Name(), ".go") {
			f := filepath.Join(d, item.Name())
			//fmt.Printf("  read: %s\n", f)
			readFile(f, prefixLen)
		}
	}
}

func readFile(f string, prefixLen int) {
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
						replaceTo := fmt.Sprintf("\"[%s:%d]\"", pos.Filename[prefixLen:], pos.Line)
						if lit.Value == replaceTo {
							continue
						}
						isModify = true
						log.Printf("Found a constant string argument:%+v, replace to \"[%s:%d]\"\n", lit.Value, pos.Filename[prefixLen:], pos.Line)
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

var excludePaths = map[string]struct{}{}

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
	if len(*exclude) > 0 {
		arr := strings.Split(*exclude, ",")
		for _, item := range arr {
			excludePaths[strings.Trim(item, " ")] = struct{}{}
		}
	}
	if info.IsDir() {
		a, err := filepath.Abs(*src)
		if err != nil {
			log.Println("get abs path error:", err, *src)
			return
		}
		readDir(a, len(a))
		return
	}
	if strings.HasSuffix(*src, ".go") {
		readFile(*src, 0)
	}
}
