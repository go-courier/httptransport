package validator

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/go-courier/x/ptr"
	typesutil "github.com/go-courier/x/types"
)

func TestMapValidator_New(t *testing.T) {
	caseSet := map[reflect.Type][]struct {
		rule   string
		expect *MapValidator
	}{
		reflect.TypeOf(map[string]string{}): {
			{"@map[1,1000]", &MapValidator{
				MinProperties: 1,
				MaxProperties: ptr.Uint64(1000),
			}},
		},
		reflect.TypeOf(map[string]map[string]string{}): {
			{"@map<,@map[1,2]>[1,]", &MapValidator{
				MinProperties: 1,
				ElemValidator: ValidatorMgrDefault.MustCompile(context.Background(), []byte("@map[1,2]"), typesutil.FromRType(reflect.TypeOf(map[string]string{})), nil),
			}},
			{"@map<@string[0,],@map[1,2]>[1,]", &MapValidator{
				MinProperties: 1,
				KeyValidator:  ValidatorMgrDefault.MustCompile(context.Background(), []byte("@string[0,]"), typesutil.FromRType(reflect.TypeOf("")), nil),
				ElemValidator: ValidatorMgrDefault.MustCompile(context.Background(), []byte("@map[1,2]"), typesutil.FromRType(reflect.TypeOf(map[string]string{})), nil),
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

func TestMapValidator_NewFailed(t *testing.T) {
	invalidRules := map[reflect.Type][]string{
		reflect.TypeOf([]string{}): {
			"@map",
		},
		reflect.TypeOf(map[string]string{}): {
			"@map<1,>",
			"@map<,2>",
			"@map<1,2,3>",
			"@map[1,0]",
			"@map[1,-2]",
			"@map[a,]",
			"@map[-1,1]",
			"@map(-1,1)",
			"@map<@unknown,>",
			"@map<,@unknown>",
			"@map<@string[0,],@unknown>",
		},
	}

	validator := &MapValidator{}

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

func TestMapValidator_Validate(t *testing.T) {
	cases := []struct {
		values    []interface{}
		validator *MapValidator
		desc      string
	}{
		{[]interface{}{
			map[string]string{"1": "", "2": ""},
			map[string]string{"1": "", "2": "", "3": ""},
			map[string]string{"1": "", "2": "", "3": "", "4": ""},
		}, &MapValidator{
			MinProperties: 2,
			MaxProperties: ptr.Uint64(4),
		}, "in range"},
		{[]interface{}{
			reflect.ValueOf(map[string]string{"1": "", "2": ""}),
			map[string]string{"1": "", "2": "", "3": ""},
		}, &MapValidator{
			MinProperties: 2,
			MaxProperties: ptr.Uint64(4),
			KeyValidator:  ValidatorMgrDefault.MustCompile(context.Background(), []byte("@string[1,]"), typesutil.FromRType(reflect.TypeOf("1")), nil),
			ElemValidator: ValidatorMgrDefault.MustCompile(context.Background(), []byte("@string[1,]?"), typesutil.FromRType(reflect.TypeOf("1")), nil),
		}, "key value validate"},
	}

	for _, c := range cases {
		for _, v := range c.values {
			t.Run(fmt.Sprintf("%s: %s validate %v", c.desc, c.validator, v), func(t *testing.T) {
				NewWithT(t).Expect(c.validator.Validate(v)).To(BeNil())
			})
		}
	}
}

func TestMapValidator_ValidateFailed(t *testing.T) {
	cases := []struct {
		values    []interface{}
		validator *MapValidator
		desc      string
	}{
		{[]interface{}{
			map[string]string{"1": ""},
			map[string]string{"1": "", "2": "", "3": "", "4": "", "5": ""},
			map[string]string{"1": "", "2": "", "3": "", "4": "", "5": "", "6": ""},
		}, &MapValidator{
			MinProperties: 2,
			MaxProperties: ptr.Uint64(4),
		}, "out of range"},
		{[]interface{}{
			map[string]string{"1": "", "2": ""},
			map[string]string{"1": "", "2": "", "3": ""},
		}, &MapValidator{
			MinProperties: 2,
			MaxProperties: ptr.Uint64(4),
			KeyValidator:  ValidatorMgrDefault.MustCompile(context.Background(), []byte("@string[2,]"), typesutil.FromRType(reflect.TypeOf("")), nil),
			ElemValidator: ValidatorMgrDefault.MustCompile(context.Background(), []byte("@string[2,]"), typesutil.FromRType(reflect.TypeOf("")), nil),
		}, "key elem validate failed"},
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
