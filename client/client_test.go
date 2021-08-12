package client

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
		require.NoError(t, err)
	})

	t.Run("direct request 404", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "https://api.github.com/xxxxn", nil)

		meta, err := ipInfoClient.Do(context.Background(), request).Into(nil)
		require.Error(t, err)

		t.Log(err)
		t.Log(meta)
	})
}
