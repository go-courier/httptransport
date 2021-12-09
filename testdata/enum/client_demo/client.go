package client_demo

import (
	context "context"

	github_com_go_courier_courier "github.com/go-courier/courier"
)

type ClientDemo interface {
	WithContext(context.Context) ClientDemo
	Context() context.Context
}

func NewClientDemo(c github_com_go_courier_courier.Client) *ClientDemoStruct {
	return &(ClientDemoStruct{
		Client: c,
	})
}

type ClientDemoStruct struct {
	Client github_com_go_courier_courier.Client
	ctx    context.Context
}

func (c *ClientDemoStruct) WithContext(ctx context.Context) ClientDemo {
	cc := new(ClientDemoStruct)
	cc.Client = c.Client
	cc.ctx = ctx
	return cc
}

func (c *ClientDemoStruct) Context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return context.Background()
}
