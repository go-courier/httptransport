package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-courier/httptransport/httpx"
	"github.com/go-courier/logr"
	"github.com/go-courier/metax"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func LogHandler() func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return &loggerHandler{
			nextHandler: handler,
		}
	}
}

type loggerHandler struct {
	nextHandler http.Handler
}

type LoggerResponseWriter struct {
	rw http.ResponseWriter

	headerWritten bool

	statusCode int
	err        error
}

func (rw *LoggerResponseWriter) Header() http.Header {
	return rw.rw.Header()
}

func (rw *LoggerResponseWriter) WriteErr(err error) {
	rw.err = err
}

func (rw *LoggerResponseWriter) WriteHeader(statusCode int) {
	rw.writeHeader(statusCode)
}

func (rw *LoggerResponseWriter) Write(data []byte) (int, error) {
	if rw.err != nil && rw.statusCode >= http.StatusBadRequest {
		rw.err = errors.New(string(data))
	}
	return rw.rw.Write(data)
}

func (rw *LoggerResponseWriter) writeHeader(statusCode int) {
	if !rw.headerWritten {
		rw.rw.WriteHeader(statusCode)
		rw.statusCode = statusCode
		rw.headerWritten = true
	}
}

func (h *loggerHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	requestID := req.Header.Get(httpx.HeaderRequestID)
	if requestID == "" {
		requestID = uuid.New().String()
	}

	loggerRw := &LoggerResponseWriter{rw: rw}

	startAt := time.Now()

	logger := logr.FromContext(req.Context())

	level, _ := logr.ParseLevel(strings.ToLower(req.Header.Get("x-log-level")))
	if level == logr.PanicLevel {
		level = logr.TraceLevel
	}

	defer func() {
		duration := time.Since(startAt)

		header := req.Header

		fields := []interface{}{
			"tag", "access",
			"cost", fmt.Sprintf("%0.3fms", float64(duration/time.Millisecond)),
			"remote_ip", httpx.ClientIP(req),
			"method", req.Method,
			"request_url", req.URL.String(),
			"user_agent", header.Get(httpx.HeaderUserAgent),
			"status", loggerRw.statusCode,
		}

		if loggerRw.err != nil {
			if loggerRw.statusCode >= http.StatusInternalServerError {
				if level >= logr.ErrorLevel {
					logger.WithValues(fields).Error(loggerRw.err)
				}
			} else {
				if level >= logr.WarnLevel {
					logger.WithValues(fields).Warn(loggerRw.err)
				}
			}
		} else {
			if level >= logr.InfoLevel {
				logger.WithValues(fields).Info("")
			}
		}
	}()

	h.nextHandler.ServeHTTP(loggerRw, req.WithContext(metax.ContextWithMeta(req.Context(), metax.ParseMeta(requestID))))
}
