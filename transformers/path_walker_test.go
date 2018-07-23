package transformers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPathWalker(t *testing.T) {
	tt := require.New(t)

	pw := &PathWalker{}
	pw.Enter("key")
	tt.Equal([]interface{}{"key"}, pw.Paths())
	tt.Equal("key", pw.String())

	pw.Enter(1)
	tt.Equal([]interface{}{"key", 1}, pw.Paths())
	tt.Equal("key[1]", pw.String())

	pw.Enter("prop")
	tt.Equal([]interface{}{"key", 1, "prop"}, pw.Paths())
	tt.Equal("key[1].prop", pw.String())

	pw.Exit()
	pw.Exit()
	tt.Equal([]interface{}{"key"}, pw.Paths())
	tt.Equal("key", pw.String())
}
