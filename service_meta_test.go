package httptransport

import (
	"context"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestServiceMeta(t *testing.T) {
	serviceMeta := &ServiceMeta{}

	os.Setenv("PROJECT_NAME", "service-example")
	serviceMeta.SetDefaults()
	require.Equal(t, "service-example", serviceMeta.String())

	os.Setenv("PROJECT_VERSION", "1.0.0")
	serviceMeta.SetDefaults()
	require.Equal(t, "service-example@1.0.0", serviceMeta.String())
}

func TestServiceMetaWithContext(t *testing.T) {
	ctx := context.Background()

	ctx = ContextWithServiceMeta(ctx, ServiceMeta{
		Name: "test",
	})
	serviceMeta := ServerMetaFromContext(ctx)
	require.Equal(t, "test", serviceMeta.Name)
}

func TestHookForServiceMeta(t *testing.T) {
	hook := ServiceMetaHook{
		ServiceMeta{
			Name: "test",
		},
	}

	entry := logrus.NewEntry(logrus.StandardLogger())

	require.Equal(t, logrus.AllLevels, hook.Levels())
	require.NoError(t, hook.Fire(entry))
	require.Equal(t, hook.ServiceMeta.String(), entry.Data["service"])
}
