package generator

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/types"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/go-courier/oas"
	"github.com/go-courier/packagesx"
)

func NewOpenAPIGenerator(pkg *packagesx.Package) *OpenAPIGenerator {
	return &OpenAPIGenerator{
		pkg:             pkg,
		openapi:         oas.NewOpenAPI(),
		routerScanner:   NewRouterScanner(pkg),
		operatorScanner: NewOperatorScanner(pkg),
	}
}

type OpenAPIGenerator struct {
	pkg             *packagesx.Package
	openapi         *oas.OpenAPI
	routerScanner   *RouterScanner
	operatorScanner *OperatorScanner
}

func rootRouter(pkgInfo *packagesx.Package, callExpr *ast.CallExpr) *types.Var {
	if len(callExpr.Args) > 0 {
		if selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
			if typesFunc, ok := pkgInfo.TypesInfo.ObjectOf(selectorExpr.Sel).(*types.Func); ok {
				if signature, ok := typesFunc.Type().(*types.Signature); ok {
					if isRouterType(signature.Params().At(0).Type()) {
						if selectorExpr.Sel.Name == "Run" || selectorExpr.Sel.Name == "Serve" {
							switch node := callExpr.Args[0].(type) {
							case *ast.SelectorExpr:
								return pkgInfo.TypesInfo.ObjectOf(node.Sel).(*types.Var)
							case *ast.Ident:
								return pkgInfo.TypesInfo.ObjectOf(node).(*types.Var)
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func (g *OpenAPIGenerator) Scan() {
	for ident, def := range g.pkg.TypesInfo.Defs {
		if typFunc, ok := def.(*types.Func); ok {
			if typFunc.Name() != "main" {
				continue
			}

			ast.Inspect(ident.Obj.Decl.(*ast.FuncDecl), func(node ast.Node) bool {
				switch n := node.(type) {
				case *ast.CallExpr:
					if rootRouterVar := rootRouter(g.pkg, n); rootRouterVar != nil {

						router := g.routerScanner.Router(rootRouterVar)
						routes := router.Routes(g.pkg)

						operationIDs := map[string]*Route{}

						for _, route := range routes {
							operation := g.getOperationByOperatorTypes(route.Method, route.Operators...)
							if _, exists := operationIDs[operation.OperationId]; exists {
								panic(fmt.Errorf("operationID %s should be unique", operation.OperationId))
							}
							operationIDs[operation.OperationId] = route
							g.openapi.AddOperation(oas.HttpMethod(strings.ToLower(route.Method)), g.patchPath(route.Path, operation), operation)
						}

						g.operatorScanner.Bind(g.openapi)
					}
				}
				return true
			})
			return
		}
	}
}

var RxHttpRouterPath = regexp.MustCompile("/:([^/]+)")

func (g *OpenAPIGenerator) patchPath(openapiPath string, operation *oas.Operation) string {
	return RxHttpRouterPath.ReplaceAllStringFunc(openapiPath, func(str string) string {
		name := RxHttpRouterPath.FindAllStringSubmatch(str, -1)[0][1]

		var isParameterDefined = false

		for _, parameter := range operation.Parameters {
			if parameter.In == "path" && parameter.Name == name {
				isParameterDefined = true
			}
		}

		if isParameterDefined {
			return "/{" + name + "}"
		}

		return "/0"
	})
}

func (g *OpenAPIGenerator) getOperationByOperatorTypes(method string, operatorTypes ...*OperatorTypeName) *oas.Operation {
	operation := &oas.Operation{}
	length := len(operatorTypes)

	for idx, operatorType := range operatorTypes {
		operator := g.operatorScanner.Operator(operatorType.TypeName)
		operator.BindOperation(method, operation, idx == length-1)
	}

	return operation
}

func (g *OpenAPIGenerator) Output(cwd string) {
	file := filepath.Join(cwd, "openapi.json")
	data, err := json.MarshalIndent(g.openapi, "", "  ")
	if err != nil {
		return
	}
	ioutil.WriteFile(file, data, os.ModePerm)
	log.Printf("generated openapi spec into %s", color.MagentaString(file))
}
