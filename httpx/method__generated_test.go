package httpx

import (
	fmt "fmt"
)

func ExampleMethodGet() {
	m := MethodGet{}

	fmt.Println(m.Method())
	// Output:
	// GET
}

func ExampleMethodHead() {
	m := MethodHead{}

	fmt.Println(m.Method())
	// Output:
	// HEAD
}

func ExampleMethodPost() {
	m := MethodPost{}

	fmt.Println(m.Method())
	// Output:
	// POST
}

func ExampleMethodPut() {
	m := MethodPut{}

	fmt.Println(m.Method())
	// Output:
	// PUT
}

func ExampleMethodPatch() {
	m := MethodPatch{}

	fmt.Println(m.Method())
	// Output:
	// PATCH
}

func ExampleMethodDelete() {
	m := MethodDelete{}

	fmt.Println(m.Method())
	// Output:
	// DELETE
}

func ExampleMethodConnect() {
	m := MethodConnect{}

	fmt.Println(m.Method())
	// Output:
	// CONNECT
}

func ExampleMethodOptions() {
	m := MethodOptions{}

	fmt.Println(m.Method())
	// Output:
	// OPTIONS
}

func ExampleMethodTrace() {
	m := MethodTrace{}

	fmt.Println(m.Method())
	// Output:
	// TRACE
}
