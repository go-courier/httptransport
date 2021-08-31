package client

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/textproto"
	"reflect"
	"time"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport"
	"github.com/go-courier/httptransport/client/roundtrippers"
	"github.com/go-courier/httptransport/httpx"
	"github.com/go-courier/httptransport/transformers"
	"github.com/go-courier/statuserror"
	contextx "github.com/go-courier/x/context"
	typesutil "github.com/go-courier/x/types"
	"github.com/pkg/errors"
	"golang.org/x/net/http2"
)

type HttpTransport func(rt http.RoundTripper) http.RoundTripper

type Client struct {
	Protocol              string
	Host                  string
	Port                  uint16
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
		c.HttpTransports = []HttpTransport{roundtrippers.NewLogRoundTripper()}
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

type contextKeyClient struct{}

func ContextWithClient(ctx context.Context, c *http.Client) context.Context {
	return contextx.WithValue(ctx, contextKeyClient{}, c)
}

func ClientFromContext(ctx context.Context) *http.Client {
	if ctx == nil {
		return nil
	}
	if c, ok := ctx.Value(contextKeyClient{}).(*http.Client); ok {
		return c
	}
	return nil
}

type contextKeyDefaultHttpTransport struct{}

func ContextWithDefaultHttpTransport(ctx context.Context, t *http.Transport) context.Context {
	return contextx.WithValue(ctx, contextKeyDefaultHttpTransport{}, t)
}

func DefaultHttpTransportFromContext(ctx context.Context) *http.Transport {
	if ctx == nil {
		return nil
	}
	if t, ok := ctx.Value(contextKeyDefaultHttpTransport{}).(*http.Transport); ok {
		return t
	}
	return nil
}

func (c *Client) Do(ctx context.Context, req interface{}, metas ...courier.Metadata) courier.Result {
	request, ok := req.(*http.Request)
	if !ok {
		request2, err := c.newRequest(ctx, req, metas...)
		if err != nil {
			return &Result{
				Err:            statuserror.Wrap(err, http.StatusInternalServerError, "RequestFailed"),
				NewError:       c.NewError,
				TransformerMgr: c.RequestTransformerMgr.TransformerMgr,
			}
		}
		request = request2
	}

	httpClient := ClientFromContext(ctx)
	if httpClient == nil {
		httpClient = GetShortConnClientContext(ctx, c.Timeout, c.HttpTransports...)
	}

	resp, err := httpClient.Do(request)
	if err != nil {
		if errors.Unwrap(err) == context.Canceled {
			return &Result{
				Err:            statuserror.Wrap(err, 499, "ClientClosedRequest"),
				NewError:       c.NewError,
				TransformerMgr: c.RequestTransformerMgr.TransformerMgr,
			}
		}

		return &Result{
			Err:            statuserror.Wrap(err, http.StatusInternalServerError, "RequestFailed"),
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

	request, err := c.RequestTransformerMgr.NewRequestWithContext(ctx, method, c.toUrl(path), req)
	if err != nil {
		return nil, statuserror.Wrap(err, http.StatusBadRequest, "RequestTransformFailed")
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

func (r *Result) Meta() courier.Metadata {
	if r.Response != nil {
		return courier.Metadata(r.Response.Header)
	}
	return courier.Metadata{}
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

		transformer, err := r.TransformerMgr.NewTransformer(context.Background(), typesutil.FromRType(rv.Type()), transformers.TransformerOption{
			MIME: contentType,
		})

		if err != nil {
			return statuserror.Wrap(err, http.StatusInternalServerError, "ReadFailed")
		}

		if e := transformer.DecodeFrom(context.Background(), r.Response.Body, rv, textproto.MIMEHeader(r.Response.Header)); e != nil {
			return statuserror.Wrap(e, http.StatusInternalServerError, "DecodeFailed")
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
		if _, err := io.Copy(v, r.Response.Body); err != nil {
			return meta, statuserror.Wrap(err, http.StatusInternalServerError, "WriteFailed")
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

// Deprecated use GetShortConnClientContext instead
func GetShortConnClient(timeout time.Duration, httpTransports ...HttpTransport) *http.Client {
	return GetShortConnClientContext(context.Background(), timeout, httpTransports...)
}

func GetShortConnClientContext(ctx context.Context, timeout time.Duration, httpTransports ...HttpTransport) *http.Client {
	t := DefaultHttpTransportFromContext(ctx)

	if t != nil {
		t = t.Clone()
	} else {
		t = &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 0,
			}).DialContext,
			DisableKeepAlives:     true,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	}

	if err := http2.ConfigureTransport(t); err != nil {
		panic(err)
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: t,
	}

	for i := range httpTransports {
		httpTransport := httpTransports[i]
		client.Transport = httpTransport(client.Transport)
	}

	return client
}
