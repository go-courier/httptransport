package transformers

import (
	"context"
	"io"
	"reflect"

	pkgerrors "github.com/pkg/errors"

	validatorerrors "github.com/go-courier/httptransport/validator"
	reflectx "github.com/go-courier/x/reflect"
)

func NewTransformerSuper(transformer Transformer, opt *CommonTransformOption) *TransformerSuper {
	return &TransformerSuper{
		transformer:           transformer,
		CommonTransformOption: *opt,
	}
}

type TransformerSuper struct {
	transformer Transformer
	CommonTransformOption
}

func (t *TransformerSuper) EncodeTo(ctx context.Context, w io.Writer, v interface{}) error {
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

		if writerCreator, ok := w.(CanNextWriter); ok {
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
		writerCreator, ok := w.(CanNextWriter)
		if ok {
			return t.transformer.EncodeTo(ctx, writerCreator.NextWriter(), rv)
		}
		return t.transformer.EncodeTo(ctx, w, rv)
	}

	return nil
}

func (t *TransformerSuper) DecodeFrom(ctx context.Context, r io.Reader, v interface{}) error {
	if rv, ok := v.(reflect.Value); ok {
		v = rv.Interface()
	}

	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return pkgerrors.Errorf("decode target must be ptr value")
	}

	if t.Explode {
		lenOfValues := 0

		if canLen, ok := r.(interface{ Len() int }); ok {
			lenOfValues = canLen.Len()
		}

		if lenOfValues == 0 {
			return nil
		}

		if x, ok := v.(*[]string); ok {
			if canInterface, ok := r.(CanInterface); ok {
				if values, ok := canInterface.Interface().([]string); ok {
					*x = values
					return nil
				}
			}
		}

		sliceOrArrayRv := reflectx.Indirect(reflect.ValueOf(v))

		if sliceOrArrayRv.Kind() == reflect.Slice {
			// only slice should set be set
			sliceOrArrayRv.Set(reflect.MakeSlice(sliceOrArrayRv.Type(), lenOfValues, lenOfValues))
		}

		readerCreator, ok := r.(CanNextReader)
		if !ok {
			return nil
		}

		es := validatorerrors.NewErrorSet()

		for i := 0; i < sliceOrArrayRv.Len(); i++ {
			if i < lenOfValues {
				itemValue := sliceOrArrayRv.Index(i)

				// ignore when values length greater than array len
				if err := t.transformer.DecodeFrom(ctx, readerCreator.NextReader(), itemValue.Addr()); err != nil {
					es.AddErr(err, i)
				}
			}
		}

		return es.Err()
	}

	readerCreator, ok := r.(CanNextReader)
	if ok {
		return t.transformer.DecodeFrom(ctx, readerCreator.NextReader(), v)
	}
	return t.transformer.DecodeFrom(ctx, r, v)
}

type CanNextWriter interface {
	NextWriter() io.Writer
}

type CanNextReader interface {
	NextReader() io.Reader
}
