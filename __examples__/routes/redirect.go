package routes

import (
	"context"
	"net/url"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/httpx"

	"github.com/go-courier/httptransport"
)

var RedirectRouter = courier.NewRouter(httptransport.Group("/redirect"))

func init() {
	RootRouter.Register(RedirectRouter)
	RedirectRouter.Register(courier.NewRouter(Redirect{}))
	RedirectRouter.Register(courier.NewRouter(RedirectWhenError{}))
}

type Redirect struct {
	httpx.MethodGet
}

func (Redirect) Output(ctx context.Context) (interface{}, error) {
	return httpx.RedirectWithStatusFound(&url.URL{
		Path: "/other",
	}), nil
}

type RedirectWhenError struct {
	httpx.MethodPost
}

func (RedirectWhenError) Output(ctx context.Context) (interface{}, error) {
	return nil, httpx.RedirectWithStatusMovedPermanently(&url.URL{
		Path: "/other",
	})
}
