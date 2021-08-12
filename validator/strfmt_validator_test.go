package validator

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	typesutil "github.com/go-courier/x/types"
	. "github.com/onsi/gomega"
)

func ExampleNewRegexpStrfmtValidator() {
	AlphaValidator := NewRegexpStrfmtValidator("^[a-zA-Z]+$", "alpha")

	fmt.Println(AlphaValidator.Validate("a"))
	fmt.Println(AlphaValidator.Validate("1"))
	// Output:
	// <nil>
	// alpha ^[a-zA-Z]+$ not match 1
}

func TestStrfmtValidator_Validate(t *testing.T) {
	cases := []struct {
		values    []interface{}
		validator *StrfmtValidator
	}{
		{
			[]interface{}{
				"abc",
				"a",
				"c",
			},
			NewRegexpStrfmtValidator("^[a-zA-Z]+$", "alpha"),
		},
	}

	for _, c := range cases {
		for _, v := range c.values {
			t.Run(fmt.Sprintf("%s validate %s", c.validator, v), func(t *testing.T) {
				f := NewValidatorFactory()
				f.Register(c.validator)

				validator, _ := f.Compile(context.Background(), []byte("@alpha"), typesutil.FromRType(reflect.TypeOf("")), nil)
				err := validator.Validate(v)
				NewWithT(t).Expect(err).To(BeNil())
			})
		}
	}
}

func TestStrfmtValidator_ValidateFailed(t *testing.T) {
	cases := []struct {
		values    []interface{}
		validator *StrfmtValidator
	}{
		{
			[]interface{}{
				1,
				"1",
				"2",
				"3",
			},
			NewRegexpStrfmtValidator("^[a-zA-Z]+$", "alpha"),
		},
	}

	for _, c := range cases {
		for _, v := range c.values {
			t.Run(fmt.Sprintf("%s validate failed %s", c.validator, v), func(t *testing.T) {
				err := c.validator.Validate(v)
				NewWithT(t).Expect(err).NotTo(BeNil())
			})
		}
	}
}
