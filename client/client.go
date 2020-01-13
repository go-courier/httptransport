package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/textproto"
	"reflect"
	"strings"
	"time"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/client/roundtrippers"
	"github.com/go-courier/httptransport/httpx"
	"github.com/go-courier/httptransport/transformers"
	"github.com/go-courier/reflectx/typesutil"
	"github.com/go-courier/statuserror"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"

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
	NewError              func(resp *http.Response) error
}

func (c *Client) SetDefaults() {
	if c.RequestTransformerMgr == nil {
		c.RequestTransformerMgr = httptransport.NewRequestTransformerMgr(nil, nil)
		c.RequestTransformerMgr.SetDefaults()
	}
	if c.HttpTransports == nil {
		c.HttpTransports = []HttpTransport{roundtrippers.NewLogRoundTripper(logrus.WithField("client", ""))}
	}
	if c.NewError == nil {
		c.NewError = func(resp *http.Response) error {
			return &statuserror.StatusErr{
				Code:    resp.StatusCode * 1e6,
				Msg:     resp.Status,
				Sources: []string{resp.Request.Host},
			}
		}
	}
}

func ContextWithClient(ctx context.Context, c *http.Client) context.Context {
	return context.WithValue(ctx, "courier.Client", c)
}

func ClientFromContext(ctx context.Context) *http.Client {
	if ctx == nil {
		return nil
	}
	if c, ok := ctx.Value("courier.Client").(*http.Client); ok {
		return c
	}
	return nil
}

func (c *Client) Do(ctx context.Context, req interface{}, metas ...courier.Metadata) courier.Result {
	request, ok := req.(*http.Request)
	if !ok {
		request2, err := c.newRequest(ctx, req, metas...)
		if err != nil {
			return &Result{
				Err:            RequestFailed.StatusErr().WithDesc(err.Error()),
				NewError:       c.NewError,
				TransformerMgr: c.RequestTransformerMgr.TransformerMgr,
			}
		}
		request = request2
	}

	httpClient := ClientFromContext(ctx)
	if httpClient == nil {
		httpClient = GetShortConnClient(c.Timeout, c.HttpTransports...)
	}

	resp, err := httpClient.Do(request)
	if err != nil {
		if errors.Unwrap(err) == context.Canceled {
			return &Result{
				Err:            ClientClosedRequest.StatusErr().WithDesc(err.Error()),
				NewError:       c.NewError,
				TransformerMgr: c.RequestTransformerMgr.TransformerMgr,
			}
		}

		return &Result{
			Err:            RequestFailed.StatusErr().WithDesc(err.Error()),
			NewError:       c.NewError,
			TransformerMgr: c.RequestTransformerMgr.TransformerMgr,
		}
	}
	return &Result{
		NewError:       c.NewError,
		TransformerMgr: c.RequestTransformerMgr.TransformerMgr,
		Response:       resp,
	}
}

func (c *Client) toUrl(path string) string {
	protocol := c.Protocol
	if protocol == "" {
		protocol = "http"
	}
	url := fmt.Sprintf("%s://%s", protocol, c.Host)
	if c.Port > 0 {
		url = fmt.Sprintf("%s:%d", url, c.Port)
	}
	return url + path
}

func (c *Client) newRequest(ctx context.Context, req interface{}, metas ...courier.Metadata) (*http.Request, error) {
	if ctx == nil {
		ctx = context.Background()
	}

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

	request = request.WithContext(ctx)

	for k, vs := range courier.FromMetas(metas...) {
		for _, v := range vs {
			request.Header.Add(k, v)
		}
	}

	return request, nil
}

type Result struct {
	TransformerMgr transformers.TransformerMgr
	Response       *http.Response
	NewError       func(resp *http.Response) error
	Err            error
}

func (r *Result) StatusCode() int {
	if r.Response != nil {
		return r.Response.StatusCode
	}
	return 0
}

func (r *Result) Into(body interface{}) (courier.Metadata, error) {
	defer func() {
		if r.Response != nil && r.Response.Body != nil {
			r.Response.Body.Close()
		}
	}()

	if r.Err != nil {
		return nil, r.Err
	}

	meta := courier.Metadata(r.Response.Header)

	if !isOk(r.Response.StatusCode) {
		body = r.NewError(r.Response)
	}

	if body == nil {
		return meta, nil
	}

	decode := func(body interface{}) error {
		contentType := meta.Get(httpx.HeaderContentType)

		if contentType != "" {
			contentType, _, _ = mime.ParseMediaType(contentType)
		}

		rv := reflect.ValueOf(body)

		transformer, err := r.TransformerMgr.NewTransformer(nil, typesutil.FromRType(rv.Type()), transformers.TransformerOption{
			MIME: contentType,
		})

		if err != nil {
			return ReadFailed.StatusErr().WithDesc(err.Error())
		}

		if e := transformer.DecodeFromReader(r.Response.Body, rv, textproto.MIMEHeader(r.Response.Header)); e != nil {
			return ReadFailed.StatusErr().WithDesc(e.Error())
		}

		return nil
	}

	switch v := body.(type) {
	case error:
		// to unmarshal status error
		if err := decode(v); err != nil {
			return meta, err
		}
		return meta, v
	case io.Writer:
		if respWriter, ok := body.(interface{ Header() http.Header }); ok {
			header := respWriter.Header()
			for k, v := range meta {
				if strings.HasPrefix(k, "Content-") {
					header[k] = v
				}
			}
		}
		if _, err := io.Copy(v, r.Response.Body); err != nil {
			return meta, ReadFailed.StatusErr().WithDesc(err.Error())
		}
	default:
		if err := decode(body); err != nil {
			return meta, err
		}
	}

	return meta, nil
}

func isOk(code int) bool {
	return code >= http.StatusOK && code < http.StatusMultipleChoices
}

func GetShortConnClient(timeout time.Duration, httpTransports ...HttpTransport) *http.Client {
	t := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: 0,
		}).DialContext,
		DisableKeepAlives: true,
	}

	if err := http2.ConfigureTransport(t); err != nil {
		panic(err)
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: t,
	}

	if httpTransports != nil {
		for i := range httpTransports {
			httpTransport := httpTransports[i]
			client.Transport = httpTransport(client.Transport)
		}
	}

	return client
}
