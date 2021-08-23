package transformers

import (
	"context"
	"encoding/xml"
	"io"
	"net/textproto"
	"reflect"

	"github.com/go-courier/httptransport/httpx"
	typesutil "github.com/go-courier/x/types"
)

func init() {
	TransformerMgrDefault.Register(&XMLTransformer{})
}

type XMLTransformer struct {
}

func (XMLTransformer) Names() []string {
	return []string{"application/xml", "xml"}
}

func (t *XMLTransformer) String() string {
	return t.Names()[0]
}

func (XMLTransformer) NamedByTag() string {
	return "xml"
}

func (XMLTransformer) New(context.Context, typesutil.Type) (Transformer, error) {
	return &XMLTransformer{}, nil
}

func (t *XMLTransformer) EncodeTo(ctx context.Context, w io.Writer, v interface{}) error {
	if rv, ok := v.(reflect.Value); ok {
		v = rv.Interface()
	}

	httpx.MaybeWriteHeader(ctx, w, t.String(), map[string]string{
		"charset": "utf-8",
	})

	return xml.NewEncoder(w).Encode(v)
}

func (XMLTransformer) DecodeFrom(ctx context.Context, r io.Reader, v interface{}, headers ...textproto.MIMEHeader) error {
	if rv, ok := v.(reflect.Value); ok {
		if rv.Kind() != reflect.Ptr && rv.CanAddr() {
			rv = rv.Addr()
		}
		v = rv.Interface()
	}
	d := xml.NewDecoder(r)
	err := d.Decode(v)
	if err != nil {
		// todo resolve field path by InputOffset()
		return err
	}
	return nil
}
