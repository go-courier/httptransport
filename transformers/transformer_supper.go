package transformers

import (
	"context"
	"io"
	"reflect"

	validatorerrors "github.com/go-courier/httptransport/validator"
	reflectx "github.com/go-courier/x/reflect"
)

func NewSupperTransformer(transformer Transformer, opt *TransformerOption) *TransformerSupper {
	return &TransformerSupper{
		transformer:       transformer,
		TransformerOption: *opt,
	}
}

type TransformerSupper struct {
	transformer Transformer
	TransformerOption
}

func (t *TransformerSupper) EncodeTo(ctx context.Context, w io.Writer, v interface{}) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	if t.Explode {
		rv = reflectx.Indirect(rv)

		// for create slice
		if canSetN, ok := w.(interface{ SetN(n int) }); ok {
			canSetN.SetN(rv.Len())
		}

		if writerCreator, ok := w.(WriterCreator); ok {
			errSet := validatorerrors.NewErrorSet()

			for i := 0; i < rv.Len(); i++ {
				w := writerCreator.NextWriter()

				if err := t.transformer.EncodeTo(ctx, w, rv.Index(i)); err != nil {
					errSet.AddErr(err, i)
				}
			}

			return errSet.Err()
		}

		return nil
	}

	// should skip empty value when omitempty
	if !(t.Omitempty && reflectx.IsEmptyValue(rv)) {
		writerCreator, ok := w.(WriterCreator)
		if ok {
			return t.transformer.EncodeTo(ctx, writerCreator.NextWriter(), rv)
		}
		return t.transformer.EncodeTo(ctx, w, rv)
	}

	return nil
}

func (t *TransformerSupper) DecodeFrom(ctx context.Context, r io.Reader, v interface{}) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(rv)
	}

	if t.Explode {
		sliceOrArrayRv := reflectx.Indirect(rv)

		lenOfValues := 0
		if canLen, ok := r.(interface{ Len() int }); ok {
			lenOfValues = canLen.Len()
		}

		if lenOfValues == 0 {
			return nil
		}

		if sliceOrArrayRv.Kind() == reflect.Slice {
			// only slice should set be set
			sliceOrArrayRv.Set(reflect.MakeSlice(sliceOrArrayRv.Type(), lenOfValues, lenOfValues))
		}

		readerCreator, ok := r.(ReaderCreator)
		if !ok {
			return nil
		}

		es := validatorerrors.NewErrorSet()

		for i := 0; i < sliceOrArrayRv.Len(); i++ {
			if i < lenOfValues {
				itemValue := sliceOrArrayRv.Index(i)

				// ensure new a ptr value
				if itemValue.Kind() == reflect.Ptr {
					if itemValue.IsNil() {
						itemValue.Set(reflectx.New(itemValue.Type()))
					}
				}

				r := readerCreator.NextReader()

				// ignore when values length greater than array len
				if err := t.transformer.DecodeFrom(ctx, r, itemValue); err != nil {
					es.AddErr(err, i)
				}
			}
		}

		return es.Err()
	}

	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflectx.New(rv.Type()))
		}
	}

	readerCreator, ok := r.(ReaderCreator)
	if ok {
		return t.transformer.DecodeFrom(ctx, readerCreator.NextReader(), rv)
	}
	return t.transformer.DecodeFrom(ctx, r, rv)
}

type WriterCreator interface {
	NextWriter() io.Writer
}

type ReaderCreator interface {
	NextReader() io.Reader
}
