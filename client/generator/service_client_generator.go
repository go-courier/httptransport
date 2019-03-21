package generator

import (
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

func (g *ServiceClientGenerator) Scan(openapi *oas.OpenAPI) {
	g.WriteClientInterface(openapi)

	g.WriteClient()

	eachOperation(openapi, func(method string, path string, op *oas.Operation) {
		g.File.WriteBlock(
			g.OperationMethod(op, false),
		)
	})
}

func (g *ServiceClientGenerator) WriteClientInterface(openapi *oas.OpenAPI) {
	snippets := make([]codegen.SnippetCanBeInterfaceMethod, 0)

	eachOperation(openapi, func(method string, path string, op *oas.Operation) {
		snippets = append(snippets, g.OperationMethod(op, true).(*codegen.FuncType))
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
			),
				g.ClientInstanceName(),
			),
		),
	)
}

func (g *ServiceClientGenerator) OperationMethod(operation *oas.Operation, asInterface bool) codegen.Snippet {
	mediaType, _ := mediaTypeAndStatusErrors(&operation.Responses)

	if mediaType != nil {
		respType, _ := NewTypeGenerator(g.ServiceName, g.File).Type(mediaType.Schema)

		if respType != nil {
			m := codegen.Func(
				codegen.Var(codegen.Star(codegen.Type(operation.OperationId)), "req"),
				codegen.Var(codegen.Ellipsis(codegen.Type(g.File.Use("github.com/go-courier/courier", "Metadata"))), "metas"),
			).
				Named(operation.OperationId).
				Return(
					codegen.Var(codegen.Star(respType)),
					codegen.Var(codegen.Type(g.File.Use("github.com/go-courier/courier", "Metadata"))),
					codegen.Var(codegen.Error),
				)

			if asInterface {
				return m
			}

			return m.
				MethodOf(codegen.Var(codegen.Star(codegen.Type(g.ClientInstanceName())), "c")).
				Do(codegen.Return(codegen.Expr("req.Invoke(c.Client)")))
		}
	}

	m := codegen.Func(
		codegen.Var(codegen.Ellipsis(codegen.Type(g.File.Use("github.com/go-courier/courier", "Metadata"))), "metas"),
	).
		Return(
			codegen.Var(codegen.Type(g.File.Use("github.com/go-courier/courier", "Metadata"))),
			codegen.Var(codegen.Error),
		).
		Named(operation.OperationId)

	if asInterface {
		return m
	}

	return m.
		MethodOf(codegen.Var(codegen.Star(codegen.Type(g.ClientInstanceName())), "c")).
		Do(codegen.Return(codegen.Expr("(&?{}).Invoke(c.Client)", codegen.Type(operation.OperationId))))
}
