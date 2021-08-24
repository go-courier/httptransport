package transformers

import (
	"context"
	"io"
	"net/textproto"

	"github.com/go-courier/httptransport/httpx"
	encodingx "github.com/go-courier/x/encoding"
	typesx "github.com/go-courier/x/types"
)

func init() {
	TransformerMgrDefault.Register(&TransformerPlainText{})
}

type TransformerPlainText struct {
}

func (t *TransformerPlainText) String() string {
	return t.Names()[0]
}

func (TransformerPlainText) Names() []string {
	return []string{"text/plain", "plain", "text", "txt"}
}

func (TransformerPlainText) New(context.Context, typesx.Type) (Transformer, error) {
	return &TransformerPlainText{}, nil
}

func (t *TransformerPlainText) EncodeTo(ctx context.Context, w io.Writer, v interface{}) error {
	httpx.MaybeWriteHeader(ctx, w, t.String(), map[string]string{
		"charset": "utf-8",
	})

	data, err := encodingx.MarshalText(v)
	if err != nil {
		return err
	}

	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}

func (t *TransformerPlainText) DecodeFrom(ctx context.Context, r io.Reader, v interface{}, headers ...textproto.MIMEHeader) error {
	switch x := r.(type) {
	case CanString:
		raw := x.String()
		if x, ok := v.(*string); ok {
			*x = raw
			return nil
		}
		return encodingx.UnmarshalText(v, []byte(raw))
	case CanInterface:
		if raw, ok := x.Interface().(string); ok {
			if x, ok := v.(*string); ok {
				*x = raw
				return nil
			}
			return encodingx.UnmarshalText(v, []byte(raw))
		}
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return encodingx.UnmarshalText(v, data)
}
