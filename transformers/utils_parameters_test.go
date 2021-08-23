package transformers

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-courier/x/ptr"
	typex "github.com/go-courier/x/types"
	. "github.com/onsi/gomega"
)

type Sub struct {
	A string `name:"a" in:"query"`
}

type PtrSub struct {
	B []string `name:"b" in:"query"`
}

type P struct {
	Sub
	*PtrSub
	C *string `name:"c" in:"query"`
}

func TestParameters(t *testing.T) {
	params := make([]*Parameter, 0)

	p := P{}
	p.A = "a"
	p.PtrSub = &PtrSub{
		B: []string{"b"},
	}
	p.C = ptr.String("c")

	EachParameter(context.Background(), typex.FromRType(reflect.TypeOf(p)), func(p *Parameter) bool {
		params = append(params, p)
		return true
	})

	rv := reflect.ValueOf(&p)

	NewWithT(t).Expect(params).To(HaveLen(3))
	NewWithT(t).Expect(params[0].FieldValue(rv).Interface()).To(Equal(p.A))
	NewWithT(t).Expect(params[1].FieldValue(rv).Interface()).To(Equal(p.B))
	NewWithT(t).Expect(params[2].FieldValue(rv).Interface()).To(Equal(p.C))

}
