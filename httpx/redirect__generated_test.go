package httpx

import (
	"fmt"
	net_url "net/url"
)

func ExampleStatusMultipleChoices() {
	m := RedirectWithStatusMultipleChoices(&(net_url.URL{
		Path: "/test",
	}))

	fmt.Println(m.StatusCode())
	fmt.Println(m.Location())
	fmt.Println(m.Error())
	// Output:
	// 300
	// /test
	// Location: /test
}

func ExampleStatusMovedPermanently() {
	m := RedirectWithStatusMovedPermanently(&(net_url.URL{
		Path: "/test",
	}))

	fmt.Println(m.StatusCode())
	fmt.Println(m.Location())
	fmt.Println(m.Error())
	// Output:
	// 301
	// /test
	// Location: /test
}

func ExampleStatusFound() {
	m := RedirectWithStatusFound(&(net_url.URL{
		Path: "/test",
	}))

	fmt.Println(m.StatusCode())
	fmt.Println(m.Location())
	fmt.Println(m.Error())
	// Output:
	// 302
	// /test
	// Location: /test
}

func ExampleStatusSeeOther() {
	m := RedirectWithStatusSeeOther(&(net_url.URL{
		Path: "/test",
	}))

	fmt.Println(m.StatusCode())
	fmt.Println(m.Location())
	fmt.Println(m.Error())
	// Output:
	// 303
	// /test
	// Location: /test
}

func ExampleStatusNotModified() {
	m := RedirectWithStatusNotModified(&(net_url.URL{
		Path: "/test",
	}))

	fmt.Println(m.StatusCode())
	fmt.Println(m.Location())
	fmt.Println(m.Error())
	// Output:
	// 304
	// /test
	// Location: /test
}

func ExampleStatusUseProxy() {
	m := RedirectWithStatusUseProxy(&(net_url.URL{
		Path: "/test",
	}))

	fmt.Println(m.StatusCode())
	fmt.Println(m.Location())
	fmt.Println(m.Error())
	// Output:
	// 305
	// /test
	// Location: /test
}

func ExampleStatusTemporaryRedirect() {
	m := RedirectWithStatusTemporaryRedirect(&(net_url.URL{
		Path: "/test",
	}))

	fmt.Println(m.StatusCode())
	fmt.Println(m.Location())
	fmt.Println(m.Error())
	// Output:
	// 307
	// /test
	// Location: /test
}

func ExampleStatusPermanentRedirect() {
	m := RedirectWithStatusPermanentRedirect(&(net_url.URL{
		Path: "/test",
	}))

	fmt.Println(m.StatusCode())
	fmt.Println(m.Location())
	fmt.Println(m.Error())
	// Output:
	// 308
	// /test
	// Location: /test
}
