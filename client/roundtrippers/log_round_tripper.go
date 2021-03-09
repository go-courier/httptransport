package roundtrippers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-courier/logr"
	"github.com/pkg/errors"
)

func NewLogRoundTripper() func(roundTripper http.RoundTripper) http.RoundTripper {
	return func(roundTripper http.RoundTripper) http.RoundTripper {
		return &LogRoundTripper{
			nextRoundTripper: roundTripper,
		}
	}
}

type LogRoundTripper struct {
	nextRoundTripper http.RoundTripper
}

func (rt *LogRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	startedAt := time.Now()

	ctx, logger := logr.Start(req.Context(), "Request")
	defer logger.End()

	resp, err := rt.nextRoundTripper.RoundTrip(req.WithContext(ctx))

	defer func() {
		cost := time.Since(startedAt)

		logger := logger.WithValues(
			"cost", fmt.Sprintf("%0.3fms", float64(cost/time.Millisecond)),
			"method", req.Method,
			"url", req.URL.String(),
			"metadata", req.Header,
		)

		if err == nil {
			logger.Info("success")
		} else {
			logger.Warn(errors.Wrap(err, "http request failed"))
		}
	}()

	return resp, err
}
