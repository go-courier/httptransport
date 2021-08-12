package httptransport

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/go-courier/logr"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/handlers"
	"github.com/go-courier/httptransport/transformers"
	"github.com/go-courier/httptransport/validator"
	"github.com/julienschmidt/httprouter"
)

func MiddlewareChain(mw ...HttpMiddleware) HttpMiddleware {
	return func(final http.Handler) http.Handler {
		last := final
		for i := len(mw) - 1; i >= 0; i-- {
			last = mw[i](last)
		}
		return last
	}
}

type HttpMiddleware func(http.Handler) http.Handler

func NewHttpTransport(serverModifiers ...ServerModifier) *HttpTransport {
	return &HttpTransport{
		ServerModifiers: serverModifiers,
	}
}

type HttpTransport struct {
	ServiceMeta

	Port int

	// for modifying http.Server
	ServerModifiers []ServerModifier

	// Middlewares
	// can use https://github.com/gorilla/handlers
	Middlewares []HttpMiddleware

	// validator mgr for parameter validating
	ValidatorMgr validator.ValidatorMgr
	// transformer mgr for parameter transforming
	TransformerMgr transformers.TransformerMgr

	CertFile string
	KeyFile  string

	httpRouter *httprouter.Router
}

type ServerModifier func(server *http.Server) error

func (t *HttpTransport) SetDefaults() {
	t.ServiceMeta.SetDefaults()

	if t.ValidatorMgr == nil {
		t.ValidatorMgr = validator.ValidatorMgrDefault
	}

	if t.TransformerMgr == nil {
		t.TransformerMgr = transformers.TransformerMgrDefault
	}

	if t.Middlewares == nil {
		t.Middlewares = []HttpMiddleware{handlers.LogHandler()}
	}

	if t.Port == 0 {
		t.Port = 80
	}
}

func (t *HttpTransport) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	t.httpRouter.ServeHTTP(w, req)
}

func courierPrintln(format string, args ...interface{}) {
	fmt.Printf(`[Courier] `+format+"\n", args...)
}

func (t *HttpTransport) Serve(router *courier.Router) error {
	return t.ServeContext(context.Background(), router)
}

func (t *HttpTransport) ServeContext(ctx context.Context, router *courier.Router) error {
	t.SetDefaults()

	logger := logr.FromContext(ctx)

	t.httpRouter = t.convertRouterToHttpRouter(router)

	srv := &http.Server{}

	srv.Addr = fmt.Sprintf(":%d", t.Port)
	srv.Handler = MiddlewareChain(t.Middlewares...)(t)

	for i := range t.ServerModifiers {
		if err := t.ServerModifiers[i](srv); err != nil {
			logger.Fatal(err)
		}
	}

	go func() {
		courierPrintln("%s listen on %s", t.ServiceMeta, srv.Addr)

		if t.CertFile != "" && t.KeyFile != "" {
			if err := srv.ListenAndServeTLS(t.CertFile, t.KeyFile); err != nil {
				if err == http.ErrServerClosed {
					logger.Error(err)
				} else {
					logger.Fatal(err)
				}
			}
			return
		}

		if err := srv.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				logger.Error(err)
			} else {
				logger.Fatal(err)
			}
		}
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)
	<-stopCh

	timeout := 10 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger.Info("shutdowning in %s", timeout)

	return srv.Shutdown(ctx)
}

func (t *HttpTransport) convertRouterToHttpRouter(router *courier.Router) *httprouter.Router {
	routes := router.Routes()

	if len(routes) == 0 {
		panic(errors.Errorf("need to register Operator to Router %#v before serve", router))
	}

	routeMetas := make([]*HttpRouteMeta, len(routes))
	for i := range routes {
		routeMetas[i] = NewHttpRouteMeta(routes[i])
	}

	httpRouter := httprouter.New()

	sort.Slice(routeMetas, func(i, j int) bool {
		return routeMetas[i].Key() < routeMetas[j].Key()
	})

	for i := range routeMetas {
		httpRoute := routeMetas[i]
		httpRoute.Log()

		if err := tryCatch(func() {
			httpRouter.HandlerFunc(
				httpRoute.Method(),
				httpRoute.Path(),
				NewHttpRouteHandler(&t.ServiceMeta, httpRoute, NewRequestTransformerMgr(t.TransformerMgr, t.ValidatorMgr)).ServeHTTP,
			)
		}); err != nil {
			panic(errors.Errorf("register http route `%s` failed: %s", httpRoute, err))
		}
	}

	return httpRouter
}

func tryCatch(fn func()) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.Errorf("%+v", e)
		}
	}()

	fn()
	return nil
}
