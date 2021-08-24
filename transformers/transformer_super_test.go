package transformers

import (
	"context"
	"testing"
)

func BenchmarkTransformerSuper(b *testing.B) {
	ts := NewTransformerSuper(&TransformerPlainText{}, &CommonTransformOption{Omitempty: true})

	b.Run("DecodeFrom by super", func(b *testing.B) {
		ret := ""
		for i := 0; i < b.N; i++ {
			_ = ts.DecodeFrom(context.Background(), NewStringReader("111"), &ret)
		}
		b.Log(ret)
	})

	b.Run("DecodeFrom direct", func(b *testing.B) {
		pt := TransformerPlainText{}

		ret := ""

		for i := 0; i < b.N; i++ {
			_ = pt.DecodeFrom(context.Background(), NewStringReader("111"), &ret)
		}

		b.Log(ret)
	})
}
