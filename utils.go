package httptransport

import "github.com/pkg/errors"

func TryCatch(fn func()) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.Errorf("%+v", e)
		}
	}()

	fn()
	return nil
}
