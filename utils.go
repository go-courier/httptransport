package httptransport

import (
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

func TryCatch(fn func()) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.Errorf("%+v", e)
		}
	}()

	fn()
	return nil
}

func isLegitimateHTTPMethod(m string) bool {
	m = strings.ToUpper(m)
	switch m {
	case http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	}
	return false
}
