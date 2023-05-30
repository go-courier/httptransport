package openapi

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-courier/logr"
	"github.com/go-courier/logr/slog"
	"github.com/go-courier/packagesx"
	. "github.com/onsi/gomega"
)

func TestOpenAPIGenerator(t *testing.T) {
	cwd, _ := os.Getwd()
	dir := filepath.Join(cwd, "../../testdata/server/cmd/app")

	ctx := logr.WithLogger(context.Background(), slog.Logger(slog.Default()))

	pkg, err := packagesx.Load(dir)
	NewWithT(t).Expect(err).To(BeNil())

	g := NewOpenAPIGenerator(pkg)

	g.Scan(ctx)
	g.Output(dir)
}
