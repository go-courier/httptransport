package generator

import (
	"context"
	"net/http"
	"regexp"

	"github.com/go-courier/codegen"
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

var reBraceToColon = regexp.MustCompile(`/\{([^/]+)\}`)

func toColonPath(path string) string {
	return reBraceToColon.ReplaceAllStringFunc(path, func(str string) string {
		name := reBraceToColon.FindAllStringSubmatch(str, -1)[0][1]
		return "/:" + name
	})
}

func (g *OperationGenerator) Scan(ctx context.Context, openapi *oas.OpenAPI) {
	eachOperation(openapi, func(method string, path string, op *oas.Operation) {
		g.WriteOperation(ctx, method, path, op)
	})
}

func (g *OperationGenerator) ID(id string) string {
	if g.ServiceName != "" {
		return g.ServiceName + "." + id
	}
	return id
}

func (g *OperationGenerator) WriteOperation(ctx context.Context, method string, path string, operation *oas.Operation) {
	id := operation.OperationId

	fields := make([]*codegen.SnippetField, 0)

	for i := range operation.Parameters {
		fields = append(fields, g.ParamField(ctx, operation.Parameters[i]))
	}

	if respBodyField := g.RequestBodyField(ctx, operation.RequestBody); respBodyField != nil {
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

	respType, statusErrors := g.ResponseType(ctx, &operation.Responses)

	g.File.Write(codegen.Comments(statusErrors...).Bytes())

	ctxWithMeta := `ctx = ` + g.File.Use("github.com/go-courier/metax", "ContextWith") + `(ctx, "operationID","` + g.ID(id) + `")`

	g.File.WriteBlock(
		codegen.Func(
			codegen.Var(codegen.Type(g.File.Use("context", "Context")), "ctx"),
			codegen.Var(codegen.Type(g.File.Use("github.com/go-courier/courier", "Client")), "c"),
			codegen.Var(codegen.Ellipsis(codegen.Type(g.File.Use("github.com/go-courier/courier", "Metadata"))), "metas"),
		).
			Return(
				codegen.Var(codegen.Type(g.File.Use("github.com/go-courier/courier", "Result"))),
			).
			Named("Do").
			MethodOf(codegen.Var(codegen.Star(codegen.Type(id)), "req")).
			Do(
				codegen.Expr(`
` + ctxWithMeta + `
return c.Do(ctx, req, metas...)
`),
			),
	)

	if respType != nil {
		g.File.WriteBlock(
			codegen.Func(
				codegen.Var(codegen.Type(g.File.Use("context", "Context")), "ctx"),
				codegen.Var(codegen.Type(g.File.Use("github.com/go-courier/courier", "Client")), "c"),
				codegen.Var(codegen.Ellipsis(codegen.Type(g.File.Use("github.com/go-courier/courier", "Metadata"))), "metas"),
			).
				Return(
					codegen.Var(codegen.Star(respType)),
					codegen.Var(codegen.Type(g.File.Use("github.com/go-courier/courier", "Metadata"))),
					codegen.Var(codegen.Error),
				).
				Named("InvokeContext").
				MethodOf(codegen.Var(codegen.Star(codegen.Type(id)), "req")).
				Do(
					codegen.Expr("resp := new(?)", respType),
					codegen.Expr(`
meta, err := req.Do(ctx, c, metas...).Into(resp)
`),
					codegen.Return(codegen.Id("resp"), codegen.Id("meta"), codegen.Id("err")),
				),
		)

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
					codegen.Return(codegen.Expr("req.InvokeContext(context.Background(), c, metas...)")),
				),
		)

		return
	}

	g.File.WriteBlock(
		codegen.Func(
			codegen.Var(codegen.Type(g.File.Use("context", "Context")), "ctx"),
			codegen.Var(codegen.Type(g.File.Use("github.com/go-courier/courier", "Client")), "c"),
			codegen.Var(codegen.Ellipsis(codegen.Type(g.File.Use("github.com/go-courier/courier", "Metadata"))), "metas"),
		).
			Return(
				codegen.Var(codegen.Type(g.File.Use("github.com/go-courier/courier", "Metadata"))),
				codegen.Var(codegen.Error),
			).
			Named("InvokeContext").
			MethodOf(codegen.Var(codegen.Star(codegen.Type(id)), "req")).
			Do(
				codegen.Return(
					codegen.Expr(`req.Do(ctx, c, metas...).Into(nil)`),
				),
			),
	)

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
				codegen.Return(codegen.Expr("req.InvokeContext(context.Background(), c, metas...)")),
			),
	)

}

func (g *OperationGenerator) ParamField(ctx context.Context, parameter *oas.Parameter) *codegen.SnippetField {
	field := NewTypeGenerator(g.ServiceName, g.File).FieldOf(ctx, parameter.Name, parameter.Schema, map[string]bool{
		parameter.Name: parameter.Required,
	})

	tag := `in:"` + string(parameter.In) + `"`
	if field.Tag != "" {
		tag = tag + " " + field.Tag
	}
	field.Tag = tag

	return field
}

func (g *OperationGenerator) RequestBodyField(ctx context.Context, requestBody *oas.RequestBody) *codegen.SnippetField {
	mediaType := requestBodyMediaType(requestBody)

	if mediaType == nil {
		return nil
	}

	field := NewTypeGenerator(g.ServiceName, g.File).FieldOf(ctx, "Data", mediaType.Schema, map[string]bool{})

	tag := `in:"body"`
	if field.Tag != "" {
		tag = tag + " " + field.Tag
	}
	field.Tag = tag

	return field
}

func isOk(code int) bool {
	return code >= http.StatusOK && code < http.StatusMultipleChoices
}

func (g *OperationGenerator) ResponseType(ctx context.Context, responses *oas.Responses) (codegen.SnippetType, []string) {
	mediaType, statusErrors := mediaTypeAndStatusErrors(responses)

	if mediaType == nil {
		return nil, nil
	}

	typ, _ := NewTypeGenerator(g.ServiceName, g.File).Type(ctx, mediaType.Schema)
	return typ, statusErrors
}
