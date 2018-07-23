package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-courier/loaderx"
	"github.com/go-courier/oas"
	"github.com/stretchr/testify/require"
)

func TestOperatorScanner(t *testing.T) {
	cwd, _ := os.Getwd()
	program, pkgInfo, _ := loaderx.LoadWithTests(filepath.Join(cwd, "./__examples__/router_scanner/auth"))

	info := loaderx.NewPackageInfo(pkgInfo)

	scanner := NewOperatorScanner(program, pkgInfo)

	cases := map[string]string{
		"RespWithDescribers": /* language=json*/ `{
  "operationId": "RespWithDescribers",
  "responses": {
    "200": {
      "content": {
        "application/json": {
          "schema": {
            "type": "null"
          }
        }
      }
    }
  }
}`,
		"NoContent": /* language=json*/ `{
  "operationId": "NoContent",
  "responses": {
    "204": {}
  }
}`,
		"Auth": /* language=json*/ `{
  "summary": "Auth",
  "description": "auth auth",
  "operationId": "Auth",
  "parameters": [
    {
      "name": "HBool",
      "in": "header",
      "required": true,
      "schema": {
        "type": "boolean",
        "x-go-field-name": "HBool"
      }
    },
    {
      "name": "HInt",
      "in": "header",
      "required": true,
      "schema": {
        "type": "integer",
        "format": "int32",
        "x-go-field-name": "HInt"
      }
    },
    {
      "name": "HString",
      "in": "header",
      "required": true,
      "schema": {
        "type": "string",
        "x-go-field-name": "HString"
      }
    },
    {
      "name": "bytes",
      "in": "query",
      "schema": {
        "type": "array",
        "items": {
          "type": "integer",
          "format": "uint8"
        },
        "x-go-field-name": "QBytes",
        "x-tag-name": "bytes,omitempty"
      }
    },
    {
      "name": "bytesKeep",
      "in": "query",
      "required": true,
      "schema": {
        "type": "array",
        "items": {
          "type": "integer",
          "format": "uint8"
        },
        "x-go-field-name": "QBytesKeepEmpty",
        "x-tag-name": "bytesKeep"
      }
    },
    {
      "name": "bytesOmit",
      "in": "query",
      "schema": {
        "type": "array",
        "items": {
          "type": "integer",
          "format": "uint8"
        },
        "x-go-field-name": "QBytesOmitEmpty",
        "x-tag-name": "bytesOmit,omitempty"
      }
    },
    {
      "name": "int",
      "in": "query",
      "required": true,
      "schema": {
        "type": "integer",
        "format": "int32",
        "x-go-field-name": "QInt",
        "x-tag-name": "int"
      }
    },
    {
      "name": "string",
      "in": "query",
      "required": true,
      "schema": {
        "type": "string",
        "x-go-field-name": "QString",
        "x-tag-name": "string"
      }
    },
    {
      "name": "a",
      "in": "cookie",
      "required": true,
      "schema": {
        "type": "string",
        "x-go-field-name": "CString",
        "x-tag-name": "a"
      }
    },
    {
      "name": "slice",
      "in": "cookie",
      "required": true,
      "schema": {
        "type": "array",
        "items": {
          "type": "string"
        },
        "x-go-field-name": "CSlice",
        "x-tag-name": "slice"
      }
    }
  ],
  "requestBody": {
    "required": true,
    "content": {
      "application/json": {
        "schema": {
          "$ref": "#/components/schemas/Data"
        }
      }
    }
  },
  "responses": {
    "200": {
      "content": {
        "application/json": {
          "schema": {
            "$ref": "#/components/schemas/Data"
          }
        }
      }
    }
  }
}`,
	}

	for n, result := range cases {
		t.Run(n, func(t *testing.T) {
			operation := &oas.Operation{}
			op := scanner.Operator(info.TypeName(n))
			op.BindOperation("", operation, true)
			data, _ := json.MarshalIndent(operation, "", "  ")
			fmt.Println(string(data))
			require.Equal(t, result, string(data))
		})
	}
}
