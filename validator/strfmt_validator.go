package validator

import (
	"context"
	"reflect"
	"regexp"
)

func NewRegexpStrfmtValidator(regexpStr string, name string, aliases ...string) *StrfmtValidator {
	re := regexp.MustCompile(regexpStr)
	validate := func(v interface{}) error {
		if !re.MatchString(v.(string)) {
			return &NotMatchError{
				Target:  name,
				Current: v,
				Pattern: re.String(),
			}
		}
		return nil
	}
	return NewStrfmtValidator(validate, name, aliases...)
}

func NewStrfmtValidator(validate func(v interface{}) error, name string, aliases ...string) *StrfmtValidator {
	return &StrfmtValidator{
		names:    append([]string{name}, aliases...),
		validate: validate,
	}
}

type StrfmtValidator struct {
	names    []string
	validate func(v interface{}) error
}

func (validator *StrfmtValidator) String() string {
	return "@" + validator.names[0]
}

func (validator *StrfmtValidator) Names() []string {
	return validator.names
}

func (validator StrfmtValidator) New(ctx context.Context, rule *Rule) (Validator, error) {
	return &validator, validator.TypeCheck(rule)
}

func (validator *StrfmtValidator) TypeCheck(rule *Rule) error {
	if rule.Type.Kind() == reflect.String {
		return nil
	}
	return NewUnsupportedTypeError(rule.String(), validator.String())
}

func (validator *StrfmtValidator) Validate(v interface{}) error {
	if rv, ok := v.(reflect.Value); ok && rv.CanInterface() {
		v = rv.Interface()
	}
	s, ok := v.(string)
	if !ok {
		return NewUnsupportedTypeError(reflect.TypeOf(v).String(), validator.String())
	}
	return validator.validate(s)
}
