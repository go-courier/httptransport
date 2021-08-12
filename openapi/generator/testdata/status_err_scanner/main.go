package main

import (
	"fmt"
	"net/http"

	"github.com/go-courier/statuserror"
)

// @StatusErr[InternalServerError][500100001][InternalServerError]
func call() {
	fn()
}

func main() {
	call()
	fmt.Println(Unauthorized)
}

func fn() error {
	return statuserror.Wrap(fmt.Errorf("test"), http.StatusInternalServerError, "Test")
}
