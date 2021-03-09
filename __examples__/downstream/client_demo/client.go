package client_demo

import (
	context "context"

	github_com_go_courier_courier "github.com/go-courier/courier"
)

type ClientDemo interface {
	WithContext(context.Context) ClientDemo
	Context() context.Context
	Cookie(req *Cookie, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
	Create(req *Create, metas ...github_com_go_courier_courier.Metadata) (*Data, github_com_go_courier_courier.Metadata, error)
	DownloadFile(metas ...github_com_go_courier_courier.Metadata) (*GithubComGoCourierHttptransportHttpxAttachment, github_com_go_courier_courier.Metadata, error)
	FormMultipartWithFile(req *FormMultipartWithFile, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
	FormMultipartWithFiles(req *FormMultipartWithFiles, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
	FormURLEncoded(req *FormURLEncoded, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
	GetByID(req *GetByID, metas ...github_com_go_courier_courier.Metadata) (*Data, github_com_go_courier_courier.Metadata, error)
	HealthCheck(req *HealthCheck, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
	Proxy(metas ...github_com_go_courier_courier.Metadata) (*IpInfo, github_com_go_courier_courier.Metadata, error)
	ProxyV2(metas ...github_com_go_courier_courier.Metadata) (*IpInfo, github_com_go_courier_courier.Metadata, error)
	Redirect(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
	RedirectWhenError(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
	RemoveByID(req *RemoveByID, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
	ShowImage(metas ...github_com_go_courier_courier.Metadata) (*GithubComGoCourierHttptransportHttpxImagePNG, github_com_go_courier_courier.Metadata, error)
	UpdateByID(req *UpdateByID, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
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

func (c *ClientDemoStruct) Cookie(req *Cookie, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(c.Context(), c.Client, metas...)
}

func (c *ClientDemoStruct) Create(req *Create, metas ...github_com_go_courier_courier.Metadata) (*Data, github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(c.Context(), c.Client, metas...)
}

func (c *ClientDemoStruct) DownloadFile(metas ...github_com_go_courier_courier.Metadata) (*GithubComGoCourierHttptransportHttpxAttachment, github_com_go_courier_courier.Metadata, error) {
	return (&DownloadFile{}).InvokeContext(c.Context(), c.Client, metas...)
}

func (c *ClientDemoStruct) FormMultipartWithFile(req *FormMultipartWithFile, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(c.Context(), c.Client, metas...)
}

func (c *ClientDemoStruct) FormMultipartWithFiles(req *FormMultipartWithFiles, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(c.Context(), c.Client, metas...)
}

func (c *ClientDemoStruct) FormURLEncoded(req *FormURLEncoded, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(c.Context(), c.Client, metas...)
}

func (c *ClientDemoStruct) GetByID(req *GetByID, metas ...github_com_go_courier_courier.Metadata) (*Data, github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(c.Context(), c.Client, metas...)
}

func (c *ClientDemoStruct) HealthCheck(req *HealthCheck, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(c.Context(), c.Client, metas...)
}

func (c *ClientDemoStruct) Proxy(metas ...github_com_go_courier_courier.Metadata) (*IpInfo, github_com_go_courier_courier.Metadata, error) {
	return (&Proxy{}).InvokeContext(c.Context(), c.Client, metas...)
}

func (c *ClientDemoStruct) ProxyV2(metas ...github_com_go_courier_courier.Metadata) (*IpInfo, github_com_go_courier_courier.Metadata, error) {
	return (&ProxyV2{}).InvokeContext(c.Context(), c.Client, metas...)
}

func (c *ClientDemoStruct) Redirect(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return (&Redirect{}).InvokeContext(c.Context(), c.Client, metas...)
}

func (c *ClientDemoStruct) RedirectWhenError(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return (&RedirectWhenError{}).InvokeContext(c.Context(), c.Client, metas...)
}

func (c *ClientDemoStruct) RemoveByID(req *RemoveByID, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(c.Context(), c.Client, metas...)
}

func (c *ClientDemoStruct) ShowImage(metas ...github_com_go_courier_courier.Metadata) (*GithubComGoCourierHttptransportHttpxImagePNG, github_com_go_courier_courier.Metadata, error) {
	return (&ShowImage{}).InvokeContext(c.Context(), c.Client, metas...)
}

func (c *ClientDemoStruct) UpdateByID(req *UpdateByID, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(c.Context(), c.Client, metas...)
}
