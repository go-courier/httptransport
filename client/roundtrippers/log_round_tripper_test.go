package roundtrippers

import (
	"context"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/go-courier/httptransport"
)

func TestLogRoundTripper(t *testing.T) {
	mgr := httptransport.NewRequestTransformerMgr(nil, nil)
	mgr.SetDefaults()

	req, _ := mgr.NewRequest(http.MethodGet, "https://github.com", nil)

	_, _ = NewLogRoundTripper(logrus.WithContext(context.Background()))(http.DefaultTransport).RoundTrip(req)
}
