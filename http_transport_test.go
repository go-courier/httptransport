package httptransport_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/go-courier/httptransport/client"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"testing"
	"time"

	"github.com/go-courier/httptransport/__examples__/routes"
	"github.com/stretchr/testify/require"

	"github.com/go-courier/httptransport"
)

func TestHttpTransport(t *testing.T) {
	ht := httptransport.NewHttpTransport(func(server *http.Server) {
		server.ReadTimeout = 15 * time.Second
	})
	ht.SetDefaults()
	ht.Port = 8080

	go func() {
		err := ht.Serve(routes.RootRouter)
		require.Error(t, err)
	}()

	time.Sleep(200 * time.Millisecond)

	resp, err := http.Get("http://127.0.0.1:8080/demo/restful/1")
	require.NoError(t, err)

	data, err := httputil.DumpResponse(resp, true)
	require.NoError(t, err)
	fmt.Println(string(data))

	time.Sleep(1 * time.Second)
	p, _ := os.FindProcess(os.Getpid())
	p.Signal(os.Interrupt)
}

func SkipTestHttpTransportWithTLS(t *testing.T) {
	ht := httptransport.NewHttpTransport(func(server *http.Server) {
		server.ReadTimeout = 15 * time.Second
	})

	ht.CertFile = "./testdata/MyCertificate.crt"
	ht.KeyFile = "./testdata/MyKey.key"

	rootCA, _ := ioutil.ReadFile(ht.CertFile)

	ht.SetDefaults()
	ht.Port = 8081

	go func() {
		err := ht.Serve(routes.RootRouter)
		require.Error(t, err)
	}()

	time.Sleep(200 * time.Millisecond)

	t.Log(rootCA)

	req, err := http.NewRequest("GET", "https://localhost:8081/demo/restful/1", nil)
	require.NoError(t, err)

	resp, err := client.GetShortConnClient(10*time.Second, NewInsecureTLSTransport(rootCA)).Do(req)
	require.NoError(t, err)

	data, err := httputil.DumpResponse(resp, true)
	require.NoError(t, err)
	fmt.Println(string(data))

	time.Sleep(2 * time.Second)
	p, _ := os.FindProcess(os.Getpid())
	p.Signal(os.Interrupt)
}

func NewInsecureTLSTransport(rootCA []byte) client.HttpTransport {
	return func(rt http.RoundTripper) http.RoundTripper {
		if httpRt, ok := rt.(*http.Transport); ok {
			if httpRt.TLSClientConfig == nil {
				httpRt.TLSClientConfig = &tls.Config{}
			}
			httpRt.TLSClientConfig.RootCAs = rootCertPool(rootCA)
			return httpRt
		}
		return rt
	}
}

func rootCertPool(caData []byte) *x509.CertPool {
	if len(caData) == 0 {
		return nil
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caData)
	return certPool
}
