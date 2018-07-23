package httptransport

import (
	"fmt"
)

func TryCatch(fn func()) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	fn()
	return nil
}
