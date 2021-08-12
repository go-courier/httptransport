package validator

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/go-courier/x/ptr"
	typesutil "github.com/go-courier/x/types"
	. "github.com/onsi/gomega"
)

func TestNewValidatorLoader(t *testing.T) {
	type SomeStruct struct {
		PtrString *string
		String    string
	}

	var val *string

	someStruct := &SomeStruct{}

	cases := []struct {
		valuesPass   []interface{}
		valuesFailed []interface{}
		rule         string
		typ          reflect.Type
		validator    *ValidatorLoader
	}{
		{
			[]interface{}{
				reflect.ValueOf(someStruct).Elem().FieldByName("String"),
				"1",
			},
			[]interface{}{"222"},
			"@string[1,2] = '1'",
			reflect.TypeOf(""),
			&ValidatorLoader{
				Optional:        true,
				DefaultValue:    []byte("1"),
				PreprocessStage: PreprocessSkip,
			},
		},
		{
			[]interface{}{
				Duration(1 * time.Second),
				Duration(1 * time.Second),
			},
			[]interface{}{},
			"@string",
			reflect.TypeOf(Duration(1 * time.Second)),
			&ValidatorLoader{
				PreprocessStage: PreprocessString,
			},
		},
		{
			[]interface{}{
				val,
				reflect.ValueOf(val),
				reflect.ValueOf(someStruct).Elem().FieldByName("String"),
				ptr.String("1"),
			},
			[]interface{}{
				ptr.String("222"),
			},
			"@string[1,2] = 2",
			reflect.TypeOf(ptr.String("")),
			&ValidatorLoader{
				Optional:        true,
				DefaultValue:    []byte("2"),
				PreprocessStage: PreprocessPtr,
			},
		},
		{
			[]interface{}{
				ptr.String("1"),
				ptr.String("22"),
			},
			[]interface{}{
				ptr.String(""),
				(*string)(nil),
			},
			"@string[1,2]",
			reflect.TypeOf(ptr.String("")),
			&ValidatorLoader{
				PreprocessStage: PreprocessPtr,
			},
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s/%s", c.typ, c.rule), func(t *testing.T) {
			validator, err := ValidatorMgrDefault.Compile(context.Background(), []byte(c.rule), typesutil.FromRType(c.typ), nil)
			NewWithT(t).Expect(err).To(BeNil())

			loader := validator.(*ValidatorLoader)

			NewWithT(t).Expect(loader.Optional).To(Equal(c.validator.Optional))
			NewWithT(t).Expect(loader.PreprocessStage).To(Equal(c.validator.PreprocessStage))
			NewWithT(t).Expect(loader.DefaultValue).To(Equal(c.validator.DefaultValue))

			for i, v := range c.valuesPass {
				t.Run(fmt.Sprintf("%d should success", i), func(t *testing.T) {
					err := loader.Validate(v)
					NewWithT(t).Expect(err).To(BeNil())
				})
			}

			for i, v := range c.valuesFailed {
				t.Run(fmt.Sprintf("%d should failed", i), func(t *testing.T) {
					err := loader.Validate(v)
					NewWithT(t).Expect(err).NotTo(BeNil())
				})
			}
		})
	}
}

func TestNewValidatorLoaderFailed(t *testing.T) {
	invalidRules := map[reflect.Type][]string{
		reflect.TypeOf(1): {
			"@string",
			"@int[1,2] = s",
		},
		reflect.TypeOf(""): {
			"@string<length, 1>",
			"@string[1,2] = 123",
		},
		reflect.TypeOf(Duration(1)): {
			"@string[,10] = 2ss",
		},
	}

	for typ := range invalidRules {
		for _, r := range invalidRules[typ] {
			t.Run(fmt.Sprintf("%s validate %s", typ, r), func(t *testing.T) {
				_, err := ValidatorMgrDefault.Compile(context.Background(), []byte(r), typesutil.FromRType(typ), nil)
				NewWithT(t).Expect(err).NotTo(BeNil())
				t.Log(err)
			})
		}
	}
}

type Duration time.Duration

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

func (d *Duration) UnmarshalText(data []byte) error {
	dur, err := time.ParseDuration(string(data))
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}
