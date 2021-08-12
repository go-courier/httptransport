package httpx

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

func NewRequestInfo(r *http.Request) *RequestInfo {
	return &RequestInfo{
		receivedAt: time.Now(),
		Request:    r,
	}
}

type RequestInfo struct {
	Request    *http.Request
	receivedAt time.Time
	query      url.Values
	cookies    []*http.Cookie
	params     httprouter.Params
}

func (info *RequestInfo) Value(in string, name string) string {
	values := info.Values(in, name)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (info *RequestInfo) Values(in string, name string) []string {
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

func (info *RequestInfo) Param(name string) string {
	if info.params == nil {
		params, ok := info.Request.Context().Value(httprouter.ParamsKey).(httprouter.Params)
		if !ok {
			params = httprouter.Params{}
		}
		info.params = params
	}
	return info.params.ByName(name)
}

func (info *RequestInfo) QueryValues(name string) []string {
	if info.query == nil {
		info.query = info.Request.URL.Query()

		if info.Request.Method == http.MethodGet && len(info.query) == 0 && info.Request.ContentLength > 0 {
			if strings.HasPrefix(info.Request.Header.Get("Content-Type"), MIME_FORM_URLENCODED) {
				data, err := ioutil.ReadAll(info.Request.Body)
				if err == nil {
					info.Request.Body.Close()

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

func (info *RequestInfo) HeaderValues(name string) []string {
	return info.Request.Header[textproto.CanonicalMIMEHeaderKey(name)]
}

func (info *RequestInfo) CookieValues(name string) []string {
	if info.cookies == nil {
		info.cookies = info.Request.Cookies()
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

func (info *RequestInfo) Body() io.Reader {
	return info.Request.Body
}
