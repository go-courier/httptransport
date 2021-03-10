package generator

import (
	"fmt"
	"go/ast"
	"go/types"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/go-courier/packagesx"
	"github.com/go-courier/statuserror"
	"github.com/go-courier/statuserror/generator"
)

func NewStatusErrScanner(pkg *packagesx.Package) *StatusErrScanner {
	statusErrorScanner := &StatusErrScanner{
		pkg:              pkg,
		statusErrorTypes: map[*types.Named][]*statuserror.StatusErr{},
		errorsUsed:       map[*types.Func][]*statuserror.StatusErr{},
	}

	statusErrorScanner.init()

	return statusErrorScanner
}

type StatusErrScanner struct {
	StatusErrType    *types.Named
	pkg              *packagesx.Package
	statusErrorTypes map[*types.Named][]*statuserror.StatusErr
	errorsUsed       map[*types.Func][]*statuserror.StatusErr
}

var statusErrPkgPath = reflect.TypeOf(statuserror.StatusErr{}).PkgPath()

func (scanner *StatusErrScanner) StatusErrorsInFunc(typeFunc *types.Func) []*statuserror.StatusErr {
	if typeFunc == nil {
		return nil
	}

	if statusErrList, ok := scanner.errorsUsed[typeFunc]; ok {
		return statusErrList
	}

	scanner.errorsUsed[typeFunc] = []*statuserror.StatusErr{}

	pkg := packagesx.NewPackage(scanner.pkg.Pkg(typeFunc.Pkg().Path()))

	funcDecl := pkg.FuncDeclOf(typeFunc)

	if funcDecl != nil {
		ast.Inspect(funcDecl, func(node ast.Node) bool {
			switch v := node.(type) {
			case *ast.CallExpr:
				identList := packagesx.GetIdentChainOfCallFunc(v.Fun)
				if len(identList) > 0 {
					callIdent := identList[len(identList)-1]
					obj := pkg.TypesInfo.ObjectOf(callIdent)

					if obj != nil {
						// pick status errors from github.com/go-courier/statuserror.Wrap
						if callIdent.Name == "Wrap" && obj.Pkg().Path() == statusErrPkgPath {

							code := 0
							key := ""
							msg := ""
							desc := make([]string, 0)

							for i, arg := range v.Args[1:] {
								tv, err := pkg.Eval(arg)
								if err != nil {
									continue
								}

								switch i {
								case 0: // code
									code, _ = strconv.Atoi(tv.Value.String())
								case 1: // key
									key, _ = strconv.Unquote(tv.Value.String())
								case 2: // msg
									msg, _ = strconv.Unquote(tv.Value.String())
								default:
									d, _ := strconv.Unquote(tv.Value.String())
									desc = append(desc, d)
								}
							}

							if code > 0 {
								if msg == "" {
									msg = key
								}

								scanner.appendStateErrs(typeFunc, statuserror.Wrap(errors.New(""), code, key, append([]string{msg}, desc...)...))
							}

						}
					}

					// Deprecated old code defined
					if obj != nil && obj.Pkg() != nil && obj.Pkg().Path() == scanner.StatusErrType.Obj().Pkg().Path() {
						for i := range identList {
							scanner.mayAddStateErrorByObject(typeFunc, pkg.TypesInfo.ObjectOf(identList[i]))
						}
						return false
					}

					if nextTypeFunc, ok := obj.(*types.Func); ok && nextTypeFunc != typeFunc && nextTypeFunc.Pkg() != nil {
						scanner.appendStateErrs(typeFunc, scanner.StatusErrorsInFunc(nextTypeFunc)...)
					}
				}
			case *ast.Ident:
				scanner.mayAddStateErrorByObject(typeFunc, pkg.TypesInfo.ObjectOf(v))
			}
			return true
		})

		doc := packagesx.StringifyCommentGroup(funcDecl.Doc)
		scanner.appendStateErrs(typeFunc, pickStatusErrorsFromDoc(doc)...)
	}

	return scanner.errorsUsed[typeFunc]
}

func (scanner *StatusErrScanner) mayAddStateErrorByObject(typeFunc *types.Func, obj types.Object) {
	if obj == nil {
		return
	}
	if typeConst, ok := obj.(*types.Const); ok {
		if named, ok := typeConst.Type().(*types.Named); ok {
			if errs, ok := scanner.statusErrorTypes[named]; ok {
				for i := range errs {
					if errs[i].Key == typeConst.Name() {
						scanner.appendStateErrs(typeFunc, errs[i])
					}
				}
			}
		}
	}
}

func (scanner *StatusErrScanner) appendStateErrs(typeFunc *types.Func, statusErrs ...*statuserror.StatusErr) {
	m := map[string]*statuserror.StatusErr{}

	errs := append(scanner.errorsUsed[typeFunc], statusErrs...)
	for i := range errs {
		s := errs[i]
		m[fmt.Sprintf("%s%d", s.Key, s.Code)] = s
	}

	next := make([]*statuserror.StatusErr, 0)
	for k := range m {
		next = append(next, m[k])
	}

	sort.Slice(next, func(i, j int) bool {
		return next[i].Code < next[j].Code
	})

	scanner.errorsUsed[typeFunc] = next
}

func (scanner *StatusErrScanner) init() {
	pkg := scanner.pkg.Pkg("github.com/go-courier/statuserror")
	if pkg == nil {
		return
	}

	scanner.StatusErrType = packagesx.NewPackage(pkg).TypeName("StatusErr").Type().(*types.Named)
	ttypeStatusError := packagesx.NewPackage(pkg).TypeName("StatusError").Type().Underlying().(*types.Interface)

	isStatusError := func(typ *types.TypeName) bool {
		return types.Implements(typ.Type(), ttypeStatusError)
	}

	s := generator.NewStatusErrorScanner(scanner.pkg)

	for _, pkgInfo := range scanner.pkg.AllPackages {
		for _, obj := range pkgInfo.TypesInfo.Defs {
			if typName, ok := obj.(*types.TypeName); ok {
				if isStatusError(typName) {
					scanner.statusErrorTypes[typName.Type().(*types.Named)] = s.StatusError(typName)
				}
			}
		}
	}
}

func pickStatusErrorsFromDoc(doc string) []*statuserror.StatusErr {
	statusErrorList := make([]*statuserror.StatusErr, 0)

	lines := strings.Split(doc, "\n")

	for _, line := range lines {
		if line != "" {
			if statusErr, err := statuserror.ParseStatusErrSummary(line); err == nil {
				statusErrorList = append(statusErrorList, statusErr)
			}
		}
	}

	return statusErrorList
}
