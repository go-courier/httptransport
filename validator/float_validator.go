package validator

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-courier/httptransport/validator/rules"
	"github.com/go-courier/x/ptr"
)

var (
	TargetFloatValue                = "float value"
	TargetDecimalDigitsOfFloatValue = "decimal digits of float value"
	TargetTotalDigitsOfFloatValue   = "total digits of float value"
)

/*
	Validator for float32 and float64

	Rules:

	ranges
		@float[min,max]
		@float[1,10] // value should large or equal than 1 and less or equal than 10
		@float(1,10] // value should large than 1 and less or equal than 10
		@float[1,10) // value should large or equal than 1

		@float[1,)  // value should large or equal than 1
		@float[,1)  // value should less than 1

	enumeration
		@float{1.1,1.2,1.3} // value should be one of these

	multiple of some float value
		@float{%multipleOf}
		@float{%2.2} // value should be multiple of 2.2

	max digits and decimal digits.
	when defined, all values in rule should be under range of them.
		@float<MAX_DIGITS,DECIMAL_DIGITS>
		@float<5,2> // will checkout these values invalid: 1.111 (decimal digits too many), 12345.6 (digits too many)

	composes
		@float<MAX_DIGITS,DECIMAL_DIGITS>[min,max]

	aliases:
		@float32 = @float<7>
		@float64 = @float<15>
*/
type FloatValidator struct {
	MaxDigits     uint
	DecimalDigits *uint

	Minimum          *float64
	Maximum          *float64
	ExclusiveMaximum bool
	ExclusiveMinimum bool

	MultipleOf float64
	Enums      []float64
}

func init() {
	ValidatorMgrDefault.Register(&FloatValidator{})
}

func (validator *FloatValidator) SetDefaults() {
	if validator != nil {
		if validator.MaxDigits == 0 {
			validator.MaxDigits = 7
		}
		if validator.DecimalDigits == nil {
			validator.DecimalDigits = ptr.Uint(2)
		}
	}
}

func (FloatValidator) Names() []string {
	return []string{"float", "double", "float32", "float64"}
}

func isFloatType(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

func (validator *FloatValidator) Validate(v interface{}) error {
	var val float64

	switch x := v.(type) {
	case reflect.Value:
		if !isFloatType(x.Type()) {
			return NewUnsupportedTypeError(x.Type().String(), validator.String())
		}
		val = x.Float()
	case float64:
		val = x
	case float32:
		val = float64(x)
	default:
		return NewUnsupportedTypeError(fmt.Sprintf("%T", v), validator.String())
	}

	decimalDigits := *validator.DecimalDigits

	m, d := lengthOfDigits(val)
	if m > validator.MaxDigits {
		return &OutOfRangeError{
			Target:  TargetTotalDigitsOfFloatValue,
			Current: m,
			Maximum: validator.MaxDigits,
		}
	}

	if d > decimalDigits {
		return &OutOfRangeError{
			Target:  TargetDecimalDigitsOfFloatValue,
			Current: d,
			Maximum: decimalDigits,
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
				Target:  TargetFloatValue,
				Current: v,
				Enums:   enums,
			}
		}

		return nil
	}

	if validator.Minimum != nil {
		mininum := *validator.Minimum
		if (validator.ExclusiveMinimum && val == mininum) || val < mininum {
			return &OutOfRangeError{
				Target:           TargetFloatValue,
				Current:          val,
				Minimum:          mininum,
				ExclusiveMinimum: validator.ExclusiveMinimum,
			}
		}
	}

	if validator.Maximum != nil {
		maxinum := *validator.Maximum
		if (validator.ExclusiveMaximum && val == maxinum) || val > maxinum {
			return &OutOfRangeError{
				Target:           TargetFloatValue,
				Current:          val,
				Maximum:          maxinum,
				ExclusiveMaximum: validator.ExclusiveMaximum,
			}
		}
	}

	if validator.MultipleOf != 0 {
		if !multipleOf(val, validator.MultipleOf, decimalDigits) {
			return &MultipleOfError{
				Target:     TargetFloatValue,
				Current:    val,
				MultipleOf: validator.MultipleOf,
			}
		}
	}

	return nil
}

func lengthOfDigits(f float64) (uint, uint) {
	s := strconv.FormatFloat(f, 'e', -1, 64)
	var n, d int

	parts := strings.Split(s, "e")
	nd := strings.Split(parts[0], ".")
	i := nd[0]
	n = len(i)

	if len(nd) == 2 {
		d = len(nd[1])
	}

	if len(parts) == 2 {
		switch parts[1][0] {
		case '+':
			v, _ := strconv.ParseUint(parts[1][1:], 10, 64)
			n = n + int(v)
			d = d - int(v)
			if d < 0 {
				d = 0
			}
		case '-':
			v, _ := strconv.ParseUint(parts[1][1:], 10, 64)
			n = n - int(v)
			if n <= 0 {
				n = 1
			}
			d = d + int(v)
		}
	}

	if math.Abs(f) < 1.0 {
		n = 0
	}

	return uint(n + d), uint(d)
}

func multipleOf(v float64, div float64, decimalDigits uint) bool {
	val := fixDecimal(v/div, int(decimalDigits))
	return val == math.Trunc(val)
}

func fixDecimal(f float64, n int) float64 {
	res, _ := strconv.ParseFloat(strconv.FormatFloat(f, 'g', n, 64), 64)
	return res
}

func (FloatValidator) New(ctx context.Context, rule *Rule) (Validator, error) {
	validator := &FloatValidator{}

	switch rule.Name {
	case "float", "float32":
		validator.MaxDigits = 7
	case "double", "float64":
		validator.MaxDigits = 15
	}

	if rule.Params != nil {
		if len(rule.Params) > 2 {
			return nil, fmt.Errorf("float should only 1 or 2 parameter, but got %d", len(rule.Params))
		}

		maxDigitsBytes := rule.Params[0].Bytes()
		if len(maxDigitsBytes) > 0 {
			maxDigits, err := strconv.ParseUint(string(maxDigitsBytes), 10, 4)
			if err != nil {
				return nil, rules.NewSyntaxError("decimal digits should be a uint value which less than 16, but got `%s`", maxDigitsBytes)
			}
			validator.MaxDigits = uint(maxDigits)
		}

		if len(rule.Params) > 1 {
			decimalDigitsBytes := rule.Params[1].Bytes()
			if len(decimalDigitsBytes) > 0 {
				decimalDigits, err := strconv.ParseUint(string(decimalDigitsBytes), 10, 4)
				if err != nil || uint(decimalDigits) >= validator.MaxDigits {
					return nil, rules.NewSyntaxError("decimal digits should be a uint value which less than %d, but got `%s`", validator.MaxDigits, decimalDigitsBytes)
				}
				validator.DecimalDigits = ptr.Uint(uint(decimalDigits))
			}

		}

	}

	validator.SetDefaults()

	validator.ExclusiveMinimum = rule.ExclusiveLeft
	validator.ExclusiveMaximum = rule.ExclusiveRight

	if rule.Range != nil {
		min, max, err := floatRange(
			"float",
			validator.MaxDigits, validator.DecimalDigits,
			rule.Range...,
		)
		if err != nil {
			return nil, err
		}

		validator.Minimum = min
		validator.Maximum = max

		validator.ExclusiveMinimum = rule.ExclusiveLeft
		validator.ExclusiveMaximum = rule.ExclusiveRight
	}

	ruleValues := rule.ComputedValues()

	if ruleValues != nil {
		if len(ruleValues) == 1 {
			mayBeMultipleOf := ruleValues[0].Bytes()
			if mayBeMultipleOf[0] == '%' {
				v := mayBeMultipleOf[1:]
				multipleOf, err := parseFloat(v, validator.MaxDigits, validator.DecimalDigits)
				if err != nil {
					return nil, rules.NewSyntaxError("multipleOf should be a valid float<%d> value, but got `%s`", validator.MaxDigits, v)
				}
				validator.MultipleOf = multipleOf
			}
		}

		if validator.MultipleOf == 0 {
			for _, v := range ruleValues {
				b := v.Bytes()
				enumValue, err := parseFloat(b, validator.MaxDigits, validator.DecimalDigits)
				if err != nil {
					return nil, rules.NewSyntaxError("enum should be a valid float<%d> value, but got `%s`", validator.MaxDigits, b)
				}
				validator.Enums = append(validator.Enums, enumValue)
			}
		}
	}

	return validator, validator.TypeCheck(rule)
}

func (validator *FloatValidator) TypeCheck(rule *Rule) error {
	switch rule.Type.Kind() {
	case reflect.Float32:
		if validator.MaxDigits > 7 {
			return fmt.Errorf("max digits too large for type %s", rule)
		}
		return nil
	case reflect.Float64:
		return nil
	}
	return NewUnsupportedTypeError(rule.String(), validator.String())
}

func floatRange(typ string, maxDigits uint, decimalDigits *uint, ranges ...*rules.RuleLit) (*float64, *float64, error) {
	fullType := fmt.Sprintf("%s<%d>", typ, maxDigits)
	if decimalDigits != nil {
		fullType = fmt.Sprintf("%s<%d,%d>", typ, maxDigits, *decimalDigits)
	}

	parseMaybeFloat := func(b []byte) (*float64, error) {
		if len(b) == 0 {
			return nil, nil
		}
		n, err := parseFloat(b, maxDigits, decimalDigits)
		if err != nil {
			return nil, fmt.Errorf("%s value is not correct: %s", fullType, err)
		}
		return &n, nil
	}

	switch len(ranges) {
	case 2:
		min, err := parseMaybeFloat(ranges[0].Bytes())
		if err != nil {
			return nil, nil, fmt.Errorf("min %s", err)
		}
		max, err := parseMaybeFloat(ranges[1].Bytes())
		if err != nil {
			return nil, nil, fmt.Errorf("max %s", err)
		}
		if min != nil && max != nil && *max < *min {
			return nil, nil, fmt.Errorf("max %s value must be equal or large than min value %v, current %v", fullType, *min, *max)
		}
		return min, max, nil
	case 1:
		min, err := parseMaybeFloat(ranges[0].Bytes())
		if err != nil {
			return nil, nil, fmt.Errorf("min %s", err)
		}
		return min, min, nil
	}
	return nil, nil, nil
}

func parseFloat(b []byte, maxDigits uint, maybeDecimalDigits *uint) (float64, error) {
	f, err := strconv.ParseFloat(string(b), 64)
	if err != nil {
		return 0, err
	}

	if b[0] == '-' {
		b = b[1:]
	}

	if b[0] == '.' {
		b = append([]byte("0"), b...)
	}

	i := bytes.IndexRune(b, '.')

	decimalDigits := maxDigits - 1
	if maybeDecimalDigits != nil && *maybeDecimalDigits < maxDigits {
		decimalDigits = *maybeDecimalDigits
	}

	m := uint(len(b) - 1)
	if uint(len(b)-1) > maxDigits {
		return 0, fmt.Errorf("max digits should be less than %d, but got %d", decimalDigits, m)
	}

	if i != -1 {
		d := uint(len(b) - i - 1)
		if d > decimalDigits {
			return 0, fmt.Errorf("decimal digits should be less than %d, but got %d", decimalDigits, d)
		}
	}
	return f, nil
}

func (validator *FloatValidator) String() string {
	validator.SetDefaults()

	rule := rules.NewRule(validator.Names()[0])

	decimalDigits := *validator.DecimalDigits

	rule.Params = []rules.RuleNode{
		rules.NewRuleLit([]byte(strconv.Itoa(int(validator.MaxDigits)))),
		rules.NewRuleLit([]byte(strconv.Itoa(int(decimalDigits)))),
	}

	if validator.Minimum != nil || validator.Maximum != nil {
		rule.Range = make([]*rules.RuleLit, 2)

		if validator.Minimum != nil {
			rule.Range[0] = rules.NewRuleLit(
				[]byte(fmt.Sprintf("%."+strconv.Itoa(int(decimalDigits))+"f", *validator.Minimum)),
			)
		}

		if validator.Maximum != nil {
			rule.Range[1] = rules.NewRuleLit(
				[]byte(fmt.Sprintf("%."+strconv.Itoa(int(decimalDigits))+"f", *validator.Maximum)),
			)
		}

		rule.ExclusiveLeft = validator.ExclusiveMinimum
		rule.ExclusiveRight = validator.ExclusiveMaximum
	}

	if validator.MultipleOf != 0 {
		rule.ValueMatrix = [][]*rules.RuleLit{
			{rules.NewRuleLit([]byte("%" + fmt.Sprintf("%."+strconv.Itoa(int(decimalDigits))+"f", validator.MultipleOf)))},
		}
	} else if validator.Enums != nil {
		ruleValues := make([]*rules.RuleLit, 0)
		for _, v := range validator.Enums {
			ruleValues = append(ruleValues, rules.NewRuleLit([]byte(strconv.FormatFloat(v, 'f', -1, 64))))
		}
		rule.ValueMatrix = [][]*rules.RuleLit{ruleValues}
	}

	return string(rule.Bytes())
}
