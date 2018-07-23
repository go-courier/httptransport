package generator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	require.Equal(t, "github.com/go-courier/courier", pkgImportPathCourier)
	require.Equal(t, "github.com/go-courier/httptransport", pkgImportPathHttpTransport)
	require.Equal(t, "github.com/go-courier/httptransport/httpx", pkgImportPathHttpx)
}
