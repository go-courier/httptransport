package errors

import (
	"net/http"
)

type StatusError int

func (StatusError) ServiceCode() int {
	return 999 * 1e3
}

const (
	// InternalServerError
	// Something wrong in server
	InternalServerError StatusError = http.StatusInternalServerError*1e6 + iota + 1
)

const (
	// @errTalk Unauthorized
	Unauthorized StatusError = http.StatusUnauthorized*1e6 + iota + 1
)
