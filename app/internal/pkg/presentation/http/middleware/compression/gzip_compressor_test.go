package compression

import (
	"bytes"
	"compress/gzip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGzipCompressor(t *testing.T) {
	c, err := NewGzipCompressor(gzip.DefaultCompression)
	require.NoError(t, err)
	assert.Equal(t, "gzip", c.ContentEncoding())

	var buf bytes.Buffer
	w := c.NewWriter(&buf)
	_, err = w.Write([]byte("ok"))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	r, err := c.NewReader(&buf)
	require.NoError(t, err)
	defer r.Close()
}

func TestNewGzipCompressor_invalidLevel(t *testing.T) {
	_, err := NewGzipCompressor(99)
	assert.Error(t, err)
}
