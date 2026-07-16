package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withArgs(t *testing.T, args ...string) {
	t.Helper()
	old := os.Args
	os.Args = append([]string{"gophprofile-processor"}, args...)
	t.Cleanup(func() { os.Args = old })
}

func writeProcessorConfig(t *testing.T, yaml string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "processor.yaml")
	require.NoError(t, os.WriteFile(path, []byte(yaml), 0o600))
	return path
}

func TestFinalizeConfig_ok(t *testing.T) {
	cfg := Config{
		Logger:    logger.Config{Level: "info", Backend: "slog", Format: "json"},
		Telemetry: Telemetry{ServiceName: "gophprofile-processor", SampleRatio: 1},
		Worker: Worker{
			ShutdownTimeout:    time.Second,
			HealthFilePath:     "/tmp/health",
			HealthFileInterval: 30 * time.Second,
		},
	}
	require.NoError(t, finalizeConfig(&cfg, "app/configs/processor.yaml"))
	assert.Equal(t, time.Second, cfg.Worker.ShutdownTimeout)
	assert.Equal(t, "/tmp/health", cfg.Worker.HealthFilePath)
	assert.Equal(t, 30*time.Second, cfg.Worker.HealthFileInterval)
}

func TestTelemetry_Validate(t *testing.T) {
	tel := Telemetry{Enabled: true, ServiceName: "gophprofile-processor", SampleRatio: 1, OTLPEndpoint: "localhost:4317"}
	require.NoError(t, (&tel).Validate())

	bad := Telemetry{ServiceName: "gophprofile-processor", SampleRatio: 2}
	assert.Error(t, bad.Validate())
}

func TestWorker_Validate(t *testing.T) {
	w := Worker{
		ShutdownTimeout:    time.Second,
		HealthFilePath:     "  /tmp/health  ",
		HealthFileInterval: 30 * time.Second,
	}
	require.NoError(t, w.Validate())
	assert.Equal(t, "/tmp/health", w.HealthFilePath)

	assert.Error(t, (&Worker{ShutdownTimeout: time.Second, HealthFileInterval: time.Second}).Validate())
	assert.Error(t, (&Worker{ShutdownTimeout: time.Second, HealthFilePath: "/tmp/health"}).Validate())
}

func TestLoadConfig_defaultValues(t *testing.T) {
	path := writeProcessorConfig(t, `
logger:
  level: info
worker:
  shutdown_timeout: "10s"
minio:
  bucket: gophprofile
broker:
  exchange: avatars
  queue: avatar-processing
`)
	withArgs(t, "-c", path)
	t.Setenv("GOPHPROFILE_WORKER_SHUTDOWN_TIMEOUT", "")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 10*time.Second, cfg.Worker.ShutdownTimeout)
	assert.Equal(t, "/tmp/health", cfg.Worker.HealthFilePath)
	assert.Equal(t, 30*time.Second, cfg.Worker.HealthFileInterval)
	assert.Equal(t, "gophprofile", cfg.MinIO.Bucket)
	assert.Equal(t, "avatars", cfg.Broker.Exchange)
	assert.Equal(t, "avatar-processing", cfg.Broker.Queue)
}

func TestLoadConfig_customYAML(t *testing.T) {
	path := writeProcessorConfig(t, `
logger:
  level: warn
worker:
  shutdown_timeout: "3s"
minio:
  bucket: custom-bucket
broker:
  exchange: custom-exchange
  queue: custom-queue
database:
  max_conns: 20
`)
	withArgs(t, "-c", path)
	t.Setenv("GOPHPROFILE_WORKER_SHUTDOWN_TIMEOUT", "")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 3*time.Second, cfg.Worker.ShutdownTimeout)
	assert.Equal(t, "warn", cfg.Logger.Level)
	assert.Equal(t, "custom-bucket", cfg.MinIO.Bucket)
	assert.Equal(t, "custom-exchange", cfg.Broker.Exchange)
	assert.Equal(t, "custom-queue", cfg.Broker.Queue)
	assert.Equal(t, int32(20), cfg.DB.Pool.MaxConns)
}

func TestLoadConfig_envShutdownTimeout(t *testing.T) {
	path := writeProcessorConfig(t, `
worker:
  shutdown_timeout: "10s"
`)
	withArgs(t, "-c", path)
	t.Setenv("GOPHPROFILE_WORKER_SHUTDOWN_TIMEOUT", "7s")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 7*time.Second, cfg.Worker.ShutdownTimeout)
}

func TestLoadConfig_flagShutdownTimeout(t *testing.T) {
	path := writeProcessorConfig(t, `
worker:
  shutdown_timeout: "10s"
`)
	withArgs(t, "-c", path, "--shutdown-timeout", "5s")
	t.Setenv("GOPHPROFILE_WORKER_SHUTDOWN_TIMEOUT", "7s")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 5*time.Second, cfg.Worker.ShutdownTimeout)
}

func TestLoadConfig_viperDefaultsWithoutYAML(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("CONFIG", "")
	t.Setenv("GOPHPROFILE_WORKER_SHUTDOWN_TIMEOUT", "")
	withArgs(t)

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 10*time.Second, cfg.Worker.ShutdownTimeout)
	assert.Equal(t, "/tmp/health", cfg.Worker.HealthFilePath)
	assert.Equal(t, 30*time.Second, cfg.Worker.HealthFileInterval)
	assert.Equal(t, "info", cfg.Logger.Level)
	assert.Equal(t, int32(10), cfg.DB.Pool.MaxConns)
	assert.Equal(t, int32(2), cfg.DB.Pool.MinConns)
	assert.Equal(t, "gophprofile", cfg.MinIO.Bucket)
	assert.Equal(t, "avatars", cfg.Broker.Exchange)
	assert.Equal(t, "avatar-processing", cfg.Broker.Queue)
	assert.False(t, cfg.Telemetry.Enabled)
	assert.Equal(t, "gophprofile-processor", cfg.Telemetry.ServiceName)
	assert.Equal(t, "localhost:4317", cfg.Telemetry.OTLPEndpoint)
	assert.True(t, cfg.Telemetry.OTLPInsecure)
	assert.Equal(t, 1.0, cfg.Telemetry.SampleRatio)
	assert.Equal(t, "development", cfg.Telemetry.Environment)
}

func TestLoadConfig_configEnvPath(t *testing.T) {
	path := writeProcessorConfig(t, `
worker:
  shutdown_timeout: "3s"
`)
	t.Setenv("CONFIG", path)
	t.Setenv("GOPHPROFILE_WORKER_SHUTDOWN_TIMEOUT", "")
	withArgs(t)

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 3*time.Second, cfg.Worker.ShutdownTimeout)
}

func TestLoadConfig_flagOverridesConfigEnv(t *testing.T) {
	envPath := writeProcessorConfig(t, `
worker:
  shutdown_timeout: "3s"
`)
	flagPath := writeProcessorConfig(t, `
worker:
  shutdown_timeout: "5s"
`)
	t.Setenv("CONFIG", envPath)
	t.Setenv("GOPHPROFILE_WORKER_SHUTDOWN_TIMEOUT", "")
	withArgs(t, "-c", flagPath)

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 5*time.Second, cfg.Worker.ShutdownTimeout)
}

func TestLoadConfig_missingConfigUsesDefaults(t *testing.T) {
	t.Setenv("CONFIG", "")
	t.Setenv("GOPHPROFILE_WORKER_SHUTDOWN_TIMEOUT", "")
	withArgs(t, "-c", filepath.Join(t.TempDir(), "missing.yaml"))

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 10*time.Second, cfg.Worker.ShutdownTimeout)
	assert.Equal(t, "info", cfg.Logger.Level)
}

func TestLoadConfig_missingConfigEnvPathUsesDefaults(t *testing.T) {
	t.Setenv("CONFIG", filepath.Join(t.TempDir(), "missing.yaml"))
	t.Setenv("GOPHPROFILE_WORKER_SHUTDOWN_TIMEOUT", "")
	withArgs(t)

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 10*time.Second, cfg.Worker.ShutdownTimeout)
	assert.Equal(t, "info", cfg.Logger.Level)
}
