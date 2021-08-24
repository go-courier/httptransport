package validator

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/go-courier/statuserror"

	reflectx "github.com/go-courier/x/reflect"
)

type Location string

func NewErrorSet(paths ...interface{}) *ErrorSet {
	return &ErrorSet{
		paths:  paths,
		errors: make([]FieldError, 0),
	}
}

type ErrorSet struct {
	paths  []interface{}
	errors []FieldError
}

func (es *ErrorSet) ToErrorFields() statuserror.ErrorFields {
	errorFields := make([]*statuserror.ErrorField, 0)

	es.Flatten().Each(func(fieldErr *FieldError) {
		if len(fieldErr.Path) > 1 {
			if l, ok := fieldErr.Path[0].(Location); ok {
				fe := &statuserror.ErrorField{
					In:    string(l),
					Field: fieldErr.Path[1:].String(),
					Msg:   fieldErr.Error.Error(),
				}
				errorFields = append(errorFields, fe)
			}
		}
	})

	return errorFields
}

func (es *ErrorSet) AddErr(err error, keyPathNodes ...interface{}) {
	if err == nil {
		return
	}
	es.errors = append(es.errors, FieldError{
		Path:  keyPathNodes,
		Error: err,
	})
}

func (es *ErrorSet) Each(cb func(fieldErr *FieldError)) {
	for i := range es.errors {
		cb(&es.errors[i])
	}
}

func (es *ErrorSet) Flatten() *ErrorSet {
	flattened := NewErrorSet(es.paths...)

	var walk func(es *ErrorSet, parents ...interface{})

	walk = func(es *ErrorSet, parents ...interface{}) {
		es.Each(func(fieldErr *FieldError) {
			if subSet, ok := fieldErr.Error.(*ErrorSet); ok {
				walk(subSet, append(parents, fieldErr.Path...)...)
			} else {
				flattened.AddErr(fieldErr.Error, append(parents, fieldErr.Path...)...)
			}
		})
	}

	walk(es)

	return flattened
}

func (es *ErrorSet) Len() int {
	return len(es.errors)
}

func (es *ErrorSet) Err() error {
	if len(es.errors) == 0 {
		return nil
	}
	return es
}

func (es *ErrorSet) Error() string {
	set := es.Flatten()

	buf := bytes.Buffer{}

	set.Each(func(fieldErr *FieldError) {
		buf.WriteString(fmt.Sprintf("%s %s", fieldErr.Path, fieldErr.Error))
		buf.WriteRune('\n')
	})

	return buf.String()
}

type FieldError struct {
	Path  KeyPath
	Error error
}

type KeyPath []interface{}

func (keyPath KeyPath) String() string {
	buf := &bytes.Buffer{}
	for i := 0; i < len(keyPath); i++ {
		switch keyOrIndex := keyPath[i].(type) {
		case string:
			if buf.Len() > 0 {
				buf.WriteRune('.')
			}
			buf.WriteString(keyOrIndex)
		case int:
			buf.WriteString(fmt.Sprintf("[%d]", keyOrIndex))
		}
	}
	return buf.String()
}

type MissingRequired struct{}

func (MissingRequired) Error() string {
	return "missing required field"
}

type NotMatchError struct {
	Target  string
	Current interface{}
	Pattern string
}

func (err *NotMatchError) Error() string {
	return fmt.Sprintf("%s %s not match %v", err.Target, err.Pattern, err.Current)
}

type MultipleOfError struct {
	Target     string
	Current    interface{}
	MultipleOf interface{}
}

func (e *MultipleOfError) Error() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(e.Target)
	buf.WriteString(fmt.Sprintf(" should be multiple of %v", e.MultipleOf))
	buf.WriteString(fmt.Sprintf(", but got invalid value %v", e.Current))
	return buf.String()
}

type NotInEnumError struct {
	Target  string
	Current interface{}
	Enums   []interface{}
}

func (e *NotInEnumError) Error() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(e.Target)
	buf.WriteString(" should be one of ")

	for i, v := range e.Enums {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%v", v))
	}

	buf.WriteString(fmt.Sprintf(", but got invalid value %v", e.Current))

	return buf.String()
}

type OutOfRangeError struct {
	Target           string
	Current          interface{}
	Minimum          interface{}
	Maximum          interface{}
	ExclusiveMaximum bool
	ExclusiveMinimum bool
}

func (e *OutOfRangeError) Error() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(e.Target)
	buf.WriteString(" should be")

	if e.Minimum != nil {
		buf.WriteString(" larger")
		if e.ExclusiveMinimum {
			buf.WriteString(" or equal")
		}

		buf.WriteString(fmt.Sprintf(" than %v", reflectx.Indirect(reflect.ValueOf(e.Minimum)).Interface()))
	}

	if e.Maximum != nil {
		if e.Minimum != nil {
			buf.WriteString(" and")
		}

		buf.WriteString(" less")
		if e.ExclusiveMaximum {
			buf.WriteString(" or equal")
		}

		buf.WriteString(fmt.Sprintf(" than %v", reflectx.Indirect(reflect.ValueOf(e.Maximum)).Interface()))
	}

	buf.WriteString(fmt.Sprintf(", but got invalid value %v", e.Current))

	return buf.String()
}

func NewUnsupportedTypeError(typ string, rule string, msgs ...string) *UnsupportedTypeError {
	return &UnsupportedTypeError{
		rule: rule,
		typ:  typ,
		msgs: msgs,
	}
}

type UnsupportedTypeError struct {
	msgs []string
	rule string
	typ  string
}

func (e *UnsupportedTypeError) Error() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(e.rule)
	buf.WriteString(" could not validate type ")
	buf.WriteString(e.typ)

	for i, msg := range e.msgs {
		if i == 0 {
			buf.WriteString(": ")
		} else {
			buf.WriteString("; ")
		}
		buf.WriteString(msg)
	}

	return buf.String()
}
