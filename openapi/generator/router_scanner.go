package generator

import (
	"bytes"
	"context"
	"go/ast"
	"go/types"
	"sort"
	"strconv"
	"strings"

	"github.com/go-courier/packagesx"
	typesutil "github.com/go-courier/x/types"
	"github.com/julienschmidt/httprouter"
)

func NewRouterScanner(pkg *packagesx.Package) *RouterScanner {
	routerScanner := &RouterScanner{
		pkg:             pkg,
		routers:         map[*types.Var]*Router{},
		operatorScanner: NewOperatorScanner(pkg),
	}

	routerScanner.init()

	return routerScanner
}

type RouterScanner struct {
	pkg             *packagesx.Package
	routers         map[*types.Var]*Router
	operatorScanner *OperatorScanner
}

func (scanner *RouterScanner) init() {
	for _, pkg := range scanner.pkg.AllPackages {
		for ident, obj := range pkg.TypesInfo.Defs {
			if typeVar, ok := obj.(*types.Var); ok {
				if typeVar != nil && !strings.HasSuffix(typeVar.Pkg().Path(), pkgImportPathCourier) {
					if isRouterType(typeVar.Type()) {
						router := NewRouter(typeVar)

						ast.Inspect(ident.Obj.Decl.(ast.Node), func(node ast.Node) bool {
							switch callExpr := node.(type) {
							case *ast.CallExpr:
								router.AppendOperators(scanner.OperatorTypeNamesFromArgs(packagesx.NewPackage(pkg), callExpr.Args...)...)
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
												switch v := routerIdent.(type) {
												case *ast.Ident:
													argTypeVar := pkg.TypesInfo.ObjectOf(v).(*types.Var)
													if r, ok := scanner.routers[argTypeVar]; ok {
														router.Register(r)
													}
												case *ast.SelectorExpr:
													argTypeVar := pkg.TypesInfo.ObjectOf(v.Sel).(*types.Var)
													if r, ok := scanner.routers[argTypeVar]; ok {
														router.Register(r)
													}
												case *ast.CallExpr:
													router.With(scanner.OperatorTypeNamesFromArgs(packagesx.NewPackage(pkg), v.Args...)...)
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

func (scanner *RouterScanner) Router(typeName *types.Var) *Router {
	return scanner.routers[typeName]
}

type OperatorWithTypeName struct {
	*Operator
	TypeName *types.TypeName
}

func (operator *OperatorWithTypeName) String() string {
	return operator.TypeName.Pkg().Name() + "." + operator.TypeName.Name()
}

func (scanner *RouterScanner) OperatorTypeNamesFromArgs(pkg *packagesx.Package, args ...ast.Expr) []*OperatorWithTypeName {
	opTypeNames := make([]*OperatorWithTypeName, 0)

	for _, arg := range args {
		opTypeName := scanner.OperatorTypeNameFromType(pkg.TypesInfo.TypeOf(arg))

		if opTypeName == nil {
			continue
		}

		// modify meta if httptransport.Group() or httptransport.BasePath()
		if callExpr, ok := arg.(*ast.CallExpr); ok {
			if selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if isFromHttpTransport(pkg.TypesInfo.ObjectOf(selectorExpr.Sel).Type()) {
					switch selectorExpr.Sel.Name {
					case "BasePath":
						switch v := callExpr.Args[0].(type) {
						case *ast.BasicLit:
							opTypeName.BasePath, _ = strconv.Unquote(v.Value)
						}
					case "Group":
						switch v := callExpr.Args[0].(type) {
						case *ast.BasicLit:
							opTypeName.Path, _ = strconv.Unquote(v.Value)
						}
					}
				}
			}
		}

		if opTypeName.TypeName != nil {
			// handle interface WithMiddleOperators
			method, ok := typesutil.FromTType(opTypeName.TypeName.Type()).MethodByName("MiddleOperators")
			if ok {
				results, n := scanner.pkg.FuncResultsOf(method.(*typesutil.TMethod).Func)
				if n == 1 {
					for _, v := range results[0] {
						if compositeLit, ok := v.Expr.(*ast.CompositeLit); ok {
							ops := scanner.OperatorTypeNamesFromArgs(pkg, compositeLit.Elts...)
							opTypeNames = append(opTypeNames, ops...)
						}

					}
				}
			}
		}

		opTypeNames = append(opTypeNames, opTypeName)
	}

	return opTypeNames
}

func (scanner *RouterScanner) OperatorTypeNameFromType(typ types.Type) *OperatorWithTypeName {
	switch t := typ.(type) {
	case *types.Pointer:
		return scanner.OperatorTypeNameFromType(t.Elem())
	case *types.Named:
		typeName := t.Obj()

		if operator := scanner.operatorScanner.Operator(context.Background(), typeName); operator != nil {
			return &OperatorWithTypeName{
				Operator: operator,
				TypeName: typeName,
			}
		}

		return nil
	default:
		return nil
	}
}

func NewRouter(typeVar *types.Var, operators ...*OperatorWithTypeName) *Router {
	return &Router{
		typeVar:   typeVar,
		operators: operators,
	}
}

func (r *Router) Name() string {
	if r.typeVar == nil {
		return "Anonymous"
	}
	return r.typeVar.Pkg().Name() + "." + r.typeVar.Name()
}

func (r *Router) String() string {
	buf := bytes.NewBufferString(r.Name())

	buf.WriteString("<")
	for i := range r.operators {
		o := r.operators[i]
		if i != 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(o.String())
	}
	buf.WriteString(">")

	buf.WriteString("[")

	i := 0
	for sub := range r.children {
		if i != 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(sub.Name())
		i++
	}
	buf.WriteString("]")

	return buf.String()
}

type Router struct {
	typeVar   *types.Var
	parent    *Router
	operators []*OperatorWithTypeName
	children  map[*Router]bool
}

func (router *Router) AppendOperators(operators ...*OperatorWithTypeName) {
	router.operators = append(router.operators, operators...)
}

func (router *Router) With(operators ...*OperatorWithTypeName) {
	router.Register(NewRouter(nil, operators...))
}

func (router *Router) Register(r *Router) {
	if router.children == nil {
		router.children = map[*Router]bool{}
	}
	r.parent = router
	router.children[r] = true
}

func (router *Router) Route() *Route {
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

	return &route
}

func (router *Router) Routes() (routes []*Route) {
	for child := range router.children {
		route := child.Route()

		if route.last {
			routes = append(routes, route)
		}

		if child.children != nil {
			routes = append(routes, child.Routes()...)
		}
	}

	sort.Slice(routes, func(i, j int) bool {
		return routes[i].String() < routes[j].String()
	})

	return routes
}

type Route struct {
	Operators []*OperatorWithTypeName
	last      bool
}

func (route *Route) String() string {
	buf := bytes.NewBufferString(route.Method())
	buf.WriteString(" ")
	buf.WriteString(route.Path())

	for i := range route.Operators {
		buf.WriteString(" ")
		buf.WriteString(route.Operators[i].String())
	}

	return buf.String()
}

func (route *Route) Method() string {
	method := ""
	for _, m := range route.Operators {
		if m.Method != "" {
			method = m.Method
		}
	}
	return method
}

func (route *Route) Path() string {
	basePath := "/"
	fullPath := ""

	for _, operator := range route.Operators {
		if operator.BasePath != "" {
			basePath = operator.BasePath
		}
		if operator.Path != "" {
			fullPath += operator.Path
		}
	}

	return httprouter.CleanPath(basePath + fullPath)
}
