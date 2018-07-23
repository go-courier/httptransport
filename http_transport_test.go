package httptransport_test

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/go-courier/httptransport"
	"github.com/go-courier/httptransport/__examples__/routes"
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
