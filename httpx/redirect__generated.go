package httpx

import (
	net_http "net/http"
	net_url "net/url"
)

func RedirectWithStatusMultipleChoices(u *net_url.URL) *StatusMultipleChoices {
	return &StatusMultipleChoices{
		redirect: &redirect{
			URL: u,
		},
	}
}

type StatusMultipleChoices struct {
	*redirect
}

func (StatusMultipleChoices) StatusCode() int {
	return net_http.StatusMultipleChoices
}

func RedirectWithStatusMovedPermanently(u *net_url.URL) *StatusMovedPermanently {
	return &StatusMovedPermanently{
		redirect: &redirect{
			URL: u,
		},
	}
}

type StatusMovedPermanently struct {
	*redirect
}

func (StatusMovedPermanently) StatusCode() int {
	return net_http.StatusMovedPermanently
}

func RedirectWithStatusFound(u *net_url.URL) *StatusFound {
	return &StatusFound{
		redirect: &redirect{
			URL: u,
		},
	}
}

type StatusFound struct {
	*redirect
}

func (StatusFound) StatusCode() int {
	return net_http.StatusFound
}

func RedirectWithStatusSeeOther(u *net_url.URL) *StatusSeeOther {
	return &StatusSeeOther{
		redirect: &redirect{
			URL: u,
		},
	}
}

type StatusSeeOther struct {
	*redirect
}

func (StatusSeeOther) StatusCode() int {
	return net_http.StatusSeeOther
}

func RedirectWithStatusNotModified(u *net_url.URL) *StatusNotModified {
	return &StatusNotModified{
		redirect: &redirect{
			URL: u,
		},
	}
}

type StatusNotModified struct {
	*redirect
}

func (StatusNotModified) StatusCode() int {
	return net_http.StatusNotModified
}

func RedirectWithStatusUseProxy(u *net_url.URL) *StatusUseProxy {
	return &StatusUseProxy{
		redirect: &redirect{
			URL: u,
		},
	}
}

type StatusUseProxy struct {
	*redirect
}

func (StatusUseProxy) StatusCode() int {
	return net_http.StatusUseProxy
}

func RedirectWithStatusTemporaryRedirect(u *net_url.URL) *StatusTemporaryRedirect {
	return &StatusTemporaryRedirect{
		redirect: &redirect{
			URL: u,
		},
	}
}

type StatusTemporaryRedirect struct {
	*redirect
}

func (StatusTemporaryRedirect) StatusCode() int {
	return net_http.StatusTemporaryRedirect
}

func RedirectWithStatusPermanentRedirect(u *net_url.URL) *StatusPermanentRedirect {
	return &StatusPermanentRedirect{
		redirect: &redirect{
			URL: u,
		},
	}
}

type StatusPermanentRedirect struct {
	*redirect
}

func (StatusPermanentRedirect) StatusCode() int {
	return net_http.StatusPermanentRedirect
}
