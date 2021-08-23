package validator

import (
	"context"
	"go/ast"
	"reflect"

	contextx "github.com/go-courier/x/context"
	reflectx "github.com/go-courier/x/reflect"
	typesx "github.com/go-courier/x/types"
)

func NewStructValidator(namedTagKey string) *StructValidator {
	return &StructValidator{
		namedTagKey:     namedTagKey,
		fieldValidators: map[string]Validator{},
	}
}

type contextKeyNamedTagKey struct{}

func ContextWithNamedTagKey(ctx context.Context, namedTagKey string) context.Context {
	return contextx.WithValue(ctx, contextKeyNamedTagKey{}, namedTagKey)
}

func NamedKeyFromContext(ctx context.Context) string {
	v := ctx.Value(contextKeyNamedTagKey{})
	if v != nil {
		if namedTagKey, ok := v.(string); ok {
			return namedTagKey
		}
	}
	return ""
}

type StructValidator struct {
	namedTagKey     string
	fieldValidators map[string]Validator
}

func init() {
	ValidatorMgrDefault.Register(&StructValidator{})
}

func (StructValidator) Names() []string {
	return []string{"struct"}
}

func (validator *StructValidator) Validate(v interface{}) error {
	switch rv := v.(type) {
	case reflect.Value:
		return validator.ValidateReflectValue(rv)
	default:
		return validator.ValidateReflectValue(reflect.ValueOf(v))
	}
}

func (validator *StructValidator) ValidateReflectValue(rv reflect.Value) error {
	errSet := NewErrorSet("")
	validator.validate(rv, errSet)
	return errSet.Err()
}

func (validator *StructValidator) validate(rv reflect.Value, errSet *ErrorSet) {
	typ := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := rv.Field(i)
		fieldName, _, exists := typesx.FieldDisplayName(field.Tag, validator.namedTagKey, field.Name)

		if !ast.IsExported(field.Name) || fieldName == "-" {
			continue
		}

		fieldType := reflectx.Deref(field.Type)
		isStructType := fieldType.Kind() == reflect.Struct

		if field.Anonymous && isStructType && !exists {
			if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
				fieldValue = reflectx.New(fieldType)
			}
			validator.validate(fieldValue, errSet)
			continue
		}

		if fieldValidator, ok := validator.fieldValidators[field.Name]; ok {
			err := fieldValidator.Validate(fieldValue)
			errSet.AddErr(err, fieldName)
		}
	}
}

const (
	TagValidate = "validate"
	TagDefault  = "default"
	TagErrMsg   = "errMsg"
)

func (validator *StructValidator) New(ctx context.Context, rule *Rule) (Validator, error) {
	if rule.Type.Kind() != reflect.Struct {
		return nil, NewUnsupportedTypeError(rule.String(), validator.String())
	}

	namedTagKey := NamedKeyFromContext(ctx)

	if rule.Rule != nil && len(rule.Params) > 0 {
		namedTagKey = string(rule.Params[0].Bytes())
	}

	if namedTagKey == "" {
		namedTagKey = validator.namedTagKey
	}

	structValidator := NewStructValidator(namedTagKey)
	errSet := NewErrorSet("")

	ctx = ContextWithNamedTagKey(ctx, structValidator.namedTagKey)

	mgr := ValidatorMgrFromContext(ctx)

	typesx.EachField(rule.Type, structValidator.namedTagKey, func(field typesx.StructField, fieldDisplayName string, omitempty bool) bool {
		tagValidateValue := field.Tag().Get(TagValidate)

		if tagValidateValue == "" && typesx.Deref(field.Type()).Kind() == reflect.Struct {
			if _, ok := typesx.EncodingTextMarshalerTypeReplacer(field.Type()); !ok {
				tagValidateValue = structValidator.String()
			}
		}

		fieldValidator, err := mgr.Compile(ContextWithNamedTagKey(ctx, namedTagKey), []byte(tagValidateValue), field.Type(), func(rule RuleModifier) {
			if omitempty {
				rule.SetOptional(omitempty)
			}
			if defaultValue, ok := field.Tag().Lookup(TagDefault); ok {
				rule.SetDefaultValue([]byte(defaultValue))
			}
			if errMsg, ok := field.Tag().Lookup(TagErrMsg); ok {
				rule.SetErrMsg([]byte(errMsg))
			}
		})

		if err != nil {
			errSet.AddErr(err, field.Name())
			return true
		}

		if fieldValidator != nil {
			structValidator.fieldValidators[field.Name()] = fieldValidator
		}
		return true
	})

	return structValidator, errSet.Err()
}

func (validator *StructValidator) String() string {
	return "@" + validator.Names()[0] + "<" + validator.namedTagKey + ">"
}
