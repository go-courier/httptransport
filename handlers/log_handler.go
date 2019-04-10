package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-courier/httptransport/httpx"
	"github.com/sirupsen/logrus"
)

func LogHandler(logger *logrus.Logger) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return &loggerHandler{
			logger:      logger,
			nextHandler: handler,
		}
	}
}

type loggerHandler struct {
	logger      *logrus.Logger
	nextHandler http.Handler
}

type LoggerResponseWriter struct {
	http.ResponseWriter
	StatusCode int
	ErrMsg     []byte
}

func (w *LoggerResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *LoggerResponseWriter) Write(data []byte) (int, error) {
	if w.StatusCode >= http.StatusBadRequest {
		w.ErrMsg = data
	}
	return w.ResponseWriter.Write(data)
}

func (h *loggerHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	loggerRw := &LoggerResponseWriter{
		ResponseWriter: rw,
	}

	startAt := time.Now()

	defer func() {
		duration := time.Now().Sub(startAt)

		logger := h.logger.WithContext(req.Context())

		header := req.Header

		fields := logrus.Fields{
			"tag":         "access",
			"request_id":  header.Get(httpx.HeaderRequestID),
			"remote_ip":   httpx.ClientIP(req),
			"method":      req.Method,
			"request_url": req.URL.String(),
			"user_agent":  header.Get(httpx.HeaderUserAgent),
			"cost":        fmt.Sprintf("%0.3fms", float64(duration/time.Millisecond)),
		}

		fields["status"] = loggerRw.StatusCode

		if loggerRw.ErrMsg != nil {
			if loggerRw.StatusCode >= http.StatusInternalServerError {
				logger.WithFields(fields).Error(string(loggerRw.ErrMsg))
			} else {
				logger.WithFields(fields).Warn(string(loggerRw.ErrMsg))
			}
		} else {
			logger.WithFields(fields).Info("")
		}
	}()

	h.nextHandler.ServeHTTP(loggerRw, req)
}
