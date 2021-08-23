package transformers

import (
	"bytes"
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/go-courier/x/ptr"
	typesutil "github.com/go-courier/x/types"
	. "github.com/onsi/gomega"
)

func TestTextTransformer(t *testing.T) {
	ct, _ := TransformerMgrDefault.NewTransformer(context.Background(), typesutil.FromRType(reflect.TypeOf("")), TransformerOption{})

	t.Run("EncodeTo", func(t *testing.T) {
		t.Run("raw value", func(t *testing.T) {
			b := bytes.NewBuffer(nil)
			h := http.Header{}
			err := ct.EncodeTo(context.Background(), WriterWithHeader(b, h), "")
			NewWithT(t).Expect(err).To(BeNil())
			NewWithT(t).Expect(h.Get("Content-Type")).To(Equal("text/plain; charset=utf-8"))
		})

		t.Run("reflect value", func(t *testing.T) {
			b := bytes.NewBuffer(nil)
			h := http.Header{}
			err := ct.EncodeTo(context.Background(), WriterWithHeader(b, h), reflect.ValueOf(1))
			NewWithT(t).Expect(err).To(BeNil())
			NewWithT(t).Expect(h.Get("Content-Type")).To(Equal("text/plain; charset=utf-8"))
		})
	})

	t.Run("DecodeFrom", func(t *testing.T) {
		t.Run("failed", func(t *testing.T) {
			b := bytes.NewBufferString("a")
			i := 0
			err := ct.DecodeFrom(context.Background(), b, &i)
			NewWithT(t).Expect(err).NotTo(BeNil())
		})

		t.Run("success", func(t *testing.T) {
			b := bytes.NewBufferString("1")
			err := ct.DecodeFrom(context.Background(), b, reflect.ValueOf(ptr.Int(0)))
			NewWithT(t).Expect(err).To(BeNil())
		})
	})
}
