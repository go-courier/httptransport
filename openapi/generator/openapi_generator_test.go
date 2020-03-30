package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-courier/packagesx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestOpenAPIGenerator(t *testing.T) {
	cwd, _ := os.Getwd()
	dir := filepath.Join(cwd, "../../__examples__")

	pkg, err := packagesx.Load(dir)
	require.NoError(t, err)

	logrus.SetLevel(logrus.DebugLevel)
	g := NewOpenAPIGenerator(pkg)

	g.Scan()
	g.Output(dir)
}
