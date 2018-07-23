package httptransport_test

import (
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/fatih/color"
	"github.com/go-courier/courier"
	"github.com/stretchr/testify/require"

	"github.com/go-courier/httptransport"
	"github.com/go-courier/httptransport/__examples__/routes"
)

func ExampleGroup() {
	g := httptransport.Group("/test")
	fmt.Println(g.Path())
	// Output:
	// /test
}

func ExampleHttpRouteMeta() {
	os.Setenv("PROJECT_NAME", "service-example")
	os.Setenv("PROJECT_VERSION", "1.0.0")
	color.NoColor = true

	routeList := routes.RootRouter.Routes()

	sort.Slice(routeList, func(i, j int) bool {
		return httptransport.NewHttpRouteMeta(routeList[i]).Key() <
			httptransport.NewHttpRouteMeta(routeList[j]).Key()
	})

	for i := range routeList {
		httpRouteMeta := httptransport.NewHttpRouteMeta(routeList[i])
		fmt.Println(fmt.Sprintf(httpRouteMeta.String()))
	}
	// Output:
	// GET /demo openapi.OpenAPI
	// GET /demo/binary/files routes.DownloadFile
	// GET /demo/binary/images routes.ShowImage
	// POS /demo/cookie routes.Cookie
	// POS /demo/forms/multipart routes.FormMultipartWithFile
	// POS /demo/forms/multipart-with-files routes.FormMultipartWithFiles
	// POS /demo/forms/urlencoded routes.FormURLEncoded
	// GET /demo/redirect routes.Redirect
	// POS /demo/redirect routes.RedirectWhenError
	// POS /demo/restful routes.Create
	// HEA /demo/restful routes.HealthCheck
	// GET /demo/restful/{id} routes.DataProvider routes.GetByID
	// DEL /demo/restful/{id} routes.DataProvider routes.RemoveByID
	// PUT /demo/restful/{id} routes.DataProvider routes.UpdateByID
}

func TestNewHttpRouteMeta(t *testing.T) {
	rootRouter := courier.NewRouter(httptransport.Group("/test"))
	rootRouter.Register(courier.NewRouter(httptransport.Group("/sub")))

	require.Error(t, httptransport.TryCatch(func() {
		for _, route := range rootRouter.Routes() {
			httptransport.NewHttpRouteMeta(route).Key()
		}
	}))

	require.Error(t, httptransport.TryCatch(func() {
		for _, route := range rootRouter.Routes() {
			httptransport.NewHttpRouteMeta(route).Method()
		}
	}))
}
