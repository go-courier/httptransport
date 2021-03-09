package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"net/url"
	"path"
	"path/filepath"
	"strconv"

	"github.com/go-courier/codegen"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

func main() {
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, path.Join(getPkgDir("net/http"), "status.go"), nil, parser.ParseComments)

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
		Response: &Response{
			Location: u,
		},
	}
}

type `+statusKey+` struct {
	*Response
}

func (`+statusKey+`) StatusCode() int {
	return ?
}

func (r `+statusKey+`) Location() *? {
	return r.Response.Location
}
`,
				codegen.Id(file.Use("net/url", "URL")),
				codegen.Id(file.Use("net/http", statusKey)),
				codegen.Id(file.Use("net/url", "URL")),
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
	// Output:
	// ?
	// /test
}`, &url.URL{
				Path: "/test",
			}, redirectStatusCodes[i]),
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
