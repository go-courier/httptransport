package httpx

import net_http "net/http"

type Method struct {
}

type MethodGet struct{}

func (MethodGet) Method() string {
	return net_http.MethodGet
}

type MethodHead struct{}

func (MethodHead) Method() string {
	return net_http.MethodHead
}

type MethodPost struct{}

func (MethodPost) Method() string {
	return net_http.MethodPost
}

type MethodPut struct{}

func (MethodPut) Method() string {
	return net_http.MethodPut
}

type MethodPatch struct{}

func (MethodPatch) Method() string {
	return net_http.MethodPatch
}

type MethodDelete struct{}

func (MethodDelete) Method() string {
	return net_http.MethodDelete
}

type MethodConnect struct{}

func (MethodConnect) Method() string {
	return net_http.MethodConnect
}

type MethodOptions struct{}

func (MethodOptions) Method() string {
	return net_http.MethodOptions
}

type MethodTrace struct{}

func (MethodTrace) Method() string {
	return net_http.MethodTrace
}
