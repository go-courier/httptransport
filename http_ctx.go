package httptransport

import (
	"context"
	"net/http"
	"os"

	"github.com/go-courier/courier"

	contextx "github.com/go-courier/x/context"
)

type ServiceMeta struct {
	Name    string
	Version string
}

func (s *ServiceMeta) SetDefaults() {
	if s.Name == "" {
		s.Name = os.Getenv("PROJECT_NAME")
	}

	if s.Version == "" {
		s.Version = os.Getenv("PROJECT_VERSION")
	}
}

func (s ServiceMeta) String() string {
	if s.Version == "" {
		return s.Name
	}
	return s.Name + "@" + s.Version
}

type contextKeyHttpRequestKey struct{}

func ContextWithHttpRequest(ctx context.Context, req *http.Request) context.Context {
	return contextx.WithValue(ctx, contextKeyHttpRequestKey{}, req)
}

func HttpRequestFromContext(ctx context.Context) *http.Request {
	p, _ := ctx.Value(contextKeyHttpRequestKey{}).(*http.Request)
	return p
}

type contextKeyServiceMetaKey struct{}

func ContextWithServiceMeta(ctx context.Context, meta ServiceMeta) context.Context {
	return contextx.WithValue(ctx, contextKeyServiceMetaKey{}, meta)
}

func ServerMetaFromContext(ctx context.Context) ServiceMeta {
	p, _ := ctx.Value(contextKeyServiceMetaKey{}).(ServiceMeta)
	return p
}

type contextKeyOperationID struct{}

func ContextWithOperationID(ctx context.Context, operationID string) context.Context {
	return contextx.WithValue(ctx, contextKeyOperationID{}, operationID)
}

func OperationIDFromContext(ctx context.Context) string {
	return ctx.Value(contextKeyOperationID{}).(string)
}

type contextKeyOperatorFactory struct{}

func ContextWithOperatorFactory(ctx context.Context, om *courier.OperatorFactory) context.Context {
	return contextx.WithValue(ctx, contextKeyOperatorFactory{}, om)
}

func OperatorFactoryFromContext(ctx context.Context) *courier.OperatorFactory {
	v, _ := ctx.Value(contextKeyOperatorFactory{}).(*courier.OperatorFactory)
	return v
}

type contextKeyQueryInBodyForHttpGet struct{}

func EnableQueryInBodyForHttpGet(ctx context.Context) context.Context {
	return contextx.WithValue(ctx, contextKeyQueryInBodyForHttpGet{}, true)
}

func ShouldQueryInBodyForHttpGet(ctx context.Context) bool {
	if v, ok := ctx.Value(contextKeyQueryInBodyForHttpGet{}).(bool); ok {
		return v
	}
	return false
}
