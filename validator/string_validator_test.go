package validator

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/go-courier/x/ptr"
	typesutil "github.com/go-courier/x/types"
	. "github.com/onsi/gomega"
)

func TestStringValidator_New(t *testing.T) {
	caseSet := map[reflect.Type][]struct {
		rule   string
		expect *StringValidator
	}{
		reflect.TypeOf(""): {
			{"@string[1,1000]", &StringValidator{
				MinLength: 1,
				MaxLength: ptr.Uint64(1000),
			}},
			{"@string[1,]", &StringValidator{
				MinLength: 1,
			}},
			{"@string<length>[1]", &StringValidator{
				MinLength: 1,
				MaxLength: ptr.Uint64(1),
			}},
			{"@char[1,]", &StringValidator{
				LenMode:   StrLenModeRuneCount,
				MinLength: 1,
			}},
			{"@string<rune_count>[1,]", &StringValidator{
				LenMode:   StrLenModeRuneCount,
				MinLength: 1,
			}},
			{"@string{KEY1,KEY2}", &StringValidator{
				Enums: []string{
					"KEY1",
					"KEY2",
				},
			}},
			{`@string/^\w+/`, &StringValidator{
				Pattern: `^\w+`,
			}},
			{`@string/^\w+\/test/`, &StringValidator{
				Pattern: `^\w+/test`,
			}},
		},
	}

	for typ, cases := range caseSet {
		for _, c := range cases {
			t.Run(fmt.Sprintf("%s %s|%s", typ, c.rule, c.expect.String()), func(t *testing.T) {
				v, err := c.expect.New(ContextWithValidatorMgr(context.Background(), ValidatorMgrDefault), MustParseRuleStringWithType(c.rule, typesutil.FromRType(typ)))
				NewWithT(t).Expect(err).To(BeNil())
				NewWithT(t).Expect(v).To(Equal(c.expect))
			})
		}
	}
}

func TestStringValidator_NewFailed(t *testing.T) {
	invalidRules := map[reflect.Type][]string{
		reflect.TypeOf(1): {
			"@string",
		},
		reflect.TypeOf(""): {
			"@string<length, 1>",
			"@string<unsupported>",
			"@string[1,0]",
			"@string[1,-2]",
			"@string[a,]",
			"@string[-1,1]",
			"@string(-1,1)",
		},
	}

	validator := &StringValidator{}

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

func TestStringValidator_Validate(t *testing.T) {
	type String string

	cases := []struct {
		values    []interface{}
		validator *StringValidator
		desc      string
	}{
		{[]interface{}{reflect.ValueOf("a"), String("aa"), "aaa", "aaaa", "aaaaa"}, &StringValidator{
			MaxLength: ptr.Uint64(5),
		}, "less than"},
		{[]interface{}{"一", "一一", "一一一"}, &StringValidator{
			LenMode:   StrLenModeRuneCount,
			MaxLength: ptr.Uint64(3),
		}, "char count less than"},
		{[]interface{}{"A", "B"}, &StringValidator{
			Enums: []string{
				"A",
				"B",
			},
		}, "in enum"},
		{[]interface{}{"word", "word1"}, &StringValidator{
			Pattern: `^\w+`,
		}, "regexp matched"},
	}

	for _, c := range cases {
		for _, v := range c.values {
			t.Run(fmt.Sprintf("%s: %s validate %v", c.desc, c.validator, v), func(t *testing.T) {
				NewWithT(t).Expect(c.validator.Validate(v)).To(BeNil())
			})
		}
	}
}

func TestStringValidator_ValidateFailed(t *testing.T) {
	type String string

	cases := []struct {
		values    []interface{}
		validator *StringValidator
		desc      string
	}{
		{[]interface{}{"C", "D", "E"}, &StringValidator{
			Enums: []string{
				"A",
				"B",
			},
		}, "enum not match"},
		{[]interface{}{"-word", "-word1"}, &StringValidator{
			Pattern: `^\w+`,
		}, "regexp not matched"},
		{[]interface{}{1.1, reflect.ValueOf(1.1)}, &StringValidator{
			MinLength: 5,
		}, "unsupported types"},
		{[]interface{}{"a", "aa", String("aaa"), []byte("aaaa")}, &StringValidator{
			MinLength: 5,
		}, "too small"},
		{[]interface{}{"aa", "aaa", "aaaa", "aaaaa"}, &StringValidator{
			MaxLength: ptr.Uint64(1),
		}, "too large"},
		{[]interface{}{"字符太多"}, &StringValidator{
			LenMode:   StrLenModeRuneCount,
			MaxLength: ptr.Uint64(3),
		}, "too many chars"},
	}

	for _, c := range cases {
		for _, v := range c.values {
			t.Run(fmt.Sprintf("%s: %s validate %v", c.desc, c.validator, v), func(t *testing.T) {
				err := c.validator.Validate(v)
				NewWithT(t).Expect(err).NotTo(BeNil())
				t.Log(err)
			})
		}
	}
}
