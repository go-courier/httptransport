package testify

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

func NewMockResponseWriter() *MockResponseWriter {
	return &MockResponseWriter{
		header: http.Header{},
	}
}

type MockResponseWriter struct {
	header     http.Header
	StatusCode int
	bytes.Buffer
}

var _ http.ResponseWriter = (*MockResponseWriter)(nil)

func (w *MockResponseWriter) Header() http.Header {
	return w.header
}

func (w *MockResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
}

func (w *MockResponseWriter) Response() *http.Response {
	resp := &http.Response{}
	resp.Header = w.header
	resp.StatusCode = w.StatusCode
	resp.Body = ioutil.NopCloser(&w.Buffer)
	return resp
}

func (w *MockResponseWriter) MustDumpResponse() []byte {
	data, err := httputil.DumpResponse(w.Response(), true)
	if err != nil {
		panic(err)
	}
	return bytes.Replace(data, []byte("\r\n"), []byte("\n"), -1)
}
