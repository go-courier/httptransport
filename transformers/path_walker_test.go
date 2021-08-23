package transformers

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestPathWalker(t *testing.T) {
	pw := &PathWalker{}
	pw.Enter("key")
	NewWithT(t).Expect(pw.Paths()).To(Equal([]interface{}{"key"}))
	NewWithT(t).Expect(pw.String()).To(Equal("key"))

	pw.Enter(1)
	NewWithT(t).Expect(pw.Paths()).To(Equal([]interface{}{"key", 1}))
	NewWithT(t).Expect(pw.String()).To(Equal("key[1]"))

	pw.Enter("prop")
	NewWithT(t).Expect(pw.Paths()).To(Equal([]interface{}{"key", 1, "prop"}))
	NewWithT(t).Expect(pw.String()).To(Equal("key[1].prop"))

	pw.Exit()
	pw.Exit()
	NewWithT(t).Expect(pw.Paths()).To(Equal([]interface{}{"key"}))
	NewWithT(t).Expect(pw.String()).To(Equal("key"))
}
