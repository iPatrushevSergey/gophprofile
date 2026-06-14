package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withArgs(t *testing.T, args ...string) {
	t.Helper()
	old := os.Args
	os.Args = append([]string{"gophprofile-server"}, args...)
	t.Cleanup(func() { os.Args = old })
}

func writeServerConfig(t *testing.T, yaml string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "server.yaml")
	require.NoError(t, os.WriteFile(path, []byte(yaml), 0o600))
	return path
}

func TestFinalizeConfig_ok(t *testing.T) {
	cfg := Config{
		Server: Server{
			Address:         "localhost:8080",
			ShutdownTimeout: time.Second,
		},
	}
	require.NoError(t, finalizeConfig(&cfg, "app/configs/server.yaml"))
	assert.Equal(t, "localhost:8080", cfg.Server.Address)
}

func TestServer_TLSEnabled(t *testing.T) {
	assert.False(t, Server{CertFile: "", KeyFile: ""}.TLSEnabled())
	assert.False(t, Server{CertFile: "cert.pem", KeyFile: ""}.TLSEnabled())
	assert.True(t, Server{CertFile: "cert.pem", KeyFile: "key.pem"}.TLSEnabled())
}

func TestFinalizeConfig_invalidAddress(t *testing.T) {
	cfg := Config{
		Server: Server{Address: "bad"},
	}
	assert.Error(t, finalizeConfig(&cfg, "app/configs/server.yaml"))
}

func TestServer_Normalize_tlsFilesMustExist(t *testing.T) {
	dir := t.TempDir()
	cert := filepath.Join(dir, "server.crt")
	key := filepath.Join(dir, "server.key")
	require.NoError(t, os.WriteFile(cert, []byte("cert"), 0o600))
	require.NoError(t, os.WriteFile(key, []byte("key"), 0o600))

	server := Server{CertFile: cert, KeyFile: key}
	require.NoError(t, server.Normalize())
}

func TestServer_Normalize_missingTLSFile(t *testing.T) {
	server := Server{
		CertFile: filepath.Join(t.TempDir(), "missing.crt"),
		KeyFile:  filepath.Join(t.TempDir(), "missing.key"),
	}
	require.Error(t, server.Normalize())
}

func TestFinalizeConfig_prodRequiresTLS(t *testing.T) {
	cfg := Config{
		Server: Server{Address: "0.0.0.0:8443"},
	}
	err := finalizeConfig(&cfg, "app/configs/server.prod.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires TLS")
}

func TestFinalizeConfig_prodOK(t *testing.T) {
	dir := t.TempDir()
	cert := filepath.Join(dir, "server.crt")
	key := filepath.Join(dir, "server.key")
	require.NoError(t, os.WriteFile(cert, []byte("cert"), 0o600))
	require.NoError(t, os.WriteFile(key, []byte("key"), 0o600))

	cfg := Config{
		Server: Server{
			Address:  "0.0.0.0:8443",
			CertFile: cert,
			KeyFile:  key,
		},
	}
	require.NoError(t, finalizeConfig(&cfg, "app/configs/server.prod.yaml"))
}

func TestLoadConfig_defaultValues(t *testing.T) {
	path := writeServerConfig(t, `
server:
  address: "127.0.0.1:8080"
  shutdown_timeout: "10s"
logger:
  level: info
minio:
  bucket: gophprofile
broker:
  exchange: avatars
`)
	withArgs(t, "-c", path)
	t.Setenv("GOPHPROFILE_SERVER_ADDRESS", "")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1:8080", cfg.Server.Address)
	assert.Equal(t, 10*time.Second, cfg.Server.ShutdownTimeout)
	assert.Equal(t, "gophprofile", cfg.MinIO.Bucket)
	assert.Equal(t, "avatars", cfg.Broker.Exchange)
}

func TestLoadConfig_customYAML(t *testing.T) {
	path := writeServerConfig(t, `
logger:
  level: warn
server:
  address: "localhost:9091"
  shutdown_timeout: "3s"
minio:
  bucket: custom-bucket
broker:
  exchange: custom-exchange
database:
  max_conns: 30
`)
	withArgs(t, "-c", path)
	t.Setenv("GOPHPROFILE_SERVER_ADDRESS", "")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "localhost:9091", cfg.Server.Address)
	assert.Equal(t, 3*time.Second, cfg.Server.ShutdownTimeout)
	assert.Equal(t, "warn", cfg.Logger.Level)
	assert.Equal(t, "custom-bucket", cfg.MinIO.Bucket)
	assert.Equal(t, "custom-exchange", cfg.Broker.Exchange)
	assert.Equal(t, int32(30), cfg.DB.Pool.MaxConns)
}

func TestLoadConfig_envAddress(t *testing.T) {
	path := writeServerConfig(t, `
server:
  address: "127.0.0.1:8080"
  shutdown_timeout: "10s"
`)
	withArgs(t, "-c", path)
	t.Setenv("GOPHPROFILE_SERVER_ADDRESS", "localhost:7070")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "localhost:7070", cfg.Server.Address)
}

func TestLoadConfig_flagAddress(t *testing.T) {
	path := writeServerConfig(t, `
server:
  address: "127.0.0.1:8080"
  shutdown_timeout: "10s"
`)
	withArgs(t, "-c", path, "-a", "localhost:9999")
	t.Setenv("GOPHPROFILE_SERVER_ADDRESS", "localhost:7070")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "localhost:9999", cfg.Server.Address)
}

func TestLoadConfig_viperDefaultsWithoutYAML(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("CONFIG", "")
	t.Setenv("GOPHPROFILE_SERVER_ADDRESS", "")
	withArgs(t)

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1:8080", cfg.Server.Address)
	assert.Equal(t, 10*time.Second, cfg.Server.ShutdownTimeout)
	assert.Equal(t, "info", cfg.Logger.Level)
	assert.Equal(t, int32(25), cfg.DB.Pool.MaxConns)
	assert.Equal(t, int32(5), cfg.DB.Pool.MinConns)
	assert.Equal(t, "gophprofile", cfg.MinIO.Bucket)
	assert.Equal(t, "avatars", cfg.Broker.Exchange)
}

func TestLoadConfig_configEnvPath(t *testing.T) {
	path := writeServerConfig(t, `
server:
  address: "localhost:9091"
  shutdown_timeout: "3s"
`)
	t.Setenv("CONFIG", path)
	t.Setenv("GOPHPROFILE_SERVER_ADDRESS", "")
	withArgs(t)

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "localhost:9091", cfg.Server.Address)
	assert.Equal(t, 3*time.Second, cfg.Server.ShutdownTimeout)
}

func TestLoadConfig_flagOverridesConfigEnv(t *testing.T) {
	envPath := writeServerConfig(t, `
server:
  address: "localhost:9091"
  shutdown_timeout: "3s"
`)
	flagPath := writeServerConfig(t, `
server:
  address: "localhost:7777"
  shutdown_timeout: "5s"
`)
	t.Setenv("CONFIG", envPath)
	t.Setenv("GOPHPROFILE_SERVER_ADDRESS", "")
	withArgs(t, "-c", flagPath)

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "localhost:7777", cfg.Server.Address)
	assert.Equal(t, 5*time.Second, cfg.Server.ShutdownTimeout)
}

func TestLoadConfig_missingConfigUsesDefaults(t *testing.T) {
	t.Setenv("CONFIG", "")
	t.Setenv("GOPHPROFILE_SERVER_ADDRESS", "")
	withArgs(t, "-c", filepath.Join(t.TempDir(), "missing.yaml"))

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1:8080", cfg.Server.Address)
	assert.Equal(t, "info", cfg.Logger.Level)
}

func TestLoadConfig_missingConfigEnvPathUsesDefaults(t *testing.T) {
	t.Setenv("CONFIG", filepath.Join(t.TempDir(), "missing.yaml"))
	t.Setenv("GOPHPROFILE_SERVER_ADDRESS", "")
	withArgs(t)

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1:8080", cfg.Server.Address)
	assert.Equal(t, "info", cfg.Logger.Level)
}
