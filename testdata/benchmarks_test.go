package testdata

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

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
	r, err := http.NewRequest(req.Method(), toPath(), nil)
	if err != nil {
		return nil, err
	}

	query := url.Values{}

	{
		b := &strings.Builder{}
		_ = (&transformers.TransformerPlainText{}).EncodeTo(b, req.Protocol)
		query["protocol"] = []string{b.String()}
	}

	{

		values := make([]string, len(req.Label))
		for i := range req.Label {
			b := &strings.Builder{}
			_ = (&transformers.TransformerPlainText{}).EncodeTo(b, req.Label[i])
			values[i] = b.String()
		}
		query["label"] = values
	}

	{
		b := &strings.Builder{}
		_ = (&transformers.TransformerPlainText{}).EncodeTo(b, req.Name)
		query["name"] = []string{b.String()}
	}

	{
		b := bytes.NewBuffer(nil)
		_ = (&transformers.TransformerPlainText{}).EncodeTo(b, req.Authorization)
		r.Header["Authorization"] = []string{b.String()}
	}

	r.URL.RawQuery = query.Encode()

	return r, nil
}

func BenchmarkToRequest(b *testing.B) {
	rtm := httptransport.NewRequestTransformerMgr(nil, nil)

	b.Run("toPath", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = toPath()
		}
	})

	b.Run("native new request", func(b *testing.B) {
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
