package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-courier/packagesx"
	"github.com/stretchr/testify/require"
)

func TestStatusErrScanner(t *testing.T) {
	cwd, _ := os.Getwd()
	pkg, _ := packagesx.Load(filepath.Join(cwd, "./__examples__/status_err_scanner"))

	scanner := NewStatusErrScanner(pkg)

	statusErrs := scanner.StatusErrorsInFunc(pkg.Func("main"))
	require.Len(t, statusErrs, 3)
}
