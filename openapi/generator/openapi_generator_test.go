package generator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-courier/logr"

	"github.com/go-courier/packagesx"
	"github.com/stretchr/testify/require"
)

func TestOpenAPIGenerator(t *testing.T) {
	cwd, _ := os.Getwd()
	dir := filepath.Join(cwd, "../../__examples__/server/cmd/app")

	ctx := logr.WithLogger(context.Background(), logr.StdLogger())

	pkg, err := packagesx.Load(dir)
	require.NoError(t, err)

	g := NewOpenAPIGenerator(pkg)

	g.Scan(ctx)
	g.Output(dir)
}
