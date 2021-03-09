package routes

import (
	"context"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/__examples__/server/pkg/errors"
	"github.com/go-courier/httptransport/__examples__/server/pkg/types"
	"github.com/go-courier/httptransport/httpx"
	perrors "github.com/pkg/errors"

	"github.com/go-courier/httptransport"
)

var RestfulRouter = courier.NewRouter(httptransport.Group("/restful"))

func init() {
	RootRouter.Register(RestfulRouter)

	RestfulRouter.Register(courier.NewRouter(HealthCheck{}))
	RestfulRouter.Register(courier.NewRouter(Create{}))
	RestfulRouter.Register(courier.NewRouter(DataProvider{}, UpdateByID{}))
	RestfulRouter.Register(courier.NewRouter(DataProvider{}, GetByID{}))
	RestfulRouter.Register(courier.NewRouter(DataProvider{}, RemoveByID{}))
}

type HealthCheck struct {
	httpx.MethodHead
	PullPolicy types.PullPolicy `name:"pullPolicy,omitempty" in:"query"`
}

func (HealthCheck) Output(ctx context.Context) (interface{}, error) {
	return nil, nil
}

// Create
type Create struct {
	httpx.MethodPost
	Data Data `in:"body"`
}

func (req Create) Output(ctx context.Context) (interface{}, error) {
	return &req.Data, nil
}

type Data struct {
	ID        string         `json:"id"`
	Label     string         `json:"label"`
	PtrString *string        `json:"ptrString,omitempty"`
	SubData   *SubData       `json:"subData,omitempty"`
	Protocol  types.Protocol `json:"protocol,omitempty"`
}

type SubData struct {
	Name string `json:"name"`
}

// get by id
type GetByID struct {
	httpx.MethodGet
	Protocol types.Protocol `name:"protocol,omitempty" in:"query"`
	Name     string         `name:"name,omitempty" in:"query"`
	Label    []string       `name:"label,omitempty" in:"query"`
}

func (req GetByID) Output(ctx context.Context) (interface{}, error) {
	data := DataFromContext(ctx)
	if len(req.Label) > 0 {
		data.Label = req.Label[0]
	}
	return data, nil
}

// remove by id
type RemoveByID struct {
	httpx.MethodDelete
}

// @StatusErr[InternalServerError][500100001][InternalServerError]
func callWithErr() error {
	return errors.Unauthorized
}

func (RemoveByID) Output(ctx context.Context) (interface{}, error) {
	if false {
		return nil, callWithErr()
	}
	return nil, httpx.WithMetadata(httpx.Metadata("X-Num", "1"))(errors.InternalServerError)
}

// update by id
type UpdateByID struct {
	httpx.MethodPut
	Data Data `in:"body"`
}

func (req UpdateByID) Output(ctx context.Context) (interface{}, error) {
	return nil, perrors.Errorf("something wrong")
}

type DataProvider struct {
	ID string `name:"id" in:"path" validate:"@string[6,]"`
}

func (DataProvider) ContextKey() string {
	return "DataProvider"
}

func (DataProvider) Path() string {
	return "/:id"
}

func DataFromContext(ctx context.Context) *Data {
	return ctx.Value("DataProvider").(*Data)
}

func (req DataProvider) Output(ctx context.Context) (interface{}, error) {
	return &Data{
		ID: req.ID,
	}, nil
}
