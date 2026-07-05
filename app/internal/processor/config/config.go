// Package config loads processor service settings.
package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	mapstructure "github.com/go-viper/mapstructure/v2"
	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/logger"
	prommetrics "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/metrics/prometheus"
	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/repository/postgres"
	processingbroker "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/adapters/broker/rabbitmq"
	processingminio "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/adapters/repository/minio"
)

// Config holds grouped processor configuration.
type Config struct {
	Logger    logger.Config           `mapstructure:"logger"`
	Telemetry Telemetry               `mapstructure:"telemetry"`
	Metrics   prommetrics.Config      `mapstructure:"metrics"`
	DB        postgres.Config         `mapstructure:"database"`
	MinIO     processingminio.Config  `mapstructure:"minio"`
	Broker    processingbroker.Config `mapstructure:"broker"`
	Worker    Worker                  `mapstructure:"worker"`
}

// Telemetry holds OpenTelemetry settings.
type Telemetry struct {
	Enabled      bool    `mapstructure:"enabled"`
	ServiceName  string  `mapstructure:"service_name"`
	OTLPEndpoint string  `mapstructure:"otlp_endpoint"`
	OTLPInsecure bool    `mapstructure:"otlp_insecure"`
	SampleRatio  float64 `mapstructure:"sample_ratio"`
	Environment  string  `mapstructure:"environment"`
}

// Validate trims and checks telemetry settings.
func (t *Telemetry) Validate() error {
	t.ServiceName = strings.TrimSpace(t.ServiceName)
	t.OTLPEndpoint = strings.TrimSpace(t.OTLPEndpoint)
	t.Environment = strings.TrimSpace(t.Environment)

	if t.ServiceName == "" {
		return fmt.Errorf("service_name is required")
	}

	if t.SampleRatio < 0 || t.SampleRatio > 1 {
		return fmt.Errorf("sample_ratio must be between 0 and 1")
	}

	return nil
}

// Worker holds background worker settings.
type Worker struct {
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// Validate checks worker settings.
func (w *Worker) Validate() error {
	if w.ShutdownTimeout <= 0 {
		return fmt.Errorf("shutdown_timeout must be positive")
	}
	return nil
}

// LoadConfig loads processor config.
// Field priority: flags > env > yaml > viper defaults.
// Config file path: flag -c > env CONFIG > default path.
func LoadConfig() (Config, error) {
	// Load flags.
	fs, err := newFlagSet()
	if err != nil {
		return Config{}, err
	}

	// Load .env file.
	dotenvPath, dotenvLoaded, err := loadDotEnv()
	if err != nil {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}
	if dotenvLoaded {
		_, _ = fmt.Fprintf(os.Stderr, "config: loaded dotenv %s\n", dotenvPath)
	} else {
		_, _ = fmt.Fprintln(os.Stderr, "config: dotenv file not found, continuing with env/yaml/defaults")
	}

	// Resolve config path.
	configPath := resolveConfigPath(fs)

	// Create viper instance.
	v := viper.New()

	// Set defaults.
	setDefaults(v)

	// Read config file.
	if err := readConfigFile(v, configPath); err != nil {
		return Config{}, err
	}

	// Bind environment variables.
	bindEnv(v)

	// Bind flags.
	if err := bindFlags(v, fs); err != nil {
		return Config{}, fmt.Errorf("bind flags: %w", err)
	}

	// Unmarshal config.
	var cfg Config
	if err := v.Unmarshal(&cfg, viper.DecodeHook(durationDecodeHook())); err != nil {
		return Config{}, fmt.Errorf("unmarshal: %w", err)
	}

	// Finalize config.
	if err := finalizeConfig(&cfg, configPath); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// newFlagSet creates a new flag set for the processor.
func newFlagSet() (*pflag.FlagSet, error) {
	fs := pflag.NewFlagSet("processor", pflag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringP("config", "c", "", "path to config file")
	fs.StringP("log-level", "l", "", "logging level")
	fs.StringP("database-uri", "d", "", "PostgreSQL DSN")
	fs.String("shutdown-timeout", "", "graceful shutdown timeout")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, fmt.Errorf("flag parsing: %w", err)
	}
	return fs, nil
}

// loadDotEnv loads the .env file.
func loadDotEnv() (string, bool, error) {
	for _, file := range []string{"app/.env", ".env"} {
		if _, err := os.Stat(file); err == nil {
			if err := godotenv.Load(file); err != nil {
				return "", false, fmt.Errorf("load %s: %w", file, err)
			}
			return file, true, nil
		}
	}
	return "", false, nil
}

// resolveConfigPath resolves the config file path.
func resolveConfigPath(fs *pflag.FlagSet) string {
	path := "app/configs/processor.yaml"
	source := "default"
	if env, ok := os.LookupEnv("CONFIG"); ok {
		if trimmedPath := strings.TrimSpace(env); trimmedPath != "" {
			path = trimmedPath
			source = "env"
		}
	}
	if fs.Changed("config") {
		flagPath, err := fs.GetString("config")
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "config: read -c flag: %v\n", err)
		} else {
			path = flagPath
			source = "flag"
		}
	}
	_, _ = fmt.Fprintf(os.Stderr, "config: config file path source=%s path=%s\n", source, path)
	return path
}

// setDefaults sets the default values for the processor.
func setDefaults(v *viper.Viper) {
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.backend", "slog")
	v.SetDefault("logger.format", "json")
	v.SetDefault("telemetry.enabled", false)
	v.SetDefault("telemetry.service_name", "gophprofile-processor")
	v.SetDefault("telemetry.otlp_endpoint", "localhost:4317")
	v.SetDefault("telemetry.otlp_insecure", true)
	v.SetDefault("telemetry.sample_ratio", 1.0)
	v.SetDefault("telemetry.environment", "development")
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.address", ":9091")
	v.SetDefault("worker.shutdown_timeout", "10s")
	v.SetDefault("database.uri", "")
	v.SetDefault("database.max_conns", 10)
	v.SetDefault("database.min_conns", 2)
	v.SetDefault("database.max_conn_life", "1h")
	v.SetDefault("database.max_conn_idle", "30m")
	v.SetDefault("database.health_check", "1m")
	v.SetDefault("database.retry.max_retries", 3)
	v.SetDefault("database.retry.base_delay", "100ms")
	v.SetDefault("database.retry.max_delay", "2s")
	v.SetDefault("minio.endpoint", "")
	v.SetDefault("minio.access_key", "")
	v.SetDefault("minio.secret_key", "")
	v.SetDefault("minio.bucket", "gophprofile")
	v.SetDefault("minio.use_ssl", false)
	v.SetDefault("broker.url", "")
	v.SetDefault("broker.exchange", "avatars")
	v.SetDefault("broker.queue", "avatar-processing")
	v.SetDefault("broker.dead_letter_exchange", "avatars.dlx")
	v.SetDefault("broker.dead_letter_queue", "avatar-processing.dlq")
	v.SetDefault("broker.dead_letter_routing_key", "avatar-processing.dlq")
	v.SetDefault("broker.retry_queue", "avatar-processing.retry")
	v.SetDefault("broker.retry_exchange", "avatars.retry.dlx")
	v.SetDefault("broker.retry_return_routing_key", "back")
	v.SetDefault("broker.retry_ttl", "30s")
	v.SetDefault("broker.max_retries", 5)
	v.SetDefault("broker.reconnect_interval", "1s")
}

// readConfigFile reads the config file.
func readConfigFile(v *viper.Viper, configPath string) error {
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_, _ = fmt.Fprintf(os.Stderr, "config: file not found path=%s\n", v.ConfigFileUsed())
			return nil
		}
		return fmt.Errorf("read config: %w", err)
	}

	_, _ = fmt.Fprintf(os.Stderr, "config: loaded %s\n", v.ConfigFileUsed())
	return nil
}

// bindEnv binds the environment variables to the processor.
func bindEnv(v *viper.Viper) {
	v.SetEnvPrefix("GOPHPROFILE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
}

// bindFlags binds the flags to the processor.
func bindFlags(v *viper.Viper, fs *pflag.FlagSet) error {
	bindings := map[string]string{
		"logger.level":            "log-level",
		"database.uri":            "database-uri",
		"worker.shutdown_timeout": "shutdown-timeout",
	}
	for key, flagName := range bindings {
		f := fs.Lookup(flagName)
		if f == nil {
			return fmt.Errorf("flag not found: %s", flagName)
		}
		if err := v.BindPFlag(key, f); err != nil {
			return fmt.Errorf("bind %s: %w", flagName, err)
		}
	}
	return nil
}

// finalizeConfig validates loaded config and applies cross-cutting rules.
func finalizeConfig(cfg *Config, configPath string) error {
	if err := cfg.Logger.Validate(); err != nil {
		return fmt.Errorf("logger: %w", err)
	}

	if err := cfg.Telemetry.Validate(); err != nil {
		return fmt.Errorf("telemetry: %w", err)
	}

	if err := cfg.Metrics.Validate(); err != nil {
		return fmt.Errorf("metrics: %w", err)
	}

	if err := cfg.DB.Validate(); err != nil {
		return fmt.Errorf("database: %w", err)
	}

	if err := cfg.MinIO.Validate(); err != nil {
		return fmt.Errorf("minio: %w", err)
	}

	if err := cfg.Broker.Validate(); err != nil {
		return fmt.Errorf("broker: %w", err)
	}

	if err := cfg.Worker.Validate(); err != nil {
		return fmt.Errorf("worker: %w", err)
	}

	return nil
}

// durationDecodeHook maps incoming scalars to time.Duration.
func durationDecodeHook() mapstructure.DecodeHookFunc {
	durType := reflect.TypeOf(time.Duration(0))
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if to != durType {
			return data, nil
		}

		var d Duration
		switch v := data.(type) {
		case string:
			if err := d.Set(v); err != nil {
				return nil, err
			}
			return d.Duration, nil
		case int:
			return time.Duration(v) * time.Second, nil
		case int64:
			return time.Duration(v) * time.Second, nil
		case float64:
			return time.Duration(v) * time.Second, nil
		default:
			return nil, fmt.Errorf("unsupported duration type %T", data)
		}
	}
}
