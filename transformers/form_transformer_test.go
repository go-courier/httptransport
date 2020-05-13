package transformers

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/go-courier/ptr"
	"github.com/go-courier/reflectx/typesutil"
	"github.com/stretchr/testify/require"
)

func TestFormTransformer(t *testing.T) {
	queryStr := `Bool=true` +
		`&Bytes=bytes` +
		`&PtrInt=1` +
		`&StringArray=1&StringArray=&StringArray=3` +
		`&StringSlice=1&StringSlice=2&StringSlice=3` +
		`&Struct=%3CSub%3E%3CName%3E%3C%2FName%3E%3C%2FSub%3E` +
		`&StructSlice=%7B%22Name%22%3A%22name%22%7D%0A` +
		`&first_name=test`

	type Sub struct {
		Name string
	}

	type TestData struct {
		PtrBool     *bool `name:",omitempty"`
		PtrInt      *int
		Bool        bool
		Bytes       []byte
		FirstName   string `name:"first_name"`
		StructSlice []Sub
		StringSlice []string
		StringArray [3]string
		Struct      Sub `mime:"xml"`
	}

	data := TestData{}
	data.FirstName = "test"
	data.Bool = true
	data.Bytes = []byte("bytes")
	data.PtrInt = ptr.Int(1)
	data.StringSlice = []string{"1", "2", "3"}
	data.StructSlice = []Sub{
		{
			Name: "name",
		},
	}
	data.StringArray = [3]string{"1", "", "3"}

	ct, _ := TransformerMgrDefault.NewTransformer(context.Background(), typesutil.FromRType(reflect.TypeOf(data)), TransformerOption{
		MIME: "urlencoded",
	})

	{
		b := bytes.NewBuffer(nil)
		contentType, err := ct.EncodeToWriter(b, data)
		require.NoError(t, err)
		require.Equal(t, "application/x-www-form-urlencoded; param=value", contentType)
		require.Equal(t, queryStr, b.String())
	}

	{
		b := bytes.NewBufferString(queryStr)
		testData := TestData{}

		err := ct.DecodeFromReader(b, &testData)
		require.NoError(t, err)
		require.Equal(t, data, testData)
	}
}
