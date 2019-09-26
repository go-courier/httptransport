package httpx

import (
	"io"
	"net/http"
	"net/textproto"
	"net/url"

	"github.com/go-courier/courier"
)

type ResponseWrapper func(v interface{}) *Response

func Compose(responseWrappers ...ResponseWrapper) ResponseWrapper {
	return func(v interface{}) *Response {
		r := ResponseFrom(v)
		for i := len(responseWrappers) - 1; i >= 0; i-- {
			r = responseWrappers[i](r)
		}
		return r
	}
}

func WithStatusCode(statusCode int) ResponseWrapper {
	return func(v interface{}) *Response {
		resp := ResponseFrom(v)
		resp.StatusCode = statusCode
		return resp
	}
}

func WithCookies(cookies ...*http.Cookie) ResponseWrapper {
	return func(v interface{}) *Response {
		resp := ResponseFrom(v)
		resp.Cookies = cookies
		return resp
	}
}

func WithContentType(contentType string) ResponseWrapper {
	return func(v interface{}) *Response {
		resp := ResponseFrom(v)
		resp.ContentType = contentType
		return resp
	}
}

func Metadata(key string, values ...string) courier.Metadata {
	return courier.Metadata{
		key: values,
	}
}

func WithMetadata(metadatas ...courier.Metadata) ResponseWrapper {
	return func(v interface{}) *Response {
		resp := ResponseFrom(v)
		resp.Metadata = courier.FromMetas(metadatas...)
		return resp
	}
}

func ResponseFrom(v interface{}) *Response {
	if r, ok := v.(*Response); ok {
		return r
	}

	response := &Response{}
	response.Value = v

	if redirectDescriber, ok := v.(RedirectDescriber); ok {
		response.Location = redirectDescriber.Location()
	}

	if metadataCarrier, ok := v.(courier.MetadataCarrier); ok {
		response.Metadata = metadataCarrier.Meta()
	}

	if cookiesDescriber, ok := v.(CookiesDescriber); ok {
		response.Cookies = cookiesDescriber.Cookies()
	}

	if contentTypeDescriber, ok := v.(ContentTypeDescriber); ok {
		response.ContentType = contentTypeDescriber.ContentType()
	}

	if statusDescriber, ok := v.(StatusCodeDescriber); ok {
		response.StatusCode = statusDescriber.StatusCode()
	}

	return response
}

type Upgrader interface {
	Upgrade(w http.ResponseWriter, r *http.Request) error
}

type Response struct {
	// value of Body
	Value       interface{}
	Metadata    courier.Metadata
	Cookies     []*http.Cookie
	Location    *url.URL
	ContentType string
	StatusCode  int
}

func (response *Response) WriteTo(rw http.ResponseWriter, r *http.Request, writeToBody func(w io.Writer, response *Response) error) error {
	if upgrader, ok := response.Value.(Upgrader); ok {
		return upgrader.Upgrade(rw, r)
	}

	if response.StatusCode == 0 {
		if response.Value == nil {
			response.StatusCode = http.StatusNoContent
		} else {
			if r.Method == http.MethodPost {
				response.StatusCode = http.StatusCreated
			} else {
				response.StatusCode = http.StatusOK
			}
		}
	}

	if response.Metadata != nil {
		header := rw.Header()
		for key, values := range response.Metadata {
			header[textproto.CanonicalMIMEHeaderKey(key)] = values
		}
	}

	if response.Cookies != nil {
		for i := range response.Cookies {
			cookie := response.Cookies[i]
			if cookie != nil {
				http.SetCookie(rw, cookie)
			}
		}
	}

	if response.Location != nil {
		http.Redirect(rw, r, response.Location.String(), response.StatusCode)
		return nil
	}

	switch response.StatusCode {
	case http.StatusNoContent:
		rw.WriteHeader(response.StatusCode)
		return nil
	default:
		rw.Header().Set(HeaderContentType, response.ContentType)
		rw.WriteHeader(response.StatusCode)

		if reader, ok := response.Value.(io.Reader); ok {
			if _, err := io.Copy(rw, reader); err != nil {
				return err
			}
		} else {
			if err := writeToBody(rw, response); err != nil {
				return err
			}
		}
	}
	return nil
}
