package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-courier/codegen"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

func main() {
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, path.Join(getPkgDir("net/http"), "method.go"), nil, parser.ParseComments)

	methods := make([]string, 0)

	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.ValueSpec:
			if len(n.Values) == 1 {
				if basicLit, ok := n.Values[0].(*ast.BasicLit); ok {
					if basicLit.Kind == token.STRING {
						name := n.Names[0].Name
						if name[0] != '_' {
							methods = append(methods, n.Names[0].Name)
						}
					}
				}
			}
			return false
		}
		return true
	})

	writeMethods(methods)
}

func writeMethods(methods []string) {
	file := codegen.NewFile("httpx", codegen.GeneratedFileSuffix("./method.go"))

	for _, method := range methods {
		file.WriteBlock(
			file.Expr(`type `+method+` struct {}

func (`+method+`) Method() string {
	return ?
}
`,
				codegen.Id(file.Use("net/http", method)),
			),
		)
	}

	file.WriteFile()

	testFile := codegen.NewFile("httpx", codegen.GeneratedFileSuffix("./method_test.go"))

	for _, method := range methods {
		testFile.WriteBlock(
			testFile.Expr(`func Example` + method + `() {
	m := ` + method + `{}

	` + testFile.Use("fmt", "Println") + `(m.Method())
	// Output:
	// ` + strings.ToUpper(strings.TrimLeft(method, "Method")) + `
}`),
		)

	}

	testFile.WriteFile()
}

func getPkgDir(importPath string) string {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.LoadFiles,
	}, importPath)
	if err != nil {
		panic(err)
	}
	if len(pkgs) == 0 {
		panic(errors.Errorf("package `%s` not found", importPath))
	}
	return filepath.Dir(pkgs[0].GoFiles[0])
}
