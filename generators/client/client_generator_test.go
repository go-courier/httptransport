package client

import (
	"bytes"
	"encoding/json"
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

func TestGenEnumInt(t *testing.T) {
	cwd, _ := os.Getwd()
	g := NewClientGenerator("demo", &url.URL{}, OptionVendorImportByGoMod())
	snippet := []byte(`
{
  "openapi": "3.0.3",
  "components": {
    "schemas": {
      "ExampleComCloudchainSrvDemoPkgConstantsErrorsStatusError": {
        "type": "integer",
        "format": "int32",
        "enum": [
          400000001,
          400000002
        ],
        "x-enum-labels": [
          "400000001",
          "400000002"
        ],
        "x-go-vendor-type": "example.com/cloudchain/srv-demo/pkg/constants/errors.StatusError",
        "x-id": "ExampleComCloudchainSrvDemoPkgConstantsErrorsStatusError"
      }
    }
  }
}
`)
	if err := json.NewDecoder(bytes.NewBuffer(snippet)).Decode(g.openAPI); err != nil {
		panic(err)
	}
	g.Output(filepath.Join(cwd, "../../testdata/enum"))
}

func TestGenEnumFloat(t *testing.T) {
	cwd, _ := os.Getwd()
	g := NewClientGenerator("demo", &url.URL{}, OptionVendorImportByGoMod())
	snippet := []byte(`
{
  "openapi": "3.0.3",
  "components": {
    "schemas": {
      "ExampleComCloudchainSrvDemoPkgConstantsErrorsStatusError": {
        "type": "number",
        "format": "double",
        "enum": [
          40000000.1,
          40000000.2
        ],
        "x-enum-labels": [
          "40000000.1",
          "40000000.2"
        ],
        "x-go-vendor-type": "example.com/cloudchain/srv-demo/pkg/constants/errors.StatusError",
        "x-id": "ExampleComCloudchainSrvDemoPkgConstantsErrorsStatusError"
      }
    }
  }
}
`)
	if err := json.NewDecoder(bytes.NewBuffer(snippet)).Decode(g.openAPI); err != nil {
		panic(err)
	}
	g.Output(filepath.Join(cwd, "../../testdata/enum"))
}
