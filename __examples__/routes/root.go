package routes

import (
	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/openapi"

	"github.com/go-courier/httptransport"
)

var RootRouter = courier.NewRouter(httptransport.Group("/demo"))

func init() {
	RootRouter.Register(openapi.OpenAPIRouter)
}
