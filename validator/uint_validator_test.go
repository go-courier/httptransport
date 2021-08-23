package validator

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	typesutil "github.com/go-courier/x/types"
	. "github.com/onsi/gomega"
)

func TestUintValidator_New(t *testing.T) {
	caseSet := map[reflect.Type][]struct {
		rule   string
		expect *UintValidator
	}{
		reflect.TypeOf(uint8(1)): {
			{"@uint8", &UintValidator{
				BitSize: 8,
				Maximum: MaxUint(8),
			}},
		},
		reflect.TypeOf(uint16(1)): {
			{"@uint16", &UintValidator{
				BitSize: 16,
				Maximum: MaxUint(16),
			}},
		},
		reflect.TypeOf(uint(1)): {
			{"@uint8[1,]", &UintValidator{
				BitSize: 8,
				Minimum: 1,
				Maximum: MaxUint(8),
			}},
			{"@uint[1,1000)", &UintValidator{
				Minimum:          1,
				Maximum:          1000,
				ExclusiveMaximum: true,
			}},
			{"@uint(1,1000]", &UintValidator{
				Minimum:          1,
				Maximum:          1000,
				ExclusiveMinimum: true,
			}},
			{"@uint[1,]", &UintValidator{
				Minimum: 1,
				Maximum: MaxUint(32),
			}},
			{"@uint16{1,2}", &UintValidator{
				BitSize: 16,
				Enums: []uint64{
					1,
					2,
				},
			}},
			{"@uint16{%2}", &UintValidator{
				BitSize:    16,
				MultipleOf: 2,
			}},
		},
		reflect.TypeOf(uint64(1)): {
			{"@uint<53>", &UintValidator{
				BitSize: 53,
				Maximum: MaxUint(53),
			}},
			{"@uint64", &UintValidator{
				BitSize: 64,
				Maximum: MaxUint(64),
			}},
		},
	}

	for typ, cases := range caseSet {
		for _, c := range cases {
			c.expect.SetDefaults()

			t.Run(fmt.Sprintf("%s %s|%s", typ, c.rule, c.expect.String()), func(t *testing.T) {
				v, err := c.expect.New(ContextWithValidatorMgr(context.Background(), ValidatorMgrDefault), MustParseRuleStringWithType(c.rule, typesutil.FromRType(typ)))
				NewWithT(t).Expect(err).To(BeNil())
				NewWithT(t).Expect(v).To(Equal(c.expect))
			})
		}
	}
}

func TestUintValidator_ParseFailed(t *testing.T) {
	invalidRules := map[reflect.Type][]string{
		reflect.TypeOf(float32(1)): {
			"@uint16",
		},
		reflect.TypeOf(uint8(1)): {
			"@uint16",
		},
		reflect.TypeOf(uint16(1)): {
			"@uint",
		},
		reflect.TypeOf(uint32(1)): {
			"@uint64",
		},
		reflect.TypeOf(uint64(1)): {
			"@uint<32,2123>",
			"@uint<@string>",
			"@uint<66>",
			"@uint[1,0]",
			"@uint[1,-2]",
			"@uint[a,]",
			"@uint[-1,1]",
			"@uint(-1,1)",
			`@uint8{%a}`,
			`@uint8{A,B,C}`,
		},
	}

	validator := &UintValidator{}

	for typ := range invalidRules {
		for _, r := range invalidRules[typ] {
			rule := MustParseRuleStringWithType(r, typesutil.FromRType(typ))

			t.Run(fmt.Sprintf("validate %s new failed: %s", typ, rule.Bytes()), func(t *testing.T) {
				_, err := validator.New(ContextWithValidatorMgr(context.Background(), ValidatorMgrDefault), rule)
				NewWithT(t).Expect(err).NotTo(BeNil())
				t.Log(err)
			})
		}
	}
}

func TestUintValidator_Validate(t *testing.T) {
	cases := []struct {
		values    []interface{}
		validator *UintValidator
		desc      string
	}{
		{[]interface{}{reflect.ValueOf(uint(1)), uint(2), uint(3)}, &UintValidator{
			Enums: []uint64{
				1,
				2,
				3,
			},
		}, "in enum"},
		{[]interface{}{uint(2), uint(3), uint(4)}, &UintValidator{
			Minimum: 2,
			Maximum: 4,
		}, "in range"},
		{[]interface{}{uint8(2), uint16(3), uint32(4), uint64(4)}, &UintValidator{
			Minimum: 2,
			Maximum: 4,
		}, "uint types"},
		{[]interface{}{uint64(2), uint64(3), uint64(4)}, &UintValidator{
			BitSize: 64,
			Minimum: 2,
			Maximum: 4,
		}, "in range"},
		{[]interface{}{uint(2), uint(4), uint(6)}, &UintValidator{
			MultipleOf: 2,
		}, "multiple of"},
	}

	for _, c := range cases {
		c.validator.SetDefaults()
		for _, v := range c.values {
			t.Run(fmt.Sprintf("%s: %s validate %v", c.desc, c.validator, v), func(t *testing.T) {
				NewWithT(t).Expect(c.validator.Validate(v)).To(BeNil())
			})
		}
	}
}

func TestUintValidator_ValidateFailed(t *testing.T) {
	cases := []struct {
		values    []interface{}
		validator *UintValidator
		desc      string
	}{
		{[]interface{}{2, "string", reflect.ValueOf(1)}, &UintValidator{
			BitSize: 64,
		}, "unsupported type"},
		{[]interface{}{uint(4), uint(5), uint(6)}, &UintValidator{
			Enums: []uint64{
				1,
				2,
				3,
			},
		}, "not in enum"},
		{[]interface{}{uint(1), uint(4), uint(5)}, &UintValidator{
			Minimum:          2,
			Maximum:          4,
			ExclusiveMaximum: true,
		}, "not in range"},
		{[]interface{}{uint(1), uint(3), uint(5)}, &UintValidator{
			MultipleOf: 2,
		}, "not multiple of"},
	}

	for _, c := range cases {
		c.validator.SetDefaults()
		for _, v := range c.values {
			t.Run(fmt.Sprintf("%s: %s validate %v", c.desc, c.validator, v), func(t *testing.T) {
				err := c.validator.Validate(v)
				NewWithT(t).Expect(err).NotTo(BeNil())
				t.Log(err)
			})
		}
	}
}
