package client_demo

import (
	context "context"
	mime_multipart "mime/multipart"

	github_com_go_courier_courier "github.com/go-courier/courier"
	github_com_go_courier_metax "github.com/go-courier/metax"
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

// @StatusErr[ContextCanceled][499000000][ContextCanceled]
// @StatusErr[UnknownError][500000000][UnknownError]
func (req *Cookie) Do(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) github_com_go_courier_courier.Result {

	ctx = github_com_go_courier_metax.ContextWith(ctx, "operationID", "demo.Cookie")
	return c.Do(ctx, req, metas...)

}

func (req *Cookie) InvokeContext(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.Do(ctx, c, metas...).Into(nil)
}

func (req *Cookie) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(context.Background(), c, metas...)
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

func (req *Create) Do(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) github_com_go_courier_courier.Result {

	ctx = github_com_go_courier_metax.ContextWith(ctx, "operationID", "demo.Create")
	return c.Do(ctx, req, metas...)

}

func (req *Create) InvokeContext(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*Data, github_com_go_courier_courier.Metadata, error) {
	resp := new(Data)

	meta, err := req.Do(ctx, c, metas...).Into(resp)

	return resp, meta, err
}

func (req *Create) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*Data, github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(context.Background(), c, metas...)
}

type DownloadFile struct {
}

func (DownloadFile) Path() string {
	return "/demo/binary/files"
}

func (DownloadFile) Method() string {
	return "GET"
}

func (req *DownloadFile) Do(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) github_com_go_courier_courier.Result {

	ctx = github_com_go_courier_metax.ContextWith(ctx, "operationID", "demo.DownloadFile")
	return c.Do(ctx, req, metas...)

}

func (req *DownloadFile) InvokeContext(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*GithubComGoCourierHttptransportHttpxAttachment, github_com_go_courier_courier.Metadata, error) {
	resp := new(GithubComGoCourierHttptransportHttpxAttachment)

	meta, err := req.Do(ctx, c, metas...).Into(resp)

	return resp, meta, err
}

func (req *DownloadFile) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*GithubComGoCourierHttptransportHttpxAttachment, github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(context.Background(), c, metas...)
}

type FormMultipartWithFile struct {
	FormData struct {
		Data  Data                                                                    `name:"data,omitempty"`
		File  *mime_multipart.FileHeader                                              `name:"file"`
		Map   map[GithubComGoCourierHttptransportExamplesServerPkgTypesProtocol]int32 `name:"map,omitempty"`
		Slice []string                                                                `name:"slice,omitempty"`
		// @deprecated
		String string `name:"string,omitempty"`
	} `in:"body" mime:"multipart"`
}

func (FormMultipartWithFile) Path() string {
	return "/demo/forms/multipart"
}

func (FormMultipartWithFile) Method() string {
	return "POST"
}

func (req *FormMultipartWithFile) Do(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) github_com_go_courier_courier.Result {

	ctx = github_com_go_courier_metax.ContextWith(ctx, "operationID", "demo.FormMultipartWithFile")
	return c.Do(ctx, req, metas...)

}

func (req *FormMultipartWithFile) InvokeContext(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.Do(ctx, c, metas...).Into(nil)
}

func (req *FormMultipartWithFile) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(context.Background(), c, metas...)
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

func (req *FormMultipartWithFiles) Do(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) github_com_go_courier_courier.Result {

	ctx = github_com_go_courier_metax.ContextWith(ctx, "operationID", "demo.FormMultipartWithFiles")
	return c.Do(ctx, req, metas...)

}

func (req *FormMultipartWithFiles) InvokeContext(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.Do(ctx, c, metas...).Into(nil)
}

func (req *FormMultipartWithFiles) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(context.Background(), c, metas...)
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

func (req *FormURLEncoded) Do(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) github_com_go_courier_courier.Result {

	ctx = github_com_go_courier_metax.ContextWith(ctx, "operationID", "demo.FormURLEncoded")
	return c.Do(ctx, req, metas...)

}

func (req *FormURLEncoded) InvokeContext(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.Do(ctx, c, metas...).Into(nil)
}

func (req *FormURLEncoded) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(context.Background(), c, metas...)
}

type GetByID struct {
	ID       string                                                        `in:"path" name:"id" validate:"@string[6,]"`
	Label    []string                                                      `in:"query" name:"label,omitempty"`
	Name     string                                                        `in:"query" name:"name,omitempty"`
	Protocol GithubComGoCourierHttptransportExamplesServerPkgTypesProtocol `in:"query" name:"protocol,omitempty"`
}

func (GetByID) Path() string {
	return "/demo/restful/:id"
}

func (GetByID) Method() string {
	return "GET"
}

func (req *GetByID) Do(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) github_com_go_courier_courier.Result {

	ctx = github_com_go_courier_metax.ContextWith(ctx, "operationID", "demo.GetByID")
	return c.Do(ctx, req, metas...)

}

func (req *GetByID) InvokeContext(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*Data, github_com_go_courier_courier.Metadata, error) {
	resp := new(Data)

	meta, err := req.Do(ctx, c, metas...).Into(resp)

	return resp, meta, err
}

func (req *GetByID) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*Data, github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(context.Background(), c, metas...)
}

type HealthCheck struct {
	PullPolicy GithubComGoCourierHttptransportExamplesServerPkgTypesPullPolicy `in:"query" name:"pullPolicy,omitempty"`
}

func (HealthCheck) Path() string {
	return "/demo/restful"
}

func (HealthCheck) Method() string {
	return "HEAD"
}

func (req *HealthCheck) Do(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) github_com_go_courier_courier.Result {

	ctx = github_com_go_courier_metax.ContextWith(ctx, "operationID", "demo.HealthCheck")
	return c.Do(ctx, req, metas...)

}

func (req *HealthCheck) InvokeContext(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.Do(ctx, c, metas...).Into(nil)
}

func (req *HealthCheck) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(context.Background(), c, metas...)
}

type Proxy struct {
}

func (Proxy) Path() string {
	return "/demo/proxy"
}

func (Proxy) Method() string {
	return "GET"
}

// @StatusErr[ClientClosedRequest][499000000][ClientClosedRequest]
// @StatusErr[RequestFailed][500000000][RequestFailed]
// @StatusErr[RequestTransformFailed][400000000][RequestTransformFailed]
func (req *Proxy) Do(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) github_com_go_courier_courier.Result {

	ctx = github_com_go_courier_metax.ContextWith(ctx, "operationID", "demo.Proxy")
	return c.Do(ctx, req, metas...)

}

func (req *Proxy) InvokeContext(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*IpInfo, github_com_go_courier_courier.Metadata, error) {
	resp := new(IpInfo)

	meta, err := req.Do(ctx, c, metas...).Into(resp)

	return resp, meta, err
}

func (req *Proxy) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*IpInfo, github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(context.Background(), c, metas...)
}

type ProxyV2 struct {
}

func (ProxyV2) Path() string {
	return "/demo/v2/proxy"
}

func (ProxyV2) Method() string {
	return "GET"
}

// @StatusErr[ClientClosedRequest][499000000][ClientClosedRequest]
// @StatusErr[ContextCanceled][499000000][ContextCanceled]
// @StatusErr[RequestFailed][500000000][RequestFailed]
// @StatusErr[RequestTransformFailed][400000000][RequestTransformFailed]
// @StatusErr[UnknownError][500000000][UnknownError]
func (req *ProxyV2) Do(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) github_com_go_courier_courier.Result {

	ctx = github_com_go_courier_metax.ContextWith(ctx, "operationID", "demo.ProxyV2")
	return c.Do(ctx, req, metas...)

}

func (req *ProxyV2) InvokeContext(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*IpInfo, github_com_go_courier_courier.Metadata, error) {
	resp := new(IpInfo)

	meta, err := req.Do(ctx, c, metas...).Into(resp)

	return resp, meta, err
}

func (req *ProxyV2) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*IpInfo, github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(context.Background(), c, metas...)
}

type Redirect struct {
}

func (Redirect) Path() string {
	return "/demo/redirect"
}

func (Redirect) Method() string {
	return "GET"
}

func (req *Redirect) Do(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) github_com_go_courier_courier.Result {

	ctx = github_com_go_courier_metax.ContextWith(ctx, "operationID", "demo.Redirect")
	return c.Do(ctx, req, metas...)

}

func (req *Redirect) InvokeContext(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.Do(ctx, c, metas...).Into(nil)
}

func (req *Redirect) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(context.Background(), c, metas...)
}

type RedirectWhenError struct {
}

func (RedirectWhenError) Path() string {
	return "/demo/redirect"
}

func (RedirectWhenError) Method() string {
	return "POST"
}

func (req *RedirectWhenError) Do(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) github_com_go_courier_courier.Result {

	ctx = github_com_go_courier_metax.ContextWith(ctx, "operationID", "demo.RedirectWhenError")
	return c.Do(ctx, req, metas...)

}

func (req *RedirectWhenError) InvokeContext(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.Do(ctx, c, metas...).Into(nil)
}

func (req *RedirectWhenError) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(context.Background(), c, metas...)
}

type RemoveByID struct {
	ID string `in:"path" name:"id" validate:"@string[6,]"`
}

func (RemoveByID) Path() string {
	return "/demo/restful/:id"
}

func (RemoveByID) Method() string {
	return "DELETE"
}

func (req *RemoveByID) Do(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) github_com_go_courier_courier.Result {

	ctx = github_com_go_courier_metax.ContextWith(ctx, "operationID", "demo.RemoveByID")
	return c.Do(ctx, req, metas...)

}

func (req *RemoveByID) InvokeContext(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.Do(ctx, c, metas...).Into(nil)
}

func (req *RemoveByID) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(context.Background(), c, metas...)
}

type ShowImage struct {
}

func (ShowImage) Path() string {
	return "/demo/binary/images"
}

func (ShowImage) Method() string {
	return "GET"
}

func (req *ShowImage) Do(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) github_com_go_courier_courier.Result {

	ctx = github_com_go_courier_metax.ContextWith(ctx, "operationID", "demo.ShowImage")
	return c.Do(ctx, req, metas...)

}

func (req *ShowImage) InvokeContext(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*GithubComGoCourierHttptransportHttpxImagePNG, github_com_go_courier_courier.Metadata, error) {
	resp := new(GithubComGoCourierHttptransportHttpxImagePNG)

	meta, err := req.Do(ctx, c, metas...).Into(resp)

	return resp, meta, err
}

func (req *ShowImage) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*GithubComGoCourierHttptransportHttpxImagePNG, github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(context.Background(), c, metas...)
}

type UpdateByID struct {
	ID   string `in:"path" name:"id" validate:"@string[6,]"`
	Data Data   `in:"body"`
}

func (UpdateByID) Path() string {
	return "/demo/restful/:id"
}

func (UpdateByID) Method() string {
	return "PUT"
}

func (req *UpdateByID) Do(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) github_com_go_courier_courier.Result {

	ctx = github_com_go_courier_metax.ContextWith(ctx, "operationID", "demo.UpdateByID")
	return c.Do(ctx, req, metas...)

}

func (req *UpdateByID) InvokeContext(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.Do(ctx, c, metas...).Into(nil)
}

func (req *UpdateByID) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(context.Background(), c, metas...)
}
