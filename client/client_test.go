package client

import (
	"context"
	"encoding/xml"
	"net/http"
	"testing"
	"time"

	"github.com/go-courier/httptransport/httpx"
	"github.com/go-courier/httptransport/testify"
	"github.com/go-courier/statuserror"
	"github.com/stretchr/testify/require"
)

type IpInfo struct {
	xml.Name    `xml:"query"`
	Country     string `json:"country" xml:"country"`
	CountryCode string `json:"countryCode" xml:"countryCode"`
}

type GetByJSON struct {
	httpx.MethodGet
}

func (GetByJSON) Path() string {
	return "/me.json"
}

type GetByXML struct {
	httpx.MethodGet
}

func (GetByXML) Path() string {
	return "/me.xml"
}

func TestClient(t *testing.T) {
	ipInfoClient := &Client{
		Protocol: "https",
		Host:     "ip.nf",
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

	t.Run("request by struct", func(t *testing.T) {
		ipInfo := IpInfo{}

		meta, err := ipInfoClient.Do(context.Background(), &GetByJSON{}).Into(&ipInfo)
		require.NoError(t, err)

		t.Log(ipInfo)
		t.Log(meta)
	})

	t.Run("request by struct as xml", func(t *testing.T) {
		ipInfo := IpInfo{}

		meta, err := ipInfoClient.Do(context.Background(), &GetByXML{}).Into(&ipInfo)
		require.NoError(t, err)

		t.Log(ipInfo)
		t.Log(meta)
	})

	t.Run("cancel request", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(1 * time.Millisecond)
			cancel()
		}()

		ipInfo := IpInfo{}
		_, err := ipInfoClient.Do(ctx, &GetByJSON{}).Into(&ipInfo)
		require.Equal(t, "ClientClosedRequest", err.(*statuserror.StatusErr).Key)
	})

	t.Run("err request", func(t *testing.T) {
		errClient := &Client{
			Timeout: 100 * time.Second,
		}
		errClient.SetDefaults()

		{
			ipInfo := IpInfo{}

			_, err := errClient.Do(ContextWithClient(context.Background(), GetShortConnClient(10*time.Second)), &GetByJSON{}).Into(&ipInfo)
			require.Error(t, err)
		}
	})

	t.Run("result pass", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "https://ip.nf/me.json", nil)
		result := ipInfoClient.Do(context.Background(), request)

		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		rw := testify.NewMockResponseWriter()

		_ = httpx.ResponseFrom(result).WriteTo(rw, req, nil)

		require.Equal(t, 200, rw.StatusCode)
	})
}
