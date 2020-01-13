package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-courier/httptransport/httpx"
	"github.com/go-courier/metax"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func LogHandler(logger *logrus.Entry) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return &loggerHandler{
			logger:      logger,
			nextHandler: handler,
		}
	}
}

type loggerHandler struct {
	logger      *logrus.Entry
	nextHandler http.Handler
}

type LoggerResponseWriter struct {
	rw http.ResponseWriter

	headerWritten bool

	StatusCode int
	ErrMsg     bytes.Buffer
}

func (rw *LoggerResponseWriter) Header() http.Header {
	return rw.rw.Header()
}

func (rw *LoggerResponseWriter) WriteHeader(statusCode int) {
	rw.writeHeader(statusCode)
}

func (rw *LoggerResponseWriter) Write(data []byte) (int, error) {
	if rw.StatusCode >= http.StatusBadRequest {
		rw.ErrMsg.Write(data)
	}
	return rw.rw.Write(data)
}

func (rw *LoggerResponseWriter) writeHeader(statusCode int) {
	if !rw.headerWritten {
		rw.rw.WriteHeader(statusCode)
		rw.StatusCode = statusCode
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

	level, _ := logrus.ParseLevel(strings.ToLower(req.Header.Get("x-log-level")))
	if level == logrus.PanicLevel {
		level = h.logger.Logger.Level
	}

	defer func() {
		duration := time.Now().Sub(startAt)

		logger := h.logger.WithContext(metax.ContextWithMeta(req.Context(), metax.ParseMeta(loggerRw.Header().Get("X-Meta"))))

		header := req.Header

		fields := logrus.Fields{
			"tag":         "access",
			"cost":        fmt.Sprintf("%0.3fms", float64(duration/time.Millisecond)),
			"remote_ip":   httpx.ClientIP(req),
			"method":      req.Method,
			"request_url": req.URL.String(),
			"user_agent":  header.Get(httpx.HeaderUserAgent),
		}

		fields["status"] = loggerRw.StatusCode

		if loggerRw.ErrMsg.Len() > 0 {
			if loggerRw.StatusCode >= http.StatusInternalServerError {
				if level >= logrus.ErrorLevel {
					logger.WithFields(fields).Error(loggerRw.ErrMsg.String())
				}
			} else {
				if level >= logrus.WarnLevel {
					logger.WithFields(fields).Warn(loggerRw.ErrMsg.String())
				}
			}
		} else {
			if level >= logrus.InfoLevel {
				logger.WithFields(fields).Info()
			}
		}
	}()

	h.nextHandler.ServeHTTP(loggerRw, req.WithContext(metax.ContextWithMeta(req.Context(), metax.ParseMeta(requestID))))
}
