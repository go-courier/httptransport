package validator

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strconv"
	"unicode"

	"github.com/go-courier/httptransport/validator/rules"
	"github.com/go-courier/x/ptr"
)

var (
	TargetIntValue = "int value"
)

/*
	Validator for int

	Rules:

	ranges
		@int[min,max]
		@int[1,10] // value should large or equal than 1 and less or equal than 10
		@int(1,10] // value should large than 1 and less or equal than 10
		@int[1,10) // value should large or equal than 1

		@int[1,)  // value should large or equal than 1 and less than the maxinum of int32
		@int[,1)  // value should less than 1 and large or equal than the mininum of int32
		@int  // value should less or equal than maxinum of int32 and large or equal than the mininum of int32

	enumeration
		@int{1,2,3} // should one of these values

	multiple of some int value
		@int{%multipleOf}
		@int{%2} // should be multiple of 2

	bit size in parameter
		@int<8>
		@int<16>
		@int<32>
		@int<64>

	composes
		@int<8>[1,]

	aliases:
		@int8 = @int<8>
		@int16 = @int<16>
		@int32 = @int<32>
		@int64 = @int<64>

	Tips:
	for JavaScript https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Number/MAX_SAFE_INTEGER and https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Number/MIN_SAFE_INTEGER
		int<53>
*/
type IntValidator struct {
	BitSize uint

	Minimum          *int64
	Maximum          *int64
	MultipleOf       int64
	ExclusiveMaximum bool
	ExclusiveMinimum bool

	Enums []int64
}

func init() {
	ValidatorMgrDefault.Register(&IntValidator{})
}

func (IntValidator) Names() []string {
	return []string{"int", "int8", "int16", "int32", "int64"}
}

func (validator *IntValidator) SetDefaults() {
	if validator != nil {
		if validator.BitSize == 0 {
			validator.BitSize = 32
		}
		if validator.Maximum == nil {
			validator.Maximum = ptr.Int64(MaxInt(validator.BitSize))
		}
		if validator.Minimum == nil {
			validator.Minimum = ptr.Int64(MinInt(validator.BitSize))
		}
	}
}

func isIntType(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	}
	return false
}

func (validator *IntValidator) Validate(v interface{}) error {
	var val int64

	switch x := v.(type) {
	case reflect.Value:
		if !isIntType(x.Type()) {
			return NewUnsupportedTypeError(x.Type().String(), validator.String())
		}
		val = x.Int()
	case int:
		val = int64(x)
	case int8:
		val = int64(x)
	case int16:
		val = int64(x)
	case int32:
		val = int64(x)
	case int64:
		val = x
	default:
		return NewUnsupportedTypeError(fmt.Sprintf("%T", v), validator.String())
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
				Target:  TargetIntValue,
				Current: val,
				Enums:   enums,
			}
		}

		return nil
	}

	mininum := *validator.Minimum
	maxinum := *validator.Maximum

	if ((validator.ExclusiveMinimum && val == mininum) || val < mininum) ||
		((validator.ExclusiveMaximum && val == maxinum) || val > maxinum) {
		return &OutOfRangeError{
			Target:           TargetFloatValue,
			Current:          val,
			Minimum:          mininum,
			ExclusiveMinimum: validator.ExclusiveMinimum,
			Maximum:          maxinum,
			ExclusiveMaximum: validator.ExclusiveMaximum,
		}
	}

	if validator.MultipleOf != 0 {
		if val%validator.MultipleOf != 0 {
			return &MultipleOfError{
				Target:     TargetFloatValue,
				Current:    val,
				MultipleOf: validator.MultipleOf,
			}
		}
	}

	return nil
}

func (IntValidator) New(ctx context.Context, rule *Rule) (Validator, error) {
	validator := &IntValidator{}

	bitSizeBuf := &bytes.Buffer{}

	for _, char := range rule.Name {
		if unicode.IsDigit(char) {
			bitSizeBuf.WriteRune(char)
		}
	}

	if bitSizeBuf.Len() == 0 && rule.Params != nil {
		if len(rule.Params) != 1 {
			return nil, fmt.Errorf("int should only 1 parameter, but got %d", len(rule.Params))
		}
		bitSizeBuf.Write(rule.Params[0].Bytes())
	}

	if bitSizeBuf.Len() != 0 {
		bitSizeStr := bitSizeBuf.String()
		bitSizeNum, err := strconv.ParseUint(bitSizeStr, 10, 8)
		if err != nil || bitSizeNum > 64 {
			return nil, rules.NewSyntaxError("int parameter should be valid bit size, but got `%s`", bitSizeStr)
		}
		validator.BitSize = uint(bitSizeNum)
	}

	if validator.BitSize == 0 {
		validator.BitSize = 32
	}

	if rule.Range != nil {
		min, max, err := intRange(fmt.Sprintf("int<%d>", validator.BitSize), validator.BitSize, rule.Range...)
		if err != nil {
			return nil, err
		}
		validator.Minimum = min
		validator.Maximum = max
		validator.ExclusiveMinimum = rule.ExclusiveLeft
		validator.ExclusiveMaximum = rule.ExclusiveRight
	}

	validator.SetDefaults()

	ruleValues := rule.ComputedValues()

	if ruleValues != nil {
		if len(ruleValues) == 1 {
			mayBeMultipleOf := ruleValues[0].Bytes()
			if mayBeMultipleOf[0] == '%' {
				v := mayBeMultipleOf[1:]
				multipleOf, err := strconv.ParseInt(string(v), 10, int(validator.BitSize))
				if err != nil {
					return nil, rules.NewSyntaxError("multipleOf should be a valid int%d value, but got `%s`", validator.BitSize, v)
				}
				validator.MultipleOf = multipleOf
			}
		}

		if validator.MultipleOf == 0 {
			for _, v := range ruleValues {
				str := string(v.Bytes())
				enumValue, err := strconv.ParseInt(str, 10, int(validator.BitSize))
				if err != nil {
					return nil, rules.NewSyntaxError("enum should be a valid int%d value, but got `%s`", validator.BitSize, v)
				}
				validator.Enums = append(validator.Enums, enumValue)
			}
		}
	}

	return validator, validator.TypeCheck(rule)
}

func (validator *IntValidator) TypeCheck(rule *Rule) error {
	switch rule.Type.Kind() {
	case reflect.Int8:
		if validator.BitSize > 8 {
			return fmt.Errorf("bit size too large for type %s", rule.Type)
		}
		return nil
	case reflect.Int16:
		if validator.BitSize > 16 {
			return fmt.Errorf("bit size too large for type %s", rule.Type)
		}
		return nil
	case reflect.Int, reflect.Int32:
		if validator.BitSize > 32 {
			return fmt.Errorf("bit size too large for type %s", rule.Type)
		}
		return nil
	case reflect.Int64:
		return nil
	}
	return NewUnsupportedTypeError(rule.String(), validator.String())
}

func intRange(typ string, bitSize uint, ranges ...*rules.RuleLit) (*int64, *int64, error) {
	parseInt := func(b []byte) (*int64, error) {
		if len(b) == 0 {
			return nil, nil
		}
		n, err := strconv.ParseInt(string(b), 10, int(bitSize))
		if err != nil {
			return nil, fmt.Errorf("%s value is not correct: %s", typ, err)
		}
		return &n, nil
	}
	switch len(ranges) {
	case 2:
		min, err := parseInt(ranges[0].Bytes())
		if err != nil {
			return nil, nil, fmt.Errorf("min %s", err)
		}
		max, err := parseInt(ranges[1].Bytes())
		if err != nil {
			return nil, nil, fmt.Errorf("max %s", err)
		}
		if min != nil && max != nil && *max < *min {
			return nil, nil, fmt.Errorf("max %s value must be equal or large than min expect %d, current %d", typ, min, max)
		}

		return min, max, nil
	case 1:
		min, err := parseInt(ranges[0].Bytes())
		if err != nil {
			return nil, nil, fmt.Errorf("min %s", err)
		}
		return min, min, nil
	}
	return nil, nil, nil
}

func (validator *IntValidator) String() string {
	rule := rules.NewRule(validator.Names()[0])

	rule.Params = []rules.RuleNode{
		rules.NewRuleLit([]byte(strconv.Itoa(int(validator.BitSize)))),
	}

	if validator.Minimum != nil || validator.Maximum != nil {
		rule.Range = make([]*rules.RuleLit, 2)

		if validator.Minimum != nil {
			rule.Range[0] = rules.NewRuleLit(
				[]byte(fmt.Sprintf("%d", *validator.Minimum)),
			)
		}

		if validator.Maximum != nil {
			rule.Range[1] = rules.NewRuleLit(
				[]byte(fmt.Sprintf("%d", *validator.Maximum)),
			)
		}

		rule.ExclusiveLeft = validator.ExclusiveMinimum
		rule.ExclusiveRight = validator.ExclusiveMaximum
	}

	rule.ExclusiveLeft = validator.ExclusiveMinimum
	rule.ExclusiveRight = validator.ExclusiveMaximum

	if validator.MultipleOf != 0 {
		rule.ValueMatrix = [][]*rules.RuleLit{{
			rules.NewRuleLit([]byte("%" + fmt.Sprintf("%d", validator.MultipleOf))),
		}}
	} else if validator.Enums != nil {
		ruleValues := make([]*rules.RuleLit, 0)
		for _, v := range validator.Enums {
			ruleValues = append(ruleValues, rules.NewRuleLit([]byte(strconv.FormatInt(v, 10))))
		}
		rule.ValueMatrix = [][]*rules.RuleLit{ruleValues}
	}

	return string(rule.Bytes())
}
