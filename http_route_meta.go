package httptransport

import (
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/httpx"
	"github.com/julienschmidt/httprouter"
)

type MethodDescriber interface {
	Method() string
}

type PathDescriber interface {
	Path() string
}

type BasePathDescriber interface {
	BasePath() string
}

var pkgPathHttpx = reflect.TypeOf(httpx.MethodGet{}).PkgPath()

func NewOperatorFactoryWithRouteMeta(op courier.Operator, last bool) *OperatorFactoryWithRouteMeta {
	f := courier.NewOperatorFactory(op, last)

	m := &OperatorFactoryWithRouteMeta{
		OperatorFactory: f,
	}

	m.ID = m.Type.Name()

	if methodDescriber, ok := op.(MethodDescriber); ok {
		m.Method = methodDescriber.Method()
	}

	if m.Type.Kind() == reflect.Struct {
		structType := m.Type

		for i := 0; i < structType.NumField(); i++ {
			f := structType.Field(i)
			if f.Anonymous && f.Type.PkgPath() == pkgPathHttpx && strings.HasPrefix(f.Name, "Method") {
				if path, ok := f.Tag.Lookup("path"); ok {
					vs := strings.Split(path, ",")
					m.Path = vs[0]

					if len(vs) > 0 {
						for i := range vs {
							switch vs[i] {
							case "deprecated":
								m.Deprecated = true
							}
						}
					}
				}

				if basePath, ok := f.Tag.Lookup("basePath"); ok {
					m.BasePath = basePath
				}

				if summary, ok := f.Tag.Lookup("summary"); ok {
					m.Summary = summary
				}

				break
			}
		}
	}

	if basePathDescriber, ok := op.(BasePathDescriber); ok {
		m.BasePath = basePathDescriber.BasePath()
	}

	if pathDescriber, ok := m.Operator.(PathDescriber); ok {
		m.Path = pathDescriber.Path()
	}

	return m
}

type RouteMeta struct {
	ID         string
	Method     string
	Path       string
	BasePath   string
	Summary    string
	Deprecated bool
}

type OperatorFactoryWithRouteMeta struct {
	*courier.OperatorFactory
	RouteMeta
}

func NewHttpRouteMeta(route *courier.Route) *HttpRouteMeta {
	operatorFactoryWithRouteMetas := make([]*OperatorFactoryWithRouteMeta, len(route.Operators))

	for i := range route.Operators {
		operatorFactoryWithRouteMetas[i] = NewOperatorFactoryWithRouteMeta(route.Operators[i], i == len(route.Operators)-1)
	}

	return &HttpRouteMeta{
		Route:                         route,
		OperatorFactoryWithRouteMetas: operatorFactoryWithRouteMetas,
	}
}

type HttpRouteMeta struct {
	Route                         *courier.Route
	OperatorFactoryWithRouteMetas []*OperatorFactoryWithRouteMeta
}

func (route *HttpRouteMeta) OperatorNames() string {
	operatorTypeNames := make([]string, 0)

	for _, opFactory := range route.OperatorFactoryWithRouteMetas {
		if opFactory.NoOutput {
			continue
		}

		if opFactory.IsLast {
			operatorTypeNames = append(operatorTypeNames, color.MagentaString(opFactory.String()))
		} else {
			operatorTypeNames = append(operatorTypeNames, color.CyanString(opFactory.String()))
		}
	}

	return strings.Join(operatorTypeNames, " ")
}

func (route *HttpRouteMeta) Key() string {
	return reHttpRouterPath.ReplaceAllString(route.Path(), "/{$1}") + " " + route.OperatorNames()
}

func (route *HttpRouteMeta) String() string {
	method := route.Method()

	return methodColor(method)("%s %s", method[0:3], route.Key())
}

func (route *HttpRouteMeta) Log() {
	method := route.Method()

	last := route.OperatorFactoryWithRouteMetas[len(route.OperatorFactoryWithRouteMetas)-1]

	firstLine := methodColor(method)("%s %s", method[0:3], reHttpRouterPath.ReplaceAllString(route.Path(), "/{$1}"))

	if last.Deprecated {
		firstLine = firstLine + " Deprecated"
	}

	if last.Summary != "" {
		firstLine = firstLine + " " + last.Summary
	}

	courierPrintln(firstLine)
	courierPrintln("\t%s", route.OperatorNames())
}

var reHttpRouterPath = regexp.MustCompile("/:([^/]+)")

func methodColor(method string) func(f string, args ...interface{}) string {
	switch method {
	case http.MethodGet:
		return color.BlueString
	case http.MethodPost:
		return color.GreenString
	case http.MethodPut:
		return color.YellowString
	case http.MethodDelete:
		return color.RedString
	default:
		return color.WhiteString
	}
}

func (route *HttpRouteMeta) Method() string {
	method := ""
	for _, m := range route.OperatorFactoryWithRouteMetas {
		if m.Method != "" {
			method = m.Method
		}
	}
	return method
}

func (route *HttpRouteMeta) Path() string {
	basePath := "/"
	p := ""

	for _, m := range route.OperatorFactoryWithRouteMetas {
		if m.BasePath != "" {
			basePath = m.BasePath
		}

		if m.Path != "" {
			p += m.Path
		}
	}

	return httprouter.CleanPath(basePath + p)
}

func BasePath(basePath string) *MetaOperator {
	return &MetaOperator{
		basePath: basePath,
	}
}

func Group(path string) *MetaOperator {
	return &MetaOperator{path: path}
}

type MetaOperator struct {
	courier.EmptyOperator
	path     string
	basePath string
}

func (g *MetaOperator) Path() string {
	return g.path
}

func (g *MetaOperator) BasePath() string {
	return g.basePath
}
