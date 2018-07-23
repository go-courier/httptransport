package httpx

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientIP(t *testing.T) {
	{
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:80"
		require.Equal(t, "127.0.0.1", ClientIP(req))
	}

	{
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(HeaderForwardedFor, "203.0.113.195, 70.41.3.18, 150.172.238.178")
		require.Equal(t, "203.0.113.195", ClientIP(req))
	}

	{
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(HeaderRealIP, "203.0.113.195")
		require.Equal(t, "203.0.113.195", ClientIP(req))
	}

	{
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		require.Equal(t, "", ClientIP(req))
	}
}

func TestClientIPByHeaderRealIP(t *testing.T) {
	require.Equal(t, "203.0.113.195", ClientIPByHeaderRealIP("203.0.113.195"))
}

func TestGetClientIPByHeaderRealIP(t *testing.T) {
	require.Equal(t, "203.0.113.195", ClientIPByHeaderForwardedFor("203.0.113.195, 70.41.3.18, 150.172.238.178"))
}
