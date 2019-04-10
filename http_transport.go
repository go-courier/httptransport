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

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/handlers"
	"github.com/go-courier/httptransport/transformers"
	"github.com/go-courier/validator"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
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

	// Logger
	Logger *logrus.Entry

	CertFile string
	KeyFile  string

	httpRouter *httprouter.Router
}

type ServerModifier func(server *http.Server)

func (t *HttpTransport) SetDefaults() {
	t.ServiceMeta.SetDefaults()

	if t.ValidatorMgr == nil {
		t.ValidatorMgr = validator.ValidatorMgrDefault
	}

	if t.TransformerMgr == nil {
		t.TransformerMgr = transformers.TransformerMgrDefault
	}

	if t.Logger == nil {
		t.Logger = logrus.WithField("service", t.ServiceMeta.String())
	}

	if t.Middlewares == nil {
		t.Middlewares = []HttpMiddleware{handlers.LogHandler(t.Logger)}
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
	t.SetDefaults()

	t.httpRouter = t.convertRouterToHttpRouter(router)

	srv := &http.Server{}

	for i := range t.ServerModifiers {
		t.ServerModifiers[i](srv)
	}

	srv.Addr = fmt.Sprintf(":%d", t.Port)
	srv.Handler = MiddlewareChain(t.Middlewares...)(t)

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() {
		courierPrintln("%s listen on %s", t.ServiceMeta, srv.Addr)

		if t.CertFile != "" && t.KeyFile != "" {
			if err := srv.ListenAndServeTLS(t.CertFile, t.KeyFile); err != nil {
				if err == http.ErrServerClosed {
					logrus.Error(err)
				} else {
					logrus.Fatal(err)
				}
			}
			return
		}

		if err := srv.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				logrus.Error(err)
			} else {
				logrus.Fatal(err)
			}
		}
	}()

	<-stopCh

	t.Logger.Infof("shutdowning in %s", 10*time.Second)

	return srv.Shutdown(ctx)
}

func (t *HttpTransport) convertRouterToHttpRouter(router *courier.Router) *httprouter.Router {
	routes := router.Routes()

	if len(routes) == 0 {
		panic(fmt.Errorf("need to register Operator to Router %#v before serve", router))
	}

	httpRouter := httprouter.New()

	sort.Slice(routes, func(i, j int) bool {
		return NewHttpRouteMeta(routes[i]).Key() < NewHttpRouteMeta(routes[j]).Key()
	})

	for i := range routes {
		httpRoute := NewHttpRouteMeta(routes[i])
		courierPrintln(httpRoute.String())

		if err := TryCatch(func() {
			httpRouter.HandlerFunc(
				httpRoute.Method(),
				httpRoute.Path(),
				NewHttpRouteHandler(&t.ServiceMeta, httpRoute, NewRequestTransformerMgr(t.TransformerMgr, t.ValidatorMgr)).ServeHTTP,
			)
		}); err != nil {
			panic(fmt.Errorf("register http route `%s` failed: %s", httpRoute, err))
		}
	}

	return httpRouter
}
