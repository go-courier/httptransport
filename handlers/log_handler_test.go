package handlers

import (
	"net/http"
	"time"

	"github.com/go-courier/httptransport/testify"
	"github.com/sirupsen/logrus"
)

func ExampleLogHandler() {
	logrus.SetLevel(logrus.DebugLevel)

	var handle http.HandlerFunc = func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(20 * time.Millisecond)

		switch req.Method {
		case http.MethodGet:
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte(`{"status":"ok"}`))
		case http.MethodPost:
			rw.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{"key":"StatusBadRequest","msg":"something wrong"}`))
		case http.MethodPut:
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(`{"key":"StatusInternalServerError","msg":"internal server error"}`))
		}
	}

	handler := LogHandler(logrus.StandardLogger())(handle).(*loggerHandler)

	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPost} {
		req, _ := http.NewRequest(method, "/", nil)
		handler.ServeHTTP(testify.NewMockResponseWriter(), req)
	}
	// Output:
}
