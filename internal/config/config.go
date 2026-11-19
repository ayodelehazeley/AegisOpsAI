// Package config provides configuration loading and validation for AegisOpsAI.
// It uses Viper to support configuration from YAML files, environment variables
// (AEGIS_ prefix), and sensible defaults for local development.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration sections for the AegisOpsAI platform.
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	NATS      NATSConfig      `mapstructure:"nats"`
	Auth      AuthConfig      `mapstructure:"auth"`
	AI        AIConfig        `mapstructure:"ai"`
	Telemetry TelemetryConfig `mapstructure:"telemetry"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // "development", "production", "testing"
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	URL      string `mapstructure:"url"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
	MaxConns int32  `mapstructure:"max_conns"`
}

// DSN returns a PostgreSQL connection string built from the database configuration
// or standard environment variables (DATABASE_URL / AEGIS_DATABASE_URL).
func (d DatabaseConfig) DSN() string {
	if d.URL != "" {
		return d.URL
	}
	if envURL := os.Getenv("DATABASE_URL"); envURL != "" {
		return envURL
	}
	if envURL := os.Getenv("AEGIS_DATABASE_URL"); envURL != "" {
		return envURL
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode,
	)
}

// NATSConfig holds NATS messaging settings.
type NATSConfig struct {
	URL       string `mapstructure:"url"`
	ClusterID string `mapstructure:"cluster_id"`
}

// AuthConfig holds authentication/authorization settings.
type AuthConfig struct {
	JWTSecret   string        `mapstructure:"jwt_secret"`
	TokenExpiry time.Duration `mapstructure:"token_expiry"`
}

// AIConfig holds AI provider integration settings.
type AIConfig struct {
	Provider string `mapstructure:"provider"` // "openai", "ollama", "anthropic"
	APIKey   string `mapstructure:"api_key"`
	Model    string `mapstructure:"model"`
	BaseURL  string `mapstructure:"base_url"`
	Enabled  bool   `mapstructure:"enabled"`
}

// TelemetryConfig holds OpenTelemetry observability settings.
type TelemetryConfig struct {
	Enabled          bool   `mapstructure:"enabled"`
	ExporterEndpoint string `mapstructure:"exporter_endpoint"`
}

// Addr returns the formatted listen address for the HTTP server.
func (s ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// Load reads configuration from file, environment variables, and defaults.
// It searches for "aegisops.yaml" in ./configs, /etc/aegisops, and $HOME/.aegisops.
// Environment variables use the AEGIS_ prefix with underscores replacing dots
// (e.g., AEGIS_SERVER_PORT=8080).
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults for local development.
	setDefaults(v)

	// Configuration file settings.
	v.SetConfigName("aegisops")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath("/etc/aegisops")
	v.AddConfigPath("$HOME/.aegisops")

	// Environment variable settings.
	v.SetEnvPrefix("AEGIS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Explicitly bind env vars for clean struct unmarshaling
	for _, k := range []string{
		"server.host", "server.port", "server.mode",
		"database.url", "database.host", "database.port", "database.user", "database.password",
		"database.dbname", "database.sslmode", "database.max_conns",
		"nats.url", "nats.cluster_id",
		"auth.jwt_secret", "auth.token_expiry",
		"ai.provider", "ai.api_key", "ai.model", "ai.base_url", "ai.enabled",
		"telemetry.enabled", "telemetry.exporter_endpoint",
	} {
		_ = v.BindEnv(k)
	}

	// Read config file (not finding one is acceptable; we have defaults).
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	// Cloud platform overrides (Render, Heroku, Cloud Run)
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
			cfg.Server.Port = p
		}
	}
	if natsURL := os.Getenv("NATS_URL"); natsURL != "" {
		cfg.NATS.URL = natsURL
	}

	return &cfg, nil
}

// setDefaults registers sensible default values for all configuration keys.
func setDefaults(v *viper.Viper) {
	// Server defaults.
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "development")

	// Database defaults.
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "aegisops")
	v.SetDefault("database.password", "aegisops")
	v.SetDefault("database.dbname", "aegisops")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.max_conns", 25)

	// NATS defaults.
	v.SetDefault("nats.url", "nats://localhost:4222")
	v.SetDefault("nats.cluster_id", "aegisops-cluster")

	// Auth defaults.
	v.SetDefault("auth.jwt_secret", "change-me-in-production")
	v.SetDefault("auth.token_expiry", "24h")

	// AI defaults.
	v.SetDefault("ai.provider", "openai")
	v.SetDefault("ai.api_key", "")
	v.SetDefault("ai.model", "gpt-4")
	v.SetDefault("ai.base_url", "https://api.openai.com/v1")
	v.SetDefault("ai.enabled", false)

	// Telemetry defaults.
	v.SetDefault("telemetry.enabled", false)
	v.SetDefault("telemetry.exporter_endpoint", "localhost:4317")
}
