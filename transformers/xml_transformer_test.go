package transformers

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/go-courier/reflectx/typesutil"
	"github.com/stretchr/testify/require"
)

func TestXMLTransformer(t *testing.T) {
	type TestData struct {
		Data struct {
			Bool        bool
			FirstName   string `xml:"name>first"`
			StructSlice []struct {
				Name string
			}
			StringSlice     []string
			StringAttrSlice []string `xml:"StringAttrSlice,attr"`
			NestedSlice     []struct {
				Names []string
			}
		}
	}

	data := TestData{}
	data.Data.FirstName = "test"
	data.Data.StringSlice = []string{"1", "2", "3"}
	data.Data.StringAttrSlice = []string{"1", "2", "3"}

	ct, _ := TransformerMgrDefault.NewTransformer(context.Background(), typesutil.FromRType(reflect.TypeOf(data)), TransformerOption{
		MIME: "xml",
	})

	{
		b := bytes.NewBuffer(nil)
		_, err := ct.EncodeToWriter(b, data)
		require.NoError(t, err)
	}

	{
		b := bytes.NewBuffer(nil)
		_, err := ct.EncodeToWriter(b, data)
		require.NoError(t, err)
	}

	{
		b := bytes.NewBufferString("<")
		err := ct.DecodeFromReader(b, &data)
		require.Error(t, err)
	}

	{
		b := bytes.NewBufferString("<TestData></TestData>")
		err := ct.DecodeFromReader(b, reflect.ValueOf(&data))
		require.NoError(t, err)
	}

	{
		b := bytes.NewBufferString("<TestData><Data><Bool>bool</Bool></Data></TestData>")
		err := ct.DecodeFromReader(b, &data)
		require.Error(t, err)
	}
}
