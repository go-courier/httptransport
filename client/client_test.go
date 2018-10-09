package client

import (
	"encoding/xml"
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

	{
		ipInfo := IpInfo{}

		meta, err := ipInfoClient.Do("json", &GetByJSON{}).Into(&ipInfo)
		require.NoError(t, err)

		t.Log(ipInfo)
		t.Log(meta)
	}

	{
		ipInfo := IpInfo{}

		meta, err := ipInfoClient.Do("xml", &GetByXML{}).Into(&ipInfo)
		require.NoError(t, err)

		t.Log(ipInfo)
		t.Log(meta)
	}

	errClient := &Client{
		Timeout: 100 * time.Second,
	}
	errClient.SetDefaults()

	{
		ipInfo := IpInfo{}

		_, err := errClient.Do("json", &GetByJSON{}).Into(&ipInfo)
		require.Error(t, err)
	}

}
