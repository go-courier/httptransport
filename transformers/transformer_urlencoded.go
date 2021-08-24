package transformers

import (
	"context"
	"io"
	"net/textproto"
	"net/url"
	"reflect"

	"github.com/go-courier/httptransport/httpx"

	reflectx "github.com/go-courier/x/reflect"

	verrors "github.com/go-courier/httptransport/validator"
	typesutil "github.com/go-courier/x/types"
	"github.com/pkg/errors"
)

func init() {
	TransformerMgrDefault.Register(&TransformerURLEncoded{})
}

/*
TransformerURLEncoded for application/x-www-form-urlencoded

	var s = struct {
		Username string `name:"username"`
		Nickname string `name:"username,omitempty"`
		Tags []string `name:"tag"`
	}{
		Username: "name",
		Tags: []string{"1","2"},
	}

will transform to

	username=name&tag=1&tag=2
*/
type TransformerURLEncoded struct {
	*FlattenParams
}

func (TransformerURLEncoded) Names() []string {
	return []string{"application/x-www-form-urlencoded", "form", "urlencoded", "url-encoded"}
}

func (TransformerURLEncoded) NamedByTag() string {
	return "name"
}

func (TransformerURLEncoded) New(ctx context.Context, typ typesutil.Type) (Transformer, error) {
	transformer := &TransformerURLEncoded{}

	typ = typesutil.Deref(typ)
	if typ.Kind() != reflect.Struct {
		return nil, errors.Errorf("content transformer `%s` should be used for struct type", transformer)
	}

	transformer.FlattenParams = &FlattenParams{}

	if err := transformer.FlattenParams.CollectParams(ctx, typ); err != nil {
		return nil, err
	}

	return transformer, nil
}

func (transformer *TransformerURLEncoded) EncodeTo(ctx context.Context, w io.Writer, v interface{}) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}
	rv = reflectx.Indirect(rv)

	values := url.Values{}
	errSet := verrors.NewErrorSet()

	for i := range transformer.Parameters {
		p := transformer.Parameters[i]

		if p.Transformer != nil {
			fieldValue := p.FieldValue(rv)
			stringBuilders := NewStringBuilders()
			if err := NewTransformerSuper(p.Transformer, &p.TransformerOption.CommonTransformOption).EncodeTo(ctx, stringBuilders, fieldValue); err != nil {
				errSet.AddErr(err, p.Name)
				continue
			}
			values[p.Name] = stringBuilders.StringSlice()
		}
	}

	if err := errSet.Err(); err != nil {
		return err
	}

	httpx.MaybeWriteHeader(ctx, w, transformer.Names()[0], map[string]string{
		"param": "value",
	})
	_, err := w.Write([]byte(values.Encode()))
	return err
}

func (transformer *TransformerURLEncoded) DecodeFrom(ctx context.Context, r io.Reader, v interface{}, headers ...textproto.MIMEHeader) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	if rv.Kind() != reflect.Ptr {
		return errors.New("decode target must be ptr value")
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	values, err := url.ParseQuery(string(data))
	if err != nil {
		return err
	}

	es := verrors.NewErrorSet()

	for i := range transformer.Parameters {
		p := transformer.Parameters[i]

		fieldValues := values[p.Name]

		if len(fieldValues) == 0 {
			continue
		}

		if p.Transformer != nil {
			if err := NewTransformerSuper(p.Transformer, &p.TransformerOption.CommonTransformOption).DecodeFrom(ctx, NewStringReaders(fieldValues), p.FieldValue(rv).Addr()); err != nil {
				es.AddErr(err, p.Name)
				continue
			}
		}
	}

	return es.Err()
}
