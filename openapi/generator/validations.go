package generator

import (
	"context"
	"go/types"

	"github.com/go-courier/oas"
	"github.com/go-courier/ptr"
	"github.com/go-courier/reflectx/typesutil"
	"github.com/go-courier/validator"
)

func BindSchemaValidationByValidateBytes(s *oas.Schema, typ types.Type, validateBytes []byte) error {
	ttype := typesutil.FromTType(typ)

	fieldValidator, err := validator.ValidatorMgrDefault.Compile(context.Background(), validateBytes, ttype, func(rule *validator.Rule) {
		rule.DefaultValue = nil
	})
	if err != nil {
		return err
	}

	if fieldValidator != nil {
		BindSchemaValidationByValidator(s, fieldValidator)
	}

	return nil
}

func BindSchemaValidationByValidator(s *oas.Schema, v validator.Validator) {
	if validatorLoader, ok := v.(*validator.ValidatorLoader); ok {
		v = validatorLoader.Validator
	}
	if s == nil {
		*s = oas.Schema{}
	}
	switch vt := v.(type) {
	case *validator.UintValidator:
		if len(vt.Enums) > 0 {
			for v := range vt.Enums {
				s.Enum = append(s.Enum, v)
			}
			return
		}

		s.Minimum = ptr.Float64(float64(vt.Minimum))
		s.Maximum = ptr.Float64(float64(vt.Maximum))
		s.ExclusiveMinimum = vt.ExclusiveMinimum
		s.ExclusiveMaximum = vt.ExclusiveMaximum
		if vt.MultipleOf > 0 {
			s.MultipleOf = ptr.Float64(float64(vt.MultipleOf))
		}
	case *validator.IntValidator:
		if len(vt.Enums) > 0 {
			for v := range vt.Enums {
				s.Enum = append(s.Enum, v)
			}
			return
		}

		if vt.Minimum != nil {
			s.Minimum = ptr.Float64(float64(*vt.Minimum))
		}
		if vt.Maximum != nil {
			s.Maximum = ptr.Float64(float64(*vt.Maximum))
		}
		s.ExclusiveMinimum = vt.ExclusiveMinimum
		s.ExclusiveMaximum = vt.ExclusiveMaximum

		if vt.MultipleOf > 0 {
			s.MultipleOf = ptr.Float64(float64(vt.MultipleOf))
		}
	case *validator.FloatValidator:
		if len(vt.Enums) > 0 {
			for v := range vt.Enums {
				s.Enum = append(s.Enum, v)
			}
			return
		}

		if vt.Minimum != nil {
			s.Minimum = ptr.Float64(float64(*vt.Minimum))
		}
		if vt.Maximum != nil {
			s.Maximum = ptr.Float64(float64(*vt.Maximum))
		}
		s.ExclusiveMinimum = vt.ExclusiveMinimum
		s.ExclusiveMaximum = vt.ExclusiveMaximum

		if vt.MultipleOf > 0 {
			s.MultipleOf = ptr.Float64(float64(vt.MultipleOf))
		}
	case *validator.StrfmtValidator:
		s.Type = oas.TypeString // force to type string for TextMarshaler
		s.Format = vt.Names()[0]
	case *validator.StringValidator:
		s.Type = oas.TypeString // force to type string for TextMarshaler

		if len(vt.Enums) > 0 {
			for v := range vt.Enums {
				s.Enum = append(s.Enum, v)
			}
			return
		}

		s.MinLength = ptr.Uint64(vt.MinLength)
		if vt.MaxLength != nil {
			s.MaxLength = ptr.Uint64(*vt.MaxLength)
		}
		if vt.Pattern != nil {
			s.Pattern = vt.Pattern.String()
		}
	case *validator.SliceValidator:
		s.MinItems = ptr.Uint64(vt.MinItems)
		if vt.MaxItems != nil {
			s.MaxItems = ptr.Uint64(*vt.MaxItems)
		}

		if vt.ElemValidator != nil {
			BindSchemaValidationByValidator(s.Items, vt.ElemValidator)
		}
	case *validator.MapValidator:
		s.MinProperties = ptr.Uint64(vt.MinProperties)
		if vt.MaxProperties != nil {
			s.MaxProperties = ptr.Uint64(*vt.MaxProperties)
		}
		if vt.ElemValidator != nil {
			BindSchemaValidationByValidator(s.AdditionalProperties.Schema, vt.ElemValidator)
		}
	}
}
