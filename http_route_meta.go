package httptransport

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/go-courier/courier"
	"github.com/julienschmidt/httprouter"
)

type MethodDescriber interface {
	Method() string
}

type PathDescriber interface {
	Path() string
}

func NewHttpRouteMeta(route *courier.Route) *HttpRouteMeta {
	return &HttpRouteMeta{
		Route: route,
	}
}

type HttpRouteMeta struct {
	*courier.Route
}

func (route *HttpRouteMeta) Key() string {
	operatorFactories := route.OperatorFactories()

	if len(operatorFactories) == 0 {
		panic(fmt.Errorf(
			"no available operator %v",
			route.Operators,
		))
	}

	operatorTypeNames := make([]string, len(operatorFactories))

	for i, opFactory := range operatorFactories {
		if opFactory.IsLast {
			operatorTypeNames[i] = color.MagentaString(opFactory.String())
		} else {
			operatorTypeNames[i] = color.CyanString(opFactory.String())
		}
	}

	return RxHttpRouterPath.ReplaceAllString(route.Path(), "/{$1}") + " " + strings.Join(operatorTypeNames, " ")
}

func (route *HttpRouteMeta) String() string {
	method := route.Method()
	return methodColor(method)("%s %s", method[0:3], route.Key())
}

var RxHttpRouterPath = regexp.MustCompile("/:([^/]+)")

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
	lastOp := route.Operators[len(route.Operators)-1]
	if methodDescriber, ok := route.Operators[len(route.Operators)-1].(MethodDescriber); ok {
		return methodDescriber.Method()
	}
	panic(fmt.Errorf("missing `Method() string` of %s", reflect.TypeOf(lastOp).Name()))
}

func (route *HttpRouteMeta) Path() string {
	p := "/"
	for _, operator := range route.Operators {
		if pathDescriber, ok := operator.(PathDescriber); ok {
			p += pathDescriber.Path()
		}
	}
	return httprouter.CleanPath(p)
}

func Group(path string) *GroupOperator {
	return &GroupOperator{
		path: path,
	}
}

type GroupOperator struct {
	courier.EmptyOperator
	path string
}

func (g *GroupOperator) OperatorParams() map[string][]string {
	return map[string][]string{
		"path": {g.path},
	}
}

func (g *GroupOperator) Path() string {
	return g.path
}
