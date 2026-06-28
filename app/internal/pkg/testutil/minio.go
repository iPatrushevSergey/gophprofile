//go:build integration || e2e || component || contract

package testutil

import (
	"context"
	"testing"

	avatarminio "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/minio"
	"github.com/stretchr/testify/require"
	tcminio "github.com/testcontainers/testcontainers-go/modules/minio"
)

const (
	testMinIOUser   = "minioadmin"
	testMinIOPass   = "minioadmin"
	testMinIOBucket = "gophprofile-test"
)

// SetupMinIO starts MinIO in a container and returns a configured adapter.
func SetupMinIO(tb testing.TB) *avatarminio.AvatarStorage {
	tb.Helper()
	ctx := context.Background()

	container, err := tcminio.Run(ctx,
		"minio/minio:latest",
		tcminio.WithUsername(testMinIOUser),
		tcminio.WithPassword(testMinIOPass),
	)
	require.NoError(tb, err, "start minio container")
	tb.Cleanup(func() {
		require.NoError(tb, container.Terminate(ctx))
	})

	endpoint, err := container.ConnectionString(ctx)
	require.NoError(tb, err)

	storage, err := avatarminio.NewAvatarStorage(avatarminio.Config{
		Endpoint:  endpoint,
		AccessKey: testMinIOUser,
		SecretKey: testMinIOPass,
		Bucket:    testMinIOBucket,
		UseSSL:    false,
	})
	require.NoError(tb, err)

	return storage
}
