package testdata

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	contextx "github.com/go-courier/x/context"
	"github.com/julienschmidt/httprouter"

	"github.com/go-courier/httptransport"
	"github.com/go-courier/httptransport/httpx"
	"github.com/go-courier/httptransport/testdata/server/pkg/types"
	"github.com/go-courier/httptransport/transformers"
)

type WithAuth struct {
	Authorization string `name:"Authorization,omitempty" in:"header"`
}

type GetByID struct {
	httpx.MethodGet

	WithAuth

	ID       string         `name:"id" in:"path"`
	Protocol types.Protocol `name:"protocol,omitempty" in:"query"`
	Label    []string       `name:"label,omitempty" in:"query"`
	Name     string         `name:"name,omitempty" in:"query"`
}

func (GetByID) Validate() error {
	return nil
}

func (GetByID) Path() string {
	return "/:id"
}

var (
	req = &GetByID{}
)

func init() {
	req.ID = "1z"
	req.Authorization = "Bearer XXX"
	req.Protocol = types.PROTOCOL__HTTP
	req.Label = []string{"label-1", "label-2"}
	req.Name = "name"
}

func toPath() string {
	return transformers.NewPathnamePattern(req.Path()).Stringify(transformers.ParamsFromMap(map[string]string{
		"id": req.ID,
	}))
}

func newRequest() (*http.Request, error) {
	r, err := http.NewRequest(req.Method(), toPath(), nil)
	if err != nil {
		return nil, err
	}

	query := url.Values{}

	p, err := req.Protocol.MarshalText()
	if err != nil {
		return nil, err
	}

	query["protocol"] = []string{string(p)}
	query["label"] = req.Label
	query["name"] = []string{req.Name}
	r.Header["Authorization"] = []string{req.Authorization}

	r.URL.RawQuery = query.Encode()

	return r, nil
}

func newRequestWithTransformers() (*http.Request, error) {
	r, err := http.NewRequestWithContext(context.Background(), req.Method(), toPath(), nil)
	if err != nil {
		return nil, err
	}

	query := url.Values{}

	ctx := r.Context()

	{
		b := &strings.Builder{}
		_ = (&transformers.TransformerPlainText{}).EncodeTo(ctx, b, req.Protocol)
		query["protocol"] = []string{b.String()}
	}

	{

		values := make([]string, len(req.Label))
		for i := range req.Label {
			b := &strings.Builder{}
			_ = (&transformers.TransformerPlainText{}).EncodeTo(ctx, b, req.Label[i])
			values[i] = b.String()
		}
		query["label"] = values
	}

	{
		b := &strings.Builder{}
		_ = (&transformers.TransformerPlainText{}).EncodeTo(ctx, b, req.Name)
		query["name"] = []string{b.String()}
	}

	{
		b := bytes.NewBuffer(nil)
		_ = (&transformers.TransformerPlainText{}).EncodeTo(ctx, b, req.Authorization)
		r.Header["Authorization"] = []string{b.String()}
	}

	r.URL.RawQuery = query.Encode()

	return r, nil
}

func fromRequest(req *http.Request, r *GetByID) error {
	ri := httpx.NewRequestInfo(req)

	if values := ri.Values("path", "id"); len(values) > 0 {
		r.ID = values[0]
	}

	if values := ri.Values("header", "Authorization"); len(values) > 0 {
		r.Authorization = values[0]
	}

	if values := ri.Values("query", "name"); len(values) > 0 {
		r.Name = values[0]
	}

	if values := ri.Values("query", "label"); len(values) > 0 {
		r.Label = values
	}

	if values := ri.Values("query", "protocol"); len(values) > 0 {
		if err := r.Protocol.UnmarshalText([]byte(values[0])); err != nil {
			return err
		}
	}

	return nil
}

func newIncomingRequest(path string) *http.Request {
	req, _ := newRequest()
	params, _ := transformers.NewPathnamePattern(path).Parse(req.URL.Path)
	return req.WithContext(contextx.WithValue(req.Context(), httprouter.ParamsKey, params))
}

func BenchmarkFromRequest(b *testing.B) {
	rtm := httptransport.NewRequestTransformerMgr(nil, nil)

	b.Run("from request directly", func(b *testing.B) {
		r := newIncomingRequest(req.Path())

		req := GetByID{}

		for i := 0; i < b.N; i++ {
			_ = fromRequest(r, &req)
		}

		b.Log(req)
	})

	b.Run("from request by reflect", func(b *testing.B) {
		r := newIncomingRequest(req.Path())

		rt, _ := rtm.NewRequestTransformer(context.Background(), reflect.TypeOf(req))

		req := GetByID{}

		for i := 0; i < b.N; i++ {
			_ = rt.DecodeAndValidate(context.Background(), httpx.NewRequestInfo(r), &req)
		}

		b.Log(req)
	})
}

func BenchmarkToRequest(b *testing.B) {
	rtm := httptransport.NewRequestTransformerMgr(nil, nil)

	b.Run("toPath", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = toPath()
		}
	})

	b.Run("new request directly", func(b *testing.B) {
		r, _ := newRequest()
		b.Log(r.URL.String())

		for i := 0; i < b.N; i++ {
			_, _ = newRequest()
		}
	})

	b.Run("native new request with transformer", func(b *testing.B) {
		r, _ := newRequestWithTransformers()
		b.Log(r.URL.String())

		for i := 0; i < b.N; i++ {
			_, _ = newRequestWithTransformers()
		}
	})

	b.Run("new request by reflect", func(b *testing.B) {
		ctx := httptransport.AsRequestOut(context.Background())

		rt, _ := rtm.NewRequestTransformer(ctx, reflect.TypeOf(req))

		r, _ := rt.NewRequestWithContext(ctx, req.Method(), req.Path(), req)
		b.Log(r.URL.String())

		for i := 0; i < b.N; i++ {
			_, _ = rt.NewRequestWithContext(ctx, req.Method(), req.Path(), req)
		}
	})
}
