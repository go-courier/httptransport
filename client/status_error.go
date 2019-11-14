package client

import (
	"net/http"
)

//go:generate courier gen error StatusError
type StatusError int

const (
	// request failed
	RequestFailed StatusError = http.StatusInternalServerError*1e6 + iota + 1
	// read failed
	ReadFailed
)

const (
	// request canceled
	ClientClosedRequest StatusError = 499*1e6 + iota + 1
)

const (
	// transform request failed
	RequestTransformFailed StatusError = http.StatusBadRequest*1e6 + iota + 1
)
