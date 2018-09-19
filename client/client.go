package client

import (
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/textproto"
	"reflect"
	"time"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/client/roundtrippers"
	"github.com/go-courier/httptransport/httpx"
	"github.com/go-courier/httptransport/transformers"
	"github.com/go-courier/reflectx/typesutil"
	"github.com/go-courier/statuserror"
	"github.com/sirupsen/logrus"

	"github.com/go-courier/httptransport"
)

type HttpTransport func(rt http.RoundTripper) http.RoundTripper

type Client struct {
	Protocol              string
	Host                  string
	Port                  int16
	Timeout               time.Duration
	RequestTransformerMgr *httptransport.RequestTransformerMgr
	HttpTransports        []HttpTransport
	NewError              func() error
}

func (c *Client) SetDefaults() {
	if c.RequestTransformerMgr == nil {
		c.RequestTransformerMgr = httptransport.NewRequestTransformerMgr(nil, nil)
		c.RequestTransformerMgr.SetDefaults()
	}
	if c.HttpTransports == nil {
		c.HttpTransports = []HttpTransport{roundtrippers.NewLogRoundTripper(logrus.StandardLogger())}
	}
	if c.NewError == nil {
		c.NewError = func() error {
			return &statuserror.StatusErr{}
		}
	}
}

func (c *Client) Do(operationID string, req interface{}, metas ...courier.Metadata) courier.Result {
	request, err := c.newRequest(operationID, req, metas...)
	if err != nil {
		return &Result{
			Err: RequestFailed.StatusErr().WithDesc(err.Error()),
		}
	}

	httpClient := GetShortConnClient(c.Timeout, c.HttpTransports...)
	resp, err := httpClient.Do(request)
	if err != nil {
		return &Result{
			Err: RequestFailed.StatusErr().WithDesc(err.Error()),
		}
	}
	return &Result{
		Response:       resp,
		transformerMgr: c.RequestTransformerMgr.TransformerMgr,
	}
}

func (c *Client) toUrl(path string) string {
	if c.Protocol == "" {
		c.Protocol = "http"
	}
	url := fmt.Sprintf("%s://%s", c.Protocol, c.Host)
	if c.Port > 0 {
		url = fmt.Sprintf("%s:%d", url, c.Port)
	}
	return url + path
}

func (c *Client) newRequest(operationID string, req interface{}, metas ...courier.Metadata) (*http.Request, error) {
	method := ""
	path := ""

	if methodDescriber, ok := req.(httptransport.MethodDescriber); ok {
		method = methodDescriber.Method()
	}

	if pathDescriber, ok := req.(httptransport.PathDescriber); ok {
		path = pathDescriber.Path()
	}

	request, err := c.RequestTransformerMgr.NewRequest(method, c.toUrl(path), req)
	if err != nil {
		return nil, RequestTransformFailed.StatusErr().WithDesc(err.Error())
	}
	for k, vs := range courier.FromMetas(metas...) {
		for _, v := range vs {
			request.Header.Add(k, v)
		}
	}
	request.Header.Add("X-Operation-Id", operationID)
	return request, nil
}

type Result struct {
	transformerMgr transformers.TransformerMgr
	NewError       func() error
	*http.Response
	Err error
}

func (r *Result) Into(body interface{}) (courier.Metadata, error) {
	defer r.Body.Close()

	if r.Err != nil {
		return nil, r.Err
	}

	meta := courier.Metadata(r.Header)

	if !isOk(r.StatusCode) {
		body = r.NewError()
	}

	if body == nil {
		return meta, nil
	}

	contentType := meta.Get(httpx.HeaderContentType)

	if contentType != "" {
		contentType, _, _ = mime.ParseMediaType(contentType)
	}

	if writer, ok := body.(io.Writer); ok {
		if _, err := io.Copy(writer, r.Body); err != nil {
			return meta, ReadFailed.StatusErr().WithDesc(err.Error())
		}
	} else {
		rv := reflect.ValueOf(body)
		transformer, err := r.transformerMgr.NewTransformer(typesutil.FromRType(rv.Type()), transformers.TransformerOption{
			MIME: contentType,
		})

		if err != nil {
			return meta, ReadFailed.StatusErr().WithDesc(err.Error())
		}

		if err := transformer.DecodeFromReader(r.Body, rv, textproto.MIMEHeader(r.Header)); err != nil {
			return meta, ReadFailed.StatusErr().WithDesc(err.Error())
		}
	}

	if err, ok := body.(error); ok {
		return meta, err
	}

	return meta, nil
}

func isOk(code int) bool {
	return code >= http.StatusOK && code < http.StatusMultipleChoices
}

func GetShortConnClient(timeout time.Duration, httpTransports ...HttpTransport) *http.Client {
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: 0,
			}).DialContext,
			DisableKeepAlives: true,
		},
	}

	if httpTransports != nil {
		for i := range httpTransports {
			httpTransport := httpTransports[i]
			client.Transport = httpTransport(client.Transport)
		}
	}

	return client
}
