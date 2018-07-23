package client_demo

import (
	mime_multipart "mime/multipart"

	github_com_go_courier_courier "github.com/go-courier/courier"
)

type Cookie struct {
	Token string `in:"cookie" name:"token,omitempty"`
}

func (Cookie) Path() string {
	return "/demo/cookie"
}

func (Cookie) Method() string {
	return "POST"
}

func (req *Cookie) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return c.Do("demo.Cookie", req, metas...).Into(nil)
}

type Create struct {
	Data Data `in:"body"`
}

func (Create) Path() string {
	return "/demo/restful"
}

func (Create) Method() string {
	return "POST"
}

func (req *Create) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*Data, github_com_go_courier_courier.Metadata, error) {
	resp := new(Data)
	meta, err := c.Do("demo.Create", req, metas...).Into(resp)
	return resp, meta, err
}

type DownloadFile struct {
}

func (DownloadFile) Path() string {
	return "/demo/binary/files"
}

func (DownloadFile) Method() string {
	return "GET"
}

func (req *DownloadFile) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*GithubComGoCourierHttptransportHttpxAttachment, github_com_go_courier_courier.Metadata, error) {
	resp := new(GithubComGoCourierHttptransportHttpxAttachment)
	meta, err := c.Do("demo.DownloadFile", req, metas...).Into(resp)
	return resp, meta, err
}

type FormMultipartWithFile struct {
	FormData struct {
		Data   Data                       `name:"data,omitempty"`
		File   *mime_multipart.FileHeader `name:"file"`
		Slice  []string                   `name:"slice,omitempty"`
		String string                     `name:"string,omitempty"`
	} `in:"body" mime:"multipart"`
}

func (FormMultipartWithFile) Path() string {
	return "/demo/forms/multipart"
}

func (FormMultipartWithFile) Method() string {
	return "POST"
}

func (req *FormMultipartWithFile) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return c.Do("demo.FormMultipartWithFile", req, metas...).Into(nil)
}

type FormMultipartWithFiles struct {
	FormData struct {
		Files []*mime_multipart.FileHeader `name:"files"`
	} `in:"body" mime:"multipart"`
}

func (FormMultipartWithFiles) Path() string {
	return "/demo/forms/multipart-with-files"
}

func (FormMultipartWithFiles) Method() string {
	return "POST"
}

func (req *FormMultipartWithFiles) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return c.Do("demo.FormMultipartWithFiles", req, metas...).Into(nil)
}

type FormURLEncoded struct {
	FormData struct {
		Data   Data     `name:"data"`
		Slice  []string `name:"slice"`
		String string   `name:"string"`
	} `in:"body" mime:"urlencoded"`
}

func (FormURLEncoded) Path() string {
	return "/demo/forms/urlencoded"
}

func (FormURLEncoded) Method() string {
	return "POST"
}

func (req *FormURLEncoded) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return c.Do("demo.FormURLEncoded", req, metas...).Into(nil)
}

type GetByID struct {
	ID    string   `in:"path" name:"id" validate:"@string[6,]"`
	Label []string `in:"query" name:"label,omitempty"`
	Name  string   `in:"query" name:"name,omitempty"`
}

func (GetByID) Path() string {
	return "/demo/restful/{id}"
}

func (GetByID) Method() string {
	return "GET"
}

func (req *GetByID) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*Data, github_com_go_courier_courier.Metadata, error) {
	resp := new(Data)
	meta, err := c.Do("demo.GetByID", req, metas...).Into(resp)
	return resp, meta, err
}

type HealthCheck struct {
}

func (HealthCheck) Path() string {
	return "/demo/restful"
}

func (HealthCheck) Method() string {
	return "HEAD"
}

func (req *HealthCheck) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return c.Do("demo.HealthCheck", req, metas...).Into(nil)
}

type Redirect struct {
}

func (Redirect) Path() string {
	return "/demo/redirect"
}

func (Redirect) Method() string {
	return "GET"
}

func (req *Redirect) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return c.Do("demo.Redirect", req, metas...).Into(nil)
}

type RedirectWhenError struct {
}

func (RedirectWhenError) Path() string {
	return "/demo/redirect"
}

func (RedirectWhenError) Method() string {
	return "POST"
}

func (req *RedirectWhenError) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return c.Do("demo.RedirectWhenError", req, metas...).Into(nil)
}

type RemoveByID struct {
	ID string `in:"path" name:"id" validate:"@string[6,]"`
}

func (RemoveByID) Path() string {
	return "/demo/restful/{id}"
}

func (RemoveByID) Method() string {
	return "DELETE"
}

// @StatusErr[InternalServerError][500100001][InternalServerError]
// @StatusErr[InternalServerError][500999001][InternalServerError]
// @StatusErr[Unauthorized][401999001][Unauthorized]!
func (req *RemoveByID) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return c.Do("demo.RemoveByID", req, metas...).Into(nil)
}

type ShowImage struct {
}

func (ShowImage) Path() string {
	return "/demo/binary/images"
}

func (ShowImage) Method() string {
	return "GET"
}

func (req *ShowImage) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*GithubComGoCourierHttptransportHttpxImagePNG, github_com_go_courier_courier.Metadata, error) {
	resp := new(GithubComGoCourierHttptransportHttpxImagePNG)
	meta, err := c.Do("demo.ShowImage", req, metas...).Into(resp)
	return resp, meta, err
}

type UpdateByID struct {
	ID   string `in:"path" name:"id" validate:"@string[6,]"`
	Data Data   `in:"body"`
}

func (UpdateByID) Path() string {
	return "/demo/restful/{id}"
}

func (UpdateByID) Method() string {
	return "PUT"
}

func (req *UpdateByID) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return c.Do("demo.UpdateByID", req, metas...).Into(nil)
}
