package generator

import (
	"bytes"
	"go/ast"
	"go/types"
	"sort"
	"strconv"
	"strings"

	"github.com/go-courier/packagesx"
	"github.com/go-courier/reflectx/typesutil"
	"github.com/julienschmidt/httprouter"
)

func NewRouterScanner(pkg *packagesx.Package) *RouterScanner {
	routerScanner := &RouterScanner{
		pkg:     pkg,
		routers: map[*types.Var]*Router{},
	}

	routerScanner.init()

	return routerScanner
}

func (scanner *RouterScanner) init() {
	for _, pkg := range scanner.pkg.AllPackages {
		for ident, obj := range pkg.TypesInfo.Defs {
			if typeVar, ok := obj.(*types.Var); ok {
				if typeVar != nil && !strings.HasSuffix(typeVar.Pkg().Path(), pkgImportPathCourier) {
					if isRouterType(typeVar.Type()) {
						router := NewRouter()

						ast.Inspect(ident.Obj.Decl.(ast.Node), func(node ast.Node) bool {
							switch node.(type) {
							case *ast.CallExpr:
								callExpr := node.(*ast.CallExpr)
								router.AppendOperators(operatorTypeNamesFromArgs(packagesx.NewPackage(pkg), callExpr.Args...)...)
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

	for _, pkg := range scanner.pkg.AllPackages {
		for selectExpr, selection := range pkg.TypesInfo.Selections {
			if selection.Obj() != nil {
				if typeFunc, ok := selection.Obj().(*types.Func); ok {
					recv := typeFunc.Type().(*types.Signature).Recv()
					if recv != nil && isRouterType(recv.Type()) {
						for typeVar, router := range scanner.routers {
							switch selectExpr.Sel.Name {
							case "Register":
								if typeVar == pkg.TypesInfo.ObjectOf(packagesx.GetIdentChainOfCallFunc(selectExpr)[0]) {
									file := scanner.pkg.FileOf(selectExpr)
									ast.Inspect(file, func(node ast.Node) bool {
										switch node.(type) {
										case *ast.CallExpr:
											callExpr := node.(*ast.CallExpr)
											if callExpr.Fun == selectExpr {
												routerIdent := callExpr.Args[0]
												switch routerIdent.(type) {
												case *ast.Ident:
													argTypeVar := pkg.TypesInfo.ObjectOf(routerIdent.(*ast.Ident)).(*types.Var)
													if r, ok := scanner.routers[argTypeVar]; ok {
														router.Register(r)
													}
												case *ast.SelectorExpr:
													argTypeVar := pkg.TypesInfo.ObjectOf(routerIdent.(*ast.SelectorExpr).Sel).(*types.Var)
													if r, ok := scanner.routers[argTypeVar]; ok {
														router.Register(r)
													}
												case *ast.CallExpr:
													callExprForRegister := routerIdent.(*ast.CallExpr)
													router.With(operatorTypeNamesFromArgs(packagesx.NewPackage(pkg), callExprForRegister.Args...)...)
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
	pkg     *packagesx.Package
	routers map[*types.Var]*Router
}

func (scanner *RouterScanner) Router(typeName *types.Var) *Router {
	return scanner.routers[typeName]
}

type OperatorTypeName struct {
	Path string
	*types.TypeName
}

func (operator *OperatorTypeName) SingleStringResultOf(pkg *packagesx.Package, name string) (string, bool) {
	for _, typ := range []types.Type{
		operator.Type(),
		types.NewPointer(operator.Type()),
	} {
		method, ok := typesutil.FromTType(typ).MethodByName(name)
		if ok {
			results, n := pkg.FuncResultsOf(method.(*typesutil.TMethod).Func)
			if n == 1 {
				for _, v := range results[0] {
					if v.Value != nil {
						s, _ := strconv.Unquote(v.Value.String())
						return s, true
					}
				}
			}
		}
	}

	return "", false
}

func operatorTypeNamesFromArgs(pkg *packagesx.Package, args ...ast.Expr) operatorTypeNames {
	opTypeNames := operatorTypeNames{}
	for _, arg := range args {
		opTypeName := operatorTypeNameFromType(pkg.TypesInfo.TypeOf(arg))
		if opTypeName == nil {
			continue
		}

		if callExpr, ok := arg.(*ast.CallExpr); ok {
			if selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if selectorExpr.Sel.Name == "Group" {
					if isGroupFunc(pkg.TypesInfo.ObjectOf(selectorExpr.Sel).Type()) {
						switch v := callExpr.Args[0].(type) {
						case *ast.BasicLit:
							opTypeName.Path, _ = strconv.Unquote(v.Value)
						}
					}
				}
			}
		}

		// handle interface WithMiddleOperators
		method, ok := typesutil.FromTType(opTypeName.Type()).MethodByName("MiddleOperators")
		if ok {
			results, n := pkg.FuncResultsOf(method.(*typesutil.TMethod).Func)
			if n == 1 {
				for _, v := range results[0] {
					if compositeLit, ok := v.Expr.(*ast.CompositeLit); ok {
						ops := operatorTypeNamesFromArgs(pkg, compositeLit.Elts...)
						opTypeNames = append(opTypeNames, ops...)
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

func (router *Router) Route(pkg *packagesx.Package) *Route {
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

	route.SetMethod(pkg)
	route.SetPath(pkg)

	return &route
}

func (router *Router) Routes(pkg *packagesx.Package) (routes []*Route) {
	for child := range router.children {
		route := child.Route(pkg)
		if route.last {
			routes = append(routes, route)
		}
		if child.children != nil {
			routes = append(routes, child.Routes(pkg)...)
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

func (route *Route) SetPath(pkg *packagesx.Package) {
	fullPath := "/"
	for _, operator := range route.Operators {
		if operator.Path != "" {
			fullPath += operator.Path
			continue
		}

		if pathPart, ok := operator.SingleStringResultOf(pkg, "Path"); ok {
			fullPath += pathPart
		}
	}
	route.Path = httprouter.CleanPath(fullPath)
}

func (route *Route) SetMethod(pkg *packagesx.Package) {
	if len(route.Operators) > 0 {
		operator := route.Operators[len(route.Operators)-1]
		route.Method, _ = operator.SingleStringResultOf(pkg, "Method")
	}
}
