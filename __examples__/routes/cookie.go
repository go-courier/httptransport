package routes

import (
	"context"
	"net/http"
	"time"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/httpx"

	"github.com/go-courier/httptransport"
)

var CookieRouter = courier.NewRouter(httptransport.Group("/cookie"))

func init() {
	RootRouter.Register(CookieRouter)

	CookieRouter.Register(courier.NewRouter(&Cookie{}))
}

type Cookie struct {
	httpx.MethodPost
	Token string `name:"token,omitempty" in:"cookie"`
}

func (req *Cookie) Output(ctx context.Context) (interface{}, error) {
	return httpx.Compose(
		httpx.WithCookies(&http.Cookie{
			Name:    "token",
			Value:   req.Token,
			Expires: time.Now().Add(24 * time.Hour),
		}),
		httpx.WithStatusCode(http.StatusNoContent),
	)(nil), nil
}
