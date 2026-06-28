//go:build integration

package minio_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcminio "github.com/testcontainers/testcontainers-go/modules/minio"

	processorminio "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/adapters/repository/minio"
)

const (
	testMinIOUser   = "minioadmin"
	testMinIOPass   = "minioadmin"
	testMinIOBucket = "gophprofile-test"
)

func setupProcessorAvatarStorage(tb testing.TB) *processorminio.AvatarStorage {
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

	storage, err := processorminio.NewAvatarStorage(processorminio.Config{
		Endpoint:  endpoint,
		AccessKey: testMinIOUser,
		SecretKey: testMinIOPass,
		Bucket:    testMinIOBucket,
		UseSSL:    false,
	})
	require.NoError(tb, err)

	return storage
}

func TestAvatarStorage_PutAndGet(t *testing.T) {
	storage := setupProcessorAvatarStorage(t)
	ctx := context.Background()
	payload := []byte("processor-avatar")
	key := "user-1/avatar-1/original"

	require.NoError(t, storage.Put(ctx, key, payload, "image/png"))

	got, err := storage.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, payload, got)
}
