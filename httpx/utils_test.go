package httpx

import (
	"net/http"
	"testing"

	. "github.com/onsi/gomega"
)

func TestClientIP(t *testing.T) {
	{
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:80"
		NewWithT(t).Expect(ClientIP(req)).To(Equal("127.0.0.1"))
	}

	{
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(HeaderForwardedFor, "203.0.113.195, 70.41.3.18, 150.172.238.178")
		NewWithT(t).Expect(ClientIP(req)).To(Equal("203.0.113.195"))
	}

	{
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(HeaderRealIP, "203.0.113.195")
		NewWithT(t).Expect(ClientIP(req)).To(Equal("203.0.113.195"))
	}

	{
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		NewWithT(t).Expect(ClientIP(req)).To(Equal(""))
	}
}

func TestClientIPByHeaderRealIP(t *testing.T) {
	NewWithT(t).Expect(ClientIPByHeaderRealIP("203.0.113.195")).To(Equal("203.0.113.195"))
}

func TestGetClientIPByHeaderRealIP(t *testing.T) {
	NewWithT(t).Expect(ClientIPByHeaderForwardedFor("203.0.113.195, 70.41.3.18, 150.172.238.178")).To(Equal("203.0.113.195"))
}
