package minio

import (
	"fmt"
	"strings"
	"time"
)

// Config holds MinIO connection settings.
type Config struct {
	Endpoint             string        `mapstructure:"endpoint"`
	AccessKey            string        `mapstructure:"access_key"`
	SecretKey            string        `mapstructure:"secret_key"`
	Bucket               string        `mapstructure:"bucket"`
	UseSSL               bool          `mapstructure:"use_ssl"`
	UploadReservationTTL time.Duration `mapstructure:"upload_reservation_ttl"`
	UploadGCInterval     time.Duration `mapstructure:"upload_gc_interval"`
}

// Validate trims string fields and checks settings when MinIO is enabled.
func (c *Config) Validate() error {
	c.Endpoint = strings.TrimSpace(c.Endpoint)
	c.AccessKey = strings.TrimSpace(c.AccessKey)
	c.SecretKey = strings.TrimSpace(c.SecretKey)
	c.Bucket = strings.TrimSpace(c.Bucket)

	if !c.Enabled() {
		return nil
	}

	if c.Bucket == "" {
		return fmt.Errorf("minio bucket is required")
	}

	if c.UploadReservationTTL <= 0 {
		return fmt.Errorf("minio upload_reservation_ttl must be positive")
	}

	if c.UploadGCInterval <= 0 {
		return fmt.Errorf("minio upload_gc_interval must be positive")
	}
	return nil
}

// Enabled reports whether MinIO integration is configured.
func (c Config) Enabled() bool {
	return c.Endpoint != "" && c.AccessKey != "" && c.SecretKey != ""
}
