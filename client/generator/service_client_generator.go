package generator

import (
	"context"

	"github.com/go-courier/codegen"
	"github.com/go-courier/oas"
)

func NewServiceClientGenerator(serviceName string, file *codegen.File) *ServiceClientGenerator {
	return &ServiceClientGenerator{
		ServiceName: serviceName,
		File:        file,
	}
}

type ServiceClientGenerator struct {
	ServiceName string
	File        *codegen.File
}

func (g *ServiceClientGenerator) Scan(ctx context.Context, openapi *oas.OpenAPI) {
	g.WriteClientInterface(ctx, openapi)

	g.WriteClient()

	g.File.WriteBlock(codegen.Expr(`

func (c *` + g.ClientInstanceName() + `) WithContext(ctx context.Context) ` + g.ClientInterfaceName() + ` {
	cc := new(` + g.ClientInstanceName() + `)
	cc.Client = c.Client
	cc.ctx = ctx
	return cc
}

func (c *` + g.ClientInstanceName() + `) Context() context.Context {
	if c.ctx != nil {
      return c.ctx
    }
	return context.Background()
}

`))

	eachOperation(openapi, func(method string, path string, op *oas.Operation) {
		g.File.WriteBlock(
			g.OperationMethod(ctx, op, false),
		)
	})
}

func (g *ServiceClientGenerator) WriteClientInterface(ctx context.Context, openapi *oas.OpenAPI) {
	varContext := codegen.Var(codegen.Type(g.File.Use("context", "Context")))

	snippets := []codegen.SnippetCanBeInterfaceMethod{
		codegen.Func(varContext).Named("WithContext").Return(codegen.Var(codegen.Type(g.ClientInterfaceName()))),
		codegen.Func().Named("Context").Return(varContext),
	}

	eachOperation(openapi, func(method string, path string, op *oas.Operation) {
		snippets = append(snippets, g.OperationMethod(ctx, op, true).(*codegen.FuncType))
	})

	g.File.WriteBlock(
		codegen.DeclType(
			codegen.Var(codegen.Interface(
				snippets...,
			), g.ClientInterfaceName()),
		),
	)
}

func (g *ServiceClientGenerator) ClientInterfaceName() string {
	return codegen.UpperCamelCase("Client-" + g.ServiceName)
}

func (g *ServiceClientGenerator) ClientInstanceName() string {
	return codegen.UpperCamelCase("Client-" + g.ServiceName + "-Struct")
}

func (g *ServiceClientGenerator) WriteClient() {
	g.File.WriteBlock(
		codegen.Func(
			codegen.Var(codegen.Type(g.File.Use("github.com/go-courier/courier", "Client")), "c"),
		).Return(
			codegen.Var(codegen.Star(codegen.Type(g.ClientInstanceName()))),
		).Named(
			"New" + g.ClientInterfaceName(),
		).Do(
			codegen.Return(codegen.Unary(codegen.Paren(codegen.Compose(
				codegen.Type(g.ClientInstanceName()),
				codegen.KeyValue(codegen.Id("Client"), codegen.Id("c")),
			)))),
		),
	)

	g.File.WriteBlock(
		codegen.DeclType(
			codegen.Var(codegen.Struct(
				codegen.Var(codegen.Type(g.File.Use("github.com/go-courier/courier", "Client")), "Client"),
				codegen.Var(codegen.Type(g.File.Use("context", "Context")), "ctx"),
			),
				g.ClientInstanceName(),
			),
		),
	)
}

func (g *ServiceClientGenerator) OperationMethod(ctx context.Context, operation *oas.Operation, asInterface bool) codegen.Snippet {
	mediaType, _ := mediaTypeAndStatusErrors(&operation.Responses)

	params := make([]*codegen.SnippetField, 0)

	hasReq := len(operation.Parameters) != 0 || requestBodyMediaType(operation.RequestBody) != nil

	if hasReq {
		params = append(params, codegen.Var(codegen.Star(codegen.Type(operation.OperationId)), "req"))
	}

	params = append(params, codegen.Var(codegen.Ellipsis(codegen.Type(g.File.Use("github.com/go-courier/courier", "Metadata"))), "metas"))

	returns := make([]*codegen.SnippetField, 0)

	if mediaType != nil {
		respType, _ := NewTypeGenerator(g.ServiceName, g.File).Type(ctx, mediaType.Schema)

		if respType != nil {
			returns = append(returns, codegen.Var(codegen.Star(respType)))
		}
	}

	returns = append(
		returns,
		codegen.Var(codegen.Type(g.File.Use("github.com/go-courier/courier", "Metadata"))),
		codegen.Var(codegen.Error),
	)

	m := codegen.Func(params...).
		Return(returns...).
		Named(operation.OperationId)

	if asInterface {
		return m
	}

	m = m.
		MethodOf(codegen.Var(codegen.Star(codegen.Type(g.ClientInstanceName())), "c"))

	if hasReq {
		return m.Do(codegen.Return(codegen.Expr("req.InvokeContext(c.Context(), c.Client, metas...)")))
	}

	return m.Do(codegen.Return(codegen.Expr("(&?{}).InvokeContext(c.Context(), c.Client, metas...)", codegen.Type(operation.OperationId))))
}
