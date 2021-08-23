package httptransport

import (
	"context"
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

func TestServiceMeta(t *testing.T) {
	serviceMeta := &ServiceMeta{}

	os.Setenv("PROJECT_NAME", "service-example")
	serviceMeta.SetDefaults()
	NewWithT(t).Expect(serviceMeta.String()).To(Equal("service-example"))

	os.Setenv("PROJECT_VERSION", "1.0.0")
	serviceMeta.SetDefaults()
	NewWithT(t).Expect(serviceMeta.String()).To(Equal("service-example@1.0.0"))
}

func TestServiceMetaWithContext(t *testing.T) {
	ctx := context.Background()

	ctx = ContextWithServiceMeta(ctx, ServiceMeta{
		Name: "test",
	})
	serviceMeta := ServerMetaFromContext(ctx)
	NewWithT(t).Expect(serviceMeta.Name).To(Equal("test"))
}
