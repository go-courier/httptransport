package client

import (
	"context"
	"net/http"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestClient(t *testing.T) {
	ipInfoClient := &Client{
		Protocol: "https",
		Host:     "api.github.com",
		Timeout:  100 * time.Second,
	}

	ipInfoClient.SetDefaults()

	t.Run("direct request", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "https://api.github.com", nil)
		_, err := ipInfoClient.Do(context.Background(), request).Into(nil)
		NewWithT(t).Expect(err).To(BeNil())
	})

	t.Run("direct request 404", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "https://api.github.com/xxxxn", nil)

		meta, err := ipInfoClient.Do(context.Background(), request).Into(nil)
		NewWithT(t).Expect(err).NotTo(BeNil())
		t.Log(err)
		t.Log(meta)
	})
}
