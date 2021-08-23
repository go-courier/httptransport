package transformers

import (
	"io"
	"strings"

	"github.com/pkg/errors"
)

func NewStringReaders(values []string) *StringReaders {
	bs := make([]io.Reader, len(values))
	for i := range values {
		bs[i] = strings.NewReader(values[i])
	}

	return &StringReaders{
		readers: bs,
	}
}

type StringReaders struct {
	idx     int
	readers []io.Reader
}

func (v *StringReaders) Len() int {
	return len(v.readers)
}

func (v *StringReaders) NextReader() io.Reader {
	r := v.readers[v.idx]
	v.idx++
	return r
}

func (v *StringReaders) Read(p []byte) (n int, err error) {
	if v.idx < len(v.readers) {
		return v.readers[v.idx].Read(p)
	}
	return -1, errors.Errorf("bounds out of range, %d", v.idx)
}

func NewStringBuilders() *StringBuilders {
	return &StringBuilders{}
}

type StringBuilders struct {
	idx     int
	buffers []*strings.Builder
}

func (v *StringBuilders) SetN(n int) {
	v.buffers = make([]*strings.Builder, n)
	v.idx = 0
	for i := range v.buffers {
		v.buffers[i] = &strings.Builder{}
	}
}
func (v *StringBuilders) NextWriter() io.Writer {
	if v.idx == 0 && len(v.buffers) == 0 {
		v.SetN(1)
	}
	r := v.buffers[v.idx]
	v.idx++
	return r
}

func (v *StringBuilders) Write(p []byte) (n int, err error) {
	if v.idx == 0 && len(v.buffers) == 0 {
		v.SetN(1)
	}
	return v.buffers[v.idx].Write(p)
}

func (v *StringBuilders) StringSlice() []string {
	values := make([]string, len(v.buffers))
	for i, b := range v.buffers {
		values[i] = b.String()
	}
	return values
}
