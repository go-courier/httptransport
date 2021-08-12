package transformers

import (
	"context"
	"io"
	"net/textproto"
	"reflect"

	"github.com/go-courier/httptransport/httpx"
	encodingx "github.com/go-courier/x/encoding"
	typesutil "github.com/go-courier/x/types"
)

func init() {
	TransformerMgrDefault.Register(&TransformerHTMLText{})
}

type TransformerHTMLText struct {
}

func (t *TransformerHTMLText) String() string {
	return t.Names()[0]
}

func (TransformerHTMLText) Names() []string {
	return []string{"text/html", "html"}
}

func (TransformerHTMLText) NamedByTag() string {
	return ""
}

func (TransformerHTMLText) New(context.Context, typesutil.Type) (Transformer, error) {
	return &TransformerHTMLText{}, nil
}

func (t *TransformerHTMLText) EncodeTo(ctx context.Context, w io.Writer, v interface{}) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	httpx.MaybeWriteHeader(ctx, w, t.String(), map[string]string{
		"charset": "utf-8",
	})

	data, err := encodingx.MarshalText(rv)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	return nil
}

func (TransformerHTMLText) DecodeFrom(ctx context.Context, r io.Reader, v interface{}, headers ...textproto.MIMEHeader) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return encodingx.UnmarshalText(rv, data)
}
