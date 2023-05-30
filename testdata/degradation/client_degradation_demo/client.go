package client_degradation_demo

import (
	context "context"

	github_com_go_courier_courier "github.com/go-courier/courier"
)

type ClientDegradationDemo interface {
	WithContext(context.Context) ClientDegradationDemo
	Context() context.Context
	DemoApi(metas ...github_com_go_courier_courier.Metadata) (*DemoApiResp, github_com_go_courier_courier.Metadata, error)
}

func NewClientDegradationDemo(c github_com_go_courier_courier.Client) *ClientDegradationDemoStruct {
	return &(ClientDegradationDemoStruct{
		Client: c,
	})
}

type ClientDegradationDemoStruct struct {
	Client github_com_go_courier_courier.Client
	ctx    context.Context
}

func (c *ClientDegradationDemoStruct) WithContext(ctx context.Context) ClientDegradationDemo {
	cc := new(ClientDegradationDemoStruct)
	cc.Client = c.Client
	cc.ctx = ctx
	return cc
}

func (c *ClientDegradationDemoStruct) Context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return context.Background()
}

func (c *ClientDegradationDemoStruct) DemoApi(metas ...github_com_go_courier_courier.Metadata) (*DemoApiResp, github_com_go_courier_courier.Metadata, error) {
	return (&DemoApi{}).InvokeContext(c.Context(), c.Client, metas...)
}
