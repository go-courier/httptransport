package client_demo

import (
	github_com_go_courier_courier "github.com/go-courier/courier"
)

type ClientDemo interface {
	Cookie(req *Cookie, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
	Create(req *Create, metas ...github_com_go_courier_courier.Metadata) (*Data, github_com_go_courier_courier.Metadata, error)
	DownloadFile(metas ...github_com_go_courier_courier.Metadata) (*GithubComGoCourierHttptransportHttpxAttachment, github_com_go_courier_courier.Metadata, error)
	FormMultipartWithFile(req *FormMultipartWithFile, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
	FormMultipartWithFiles(req *FormMultipartWithFiles, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
	FormURLEncoded(req *FormURLEncoded, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
	GetByID(req *GetByID, metas ...github_com_go_courier_courier.Metadata) (*Data, github_com_go_courier_courier.Metadata, error)
	HealthCheck(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
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
}

func (c *ClientDemoStruct) Cookie(req *Cookie, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.Invoke(c.Client)
}

func (c *ClientDemoStruct) Create(req *Create, metas ...github_com_go_courier_courier.Metadata) (*Data, github_com_go_courier_courier.Metadata, error) {
	return req.Invoke(c.Client)
}

func (c *ClientDemoStruct) DownloadFile(metas ...github_com_go_courier_courier.Metadata) (*GithubComGoCourierHttptransportHttpxAttachment, github_com_go_courier_courier.Metadata, error) {
	return (&DownloadFile{}).Invoke(c.Client)
}

func (c *ClientDemoStruct) FormMultipartWithFile(req *FormMultipartWithFile, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.Invoke(c.Client)
}

func (c *ClientDemoStruct) FormMultipartWithFiles(req *FormMultipartWithFiles, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.Invoke(c.Client)
}

func (c *ClientDemoStruct) FormURLEncoded(req *FormURLEncoded, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.Invoke(c.Client)
}

func (c *ClientDemoStruct) GetByID(req *GetByID, metas ...github_com_go_courier_courier.Metadata) (*Data, github_com_go_courier_courier.Metadata, error) {
	return req.Invoke(c.Client)
}

func (c *ClientDemoStruct) HealthCheck(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return (&HealthCheck{}).Invoke(c.Client)
}

func (c *ClientDemoStruct) Redirect(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return (&Redirect{}).Invoke(c.Client)
}

func (c *ClientDemoStruct) RedirectWhenError(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return (&RedirectWhenError{}).Invoke(c.Client)
}

func (c *ClientDemoStruct) RemoveByID(req *RemoveByID, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.Invoke(c.Client)
}

func (c *ClientDemoStruct) ShowImage(metas ...github_com_go_courier_courier.Metadata) (*GithubComGoCourierHttptransportHttpxImagePNG, github_com_go_courier_courier.Metadata, error) {
	return (&ShowImage{}).Invoke(c.Client)
}

func (c *ClientDemoStruct) UpdateByID(req *UpdateByID, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.Invoke(c.Client)
}
