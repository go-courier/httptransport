package client_demo

import (
	github_com_go_courier_courier "github.com/go-courier/courier"
)

type ClientDemo interface {
	Cookie(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
	Create(req *Create, metas ...github_com_go_courier_courier.Metadata) (*Data, github_com_go_courier_courier.Metadata, error)
	DownloadFile(req *DownloadFile, metas ...github_com_go_courier_courier.Metadata) (*GithubComGoCourierHttptransportHttpxAttachment, github_com_go_courier_courier.Metadata, error)
	FormMultipartWithFile(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
	FormMultipartWithFiles(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
	FormURLEncoded(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
	GetByID(req *GetByID, metas ...github_com_go_courier_courier.Metadata) (*Data, github_com_go_courier_courier.Metadata, error)
	HealthCheck(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
	Redirect(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
	RedirectWhenError(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
	RemoveByID(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
	ShowImage(req *ShowImage, metas ...github_com_go_courier_courier.Metadata) (*GithubComGoCourierHttptransportHttpxImagePNG, github_com_go_courier_courier.Metadata, error)
	UpdateByID(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error)
}

func NewClientDemo(c github_com_go_courier_courier.Client) *ClientDemoStruct {
	return &(ClientDemoStruct{
		Client: c,
	})
}

type ClientDemoStruct struct {
	Client github_com_go_courier_courier.Client
}

func (c *ClientDemoStruct) Cookie(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return (&Cookie{}).Invoke(c.Client)
}

func (c *ClientDemoStruct) Create(req *Create, metas ...github_com_go_courier_courier.Metadata) (*Data, github_com_go_courier_courier.Metadata, error) {
	return req.Invoke(c.Client)
}

func (c *ClientDemoStruct) DownloadFile(req *DownloadFile, metas ...github_com_go_courier_courier.Metadata) (*GithubComGoCourierHttptransportHttpxAttachment, github_com_go_courier_courier.Metadata, error) {
	return req.Invoke(c.Client)
}

func (c *ClientDemoStruct) FormMultipartWithFile(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return (&FormMultipartWithFile{}).Invoke(c.Client)
}

func (c *ClientDemoStruct) FormMultipartWithFiles(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return (&FormMultipartWithFiles{}).Invoke(c.Client)
}

func (c *ClientDemoStruct) FormURLEncoded(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return (&FormURLEncoded{}).Invoke(c.Client)
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

func (c *ClientDemoStruct) RemoveByID(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return (&RemoveByID{}).Invoke(c.Client)
}

func (c *ClientDemoStruct) ShowImage(req *ShowImage, metas ...github_com_go_courier_courier.Metadata) (*GithubComGoCourierHttptransportHttpxImagePNG, github_com_go_courier_courier.Metadata, error) {
	return req.Invoke(c.Client)
}

func (c *ClientDemoStruct) UpdateByID(metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return (&UpdateByID{}).Invoke(c.Client)
}
