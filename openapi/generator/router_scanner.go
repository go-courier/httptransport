package generator

import (
	"bytes"
	"go/ast"
	"go/types"
	"sort"
	"strconv"
	"strings"

	"github.com/go-courier/loaderx"
	"github.com/go-courier/reflectx/typesutil"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/tools/go/loader"
)

func NewRouterScanner(program *loader.Program) *RouterScanner {
	routerScanner := &RouterScanner{
		program: program,
		routers: map[*types.Var]*Router{},
	}

	routerScanner.init()

	return routerScanner
}

func (scanner *RouterScanner) init() {
	for _, pkgInfo := range scanner.program.AllPackages {
		for ident, obj := range pkgInfo.Defs {
			if typeVar, ok := obj.(*types.Var); ok {
				if typeVar != nil && !strings.HasSuffix(typeVar.Pkg().Path(), pkgImportPathCourier) {
					if isRouterType(typeVar.Type()) {
						router := NewRouter()

						ast.Inspect(ident.Obj.Decl.(ast.Node), func(node ast.Node) bool {
							switch node.(type) {
							case *ast.CallExpr:
								callExpr := node.(*ast.CallExpr)
								router.AppendOperators(operatorTypeNamesFromArgs(pkgInfo, callExpr.Args...)...)
								return false
							}
							return true
						})

						scanner.routers[typeVar] = router
					}
				}
			}
		}
	}

	for _, pkgInfo := range scanner.program.AllPackages {
		for selectExpr, selection := range pkgInfo.Selections {
			if selection.Obj() != nil {
				if typeFunc, ok := selection.Obj().(*types.Func); ok {
					recv := typeFunc.Type().(*types.Signature).Recv()
					if recv != nil && isRouterType(recv.Type()) {
						for typeVar, router := range scanner.routers {
							switch selectExpr.Sel.Name {
							case "Register":
								if typeVar == pkgInfo.ObjectOf(loaderx.GetIdentChainOfCallFunc(selectExpr)[0]) {
									file := loaderx.NewProgram(scanner.program).FileOf(selectExpr)
									ast.Inspect(file, func(node ast.Node) bool {
										switch node.(type) {
										case *ast.CallExpr:
											callExpr := node.(*ast.CallExpr)
											if callExpr.Fun == selectExpr {
												routerIdent := callExpr.Args[0]
												switch routerIdent.(type) {
												case *ast.Ident:
													argTypeVar := pkgInfo.ObjectOf(routerIdent.(*ast.Ident)).(*types.Var)
													if r, ok := scanner.routers[argTypeVar]; ok {
														router.Register(r)
													}
												case *ast.SelectorExpr:
													argTypeVar := pkgInfo.ObjectOf(routerIdent.(*ast.SelectorExpr).Sel).(*types.Var)
													if r, ok := scanner.routers[argTypeVar]; ok {
														router.Register(r)
													}
												case *ast.CallExpr:
													callExprForRegister := routerIdent.(*ast.CallExpr)
													router.With(operatorTypeNamesFromArgs(pkgInfo, callExprForRegister.Args...)...)
												}
												return false
											}
										}
										return true
									})
								}
							}
						}
					}
				}
			}
		}
	}
}

type RouterScanner struct {
	program *loader.Program
	routers map[*types.Var]*Router
}

func (scanner *RouterScanner) Router(typeName *types.Var) *Router {
	return scanner.routers[typeName]
}

type OperatorTypeName struct {
	Path string
	*types.TypeName
}

func (operator *OperatorTypeName) SingleStringResultOf(program *loader.Program, name string) (string, bool) {
	method, ok := typesutil.FromTType(operator.Type()).MethodByName(name)
	if ok {
		results, n := loaderx.NewProgram(program).FuncResultsOf(method.(*typesutil.TMethod).Func)
		if n == 1 {
			for _, v := range results[0] {
				if v.Value != nil {
					s, _ := strconv.Unquote(v.Value.String())
					return s, true
				}
			}
		}
	}
	return "", false
}

func operatorTypeNamesFromArgs(pkgInfo *loader.PackageInfo, args ...ast.Expr) operatorTypeNames {
	opTypeNames := operatorTypeNames{}
	for _, arg := range args {
		opTypeName := operatorTypeNameFromType(pkgInfo.TypeOf(arg))
		if opTypeName == nil {
			continue
		}
		if callExpr, ok := arg.(*ast.CallExpr); ok {
			if selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if selectorExpr.Sel.Name == "Group" {
					if isGroupFunc(pkgInfo.ObjectOf(selectorExpr.Sel).Type()) {
						switch v := callExpr.Args[0].(type) {
						case *ast.BasicLit:
							opTypeName.Path, _ = strconv.Unquote(v.Value)
						}
					}
				}
			}
		}
		opTypeNames = append(opTypeNames, opTypeName)

	}
	return opTypeNames
}

type operatorTypeNames []*OperatorTypeName

func (names operatorTypeNames) String() string {
	buf := bytes.NewBuffer(nil)
	for i, name := range names {
		if i > 0 {
			buf.WriteRune(' ')
		}
		buf.WriteString(name.Pkg().Name() + "." + name.Name())
	}
	return buf.String()
}

func operatorTypeNameFromType(typ types.Type) *OperatorTypeName {
	switch typ.(type) {
	case *types.Pointer:
		return operatorTypeNameFromType(typ.(*types.Pointer).Elem())
	case *types.Named:
		return &OperatorTypeName{
			TypeName: typ.(*types.Named).Obj(),
		}
	default:
		return nil
	}
}

func NewRouter(operators ...*OperatorTypeName) *Router {
	return &Router{
		operators: operators,
	}
}

type Router struct {
	parent    *Router
	operators []*OperatorTypeName
	children  map[*Router]bool
}

func (router *Router) AppendOperators(operators ...*OperatorTypeName) {
	router.operators = append(router.operators, operators...)
}

func (router *Router) With(operators ...*OperatorTypeName) {
	router.Register(NewRouter(operators...))
}

func (router *Router) Register(r *Router) {
	if router.children == nil {
		router.children = map[*Router]bool{}
	}
	r.parent = router
	router.children[r] = true
}

func (router *Router) Route(program *loader.Program) *Route {
	parent := router.parent
	operators := router.operators

	for parent != nil {
		operators = append(parent.operators, operators...)
		parent = parent.parent
	}

	route := Route{
		last:      router.children == nil,
		Operators: operators,
	}

	route.SetMethod(program)
	route.SetPath(program)

	return &route
}

func (router *Router) Routes(program *loader.Program) (routes []*Route) {
	for child := range router.children {
		route := child.Route(program)
		if route.last {
			routes = append(routes, route)
		}
		if child.children != nil {
			routes = append(routes, child.Routes(program)...)
		}
	}

	sort.Slice(routes, func(i, j int) bool {
		return routes[i].String() < routes[j].String()
	})

	return routes
}

type Route struct {
	Method    string
	Path      string
	Operators operatorTypeNames
	last      bool
}

func (route *Route) String() string {
	return route.Method + " " + route.Path + " " + route.Operators.String()
}

func (route *Route) SetPath(program *loader.Program) {
	fullPath := "/"
	for _, operator := range route.Operators {
		if operator.Path != "" {
			fullPath += operator.Path
			continue
		}

		if pathPart, ok := operator.SingleStringResultOf(program, "Path"); ok {
			fullPath += pathPart
		}
	}
	route.Path = httprouter.CleanPath(fullPath)
}

func (route *Route) SetMethod(program *loader.Program) {
	if len(route.Operators) > 0 {
		operator := route.Operators[len(route.Operators)-1]
		route.Method, _ = operator.SingleStringResultOf(program, "Method")
	}
}
