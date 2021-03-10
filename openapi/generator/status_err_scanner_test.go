package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/gomega"

	"github.com/go-courier/packagesx"
)

func TestStatusErrScanner(t *testing.T) {
	cwd, _ := os.Getwd()
	pkg, _ := packagesx.Load(filepath.Join(cwd, "./__examples__/status_err_scanner"))

	scanner := NewStatusErrScanner(pkg)

	t.Run("should scan from comments", func(t *testing.T) {
		statusErrs := scanner.StatusErrorsInFunc(pkg.Func("call"))
		gomega.NewWithT(t).Expect(statusErrs).To(gomega.HaveLen(2))
	})

	t.Run("should scan all", func(t *testing.T) {
		statusErrs := scanner.StatusErrorsInFunc(pkg.Func("main"))
		gomega.NewWithT(t).Expect(statusErrs).To(gomega.HaveLen(3))
	})
}
