package httptransport_test

import (
	"fmt"
	"os"
	"sort"

	"github.com/fatih/color"
	"github.com/go-courier/httptransport"
	"github.com/go-courier/httptransport/__examples__/server/cmd/app/routes"
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
		fmt.Println(httpRouteMeta.String())
	}
	// Output:
	// GET /demo openapi.OpenAPI
	// GET /demo/binary/files routes.DownloadFile
	// GET /demo/binary/images routes.ShowImage
	// POS /demo/cookie routes.Cookie
	// POS /demo/forms/multipart routes.FormMultipartWithFile
	// POS /demo/forms/multipart-with-files routes.FormMultipartWithFiles
	// POS /demo/forms/urlencoded routes.FormURLEncoded
	// GET /demo/proxy routes.Proxy
	// GET /demo/redirect routes.Redirect
	// POS /demo/redirect routes.RedirectWhenError
	// POS /demo/restful routes.Create
	// HEA /demo/restful routes.HealthCheck
	// GET /demo/restful/{id} routes.DataProvider routes.GetByID
	// DEL /demo/restful/{id} routes.DataProvider routes.RemoveByID
	// PUT /demo/restful/{id} routes.DataProvider routes.UpdateByID
	// GET /demo/v2/proxy routes.ProxyV2
}
