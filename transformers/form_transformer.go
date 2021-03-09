package transformers

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"mime"
	"net/textproto"
	"net/url"
	"reflect"

	"github.com/go-courier/reflectx/typesutil"
	verrors "github.com/go-courier/validator/errors"
	"github.com/pkg/errors"
)

func init() {
	TransformerMgrDefault.Register(&FormTransformer{})
}

type FormTransformer struct {
	*FlattenParams
}

/*
transformer for application/x-www-form-urlencoded

	var s = struct {
		Username string `name:"username"`
		Nickname string `name:"username,omitempty"`
		Tags []string `name:"tag"`
	}{
		Username: "name",
		Tags: []string{"1","2"},
	}

will be transform to

	username=name&tag=1&tag=2
*/
func (FormTransformer) Names() []string {
	return []string{"application/x-www-form-urlencoded", "form", "urlencoded", "url-encoded"}
}

func (FormTransformer) NamedByTag() string {
	return "name"
}

func (transformer *FormTransformer) String() string {
	return transformer.Names()[0]
}

func (FormTransformer) New(ctx context.Context, typ typesutil.Type) (Transformer, error) {
	transformer := &FormTransformer{}

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

func (transformer *FormTransformer) EncodeToWriter(w io.Writer, v interface{}) (string, error) {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	valueAdder := url.Values{}

	errSet := verrors.NewErrorSet("")

	NamedStructFieldValueRange(reflect.Indirect(rv), func(fieldValue reflect.Value, field *reflect.StructField) {
		fieldOpt := transformer.fieldOpts[field.Name]
		fieldTransformer := transformer.fieldTransformers[field.Name]

		maybe := NewMaybeTransformer(fieldTransformer, &fieldOpt.CommonTransformOption)

		if fieldOpt.Explode {
			for i := 0; i < fieldValue.Len(); i++ {
				if err := maybe.Add(fieldOpt.FieldName, fieldValue.Index(i), valueAdder); err != nil {
					errSet.AddErr(err, fieldOpt.FieldName, i)
				}
			}
		} else {
			if err := maybe.Add(fieldOpt.FieldName, fieldValue, valueAdder); err != nil {
				errSet.AddErr(err, fieldOpt.FieldName)
			}
		}
	})

	if err := errSet.Err(); err != nil {
		return "", err
	}

	return superWrite(w, func(w io.Writer) error {
		_, err := w.Write([]byte(valueAdder.Encode()))
		return err
	}, mime.FormatMediaType(transformer.String(), map[string]string{
		"param": "value",
	}))
}

func (transformer *FormTransformer) DecodeFromReader(r io.Reader, v interface{}, headers ...textproto.MIMEHeader) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	values, err := url.ParseQuery(string(data))
	if err != nil {
		return err
	}

	errSet := verrors.NewErrorSet("")

	NamedStructFieldValueRange(reflect.Indirect(rv), func(fieldValue reflect.Value, field *reflect.StructField) {
		fieldOpt := transformer.fieldOpts[field.Name]
		fieldTransformer := transformer.fieldTransformers[field.Name]
		maybe := NewMaybeTransformer(fieldTransformer, &fieldOpt.CommonTransformOption)

		if fieldOpt.Explode {
			valueList := values[fieldOpt.FieldName]
			lenOfValues := len(valueList)

			if fieldOpt.Omitempty && lenOfValues == 0 {
				return
			}

			if field.Type.Kind() == reflect.Slice {
				fieldValue.Set(reflect.MakeSlice(field.Type, lenOfValues, lenOfValues))
			}

			for idx := 0; idx < fieldValue.Len(); idx++ {
				if lenOfValues > idx {
					if err := maybe.DecodeFromReader(bytes.NewBufferString(valueList[idx]), fieldValue.Index(idx)); err != nil {
						errSet.AddErr(err, fieldOpt.FieldName, idx)
					}
				}
			}
		} else {
			if err := maybe.DecodeFromReader(bytes.NewBufferString(values.Get(fieldOpt.FieldName)), fieldValue); err != nil {
				errSet.AddErr(err, fieldOpt.FieldName)
				return
			}
		}

	})

	return errSet.Err()
}
