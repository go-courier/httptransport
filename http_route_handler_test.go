package httptransport_test

import (
	"context"
	"net/http"
	"net/http/httputil"
	"testing"
	"time"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport"
	"github.com/go-courier/httptransport/testdata/server/cmd/app/routes"
	"github.com/go-courier/httptransport/testify"
	. "github.com/onsi/gomega"
)

var rtMgr = httptransport.NewRequestTransformerMgr(nil, nil)
var serviceMeta = &httptransport.ServiceMeta{Name: "service-test", Version: "1.0.0"}

func init() {
	rtMgr.SetDefaults()
}

func TestHttpRouteHandler(t *testing.T) {
	t.Run("redirect", func(t *testing.T) {
		rootRouter := courier.NewRouter(httptransport.Group("/root"))
		rootRouter.Register(courier.NewRouter(routes.Redirect{}))

		httpRoute := httptransport.NewHttpRouteMeta(rootRouter.Routes()[0])
		httpRouterHandler := httptransport.NewHttpRouteHandler(serviceMeta, httpRoute, rtMgr)

		req, err := rtMgr.NewRequest((routes.Redirect{}).Method(), "/", routes.Redirect{})
		NewWithT(t).Expect(err).To(BeNil())

		rw := testify.NewMockResponseWriter()
		httpRouterHandler.ServeHTTP(rw, req)

		NewWithT(t).Expect(string(rw.MustDumpResponse())).To(Equal(`HTTP/0.0 302 Found
Content-Type: text/html; charset=utf-8
Location: /other
X-Meta: service-test@1.0.0/Redirect

<a href="/other">Found</a>.

`))
	})

	t.Run("redirect when error", func(t *testing.T) {
		rootRouter := courier.NewRouter(httptransport.Group("/root"))
		rootRouter.Register(courier.NewRouter(routes.RedirectWhenError{}))

		httpRoute := httptransport.NewHttpRouteMeta(rootRouter.Routes()[0])
		httpRouterHandler := httptransport.NewHttpRouteHandler(serviceMeta, httpRoute, rtMgr)

		req, err := rtMgr.NewRequest((routes.RedirectWhenError{}).Method(), "/", routes.RedirectWhenError{})
		NewWithT(t).Expect(err).To(BeNil())

		rw := testify.NewMockResponseWriter()
		httpRouterHandler.ServeHTTP(rw, req)

		NewWithT(t).Expect(string(rw.MustDumpResponse())).To(Equal(`HTTP/0.0 301 Moved Permanently
Location: /other
X-Meta: service-test@1.0.0/RedirectWhenError
Content-Length: 0

`))
	})

	t.Run("cookies", func(t *testing.T) {
		rootRouter := courier.NewRouter(httptransport.Group("/root"))
		rootRouter.Register(courier.NewRouter(&routes.Cookie{}))

		httpRoute := httptransport.NewHttpRouteMeta(rootRouter.Routes()[0])
		httpRouterHandler := httptransport.NewHttpRouteHandler(serviceMeta, httpRoute, rtMgr)

		req, err := rtMgr.NewRequest((routes.Cookie{}).Method(), "/", routes.Cookie{})
		NewWithT(t).Expect(err).To(BeNil())

		cookie := &http.Cookie{
			Name:    "token",
			Value:   "test",
			Expires: time.Now().Add(24 * time.Hour),
		}

		req.AddCookie(cookie)

		rw := testify.NewMockResponseWriter()
		httpRouterHandler.ServeHTTP(rw, req)

		NewWithT(t).Expect(string(rw.MustDumpResponse())).To(Equal(`HTTP/0.0 204 No Content
Set-Cookie: ` + cookie.String() + `
X-Meta: service-test@1.0.0/Cookie

`))
	})

	t.Run("return ok", func(t *testing.T) {
		rootRouter := courier.NewRouter(httptransport.Group("/root"))
		rootRouter.Register(courier.NewRouter(routes.DataProvider{}, routes.GetByID{}))

		httpRoute := httptransport.NewHttpRouteMeta(rootRouter.Routes()[0])
		httpRouterHandler := httptransport.NewHttpRouteHandler(serviceMeta, httpRoute, rtMgr)

		reqData := struct {
			routes.DataProvider
			routes.GetByID
		}{
			DataProvider: routes.DataProvider{
				ID: "123456",
			},
			GetByID: routes.GetByID{
				Label: []string{"label"},
			},
		}

		req, err := rtMgr.NewRequestWithContext(httptransport.EnableQueryInBodyForHttpGet(context.Background()), (routes.GetByID{}).Method(), reqData.Path(), reqData)
		NewWithT(t).Expect(err).To(BeNil())

		httpReq, _ := httputil.DumpRequest(req, true)
		NewWithT(t).Expect(string(httpReq)).To(Equal("GET /123456 HTTP/1.1\r\nContent-Type: application/x-www-form-urlencoded; param=value\r\n\r\nlabel=label"))

		rw := testify.NewMockResponseWriter()
		httpRouterHandler.ServeHTTP(rw, req)

		NewWithT(t).Expect(string(rw.MustDumpResponse())).To(Equal(`HTTP/0.0 200 OK
Content-Type: application/json; charset=utf-8
X-Meta: service-test@1.0.0/GetByID

{"id":"123456","label":"label"}
`))
	})

	t.Run("POST return ok", func(t *testing.T) {
		rootRouter := courier.NewRouter(httptransport.Group("/root"))
		rootRouter.Register(courier.NewRouter(routes.Create{}))

		httpRoute := httptransport.NewHttpRouteMeta(rootRouter.Routes()[0])
		httpRouterHandler := httptransport.NewHttpRouteHandler(serviceMeta, httpRoute, rtMgr)

		reqData := routes.Create{
			Data: routes.Data{
				ID:    "123456",
				Label: "123",
			},
		}

		req, err := rtMgr.NewRequest((routes.Create{}).Method(), "/", reqData)
		NewWithT(t).Expect(err).To(BeNil())

		rw := testify.NewMockResponseWriter()
		httpRouterHandler.ServeHTTP(rw, req)

		NewWithT(t).Expect(string(rw.MustDumpResponse())).To(Equal(`HTTP/0.0 201 Created
Content-Type: application/json; charset=utf-8
X-Meta: service-test@1.0.0/Create

{"id":"123456","label":"123"}
`))
	})

	t.Run("POST return bad request", func(t *testing.T) {
		rootRouter := courier.NewRouter(httptransport.Group("/root"))
		rootRouter.Register(courier.NewRouter(routes.Create{}))

		httpRoute := httptransport.NewHttpRouteMeta(rootRouter.Routes()[0])
		httpRouterHandler := httptransport.NewHttpRouteHandler(serviceMeta, httpRoute, rtMgr)

		reqData := routes.Create{
			Data: routes.Data{
				ID: "123456",
			},
		}

		req, err := rtMgr.NewRequest((routes.Create{}).Method(), "/", reqData)
		NewWithT(t).Expect(err).To(BeNil())

		rw := testify.NewMockResponseWriter()
		httpRouterHandler.ServeHTTP(rw, req)

		NewWithT(t).Expect(string(rw.MustDumpResponse())).To(Equal(`HTTP/0.0 400 Bad Request
Content-Type: application/json; charset=utf-8
X-Meta: service-test@1.0.0/Create

{"key":"BadRequest","code":400000000,"msg":"invalid parameters","desc":"","canBeTalkError":false,"id":"","sources":["service-test@1.0.0"],"errorFields":[{"field":"label","msg":"missing required field","in":"body"}]}
`))
	})

	t.Run("return nil", func(t *testing.T) {
		rootRouter := courier.NewRouter(httptransport.Group("/root"))
		rootRouter.Register(courier.NewRouter(routes.DataProvider{}, routes.RemoveByID{}))

		httpRoute := httptransport.NewHttpRouteMeta(rootRouter.Routes()[0])
		httpRouterHandler := httptransport.NewHttpRouteHandler(serviceMeta, httpRoute, rtMgr)

		reqData := routes.DataProvider{
			ID: "123456",
		}

		req, err := rtMgr.NewRequest((routes.RemoveByID{}).Method(), reqData.Path(), reqData)
		NewWithT(t).Expect(err).To(BeNil())

		rw := testify.NewMockResponseWriter()
		httpRouterHandler.ServeHTTP(rw, req)

		NewWithT(t).Expect(string(rw.MustDumpResponse())).To(Equal(`HTTP/0.0 500 Internal Server Error
Content-Type: application/json; charset=utf-8
X-Meta: service-test@1.0.0/RemoveByID
X-Num: 1

{"key":"InternalServerError","code":500999001,"msg":"InternalServerError","desc":"","canBeTalkError":false,"id":"","sources":["service-test@1.0.0"],"errorFields":null}
`))
	})

	t.Run("return attachment", func(t *testing.T) {
		rootRouter := courier.NewRouter(httptransport.Group("/root"))
		rootRouter.Register(courier.NewRouter(routes.DownloadFile{}))

		httpRoute := httptransport.NewHttpRouteMeta(rootRouter.Routes()[0])
		httpRouterHandler := httptransport.NewHttpRouteHandler(serviceMeta, httpRoute, rtMgr)

		req, err := rtMgr.NewRequest((routes.DownloadFile{}).Method(), (routes.DownloadFile{}).Path(), routes.DownloadFile{})
		NewWithT(t).Expect(err).To(BeNil())

		rw := testify.NewMockResponseWriter()
		httpRouterHandler.ServeHTTP(rw, req)

		NewWithT(t).Expect(string(rw.MustDumpResponse())).To(Equal(`HTTP/0.0 200 OK
Content-Disposition: attachment; filename=text.txt
Content-Type: text/plain
X-Meta: service-test@1.0.0/DownloadFile

123123123`))
	})

	t.Run("return with process error", func(t *testing.T) {
		rootRouter := courier.NewRouter(httptransport.Group("/root"))
		rootRouter.Register(courier.NewRouter(routes.DataProvider{}, routes.UpdateByID{}))

		httpRoute := httptransport.NewHttpRouteMeta(rootRouter.Routes()[0])
		httpRouterHandler := httptransport.NewHttpRouteHandler(serviceMeta, httpRoute, rtMgr)

		reqData := routes.DataProvider{
			ID: "123456",
		}

		req, err := rtMgr.NewRequest((routes.GetByID{}).Method(), reqData.Path(), struct {
			routes.DataProvider
			routes.UpdateByID
		}{
			DataProvider: reqData,
			UpdateByID: routes.UpdateByID{
				Data: routes.Data{
					ID:    "11",
					Label: "11",
				},
			},
		})
		NewWithT(t).Expect(err).To(BeNil())

		rw := testify.NewMockResponseWriter()
		httpRouterHandler.ServeHTTP(rw, req)

		NewWithT(t).Expect(string(rw.MustDumpResponse())).To(Equal(`HTTP/0.0 500 Internal Server Error
Content-Type: application/json; charset=utf-8
X-Meta: service-test@1.0.0/UpdateByID

{"key":"UnknownError","code":500000000,"msg":"UnknownError","desc":"something wrong","canBeTalkError":false,"id":"","sources":["service-test@1.0.0"],"errorFields":null}
`))
	})

	t.Run("return with validate err", func(t *testing.T) {
		rootRouter := courier.NewRouter(httptransport.Group("/root"))
		rootRouter.Register(courier.NewRouter(routes.DataProvider{}, routes.GetByID{}))

		httpRoute := httptransport.NewHttpRouteMeta(rootRouter.Routes()[0])
		httpRouterHandler := httptransport.NewHttpRouteHandler(serviceMeta, httpRoute, rtMgr)

		reqData := routes.DataProvider{
			ID: "10",
		}

		req, err := rtMgr.NewRequest((routes.GetByID{}).Method(), reqData.Path(), reqData)
		NewWithT(t).Expect(err).To(BeNil())

		rw := testify.NewMockResponseWriter()
		httpRouterHandler.ServeHTTP(rw, req)

		NewWithT(t).Expect(string(rw.MustDumpResponse())).To(Equal(`HTTP/0.0 400 Bad Request
Content-Type: application/json; charset=utf-8
X-Meta: service-test@1.0.0/GetByID

{"key":"BadRequest","code":400000000,"msg":"invalid parameters","desc":"","canBeTalkError":false,"id":"","sources":["service-test@1.0.0"],"errorFields":[{"field":"id","msg":"string length should be larger than 6, but got invalid value 2","in":"path"}]}
`))
	})
}
