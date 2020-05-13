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

func TestTextTransformer(t *testing.T) {
	ct, _ := TransformerMgrDefault.NewTransformer(context.Background(), typesutil.FromRType(reflect.TypeOf("")), TransformerOption{})

	{
		b := bytes.NewBuffer(nil)
		_, err := ct.EncodeToWriter(b, "")
		require.NoError(t, err)
	}

	{
		b := bytes.NewBuffer(nil)
		_, err := ct.EncodeToWriter(b, reflect.ValueOf(1))
		require.NoError(t, err)
	}

	{
		b := bytes.NewBufferString("a")
		err := ct.DecodeFromReader(b, ptr.Int(0))
		require.Error(t, err)
	}

	{
		b := bytes.NewBufferString("1")
		err := ct.DecodeFromReader(b, reflect.ValueOf(ptr.Int(0)))
		require.NoError(t, err)
	}
}
