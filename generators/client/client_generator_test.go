package client

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
)

func TestOpenAPIGenerator(t *testing.T) {
	cwd, _ := os.Getwd()

	openAPISchema := &url.URL{Scheme: "file", Path: filepath.Join(cwd, "../../testdata/server/cmd/app/openapi.json")}

	g := NewClientGenerator("demo", openAPISchema, OptionVendorImportByGoMod())

	g.Load()
	g.Output(filepath.Join(cwd, "../../testdata/downstream"))
}

func TestToColonPath(t *testing.T) {
	NewWithT(t).Expect(toColonPath("/user/{userID}/tags/{tagID}")).To(Equal("/user/:userID/tags/:tagID"))
	NewWithT(t).Expect(toColonPath("/user/{userID}")).To(Equal("/user/:userID"))
}
