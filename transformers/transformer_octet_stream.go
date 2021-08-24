package transformers

import (
	"context"
	"io"
	"mime/multipart"
	"net/textproto"
	"reflect"

	"github.com/go-courier/httptransport/httpx"
	typesx "github.com/go-courier/x/types"
)

func init() {
	TransformerMgrDefault.Register(&TransformerOctetStream{})
}

type TransformerOctetStream struct {
}

func (t *TransformerOctetStream) String() string {
	return t.Names()[0]
}

func (TransformerOctetStream) Names() []string {
	return []string{"application/octet-stream", "stream", "octet-stream"}
}

func (TransformerOctetStream) New(context.Context, typesx.Type) (Transformer, error) {
	return &TransformerOctetStream{}, nil
}

func (t *TransformerOctetStream) EncodeTo(ctx context.Context, w io.Writer, v interface{}) error {
	rv, ok := v.(reflect.Value)
	if ok {
		v = rv.Interface()
	}

	switch x := v.(type) {
	case io.Reader:
		httpx.MaybeWriteHeader(ctx, w, t.Names()[0], nil)
		if _, err := io.Copy(w, x); err != nil {
			return err
		}
	case *multipart.FileHeader:
		file, err := x.Open()
		if err != nil {
			return err
		}
		defer file.Close()

		if rw, ok := w.(httpx.WithHeader); ok {
			for k := range x.Header {
				rw.Header()[k] = x.Header[k]
			}
		}

		if _, err := io.Copy(w, file); err != nil {
			return err
		}
	}

	return nil
}

func (TransformerOctetStream) DecodeFrom(ctx context.Context, r io.Reader, v interface{}, headers ...textproto.MIMEHeader) error {
	rv, ok := v.(reflect.Value)
	if ok {
		v = rv.Interface()
	}

	switch x := v.(type) {
	case io.Writer:
		if _, err := io.Copy(x, r); err != nil {
			return err
		}
	case *multipart.FileHeader:
		if canInterface, ok := r.(CanInterface); ok {
			if fh, ok := canInterface.Interface().(*multipart.FileHeader); ok {
				*x = *fh
			}
		}
	case **multipart.FileHeader:
		if canInterface, ok := r.(CanInterface); ok {
			if fh, ok := canInterface.Interface().(*multipart.FileHeader); ok {
				*x = fh
			}
		}
	}

	return nil
}
