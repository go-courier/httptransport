package httpx

import (
	"fmt"
)

func ExampleNewAttachment_withDefaultContentType() {
	a := NewAttachment("test.txt", "")
	a.WriteString("text")

	fmt.Println(a.ContentType())
	fmt.Println(a.Meta())
	fmt.Println(a.String())
	// Output:
	// application/octet-stream
	// Content-Disposition=attachment%3B+filename%3Dtest.txt
	// text
}

func ExampleNewAttachment() {
	a := NewAttachment("test.txt", MIME_JSON)
	a.WriteString("{}")

	fmt.Println(a.ContentType())
	fmt.Println(a.Meta())
	fmt.Println(a.String())
	// Output:
	// application/json
	// Content-Disposition=attachment%3B+filename%3Dtest.txt
	// {}
}
