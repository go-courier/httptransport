package client_degradation_demo

import (
	context "context"

	github_com_go_courier_courier "github.com/go-courier/courier"
	github_com_go_courier_metax "github.com/go-courier/metax"
)

type DemoApi struct {
}

func (DemoApi) Path() string {
	return "/peer/version"
}

func (DemoApi) Method() string {
	return "GET"
}

func (req *DemoApi) Do(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) github_com_go_courier_courier.Result {

	ctx = github_com_go_courier_metax.ContextWith(ctx, "operationID", "degradationDemo.DemoApi")
	return c.Do(ctx, req, metas...)

}

func (req *DemoApi) InvokeContext(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*DemoApiResp, github_com_go_courier_courier.Metadata, error) {
	resp := new(DemoApiResp)

	meta, err := req.Do(ctx, c, metas...).Into(resp)

	return resp, meta, err
}

func (req *DemoApi) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*DemoApiResp, github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(context.Background(), c, metas...)
}
