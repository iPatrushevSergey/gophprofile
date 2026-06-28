package minio

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAvatarStorage_requiresEndpoint(t *testing.T) {
	_, err := NewAvatarStorage(Config{Bucket: "test"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint")
}

func TestNewAvatarStorage_requiresBucket(t *testing.T) {
	_, err := NewAvatarStorage(Config{Endpoint: "localhost:9000"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bucket")
}
