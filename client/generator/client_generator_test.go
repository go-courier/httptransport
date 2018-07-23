package generator

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"
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
