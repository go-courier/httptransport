package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-courier/loaderx"
	"github.com/stretchr/testify/require"
)

func TestOpenAPIGenerator(t *testing.T) {
	cwd, _ := os.Getwd()
	dir := filepath.Join(cwd, "../../__examples__")

	p, pkgInfo, err := loaderx.LoadWithTests(dir)
	require.NoError(t, err)

	g := NewOpenAPIGenerator(p, pkgInfo)

	g.Scan()
	g.Output(dir)
}
