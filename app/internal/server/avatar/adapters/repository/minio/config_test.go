package minio

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	cfg := Config{
		Endpoint:             " localhost:9000 ",
		AccessKey:            " access ",
		SecretKey:            " secret ",
		Bucket:               " gophprofile ",
		UploadReservationTTL: 30 * time.Minute,
		UploadGCInterval:     5 * time.Minute,
	}

	require.NoError(t, cfg.Validate())
	assert.Equal(t, "localhost:9000", cfg.Endpoint)
	assert.Equal(t, "access", cfg.AccessKey)
	assert.Equal(t, "secret", cfg.SecretKey)
	assert.Equal(t, "gophprofile", cfg.Bucket)
}

func TestConfig_Validate_skipsWhenDisabled(t *testing.T) {
	cfg := Config{}

	require.NoError(t, cfg.Validate())
}

func TestConfig_Validate_requiresBucketWhenEnabled(t *testing.T) {
	cfg := Config{
		Endpoint:  "localhost:9000",
		AccessKey: "access",
		SecretKey: "secret",
	}

	require.Error(t, cfg.Validate())
}
