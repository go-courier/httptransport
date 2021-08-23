package validator

import (
	"context"
	"encoding"
	stderrors "errors"
	"fmt"
	"reflect"

	"github.com/go-courier/httptransport/validator/rules"
	encodingx "github.com/go-courier/x/encoding"
	reflectx "github.com/go-courier/x/reflect"
	typesutil "github.com/go-courier/x/types"
)

func NewValidatorLoader(validatorCreator ValidatorCreator) *ValidatorLoader {
	return &ValidatorLoader{
		ValidatorCreator: validatorCreator,
	}
}

type ValidatorLoader struct {
	ValidatorCreator ValidatorCreator
	Validator
	PreprocessStage

	DefaultValue []byte
	Optional     bool
	ErrMsg       []byte
}

type PreprocessStage int

const (
	PreprocessSkip PreprocessStage = iota
	PreprocessString
	PreprocessPtr
)

func normalize(typ typesutil.Type) (typesutil.Type, PreprocessStage) {
	if t, ok := typesutil.EncodingTextMarshalerTypeReplacer(typ); ok {
		return t, PreprocessString
	}
	if typ.Kind() == reflect.Ptr {
		return typesutil.Deref(typ), PreprocessPtr
	}
	return typ, PreprocessSkip
}

func (loader *ValidatorLoader) String() string {
	if loader.Validator != nil {
		v := loader.Validator.String()

		if loader.Optional {
			if loader.DefaultValue != nil {
				return v + " = " + string(rules.SingleQuote(loader.DefaultValue))
			}
			return v + "?"
		}

		return v
	}
	return "nil"
}

func (loader *ValidatorLoader) New(ctx context.Context, rule *Rule) (Validator, error) {
	l := NewValidatorLoader(loader.ValidatorCreator)

	l.Optional = rule.Optional
	l.DefaultValue = rule.DefaultValue
	l.ErrMsg = rule.ErrMsg

	typ := rule.Type

	rule.Type, l.PreprocessStage = normalize(rule.Type)

	if loader.ValidatorCreator != nil {
		v, err := loader.ValidatorCreator.New(ctx, rule)
		if err != nil {
			return nil, err
		}
		l.Validator = v

		if l.DefaultValue != nil {
			if rv, ok := typesutil.TryNew(typ); ok {
				if err := encodingx.UnmarshalText(rv, l.DefaultValue); err != nil {
					return nil, fmt.Errorf("default value `%s` can not unmarshal to %s: %s", l.DefaultValue, typ, err)
				}

				if err := l.Validate(rv); err != nil {
					return nil, fmt.Errorf("default value `%s` is not a valid value of %s: %s", l.DefaultValue, v, err)
				}
			}
		}
	}

	return l, nil
}

func (loader *ValidatorLoader) Validate(v interface{}) error {
	err := loader.validate(v)
	if err == nil {
		return nil
	}
	if loader.ErrMsg != nil && len(loader.ErrMsg) != 0 {
		return stderrors.New(string(loader.ErrMsg))
	}
	return err
}

func (loader *ValidatorLoader) validate(v interface{}) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	if reflectx.IsEmptyValue(rv) {
		if !loader.Optional {
			return MissingRequired{}
		}

		if loader.DefaultValue != nil && rv.CanSet() {
			err := encodingx.UnmarshalText(rv, loader.DefaultValue)
			if err != nil {
				return fmt.Errorf("unmarshal default value failed")
			}
		}
		return nil
	}

	if loader.Validator == nil {
		return nil
	}

	if loader.PreprocessStage == PreprocessString {
		// make sure value over reflect.Value
		if rv.CanInterface() {
			v = rv.Interface()
		}

		if textMarshaller, ok := v.(encoding.TextMarshaler); ok {
			data, err := textMarshaller.MarshalText()
			if err != nil {
				return err
			}
			return loader.Validator.Validate(string(data))
		}
	}

	if rv.Kind() == reflect.Interface {
		rv = rv.Elem()
	}

	return loader.Validator.Validate(reflectx.Indirect(rv))
}
