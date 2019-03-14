package roundtrippers

import (
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/go-courier/httptransport"
)

func TestLogRoundTripper(t *testing.T) {
	mgr := httptransport.NewRequestTransformerMgr(nil, nil)
	mgr.SetDefaults()

	req, _ := mgr.NewRequest(http.MethodGet, "https://github.com", nil)

	NewLogRoundTripper(logrus.StandardLogger())(http.DefaultTransport).RoundTrip(req)
}
