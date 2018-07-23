package httpx

import (
	"net/http"
	"net/url"
)

type ContentTypeDescriber interface {
	ContentType() string
}

type StatusCodeDescriber interface {
	StatusCode() int
}

type CookiesDescriber interface {
	Cookies() []*http.Cookie
}

type RedirectDescriber interface {
	StatusCodeDescriber
	Location() *url.URL
}
