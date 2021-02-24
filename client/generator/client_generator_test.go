package generator

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAPIGenerator(t *testing.T) {
	cwd, _ := os.Getwd()

	openAPISchema := &url.URL{Scheme: "file", Path: filepath.Join(cwd, "../../__examples__/server/cmd/app/openapi.json")}

	g := NewClientGenerator("demo", openAPISchema, OptionVendorImportByGoMod())

	g.Load()
	g.Output(filepath.Join(cwd, "../../__examples__/downstream"))
}

func TestToColonPath(t *testing.T) {
	require.Equal(t, "/user/:userID/tags/:tagID", toColonPath("/user/{userID}/tags/{tagID}"))
	require.Equal(t, "/user/:userID", toColonPath("/user/{userID}"))
}
