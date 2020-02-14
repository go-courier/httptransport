package httpx

import (
	fmt "fmt"
	net_url "net/url"
)

func ExampleStatusMultipleChoices() {
	m := RedirectWithStatusMultipleChoices(&(net_url.URL{
		Path: "/test",
	}))

	fmt.Println(m.StatusCode())
	fmt.Println(m.Location())
	// Output:
	// 300
	// /test
}

func ExampleStatusMovedPermanently() {
	m := RedirectWithStatusMovedPermanently(&(net_url.URL{
		Path: "/test",
	}))

	fmt.Println(m.StatusCode())
	fmt.Println(m.Location())
	// Output:
	// 301
	// /test
}

func ExampleStatusFound() {
	m := RedirectWithStatusFound(&(net_url.URL{
		Path: "/test",
	}))

	fmt.Println(m.StatusCode())
	fmt.Println(m.Location())
	// Output:
	// 302
	// /test
}

func ExampleStatusSeeOther() {
	m := RedirectWithStatusSeeOther(&(net_url.URL{
		Path: "/test",
	}))

	fmt.Println(m.StatusCode())
	fmt.Println(m.Location())
	// Output:
	// 303
	// /test
}

func ExampleStatusNotModified() {
	m := RedirectWithStatusNotModified(&(net_url.URL{
		Path: "/test",
	}))

	fmt.Println(m.StatusCode())
	fmt.Println(m.Location())
	// Output:
	// 304
	// /test
}

func ExampleStatusUseProxy() {
	m := RedirectWithStatusUseProxy(&(net_url.URL{
		Path: "/test",
	}))

	fmt.Println(m.StatusCode())
	fmt.Println(m.Location())
	// Output:
	// 305
	// /test
}

func ExampleStatusTemporaryRedirect() {
	m := RedirectWithStatusTemporaryRedirect(&(net_url.URL{
		Path: "/test",
	}))

	fmt.Println(m.StatusCode())
	fmt.Println(m.Location())
	// Output:
	// 307
	// /test
}

func ExampleStatusPermanentRedirect() {
	m := RedirectWithStatusPermanentRedirect(&(net_url.URL{
		Path: "/test",
	}))

	fmt.Println(m.StatusCode())
	fmt.Println(m.Location())
	// Output:
	// 308
	// /test
}
