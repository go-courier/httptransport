package client

import (
	"encoding/xml"
	"net/http"
	"testing"
	"time"

	"github.com/go-courier/httptransport/httpx"
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
	return "/json"
}

type GetByXML struct {
	httpx.MethodGet
}

func (GetByXML) Path() string {
	return "/xml"
}

func TestClient(t *testing.T) {
	ipInfoClient := &Client{
		Host:    "ip-api.com",
		Timeout: 100 * time.Second,
	}
	ipInfoClient.SetDefaults()

	t.Run("direct request", func(t *testing.T) {
		ipInfo := IpInfo{}

		request, _ := http.NewRequest("GET", "http://ip-api.com/json", nil)

		meta, err := ipInfoClient.Do(nil, request).Into(&ipInfo)
		require.NoError(t, err)

		t.Log(ipInfo)
		t.Log(meta)
	})

	t.Run("direct request 404", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "https://api.github.com/xxxxn", nil)

		meta, err := ipInfoClient.Do(nil, request).Into(nil)
		require.Error(t, err)

		t.Log(err)
		t.Log(meta)
	})

	t.Run("request by struct", func(t *testing.T) {
		ipInfo := IpInfo{}

		meta, err := ipInfoClient.Do(nil, &GetByJSON{}).Into(&ipInfo)
		require.NoError(t, err)

		t.Log(ipInfo)
		t.Log(meta)
	})

	t.Run("request by struct as xml", func(t *testing.T) {
		ipInfo := IpInfo{}

		meta, err := ipInfoClient.Do(nil, &GetByXML{}).Into(&ipInfo)
		require.NoError(t, err)

		t.Log(ipInfo)
		t.Log(meta)
	})

	t.Run("err request", func(t *testing.T) {
		errClient := &Client{
			Timeout: 100 * time.Second,
		}
		errClient.SetDefaults()

		{
			ipInfo := IpInfo{}

			_, err := errClient.Do(nil, &GetByJSON{}).Into(&ipInfo)
			require.Error(t, err)
		}
	})
}
