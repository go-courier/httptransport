package transformers

import (
	"context"
	"io"
	"io/ioutil"
	"mime"
	"net/textproto"
	"reflect"

	"github.com/go-courier/reflectx"
	"github.com/go-courier/reflectx/typesutil"
)

func init() {
	TransformerMgrDefault.Register(&PlainTextTransformer{})
}

type PlainTextTransformer struct {
}

func (t *PlainTextTransformer) String() string {
	return t.Names()[0]
}

func (PlainTextTransformer) Names() []string {
	return []string{"text/plain", "plain", "text", "txt"}
}

func (PlainTextTransformer) NamedByTag() string {
	return ""
}

func (PlainTextTransformer) New(context.Context, typesutil.Type) (Transformer, error) {
	return &PlainTextTransformer{}, nil
}

func (t *PlainTextTransformer) EncodeToWriter(w io.Writer, v interface{}) (string, error) {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	return superWrite(w, func(w io.Writer) error {
		if reflectx.IsBytes(rv.Type()) {
			_, err := w.Write(rv.Bytes())
			return err
		}

		data, err := reflectx.MarshalText(rv)
		if err != nil {
			return err
		}
		if _, err := w.Write(data); err != nil {
			return err
		}
		return nil
	}, mime.FormatMediaType(t.String(), map[string]string{
		"charset": "utf-8",
	}))
}

func (PlainTextTransformer) DecodeFromReader(r io.Reader, v interface{}, headers ...textproto.MIMEHeader) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	if reflectx.IsBytes(rv.Type()) {
		rv.SetBytes(data)
		return nil
	}
	return reflectx.UnmarshalText(rv, data)
}
