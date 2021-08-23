package validator

import (
	"fmt"
)

func ExampleErrorSet() {
	subErrSet := NewErrorSet("")
	subErrSet.AddErr(fmt.Errorf("err"), "PropA")
	subErrSet.AddErr(fmt.Errorf("err"), "PropB")

	errSet := NewErrorSet("")
	errSet.AddErr(fmt.Errorf("err"), "Key")
	errSet.AddErr(subErrSet.Err(), "Key", 1)
	errSet.AddErr(NewErrorSet("").Err(), "Key", 1)

	fmt.Println(errSet.Flatten().Len())
	fmt.Println(errSet)
	// Output:
	// 3
	// Key err
	// Key[1].PropA err
	// Key[1].PropB err
}

func ExampleMissingRequired() {
	fmt.Println(MissingRequired{})
	// Output:
	// missing required field
}

func ExampleNotMatchError() {
	fmt.Println(&NotMatchError{
		Target:  "number",
		Current: "1",
		Pattern: `/\d+/`,
	})
	// Output:
	// number /\d+/ not match 1
}

func ExampleMultipleOfError() {
	fmt.Println(&MultipleOfError{
		Target:     "int value",
		Current:    "11",
		MultipleOf: 2,
	})
	// Output:
	// int value should be multiple of 2, but got invalid value 11
}

func ExampleNotInEnumError() {
	fmt.Println(&NotInEnumError{
		Target:  "int value",
		Current: "11",
		Enums: []interface{}{
			"1", "2", "3",
		},
	})
	// Output:
	// int value should be one of 1, 2, 3, but got invalid value 11
}

func ExampleOutOfRangeError() {
	fmt.Println(&OutOfRangeError{
		Target:           "int value",
		Minimum:          "1",
		Maximum:          "10",
		Current:          "11",
		ExclusiveMinimum: true,
		ExclusiveMaximum: true,
	})
	// Output:
	// int value should be larger or equal than 1 and less or equal than 10, but got invalid value 11
}
