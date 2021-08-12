package transformers

import (
	"bytes"
	"context"
	"net/http"
	"reflect"
	"testing"

	typesutil "github.com/go-courier/x/types"
	. "github.com/onsi/gomega"
)

func TestXMLTransformer(t *testing.T) {
	type TestData struct {
		Data struct {
			Bool        bool
			FirstName   string `xml:"name>first"`
			StructSlice []struct {
				Name string
			}
			StringSlice     []string
			StringAttrSlice []string `xml:"StringAttrSlice,attr"`
			NestedSlice     []struct {
				Names []string
			}
		}
	}

	data := TestData{}
	data.Data.FirstName = "test"
	data.Data.StringSlice = []string{"1", "2", "3"}
	data.Data.StringAttrSlice = []string{"1", "2", "3"}

	ct, _ := TransformerMgrDefault.NewTransformer(context.Background(), typesutil.FromRType(reflect.TypeOf(data)), TransformerOption{
		MIME: "xml",
	})

	t.Run("EncodeTo", func(t *testing.T) {
		t.Run("raw value", func(t *testing.T) {
			b := bytes.NewBuffer(nil)
			h := http.Header{}
			err := ct.EncodeTo(context.Background(), WriterWithHeader(b, h), data)
			NewWithT(t).Expect(err).To(BeNil())
			NewWithT(t).Expect(h.Get("Content-Type")).To(Equal("application/xml; charset=utf-8"))
		})

		t.Run("reflect value", func(t *testing.T) {
			b := bytes.NewBuffer(nil)
			h := http.Header{}
			err := ct.EncodeTo(context.Background(), WriterWithHeader(b, h), reflect.ValueOf(data))
			NewWithT(t).Expect(err).To(BeNil())
			NewWithT(t).Expect(h.Get("Content-Type")).To(Equal("application/xml; charset=utf-8"))
		})
	})

	t.Run("DecodeFrom", func(t *testing.T) {
		t.Run("failed", func(t *testing.T) {
			b := bytes.NewBufferString("<")
			err := ct.DecodeFrom(context.Background(), b, &data)
			NewWithT(t).Expect(err).NotTo(BeNil())
		})

		t.Run("success", func(t *testing.T) {
			b := bytes.NewBufferString("<TestData></TestData>")
			err := ct.DecodeFrom(context.Background(), b, reflect.ValueOf(&data))
			NewWithT(t).Expect(err).To(BeNil())
		})

		t.Run("failed with wrong type", func(t *testing.T) {
			b := bytes.NewBufferString("<TestData><Data><Bool>bool</Bool></Data></TestData>")
			err := ct.DecodeFrom(context.Background(), b, &data)
			NewWithT(t).Expect(err).NotTo(BeNil())
		})
	})
}
