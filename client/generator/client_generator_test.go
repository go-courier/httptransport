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
	dir := filepath.Join(cwd, "../../__examples__")

	g := NewClientGenerator("demo", &url.URL{
		Scheme: "file",
		Path:   dir + "/openapi.json",
	})

	g.Load()
	g.Output(dir)
}

func TestToColonPath(t *testing.T) {
	require.Equal(t, "/user/:userID/tags/:tagID", toColonPath("/user/{userID}/tags/{tagID}"))
	require.Equal(t, "/user/:userID", toColonPath("/user/{userID}"))
}
