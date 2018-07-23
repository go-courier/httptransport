package httpx

import (
	"fmt"
	"net/url"
)

//go:generate go run __codegen__/redirect/main.go

type redirect struct {
	*url.URL
}

// Redirect could be an error
func (r *redirect) Error() string {
	return fmt.Sprintf("Location: %s", r.Location())
}

func (r *redirect) Location() *url.URL {
	return r.URL
}
