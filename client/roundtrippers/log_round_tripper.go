package roundtrippers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

func NewLogRoundTripper(logger *logrus.Entry) func(roundTripper http.RoundTripper) http.RoundTripper {
	return func(roundTripper http.RoundTripper) http.RoundTripper {
		return &LogRoundTripper{
			logger:           logger,
			nextRoundTripper: roundTripper,
		}
	}
}

type LogRoundTripper struct {
	logger           *logrus.Entry
	nextRoundTripper http.RoundTripper
}

func (rt *LogRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	startedAt := time.Now()

	resp, err := rt.nextRoundTripper.RoundTrip(req)

	defer func() {
		cost := time.Since(startedAt)

		logger := rt.logger.WithContext(req.Context()).WithFields(logrus.Fields{
			"cost":     fmt.Sprintf("%0.3fms", float64(cost/time.Millisecond)),
			"method":   req.Method,
			"url":      req.URL.String(),
			"metadata": req.Header,
		})

		if err == nil {
			logger.Infof("success")
		} else {
			logger.Warnf("do http request failed %s", err)
		}
	}()

	return resp, err
}
