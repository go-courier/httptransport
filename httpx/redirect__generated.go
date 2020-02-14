package httpx

import (
	net_http "net/http"
	net_url "net/url"
)

func RedirectWithStatusMultipleChoices(u *net_url.URL) *StatusMultipleChoices {
	return &StatusMultipleChoices{
		Response: &Response{
			Location: u,
		},
	}
}

type StatusMultipleChoices struct {
	*Response
}

func (StatusMultipleChoices) StatusCode() int {
	return net_http.StatusMultipleChoices
}

func (r StatusMultipleChoices) Location() *net_url.URL {
	return r.Response.Location
}

func RedirectWithStatusMovedPermanently(u *net_url.URL) *StatusMovedPermanently {
	return &StatusMovedPermanently{
		Response: &Response{
			Location: u,
		},
	}
}

type StatusMovedPermanently struct {
	*Response
}

func (StatusMovedPermanently) StatusCode() int {
	return net_http.StatusMovedPermanently
}

func (r StatusMovedPermanently) Location() *net_url.URL {
	return r.Response.Location
}

func RedirectWithStatusFound(u *net_url.URL) *StatusFound {
	return &StatusFound{
		Response: &Response{
			Location: u,
		},
	}
}

type StatusFound struct {
	*Response
}

func (StatusFound) StatusCode() int {
	return net_http.StatusFound
}

func (r StatusFound) Location() *net_url.URL {
	return r.Response.Location
}

func RedirectWithStatusSeeOther(u *net_url.URL) *StatusSeeOther {
	return &StatusSeeOther{
		Response: &Response{
			Location: u,
		},
	}
}

type StatusSeeOther struct {
	*Response
}

func (StatusSeeOther) StatusCode() int {
	return net_http.StatusSeeOther
}

func (r StatusSeeOther) Location() *net_url.URL {
	return r.Response.Location
}

func RedirectWithStatusNotModified(u *net_url.URL) *StatusNotModified {
	return &StatusNotModified{
		Response: &Response{
			Location: u,
		},
	}
}

type StatusNotModified struct {
	*Response
}

func (StatusNotModified) StatusCode() int {
	return net_http.StatusNotModified
}

func (r StatusNotModified) Location() *net_url.URL {
	return r.Response.Location
}

func RedirectWithStatusUseProxy(u *net_url.URL) *StatusUseProxy {
	return &StatusUseProxy{
		Response: &Response{
			Location: u,
		},
	}
}

type StatusUseProxy struct {
	*Response
}

func (StatusUseProxy) StatusCode() int {
	return net_http.StatusUseProxy
}

func (r StatusUseProxy) Location() *net_url.URL {
	return r.Response.Location
}

func RedirectWithStatusTemporaryRedirect(u *net_url.URL) *StatusTemporaryRedirect {
	return &StatusTemporaryRedirect{
		Response: &Response{
			Location: u,
		},
	}
}

type StatusTemporaryRedirect struct {
	*Response
}

func (StatusTemporaryRedirect) StatusCode() int {
	return net_http.StatusTemporaryRedirect
}

func (r StatusTemporaryRedirect) Location() *net_url.URL {
	return r.Response.Location
}

func RedirectWithStatusPermanentRedirect(u *net_url.URL) *StatusPermanentRedirect {
	return &StatusPermanentRedirect{
		Response: &Response{
			Location: u,
		},
	}
}

type StatusPermanentRedirect struct {
	*Response
}

func (StatusPermanentRedirect) StatusCode() int {
	return net_http.StatusPermanentRedirect
}

func (r StatusPermanentRedirect) Location() *net_url.URL {
	return r.Response.Location
}
