package routes

import (
	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport"
	"github.com/go-courier/httptransport/openapi"
)

var RootRouter = courier.NewRouter(httptransport.BasePath("/demo"))

func init() {
	RootRouter.Register(openapi.OpenAPIRouter)
}
