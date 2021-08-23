package validator

import (
	"context"
	"encoding"
	"fmt"
	"reflect"
	"regexp"
	"unicode/utf8"

	"github.com/go-courier/httptransport/validator/rules"
)

var (
	TargetStringLength = "string length"
)

type StrLenMode string

const (
	StrLenModeLength    StrLenMode = "length"
	StrLenModeRuneCount StrLenMode = "rune_count"
)

var strLenModes = map[StrLenMode]func(s string) uint64{
	StrLenModeLength: func(s string) uint64 {
		return uint64(len(s))
	},
	StrLenModeRuneCount: func(s string) uint64 {
		return uint64(utf8.RuneCount([]byte(s)))
	},
}

/*
	Validator for string

	Rules:
		@string/regexp/
		@string{VALUE_1,VALUE_2,VALUE_3}
		@string<StrLenMode>[from,to]
		@string<StrLenMode>[length]

	ranges
		@string[min,max]
		@string[length]
		@string[1,10] // string length should large or equal than 1 and less or equal than 10
		@string[1,]  // string length should large or equal than 1 and less than the maxinum of int32
		@string[,1]  // string length should less than 1 and large or equal than 0
		@string[10]  // string length should be equal 10

	enumeration
		@string{A,B,C} // should one of these values

	regexp
		@string/\w+/ // string values should match \w+
	since we use / as wrapper for regexp, we need to use \ to escape /

	length mode in parameter
		@string<length> // use string length directly
		@string<rune_count> // use rune count as string length

	composes
		@string<>[1,]

	aliases:
		@char = @string<rune_count>
*/
type StringValidator struct {
	Pattern   string
	LenMode   StrLenMode
	MinLength uint64
	MaxLength *uint64
	Enums     []string
}

func init() {
	ValidatorMgrDefault.Register(&StringValidator{})
}

func (StringValidator) Names() []string {
	return []string{"string", "char"}
}

var typString = reflect.TypeOf("")

func (validator *StringValidator) Validate(v interface{}) error {
	var val string

	switch x := v.(type) {
	case reflect.Value:
		if !x.Type().ConvertibleTo(typString) {
			return NewUnsupportedTypeError(x.Type().String(), validator.String())
		}
		val = x.Convert(typString).String()
	case string:
		val = x
	default:
		if tm, ok := v.(encoding.TextMarshaler); ok {
			v, err := tm.MarshalText()
			if err != nil {
				return err
			}
			val = string(v)
		} else {
			rv := reflect.ValueOf(v)
			if !rv.Type().ConvertibleTo(typString) {
				return NewUnsupportedTypeError(rv.Type().String(), validator.String())
			}
			val = rv.Convert(typString).String()
		}
	}

	if validator.Enums != nil {
		enums := make([]interface{}, len(validator.Enums))
		in := false

		for i := range validator.Enums {
			enums[i] = validator.Enums[i]

			if validator.Enums[i] == val {
				in = true
				break
			}
		}

		if !in {
			return &NotInEnumError{
				Target:  "string value",
				Current: v,
				Enums:   enums,
			}
		}
		return nil
	}

	if validator.Pattern != "" {
		matched, _ := regexp.MatchString(validator.Pattern, val)
		if !matched {
			return &NotMatchError{
				Target:  TargetStringLength,
				Pattern: validator.Pattern,
				Current: v,
			}
		}
		return nil
	}

	lenMode := validator.LenMode

	if lenMode == "" {
		lenMode = StrLenModeLength
	}

	strLen := strLenModes[lenMode](val)

	if strLen < validator.MinLength {
		return &OutOfRangeError{
			Target:  TargetStringLength,
			Current: strLen,
			Minimum: validator.MinLength,
		}
	}

	if validator.MaxLength != nil && strLen > *validator.MaxLength {
		return &OutOfRangeError{
			Target:  TargetStringLength,
			Current: strLen,
			Maximum: validator.MaxLength,
		}
	}
	return nil
}

func (StringValidator) New(ctx context.Context, rule *Rule) (Validator, error) {
	validator := &StringValidator{}

	if rule.ExclusiveLeft || rule.ExclusiveRight {
		return nil, rules.NewSyntaxError("range mark of %s should not be `(` or `)`", validator.Names()[0])
	}

	if rule.Params != nil {
		if len(rule.Params) != 1 {
			return nil, fmt.Errorf("string should only 1 parameter, but got %d", len(rule.Params))
		}
		lenMode := StrLenMode(rule.Params[0].Bytes())
		if lenMode != StrLenModeLength && lenMode != StrLenModeRuneCount {
			return nil, fmt.Errorf("invalid len mode %s", lenMode)
		}
		if lenMode != StrLenModeLength {
			validator.LenMode = lenMode
		}
	} else if rule.Name == "char" {
		validator.LenMode = StrLenModeRuneCount
	}

	if rule.Pattern != "" {
		validator.Pattern = regexp.MustCompile(rule.Pattern).String()
		return validator, validator.TypeCheck(rule)
	}

	ruleValues := rule.ComputedValues()

	for _, v := range ruleValues {
		validator.Enums = append(validator.Enums, string(v.Bytes()))
	}

	if rule.Range != nil {
		min, max, err := UintRange(fmt.Sprintf("%s of string", validator.LenMode), 64, rule.Range...)
		if err != nil {
			return nil, err
		}
		validator.MinLength = min
		validator.MaxLength = max
	}

	return validator, validator.TypeCheck(rule)
}

func (validator *StringValidator) TypeCheck(rule *Rule) error {
	if rule.Type.Kind() == reflect.String {
		return nil
	}
	return NewUnsupportedTypeError(rule.String(), validator.String())
}

func (validator *StringValidator) String() string {
	rule := rules.NewRule(validator.Names()[0])

	if validator.Enums != nil {
		ruleValues := make([]*rules.RuleLit, 0)
		for _, e := range validator.Enums {
			ruleValues = append(ruleValues, rules.NewRuleLit([]byte(e)))
		}
		rule.ValueMatrix = [][]*rules.RuleLit{ruleValues}
	}

	rule.Params = []rules.RuleNode{
		rules.NewRuleLit([]byte(validator.LenMode)),
	}

	if validator.Pattern != "" {
		rule.Pattern = validator.Pattern
	}

	rule.Range = RangeFromUint(validator.MinLength, validator.MaxLength)

	return string(rule.Bytes())
}
