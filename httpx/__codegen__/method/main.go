package main

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path"
	"strings"

	"github.com/go-courier/codegen"
)

func main() {
	pkg, _ := build.Default.Import("net/http", "", build.FindOnly)

	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, path.Join(pkg.Dir, "method.go"), nil, parser.ParseComments)

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
