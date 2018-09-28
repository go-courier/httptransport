package generator

import (
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/go-courier/codegen"
	"github.com/go-courier/httptransport/openapi/generator"
	"github.com/go-courier/oas"
)

func NewOperationGenerator(serviceName string, file *codegen.File) *OperationGenerator {
	return &OperationGenerator{
		ServiceName: serviceName,
		File:        file,
	}
}

type OperationGenerator struct {
	ServiceName string
	File        *codegen.File
}

func (g *OperationGenerator) Scan(openapi *oas.OpenAPI) {
	ops := map[string]struct {
		Method string
		Path   string
		*oas.Operation
	}{}
	operationIDs := make([]string, 0)
	braceToColonRe := regexp.MustCompile(`(.*){(.*)}(.*)`)

	for path := range openapi.Paths.Paths {
		pathItems := openapi.Paths.Paths[path]
		for method := range pathItems.Operations.Operations {
			op := pathItems.Operations.Operations[method]

			if strings.HasPrefix(op.OperationId, "OpenAPI") {
				continue
			}

			ops[op.OperationId] = struct {
				Method string
				Path   string
				*oas.Operation
			}{
				Method:    strings.ToUpper(string(method)),
				Path:      braceToColonRe.ReplaceAllString(path, `$1:$2$3`),
				Operation: op,
			}
			operationIDs = append(operationIDs, op.OperationId)
		}
	}

	sort.Strings(operationIDs)

	for _, id := range operationIDs {
		op := ops[id]
		g.WriteOperation(op.Method, op.Path, op.Operation)
	}
}

func (g *OperationGenerator) ID(id string) string {
	if g.ServiceName != "" {
		return g.ServiceName + "." + id
	}
	return id
}

func (g *OperationGenerator) WriteOperation(method string, path string, operation *oas.Operation) {
	id := operation.OperationId

	fields := make([]*codegen.SnippetField, 0)

	for i := range operation.Parameters {
		fields = append(fields, g.ParamField(operation.Parameters[i]))
	}

	if respBodyField := g.RequestBodyField(operation.RequestBody); respBodyField != nil {
		fields = append(fields, respBodyField)
	}

	g.File.WriteBlock(
		codegen.DeclType(
			codegen.Var(codegen.Struct(fields...), id),
		),
	)

	g.File.WriteBlock(
		codegen.Func().
			Named("Path").Return(codegen.Var(codegen.String)).
			MethodOf(codegen.Var(codegen.Type(id))).
			Do(codegen.Return(g.File.Val(path))),
	)

	g.File.WriteBlock(
		codegen.Func().
			Named("Method").Return(codegen.Var(codegen.String)).
			MethodOf(codegen.Var(codegen.Type(id))).
			Do(codegen.Return(g.File.Val(method))),
	)

	respType, statusErrors := g.ResponseType(&operation.Responses)

	g.File.Write(codegen.Comments(statusErrors...).Bytes())

	if respType != nil {
		g.File.WriteBlock(
			codegen.Func(
				codegen.Var(codegen.Type(g.File.Use("github.com/go-courier/courier", "Client")), "c"),
				codegen.Var(codegen.Ellipsis(codegen.Type(g.File.Use("github.com/go-courier/courier", "Metadata"))), "metas"),
			).
				Return(
					codegen.Var(codegen.Star(respType)),
					codegen.Var(codegen.Type(g.File.Use("github.com/go-courier/courier", "Metadata"))),
					codegen.Var(codegen.Error),
				).
				Named("Invoke").
				MethodOf(codegen.Var(codegen.Star(codegen.Type(id)), "req")).
				Do(
					codegen.Expr("resp := new(?)", respType),
					codegen.Expr("meta, err := c.Do(?, req, metas...).Into(resp)", g.File.Val(g.ID(id)), respType),
					codegen.Return(codegen.Id("resp"), codegen.Id("meta"), codegen.Id("err")),
				),
		)
		return
	}

	g.File.WriteBlock(
		codegen.Func(
			codegen.Var(codegen.Type(g.File.Use("github.com/go-courier/courier", "Client")), "c"),
			codegen.Var(codegen.Ellipsis(codegen.Type(g.File.Use("github.com/go-courier/courier", "Metadata"))), "metas"),
		).
			Return(
				codegen.Var(codegen.Type(g.File.Use("github.com/go-courier/courier", "Metadata"))),
				codegen.Var(codegen.Error),
			).
			Named("Invoke").
			MethodOf(codegen.Var(codegen.Star(codegen.Type(id)), "req")).
			Do(
				codegen.Return(codegen.Expr("c.Do(?, req, metas...).Into(nil)", g.File.Val(g.ID(id)))),
			),
	)

}

func (g *OperationGenerator) ParamField(parameter *oas.Parameter) *codegen.SnippetField {
	field := NewTypeGenerator(g.ServiceName, g.File).FieldOf(parameter.Name, parameter.Schema, map[string]bool{
		parameter.Name: parameter.Required,
	})

	tag := `in:"` + string(parameter.In) + `"`
	if field.Tag != "" {
		tag = tag + " " + field.Tag
	}
	field.Tag = tag

	return field
}

func (g *OperationGenerator) RequestBodyField(requestBody *oas.RequestBody) *codegen.SnippetField {
	if requestBody == nil {
		return nil
	}

	for contentType := range requestBody.Content {
		mediaType := requestBody.Content[contentType]

		field := NewTypeGenerator(g.ServiceName, g.File).FieldOf("Data", mediaType.Schema, map[string]bool{})

		tag := `in:"body"`
		if field.Tag != "" {
			tag = tag + " " + field.Tag
		}
		field.Tag = tag

		return field
	}

	return nil
}

func isOk(code int) bool {
	return code >= http.StatusOK && code < http.StatusMultipleChoices
}

func (g *OperationGenerator) ResponseType(responses *oas.Responses) (codegen.SnippetType, []string) {
	if responses == nil {
		return nil, nil
	}

	response := (*oas.Response)(nil)

	statusErrors := make([]string, 0)

	for code := range responses.Responses {
		if isOk(code) {
			response = responses.Responses[code]
		} else {
			extensions := responses.Responses[code].Extensions

			if extensions != nil {
				if errors, ok := extensions[generator.XStatusErrs]; ok {
					if errs, ok := errors.([]interface{}); ok {
						for _, err := range errs {
							statusErrors = append(statusErrors, err.(string))
						}
					}
				}
			}
		}
	}

	if response == nil {
		return nil, nil
	}

	for contentType := range response.Content {
		mediaType := response.Content[contentType]
		typ, _ := NewTypeGenerator(g.ServiceName, g.File).Type(mediaType.Schema)
		return typ, statusErrors
	}

	return nil, statusErrors
}
