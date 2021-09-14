package transformers

import (
	"context"
	"reflect"

	"github.com/go-courier/httptransport/validator"
	typesx "github.com/go-courier/x/types"
)

type RequestParameter struct {
	Parameter

	TransformerOption TransformerOption
	Transformer       Transformer

	Validator validator.Validator
}

func EachRequestParameter(ctx context.Context, tpe typesx.Type, each func(rp *RequestParameter)) error {
	errSet := validator.NewErrorSet()

	EachParameter(ctx, tpe, func(p *Parameter) bool {
		rp := &RequestParameter{}
		rp.Parameter = *p

		rp.TransformerOption.Name = rp.Name

		if flagTags, ok := rp.Tags["name"]; ok {
			rp.TransformerOption.Omitempty = flagTags.HasFlag("omitempty")
		}

		if tagFlags, ok := rp.Tags["mime"]; ok {
			rp.TransformerOption.MIME = tagFlags.Name()
		}

		if rp.In == "path" {
			rp.TransformerOption.Omitempty = false
		}

		switch rp.Type.Kind() {
		case reflect.Array, reflect.Slice:
			if !(rp.Type.Elem().PkgPath() == "" && rp.Type.Elem().Kind() == reflect.Uint8) {
				rp.TransformerOption.Explode = true
			}
		}

		getTransformer := func() (Transformer, error) {
			if rp.TransformerOption.Explode {
				return NewTransformer(ctx, rp.Type.Elem(), rp.TransformerOption)
			}
			return NewTransformer(ctx, rp.Type, rp.TransformerOption)
		}

		transformer, err := getTransformer()
		if err != nil {
			errSet.AddErr(err, rp.Name)
			return false
		}
		rp.Transformer = transformer

		parameterValidator, err := NewValidator(ctx, rp.Type, rp.Tags, rp.TransformerOption.Omitempty, transformer)
		if err != nil {
			errSet.AddErr(err, rp.Name)
			return false
		}
		rp.Validator = parameterValidator

		each(rp)

		return true
	})

	if errSet.Len() == 0 {
		return nil
	}

	return errSet.Err()
}
