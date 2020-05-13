package httpx

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/testify"
	"github.com/stretchr/testify/require"
)

type Data struct{}

func (Data) Cookies() []*http.Cookie {
	return []*http.Cookie{{
		Name:  "token",
		Value: "xxx",
	}}
}

func (Data) Meta() courier.Metadata {
	return courier.Metadata{
		"X": []string{"xxx"},
	}
}

func (Data) StatusCode() int {
	return http.StatusOK
}

func (Data) ContentType() string {
	return MIME_JSON
}

func TestResponseWrapper(t *testing.T) {
	require.Equal(t, &Response{
		Value:       nil,
		Cookies:     (&Data{}).Cookies(),
		Metadata:    (&Data{}).Meta(),
		ContentType: (&Data{}).ContentType(),
		StatusCode:  (&Data{}).StatusCode(),
	}, Compose(
		WithCookies((&Data{}).Cookies()...),
		WithMetadata((&Data{}).Meta()),
		WithContentType((&Data{}).ContentType()),
		WithStatusCode((&Data{}).StatusCode()),
	)(nil))
}

func TestResponseFrom(t *testing.T) {
	resp := ResponseFrom(&Data{})

	require.Equal(t, &Response{
		Value:       &Data{},
		Metadata:    (&Data{}).Meta(),
		Cookies:     (&Data{}).Cookies(),
		ContentType: (&Data{}).ContentType(),
		StatusCode:  (&Data{}).StatusCode(),
	}, resp)
}

func TestResponse_WriteTo(t *testing.T) {
	t.Run("redirect", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/", nil)
		rw := testify.NewMockResponseWriter()

		_ = ResponseFrom(RedirectWithStatusFound(&url.URL{
			Path: "/other",
		})).WriteTo(rw, req, nil)

		require.Equal(t, `HTTP/0.0 302 Found
Location: /other
Content-Length: 0

`, string(rw.MustDumpResponse()))
	})

	t.Run("redirect when error", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/", nil)
		rw := testify.NewMockResponseWriter()

		_ = ResponseFrom(RedirectWithStatusMovedPermanently(&url.URL{
			Path: "/other",
		})).WriteTo(rw, req, nil)

		require.Equal(t, `HTTP/0.0 301 Moved Permanently
Location: /other
Content-Length: 0

`, string(rw.MustDumpResponse()))
	})

	t.Run("cookies", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/", nil)
		rw := testify.NewMockResponseWriter()

		cookie := &http.Cookie{
			Name:    "token",
			Value:   "test",
			Expires: time.Now().Add(24 * time.Hour),
		}

		_ = ResponseFrom(WithCookies(cookie)(nil)).WriteTo(rw, req, nil)

		require.Equal(t, `HTTP/0.0 204 No Content
Set-Cookie: `+cookie.String()+`

`, string(rw.MustDumpResponse()))
	})

	t.Run("return ok", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		rw := testify.NewMockResponseWriter()

		type Data struct {
			ID string
		}

		_ = ResponseFrom(&Data{
			ID: "123456",
		}).WriteTo(rw, req, func(response *Response) (string, Encode, error) {
			return "application/json", func(w io.Writer, v interface{}) error {
				return json.NewEncoder(w).Encode(v)
			}, nil
		})

		require.Equal(t, `HTTP/0.0 200 OK
Content-Type: application/json; charset=utf-8

{"ID":"123456"}
`, string(rw.MustDumpResponse()))
	})

	t.Run("POST return ok", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/", nil)
		rw := testify.NewMockResponseWriter()

		type Data struct {
			ID string
		}

		_ = ResponseFrom(&Data{
			ID: "123456",
		}).WriteTo(rw, req, func(response *Response) (string, Encode, error) {
			return "application/json", func(w io.Writer, v interface{}) error {
				return json.NewEncoder(w).Encode(v)
			}, nil
		})

		require.Equal(t, `HTTP/0.0 201 Created
Content-Type: application/json; charset=utf-8

{"ID":"123456"}
`, string(rw.MustDumpResponse()))
	})

	t.Run("return nil", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/", nil)
		rw := testify.NewMockResponseWriter()

		_ = ResponseFrom(nil).WriteTo(rw, req, nil)

		require.Equal(t, `HTTP/0.0 204 No Content

`, string(rw.MustDumpResponse()))
	})

	t.Run("return attachment", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		rw := testify.NewMockResponseWriter()

		attachment := NewAttachment("text.txt", "text/plain")
		_, _ = attachment.WriteString("123123123")

		_ = ResponseFrom(attachment).WriteTo(rw, req, nil)

		require.Equal(t, `HTTP/0.0 200 OK
Content-Disposition: attachment; filename=text.txt
Content-Type: text/plain

123123123`, string(rw.MustDumpResponse()))
	})
}
