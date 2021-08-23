package rules

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestSlashUnslash(t *testing.T) {
	cases := [][]string{
		{`/\w+\/test/`, `\w+/test`},
		{`/a/`, `a`},
		{`/abc/`, `abc`},
		{`/☺/`, `☺`},
		{`/\xFF/`, `\xFF`},
		{`/\377/`, `\377`},
		{`/\u1234/`, `\u1234`},
		{`/\U00010111/`, `\U00010111`},
		{`/\U0001011111/`, `\U0001011111`},
		{`/\a\b\f\n\r\t\v\\\"/`, `\a\b\f\n\r\t\v\\\"`},
		{`/\//`, `/`},
	}

	for i := range cases {
		c := cases[i]

		t.Run("unslash:"+c[0], func(t *testing.T) {
			r, err := Unslash([]byte(c[0]))
			NewWithT(t).Expect(err).To(BeNil())
			NewWithT(t).Expect(string(r)).To(Equal(c[1]))
		})

		t.Run("slash:"+c[1], func(t *testing.T) {
			NewWithT(t).Expect(string(Slash([]byte(c[1])))).To(Equal(c[0]))
		})
	}

	casesForFailed := [][]string{
		{`/`, ``},
		{`/adfadf`, ``},
	}

	for i := range casesForFailed {
		c := casesForFailed[i]

		t.Run("unslash:"+c[0], func(t *testing.T) {
			_, err := Unslash([]byte(c[0]))
			NewWithT(t).Expect(err).NotTo(BeNil())
		})
	}

}
