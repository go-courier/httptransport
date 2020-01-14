package openapi

import (
	"bytes"
	"context"
	"io/ioutil"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/httpx"
)

var openAPIJSONData = bytes.NewBuffer(nil)

func init() {
	data, err := ioutil.ReadFile("./openapi.json")
	if err == nil {
		openAPIJSONData.Write(data)
	} else {
		openAPIJSONData.Write([]byte("{}"))
	}
}

var OpenAPIRouter = courier.NewRouter(OpenAPI{})

type OpenAPI struct {
	httpx.MethodGet
}

func (s OpenAPI) Output(c context.Context) (interface{}, error) {
	return httpx.WithContentType(httpx.MIME_JSON)(bytes.NewBuffer(openAPIJSONData.Bytes())), nil
}
