package httptransport

import (
	"context"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
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

var httpRequestKey = &struct{}{}

func ContextWithHttpRequest(ctx context.Context, req *http.Request) context.Context {
	return context.WithValue(ctx, httpRequestKey, req)
}

func HttpRequestFromContext(ctx context.Context) *http.Request {
	p, _ := ctx.Value(httpRequestKey).(*http.Request)
	return p
}

var contextKeyServiceMetaKey = &struct{}{}

func ContextWithServiceMeta(ctx context.Context, meta ServiceMeta) context.Context {
	return context.WithValue(ctx, contextKeyServiceMetaKey, meta)
}

func ServerMetaFromContext(ctx context.Context) ServiceMeta {
	p, _ := ctx.Value(contextKeyServiceMetaKey).(ServiceMeta)
	return p
}

type ServiceMetaHook struct {
	ServiceMeta
}

func (ServiceMetaHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (s *ServiceMetaHook) Fire(entry *logrus.Entry) error {
	entry.Data["service"] = s.String()
	return nil
}
