package validator

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"testing"

	typesutil "github.com/go-courier/x/types"

	. "github.com/onsi/gomega"
)

type SomeTextMarshaler struct {
}

func (*SomeTextMarshaler) MarshalText() ([]byte, error) {
	return []byte("SomeTextMarshaler"), nil
}

func TestStructValidator_New(t *testing.T) {
	type Named string

	type SubPtrStruct struct {
		PtrInt   *int     `validate:"@int[1,]"`
		PtrFloat *float32 `validate:"@float[1,]"`
		PtrUint  *uint    `validate:"@uint[1,]"`
	}

	type SubStruct struct {
		Int   int     `validate:"@int[1,]"`
		Float float32 `validate:"@float[1,]"`
		Uint  uint    `validate:"@uint[1,]"`
	}

	type SomeStruct struct {
		String       string             `json:"String" validate:"@string[1,]"`
		Named        Named              `json:"Named,omitempty" validate:"@string[2,]"`
		PtrString    *string            `json:",omitempty" validate:"@string[3,]"`
		SomeStringer *SomeTextMarshaler `validate:"@string[4,]"`
		Slice        []string           `validate:"@slice<@string[1,]>"`
		SubStruct
		*SubPtrStruct
	}

	validator := NewStructValidator("json")

	_, err := validator.New(ContextWithValidatorMgr(context.Background(), ValidatorMgrDefault), &Rule{
		Type: typesutil.FromRType(reflect.TypeOf(&SomeStruct{}).Elem()),
	})
	NewWithT(t).Expect(err).To(BeNil())
}

func TestStructValidator_NewFailed(t *testing.T) {
	type Named string

	type Struct struct {
		Int    int  `validate:"@int[1,"`
		PtrInt *int `validate:"@uint[2,"`
	}

	type SubStruct struct {
		Float    float32  `validate:"@string"`
		PtrFloat *float32 `validate:"@unknown"`
	}

	type SomeStruct struct {
		String string   `validate:"@string[1,"`
		Named  Named    `validate:"@int"`
		Slice  []string `validate:"@slice<@int>"`
		SubStruct
		Struct Struct
	}

	validator := NewStructValidator("json")

	_, err := validator.New(ContextWithValidatorMgr(context.Background(), ValidatorMgrDefault), &Rule{
		Type: typesutil.FromRType(reflect.TypeOf(&SomeStruct{}).Elem()),
	})
	NewWithT(t).Expect(err).NotTo(BeNil())
	t.Log(err)

	{
		_, err := validator.New(ContextWithValidatorMgr(context.Background(), ValidatorMgrDefault), &Rule{
			Type: typesutil.FromRType(reflect.TypeOf("")),
		})
		NewWithT(t).Expect(err).NotTo(BeNil())
	}
}

func ExampleNewStructValidator() {
	type Named string

	type SubPtrStruct struct {
		PtrInt   *int     `validate:"@int[1,]"`
		PtrFloat *float32 `validate:"@float[1,]"`
		PtrUint  *uint    `validate:"@uint[1,]"`
	}

	type SubStruct struct {
		Int   int     `json:"int" validate:"@int[1,]"`
		Float float32 `json:"float" validate:"@float[1,]"`
		Uint  uint    `json:"uint" validate:"@uint[1,]"`
	}

	type SomeStruct struct {
		JustRequired string
		CanEmpty     *string              `validate:"@string[0,]?"`
		String       string               `validate:"@string[1,]"`
		Named        Named                `validate:"@string[2,]"`
		PtrString    *string              `validate:"@string[3,]" default:"123"`
		SomeStringer *SomeTextMarshaler   `validate:"@string[20,]"`
		Slice        []string             `validate:"@slice<@string[1,]>"`
		SliceStruct  []SubStruct          `validate:"@slice"`
		Map          map[string]string    `validate:"@map<@string[2,],@string[1,]>"`
		MapStruct    map[string]SubStruct `validate:"@map<@string[2,],>"`
		Struct       SubStruct
		SubStruct
		*SubPtrStruct
	}

	validator := NewStructValidator("json")

	ctx := ContextWithValidatorMgr(context.Background(), ValidatorMgrDefault)

	structValidator, err := validator.New(ContextWithValidatorMgr(ctx, ValidatorMgrDefault), &Rule{
		Type: typesutil.FromRType(reflect.TypeOf(&SomeStruct{}).Elem()),
	})
	if err != nil {
		return
	}

	s := SomeStruct{
		Slice: []string{"", ""},
		SliceStruct: []SubStruct{
			{Int: 0},
		},
		Map: map[string]string{
			"1":  "",
			"11": "",
			"12": "",
		},
		MapStruct: map[string]SubStruct{
			"222": SubStruct{},
		},
	}

	errForValidate := structValidator.Validate(s)

	errSet := map[string]string{}
	errKeyPaths := make([]string, 0)

	errForValidate.(*ErrorSet).Flatten().Each(func(fieldErr *FieldError) {
		errSet[fieldErr.Path.String()] = strconv.Quote(fieldErr.Error.Error())
		errKeyPaths = append(errKeyPaths, fieldErr.Path.String())
	})

	sort.Strings(errKeyPaths)

	for i := range errKeyPaths {
		k := errKeyPaths[i]
		fmt.Println(k, errSet[k])
	}

	// Output:
	// JustRequired "missing required field"
	// Map.1 "missing required field"
	// Map.1/key "string length should be larger than 2, but got invalid value 1"
	// Map.11 "missing required field"
	// Map.12 "missing required field"
	// MapStruct.222.float "missing required field"
	// MapStruct.222.int "missing required field"
	// MapStruct.222.uint "missing required field"
	// Named "missing required field"
	// PtrFloat "missing required field"
	// PtrInt "missing required field"
	// PtrString "missing required field"
	// PtrUint "missing required field"
	// SliceStruct[0].float "missing required field"
	// SliceStruct[0].int "missing required field"
	// SliceStruct[0].uint "missing required field"
	// Slice[0] "missing required field"
	// Slice[1] "missing required field"
	// SomeStringer "missing required field"
	// String "missing required field"
	// Struct.float "missing required field"
	// Struct.int "missing required field"
	// Struct.uint "missing required field"
	// float "missing required field"
	// int "missing required field"
	// uint "missing required field"
}
