package generator

import (
	"fmt"
	"go/ast"
	"go/types"
	"sort"
	"strings"

	"github.com/go-courier/loaderx"
	"github.com/go-courier/statuserror"
	"github.com/go-courier/statuserror/generator"
	"golang.org/x/tools/go/loader"
)

func NewStatusErrScanner(program *loader.Program) *StatusErrScanner {
	statusErrorScanner := &StatusErrScanner{
		program:          program,
		statusErrorTypes: map[*types.Named][]*statuserror.StatusErr{},
		errorsUsed:       map[*types.Func][]*statuserror.StatusErr{},
	}

	statusErrorScanner.init()

	return statusErrorScanner
}

type StatusErrScanner struct {
	StatusErrType    *types.Named
	program          *loader.Program
	statusErrorTypes map[*types.Named][]*statuserror.StatusErr
	errorsUsed       map[*types.Func][]*statuserror.StatusErr
}

func (scanner *StatusErrScanner) StatusErrorsInFunc(typeFunc *types.Func) []*statuserror.StatusErr {
	if typeFunc == nil {
		return nil
	}

	if statusErrList, ok := scanner.errorsUsed[typeFunc]; ok {
		return statusErrList
	}

	scanner.errorsUsed[typeFunc] = []*statuserror.StatusErr{}

	pkgInfo := scanner.program.Package(typeFunc.Pkg().Path())
	p := loaderx.NewProgram(scanner.program)

	funcDecl := p.FuncDeclOf(typeFunc)

	if funcDecl != nil {
		ast.Inspect(funcDecl, func(node ast.Node) bool {
			switch node.(type) {
			case *ast.CallExpr:
				identList := loaderx.GetIdentChainOfCallFunc(node.(*ast.CallExpr).Fun)
				if len(identList) > 0 {
					callIdent := identList[len(identList)-1]
					obj := pkgInfo.ObjectOf(callIdent)

					if nextTypeFunc, ok := obj.(*types.Func); ok && nextTypeFunc != typeFunc && nextTypeFunc.Pkg() != nil {
						scanner.appendStateErrs(typeFunc, scanner.StatusErrorsInFunc(nextTypeFunc)...)
					}
				}
			case *ast.Ident:
				scanner.mayAddStateErrorByObject(typeFunc, pkgInfo.ObjectOf(node.(*ast.Ident)))
			}
			return true
		})

		doc := loaderx.StringifyCommentGroup(funcDecl.Doc)
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
	pkg := loaderx.NewProgram(scanner.program).Package("github.com/go-courier/statuserror")
	if pkg == nil {
		return
	}

	scanner.StatusErrType = loaderx.NewPackageInfo(pkg).TypeName("StatusErr").Type().(*types.Named)
	ttypeStatusError := loaderx.NewPackageInfo(pkg).TypeName("StatusError").Type().Underlying().(*types.Interface)

	isStatusError := func(typ *types.TypeName) bool {
		return types.Implements(typ.Type(), ttypeStatusError)
	}

	s := generator.NewStatusErrorScanner(scanner.program)

	for _, pkgInfo := range scanner.program.AllPackages {
		for _, obj := range pkgInfo.Defs {
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
