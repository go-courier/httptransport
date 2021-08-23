package transformers

import (
	"context"
	"reflect"

	reflectx "github.com/go-courier/x/reflect"

	"github.com/go-courier/httptransport/validator"
	typesx "github.com/go-courier/x/types"
)

type FlattenParams struct {
	Parameters []RequestParameter
}

type TransformerAndOption struct {
	Transformer
	Option TransformerOption
}

func (FlattenParams) NewValidator(ctx context.Context, typ typesx.Type) (validator.Validator, error) {
	p := &FlattenParams{}
	err := p.CollectParams(ctx, typ)
	return p, err
}

func (FlattenParams) String() string {
	return "@flatten"
}

func (params *FlattenParams) Validate(v interface{}) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}
	errSet := validator.NewErrorSet()
	rv = reflectx.Indirect(rv)

	for i := range params.Parameters {
		p := params.Parameters[i]

		fieldValue := p.FieldValue(rv)

		if p.Validator != nil {
			if err := p.Validator.Validate(fieldValue); err != nil {
				errSet.AddErr(err, p.Name)
			}
		}
	}

	return errSet.Err()
}

func (params *FlattenParams) CollectParams(ctx context.Context, typ typesx.Type) error {
	err := EachRequestParameter(ctx, typesx.Deref(typ), func(rp *RequestParameter) {
		params.Parameters = append(params.Parameters, *rp)
	})
	return err
}
