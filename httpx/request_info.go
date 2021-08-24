package httpx

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

func NewRequestInfo(r *http.Request) RequestInfo {
	return &requestInfo{
		receivedAt: time.Now(),
		request:    r,
	}
}

type WithFromRequestInfo interface {
	FromRequestInfo(ri RequestInfo) error
}

type RequestInfo interface {
	Context() context.Context
	Values(in string, name string) []string
	Header() http.Header
	Body() io.ReadCloser
}

type requestInfo struct {
	request    *http.Request
	receivedAt time.Time
	query      url.Values
	cookies    []*http.Cookie
	params     httprouter.Params
}

func (info *requestInfo) Header() http.Header {
	return info.request.Header
}

func (info *requestInfo) Context() context.Context {
	return info.request.Context()
}

func (info *requestInfo) Body() io.ReadCloser {
	return info.request.Body
}

func (info *requestInfo) Value(in string, name string) string {
	values := info.Values(in, name)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (info *requestInfo) Values(in string, name string) []string {
	switch in {
	case "path":
		v := info.Param(name)
		if v == "" {
			return []string{}
		}
		return []string{v}
	case "query":
		return info.QueryValues(name)
	case "cookie":
		return info.CookieValues(name)
	case "header":
		return info.HeaderValues(name)
	}
	return []string{}
}

func (info *requestInfo) Param(name string) string {
	if info.params == nil {
		params, ok := info.request.Context().Value(httprouter.ParamsKey).(httprouter.Params)
		if !ok {
			params = httprouter.Params{}
		}
		info.params = params
	}
	return info.params.ByName(name)
}

func (info *requestInfo) QueryValues(name string) []string {
	if info.query == nil {
		info.query = info.request.URL.Query()

		if info.request.Method == http.MethodGet && len(info.query) == 0 && info.request.ContentLength > 0 {
			if strings.HasPrefix(info.request.Header.Get("Content-Type"), MIME_FORM_URLENCODED) {
				data, err := ioutil.ReadAll(info.request.Body)
				if err == nil {
					info.request.Body.Close()

					query, e := url.ParseQuery(string(data))
					if e == nil {
						info.query = query
					}
				}
			}
		}
	}
	return info.query[name]
}

func (info *requestInfo) HeaderValues(name string) []string {
	return info.request.Header[textproto.CanonicalMIMEHeaderKey(name)]
}

func (info *requestInfo) CookieValues(name string) []string {
	if info.cookies == nil {
		info.cookies = info.request.Cookies()
	}

	values := make([]string, 0)
	for _, c := range info.cookies {
		if c.Name == name {
			if c.Expires.IsZero() {
				values = append(values, c.Value)
			} else if c.Expires.After(info.receivedAt) {
				values = append(values, c.Value)
			}
		}
	}
	return values
}
