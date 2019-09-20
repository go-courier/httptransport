package transformers

import (
	"context"
	"go/ast"
	"reflect"

	"github.com/go-courier/reflectx"
	"github.com/go-courier/reflectx/typesutil"
	"github.com/go-courier/validator"
	"github.com/go-courier/validator/errors"
)

type FlattenParams struct {
	fieldTransformers map[string]Transformer
	fieldOpts         map[string]TransformerOption
	validators        map[string]validator.Validator
}

func (params *FlattenParams) NewValidator(ctx context.Context, typ typesutil.Type) (validator.Validator, error) {
	typ = typesutil.Deref(typ)
	params.validators = map[string]validator.Validator{}

	errSet := errors.NewErrorSet("")

	typesutil.EachField(typ, "name", func(field typesutil.StructField, fieldDisplayName string, omitempty bool) bool {
		fieldName := field.Name()
		fieldValidator, err := NewValidator(ctx, field, field.Tag().Get("validate"), omitempty, params.fieldTransformers[fieldName])
		if err != nil {
			errSet.AddErr(err, fieldName)
			return true
		}
		params.validators[fieldName] = fieldValidator
		return true
	})

	return params, errSet.Err()
}

func (FlattenParams) String() string {
	return "@flatten"
}

func (params *FlattenParams) Validate(v interface{}) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}
	errSet := errors.NewErrorSet("")
	params.validate(rv, errSet)
	return errSet.Err()
}

func (params *FlattenParams) validate(rv reflect.Value, errSet *errors.ErrorSet) {
	typ := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := rv.Field(i)
		fieldName, _, exists := typesutil.FieldDisplayName(field.Tag, "name", field.Name)

		if !ast.IsExported(field.Name) || fieldName == "-" {
			continue
		}

		fieldType := reflectx.Deref(field.Type)
		isStructType := fieldType.Kind() == reflect.Struct

		if field.Anonymous && isStructType && !exists {
			if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
				fieldValue = reflectx.New(fieldType)
			}
			params.validate(fieldValue, errSet)
			continue
		}

		if fieldValidator, ok := params.validators[field.Name]; ok && fieldValidator != nil {
			err := fieldValidator.Validate(fieldValue)
			errSet.AddErr(err, fieldName)
		}
	}
}
func (params *FlattenParams) CollectParams(ctx context.Context, typ typesutil.Type) error {
	params.fieldTransformers = map[string]Transformer{}
	params.fieldOpts = map[string]TransformerOption{}

	errSet := errors.NewErrorSet("")

	mgr := TransformerMgrFromContext(ctx)

	typesutil.EachField(typ, "name", func(field typesutil.StructField, fieldDisplayName string, omitempty bool) bool {
		opt := TransformerOptionFromStructField(field)
		targetType := field.Type()
		fieldName := field.Name()

		if !IsBytes(targetType) {
			switch targetType.Kind() {
			case reflect.Array, reflect.Slice:
				opt.Explode = true
				targetType = targetType.Elem()
			}
		}

		fieldTransformer, err := mgr.NewTransformer(ctx, targetType, opt)
		if err != nil {
			errSet.AddErr(err, fieldName)
			return true
		}

		params.fieldTransformers[fieldName] = fieldTransformer
		params.fieldOpts[fieldName] = opt
		return true
	})

	return errSet.Err()
}

type MayValidator interface {
	NewValidator(ctx context.Context, typ typesutil.Type) (validator.Validator, error)
}

func NewValidator(ctx context.Context, field typesutil.StructField, validateStr string, omitempty bool, transformer Transformer) (validator.Validator, error) {
	namedTagKey := transformer.NamedByTag()
	if namedTagKey != "" {
		ctx = validator.ContextWithNamedTagKey(ctx, namedTagKey)
	}

	if t, ok := transformer.(MayValidator); ok {
		return t.NewValidator(ctx, field.Type())
	}

	mgr := validator.ValidatorMgrFromContext(ctx)

	return mgr.Compile(ctx, []byte(validateStr), field.Type(), func(rule *validator.Rule) {
		if omitempty {
			rule.Optional = true
		}
		if defaultValue, ok := field.Tag().Lookup("default"); ok {
			rule.DefaultValue = []byte(defaultValue)
		}
	})
}
