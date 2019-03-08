package transformers

import (
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

func (params *FlattenParams) NewValidator(typ typesutil.Type, mgr validator.ValidatorMgr) (validator.Validator, error) {
	typ = typesutil.Deref(typ)
	params.validators = map[string]validator.Validator{}

	errSet := errors.NewErrorSet("")

	typesutil.EachField(typ, "name", func(field typesutil.StructField, fieldDisplayName string, omitempty bool) bool {
		fieldName := field.Name()
		fieldValidator, err := NewValidator(field, field.Tag().Get("validate"), omitempty, params.fieldTransformers[fieldName], mgr)
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

		if fieldValidator, ok := params.validators[field.Name]; ok {
			err := fieldValidator.Validate(fieldValue)
			errSet.AddErr(err, fieldName)
		}
	}
}
func (params *FlattenParams) CollectParams(typ typesutil.Type, mgr TransformerMgr) error {
	params.fieldTransformers = map[string]Transformer{}
	params.fieldOpts = map[string]TransformerOption{}

	errSet := errors.NewErrorSet("")

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

		fieldTransformer, err := mgr.NewTransformer(targetType, opt)
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

func NewValidator(field typesutil.StructField, validateStr string, omitempty bool, transformer Transformer, mgr validator.ValidatorMgr) (validator.Validator, error) {
	if validateStr == "" && typesutil.Deref(field.Type()).Kind() == reflect.Struct {
		if _, ok := typesutil.EncodingTextMarshalerTypeReplacer(field.Type()); !ok {
			validateStr = "@struct" + "<" + transformer.NamedByTag() + ">"
		}
	}

	if t, ok := transformer.(interface {
		NewValidator(typ typesutil.Type, mgr validator.ValidatorMgr) (validator.Validator, error)
	}); ok {
		return t.NewValidator(field.Type(), mgr)
	}

	return mgr.Compile([]byte(validateStr), field.Type(), func(rule *validator.Rule) {
		if omitempty {
			rule.Optional = true
		}
		if defaultValue, ok := field.Tag().Lookup("default"); ok {
			rule.DefaultValue = []byte(defaultValue)
		}
	})
}
