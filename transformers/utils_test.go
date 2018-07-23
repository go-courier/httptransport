package transformers

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type SubA struct {
	KeyA string `name:"keyA"`
}

type SubB struct {
	KeyB string `name:"keyB"`
}

type SomeStruct struct {
	NameA string
	NameB string
	SubA
	SubB   `name:"sub"`
	Ignore string `name:"-"`
}

func TestNamedStructFieldValueRange(t *testing.T) {
	v := &SomeStruct{}
	rv := reflect.ValueOf(v).Elem()

	names := make([]string, 0)

	NamedStructFieldValueRange(rv, func(fieldValue reflect.Value, field *reflect.StructField) {
		if field.Type.Kind() == reflect.String {
			fieldValue.SetString(field.Name)
		}
		names = append(names, field.Name)
	})

	require.Equal(t, []string{
		"NameA",
		"NameB",
		"KeyA",
		"SubB",
	}, names)

	require.Equal(t, &SomeStruct{
		NameA: "NameA",
		NameB: "NameB",
		SubA: SubA{
			KeyA: "KeyA",
		},
	}, v)
}
