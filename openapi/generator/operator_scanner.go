package generator

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	"net/http"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"

	"github.com/go-courier/httptransport/httpx"
	"github.com/go-courier/httptransport/transformers"
	"github.com/go-courier/loaderx"
	"github.com/go-courier/oas"
	"github.com/go-courier/reflectx/typesutil"
	"github.com/go-courier/statuserror"
	"github.com/sirupsen/logrus"
	"golang.org/x/tools/go/loader"
)

func NewOperatorScanner(program *loader.Program, pkgInfo *loader.PackageInfo) *OperatorScanner {
	return &OperatorScanner{
		DefinitionScanner: NewDefinitionScanner(program, pkgInfo),
		program:           loaderx.NewProgram(program),
		pkgInfo:           pkgInfo,
		StatusErrScanner:  NewStatusErrScanner(program),
	}
}

type OperatorScanner struct {
	*DefinitionScanner
	*StatusErrScanner
	program   *loaderx.Program
	pkgInfo   *loader.PackageInfo
	operators map[*types.TypeName]*Operator
}

func (scanner *OperatorScanner) tagFrom(pkgPath string) string {
	tag := strings.TrimPrefix(pkgPath, scanner.pkgInfo.Pkg.Path())
	return strings.TrimPrefix(tag, "/")
}

func (scanner *OperatorScanner) Operator(typeName *types.TypeName) *Operator {
	if typeName == nil {
		return nil
	}

	if operator, ok := scanner.operators[typeName]; ok {
		return operator
	}

	defer func() {
		if e := recover(); e != nil {
			panic(fmt.Errorf("scan Operator `%s` failed, panic: %s; calltrace: %s", fullTypeName(typeName), fmt.Sprint(e), string(debug.Stack())))
		}
	}()

	if typeStruct, ok := typeName.Type().Underlying().(*types.Struct); ok {
		operator := &Operator{
			ID:  typeName.Name(),
			Tag: scanner.tagFrom(typeName.Pkg().Path()),
		}

		scanner.scanParameterOrRequestBody(operator, typeStruct)
		scanner.scanSummaryAndDescription(operator, typeName)
		scanner.scanReturns(operator, typeName)

		if scanner.operators == nil {
			scanner.operators = map[*types.TypeName]*Operator{}
		}

		scanner.operators[typeName] = operator

		return operator
	}

	return nil
}

func (scanner *OperatorScanner) scanSummaryAndDescription(op *Operator, typeName *types.TypeName) {
	comments := strings.Split(scanner.program.CommentsOf(scanner.program.IdentOf(typeName)), "\n")

	if comments[0] != "" {
		op.Summary = comments[0]
		if len(comments) > 1 {
			op.Description = strings.Join(comments[1:], "\n")
		}
	}
}

func (scanner *OperatorScanner) scanReturns(op *Operator, typeName *types.TypeName) {
	method, ok := typesutil.FromTType(typeName.Type()).MethodByName("Output")
	if ok {
		results, n := scanner.program.FuncResultsOf(method.(*typesutil.TMethod).Func)
		if n == 2 {
			for _, v := range results[0] {
				if v.Type != nil {
					if v.Type.String() != types.Typ[types.UntypedNil].String() {
						if op.SuccessType != nil && op.SuccessType.String() != v.Type.String() {
							logrus.Warnf(fmt.Sprintf("%s success result must be same struct, but got %v, already set %v", op.ID, v.Type, op.SuccessType))
						}
						op.SuccessType = v.Type
						op.SuccessStatus, op.SuccessResponse = scanner.getResponse(v.Type, v.Expr)
					}
				}
			}
		}

		if scanner.StatusErrScanner.StatusErrType != nil {
			op.StatusErrors = scanner.StatusErrScanner.StatusErrorsInFunc(method.(*typesutil.TMethod).Func)
			op.StatusErrorSchema = scanner.DefinitionScanner.getSchemaByType(scanner.StatusErrScanner.StatusErrType)
		}
	}
}

func (scanner *OperatorScanner) firstValueOfFunc(named *types.Named, name string) (interface{}, bool) {
	method, ok := typesutil.FromTType(named).MethodByName(name)
	if ok {
		results, n := scanner.program.FuncResultsOf(method.(*typesutil.TMethod).Func)
		if n == 1 {
			for _, r := range results[0] {
				if r.IsValue() {
					if v := valueOf(r.Value); v != nil {
						return v, true
					}
				}
			}
		}
	}
	return nil, false
}

func (scanner *OperatorScanner) getResponse(tpe types.Type, expr ast.Expr) (statusCode int, response *oas.Response) {
	response = &oas.Response{}

	if tpe.String() == "error" {
		statusCode = http.StatusNoContent
		return
	}

	contentType := ""

	if isHttpxResponse(tpe) {
		scanResponseWrapper := func(expr ast.Expr) {
			firstCallExpr := true

			ast.Inspect(expr, func(node ast.Node) bool {
				switch callExpr := node.(type) {
				case *ast.CallExpr:
					if firstCallExpr {
						firstCallExpr = false
						v, _ := scanner.program.Eval(callExpr.Args[0])
						tpe = v.Type
					}
					switch e := callExpr.Fun.(type) {
					case *ast.SelectorExpr:
						switch e.Sel.Name {
						case "WithStatusCode":
							v, _ := scanner.program.Eval(callExpr.Args[0])
							if code, ok := valueOf(v.Value).(int); ok {
								statusCode = code
							}
							return false
						case "WithContentType":
							v, _ := scanner.program.Eval(callExpr.Args[0])
							if code, ok := valueOf(v.Value).(string); ok {
								contentType = code
							}
							return false
						}
					}
				}
				return true
			})
		}

		if ident, ok := expr.(*ast.Ident); ok && ident.Obj != nil {
			if stmt, ok := ident.Obj.Decl.(*ast.AssignStmt); ok {
				for _, e := range stmt.Rhs {
					scanResponseWrapper(e)
				}
			}
		} else {
			scanResponseWrapper(expr)
		}
	}

	if pointer, ok := tpe.(*types.Pointer); ok {
		tpe = pointer.Elem()
	}

	if named, ok := tpe.(*types.Named); ok {
		if v, ok := scanner.firstValueOfFunc(named, "ContentType"); ok {
			contentType = v.(string)
		}
		if v, ok := scanner.firstValueOfFunc(named, "StatusCode"); ok {
			statusCode = int(v.(int64))
		}
	}

	response.AddContent(contentType, oas.NewMediaTypeWithSchema(scanner.DefinitionScanner.getSchemaByType(tpe)))

	return
}

func (scanner *OperatorScanner) scanParameterOrRequestBody(op *Operator, typeStruct *types.Struct) {
	typesutil.EachField(typesutil.FromTType(typeStruct), "name", func(field typesutil.StructField, fieldDisplayName string, omitempty bool) bool {
		location, _ := tagValueAndFlagsByTagString(field.Tag().Get("in"))

		if location == "" {
			panic(fmt.Errorf("missing tag `in` for %s of %s", field.Name(), op.ID))
		}

		schema := scanner.DefinitionScanner.getSchemaByType(field.Type().(*typesutil.TType).Type)

		transformer, err := transformers.TransformerMgrDefault.NewTransformer(field.Type(), transformers.TransformerOption{
			MIME: field.Tag().Get("mime"),
		})
		if err != nil {
			panic(err)
		}

		defaultValue, hasDefault := field.Tag().Lookup("default")
		if hasDefault {
			schema.Default = defaultValue
		}

		validate, hasValidate := field.Tag().Lookup("validate")
		if hasValidate {
			if err := BindSchemaValidationByValidateBytes(schema, field.Type().(*typesutil.TType).Type, []byte(validate)); err != nil {
				panic(err)
			}
		}
		schema.Description = scanner.program.CommentsOf(scanner.program.IdentOf(field.(*typesutil.TStructField).Var))

		schema.AddExtension(XGoFieldName, field.Name())

		tagKeys := map[string]string{
			"name":     XTagName,
			"mime":     XTagMime,
			"json":     XTagJSON,
			"xml":      XTagXML,
			"validate": XTagValidate,
		}

		for k, extKey := range tagKeys {
			if v, ok := field.Tag().Lookup(k); ok {
				schema.AddExtension(extKey, v)
			}
		}

		switch location {
		case "body":
			reqBody := oas.NewRequestBody("", true)
			reqBody.AddContent(transformer.Names()[0], oas.NewMediaTypeWithSchema(schema))
			op.SetRequestBody(reqBody)
		case "query":
			op.AddNonBodyParameter(oas.QueryParameter(fieldDisplayName, schema, !omitempty))
		case "cookie":
			op.AddNonBodyParameter(oas.CookieParameter(fieldDisplayName, schema, !omitempty))
		case "header":
			op.AddNonBodyParameter(oas.HeaderParameter(fieldDisplayName, schema, !omitempty))
		case "path":
			op.AddNonBodyParameter(oas.PathParameter(fieldDisplayName, schema))
		}

		return true
	}, "in")
}

type Operator struct {
	ID string

	Tag         string
	Summary     string
	Description string

	NonBodyParameters map[string]*oas.Parameter
	RequestBody       *oas.RequestBody

	StatusErrors      []*statuserror.StatusErr
	StatusErrorSchema *oas.Schema

	SuccessStatus   int
	SuccessType     types.Type
	SuccessResponse *oas.Response
}

func (operator *Operator) AddNonBodyParameter(parameter *oas.Parameter) {
	if operator.NonBodyParameters == nil {
		operator.NonBodyParameters = map[string]*oas.Parameter{}
	}
	operator.NonBodyParameters[parameter.Name] = parameter
}

func (operator *Operator) SetRequestBody(requestBody *oas.RequestBody) {
	operator.RequestBody = requestBody
}

func (operator *Operator) BindOperation(method string, operation *oas.Operation, last bool) {
	parameterNames := map[string]bool{}
	for _, parameter := range operation.Parameters {
		parameterNames[parameter.Name] = true
	}

	for _, parameter := range operator.NonBodyParameters {
		if !parameterNames[parameter.Name] {
			operation.Parameters = append(operation.Parameters, parameter)
		}
	}

	if operator.RequestBody != nil {
		operation.SetRequestBody(operator.RequestBody)
	}

	for _, statusError := range operator.StatusErrors {
		statusErrorList := make([]string, 0)
		if operation.Responses.Responses != nil {
			if resp, ok := operation.Responses.Responses[statusError.StatusCode()]; ok {
				if resp.Extensions != nil {
					if v, ok := resp.Extensions[XStatusErrs]; ok {
						if list, ok := v.([]string); ok {
							statusErrorList = append(statusErrorList, list...)
						}
					}
				}
			}
		}
		statusErrorList = append(statusErrorList, statusError.Summary())

		resp := oas.NewResponse("")
		resp.AddExtension(XStatusErrs, statusErrorList)
		resp.AddContent(httpx.MIME_JSON, oas.NewMediaTypeWithSchema(operator.StatusErrorSchema))
		operation.AddResponse(statusError.StatusCode(), resp)
	}

	if last {
		operation.OperationId = operator.ID

		operation.Summary = operator.Summary
		operation.Description = operator.Description

		if operator.Tag != "" {
			operation.Tags = []string{operator.Tag}
		}

		if operator.SuccessType == nil {
			operation.Responses.AddResponse(http.StatusNoContent, &oas.Response{})
		} else {
			status := operator.SuccessStatus
			if status == 0 {
				status = http.StatusOK
				if method == http.MethodPost {
					status = http.StatusCreated
				}
			}
			if status >= http.StatusMultipleChoices && status < http.StatusBadRequest {
				operator.SuccessResponse = oas.NewResponse(operator.SuccessResponse.Description)
			}
			operation.Responses.AddResponse(status, operator.SuccessResponse)
		}
	}

	// sort all parameters by postion and name
	if len(operation.Parameters) > 0 {
		sort.Slice(operation.Parameters, func(i, j int) bool {
			return positionOrders[operation.Parameters[i].In]+operation.Parameters[i].Name <
				positionOrders[operation.Parameters[j].In]+operation.Parameters[j].Name
		})
	}
}

var positionOrders = map[oas.Position]string{
	"path":   "1",
	"header": "2",
	"query":  "3",
	"cookie": "4",
}

func valueOf(v constant.Value) interface{} {
	switch v.Kind() {
	case constant.Float:
		v, _ := strconv.ParseFloat(v.String(), 10)
		return v
	case constant.Bool:
		v, _ := strconv.ParseBool(v.String())
		return v
	case constant.String:
		v, _ := strconv.Unquote(v.String())
		return v
	case constant.Int:
		v, _ := strconv.ParseInt(v.String(), 10, 64)
		return v
	}
	return nil
}
