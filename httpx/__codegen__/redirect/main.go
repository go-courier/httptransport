package main

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"net/url"
	"path"
	"strconv"

	"github.com/go-courier/codegen"
)

func main() {
	pkg, _ := build.Default.Import("net/http", "", build.FindOnly)

	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, path.Join(pkg.Dir, "status.go"), nil, parser.ParseComments)

	redirectStatuses := make([]string, 0)
	redirectStatusCodes := make([]int, 0)

	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.ValueSpec:
			if len(n.Values) == 1 {
				if basicLit, ok := n.Values[0].(*ast.BasicLit); ok {
					if basicLit.Kind == token.INT {
						if basicLit.Value[0] == '3' {
							name := n.Names[0].Name
							if name[0] != '_' {
								redirectStatuses = append(redirectStatuses, n.Names[0].Name)
								redirectStatusCode, _ := strconv.ParseInt(basicLit.Value, 10, 64)
								redirectStatusCodes = append(redirectStatusCodes, int(redirectStatusCode))
							}
						}
					}
				}
			}
			return false
		}
		return true
	})

	writeFuncs(redirectStatuses, redirectStatusCodes)
}

func writeFuncs(redirectStatuses []string, redirectStatusCodes []int) {
	file := codegen.NewFile("httpx", codegen.GeneratedFileSuffix("./redirect.go"))

	for _, statusKey := range redirectStatuses {
		file.WriteBlock(
			file.Expr(`
func RedirectWith`+statusKey+`(u *?) *`+statusKey+` {
	return &`+statusKey+`{
		redirect: &redirect{
			URL: u,
		},
	}
}

type `+statusKey+` struct {
	*redirect
}

func (`+statusKey+`) StatusCode() int {
	return ?
}`,
				codegen.Id(file.Use("net/url", "URL")),
				codegen.Id(file.Use("net/http", statusKey)),
			),
		)
	}

	file.WriteFile()

	testFile := codegen.NewFile("httpx", codegen.GeneratedFileSuffix("./redirect_test.go"))

	for i, statusKey := range redirectStatuses {
		testFile.WriteBlock(
			testFile.Expr(`func Example`+statusKey+`() {
	m := RedirectWith`+statusKey+`(?)

	`+testFile.Use("fmt", "Println")+`(m.StatusCode())
	`+testFile.Use("fmt", "Println")+`(m.Location())
	`+testFile.Use("fmt", "Println")+`(m.Error())
	// Output:
	// ?
	// /test
	// Location: /test
}`, &url.URL{
				Path: "/test",
			}, redirectStatusCodes[i]),
		)

	}

	testFile.WriteFile()
}
