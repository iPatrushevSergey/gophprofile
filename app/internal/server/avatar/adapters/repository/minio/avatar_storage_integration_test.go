//go:build integration

package minio_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/testutil"
)

func TestAvatarStorage_PutAndGet(t *testing.T) {
	storage := testutil.SetupMinIO(t)
	ctx := context.Background()
	payload := []byte("avatar-data")
	key := "user-1/avatar-1/original"

	require.NoError(t, storage.Put(ctx, key, payload, "image/png"))

	got, err := storage.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, payload, got)
	assert.Equal(t, "gophprofile-test", storage.Bucket())
}

func TestAvatarStorage_Delete(t *testing.T) {
	storage := testutil.SetupMinIO(t)
	ctx := context.Background()
	key := "user-1/avatar-1/delete"

	require.NoError(t, storage.Put(ctx, key, []byte("data"), "image/png"))
	require.NoError(t, storage.Delete(ctx, key))
}

func TestAvatarStorage_Ping(t *testing.T) {
	storage := testutil.SetupMinIO(t)
	require.NoError(t, storage.Ping(context.Background()))
}
