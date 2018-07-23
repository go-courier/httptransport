package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-courier/loaderx"
	"github.com/stretchr/testify/require"
)

func TestStatusErrScanner(t *testing.T) {
	cwd, _ := os.Getwd()
	program, pkgInfo, _ := loaderx.LoadWithTests(filepath.Join(cwd, "./__examples__/status_err_scanner"))

	p := loaderx.NewPackageInfo(pkgInfo)

	scanner := NewStatusErrScanner(program)

	statusErrs := scanner.StatusErrorsInFunc(p.Func("main"))
	require.Len(t, statusErrs, 3)
}
