package openapi

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	NewWithT(t).Expect(pkgImportPathCourier).To(Equal("github.com/go-courier/courier"))
	NewWithT(t).Expect(pkgImportPathHttpTransport).To(Equal("github.com/go-courier/httptransport"))
	NewWithT(t).Expect(pkgImportPathHttpx).To(Equal("github.com/go-courier/httptransport/httpx"))
}
