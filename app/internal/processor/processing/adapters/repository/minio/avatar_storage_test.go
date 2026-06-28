package minio

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAvatarStorage_requiresEndpoint(t *testing.T) {
	_, err := NewAvatarStorage(Config{Bucket: "test", AccessKey: "a", SecretKey: "b"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint")
}
