// Package config loads all environment-backed settings for the service.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

const defaultJWTSecret = "costguard-dev-secret-change-me"

// knownWeakSecrets are example/placeholder values that must never be used in
// production.
var knownWeakSecrets = []string{
	defaultJWTSecret,
	"please-change-me-to-a-long-random-string",
	"change-me",
	"changeme",
	"secret",
	"jwt-secret",
}

// PostgresConfig holds database connection settings.
type PostgresConfig struct {
	Host     string `env:"DB_HOST" envDefault:"127.0.0.1"`
	Port     int    `env:"DB_PORT" envDefault:"33334"`
	User     string `env:"DB_USER" envDefault:"costguard"`
	Password string `env:"DB_PASSWORD" envDefault:"costguard_pwd"`
	Database string `env:"DB_NAME" envDefault:"costguard"`
	SSLMode  string `env:"DB_SSLMODE" envDefault:"disable"`
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host     string `env:"REDIS_HOST" envDefault:"127.0.0.1"`
	Port     int    `env:"REDIS_PORT" envDefault:"36334"`
	Password string `env:"REDIS_PASSWORD" envDefault:""`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
}

// Config aggregates every environment-backed setting.
type Config struct {
	Env        string `env:"APP_ENV" envDefault:"development"`
	ServerPort int    `env:"SERVER_PORT" envDefault:"8080"`
	JWTSecret  string `env:"JWT_SECRET" envDefault:"costguard-dev-secret-change-me"`
	JWTExpire  int    `env:"JWT_EXPIRE_HOURS" envDefault:"72"`

	// CORS_ALLOWED_ORIGINS is a comma-separated origin list.
	CORSAllowedOrigins string `env:"CORS_ALLOWED_ORIGINS" envDefault:"http://localhost:19203"`

	// Rate limits are applied per client IP.
	AuthRateLimit int `env:"AUTH_RATE_LIMIT" envDefault:"10"`
	APIRateLimit  int `env:"API_RATE_LIMIT" envDefault:"60"`

	// Database pool tuning.
	DBMaxOpenConns            int `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	DBMaxIdleConns            int `env:"DB_MAX_IDLE_CONNS" envDefault:"5"`
	DBConnMaxLifetimeMin      int `env:"DB_CONN_MAX_LIFETIME_MIN" envDefault:"5"`
	DBConnectRetries          int `env:"DB_CONNECT_RETRIES" envDefault:"10"`
	DBConnectRetryIntervalSec int `env:"DB_CONNECT_RETRY_INTERVAL_SEC" envDefault:"3"`
	RedisConnectRetries       int `env:"REDIS_CONNECT_RETRIES" envDefault:"10"`
	RedisConnectRetryInterval int `env:"REDIS_CONNECT_RETRY_INTERVAL_SEC" envDefault:"3"`

	Postgres PostgresConfig
	Redis    RedisConfig
}

// Load reads configuration from environment variables and validates it.
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.ServerPort < 1 || c.ServerPort > 65535 {
		return fmt.Errorf("SERVER_PORT must be between 1 and 65535")
	}
	if c.JWTExpire <= 0 {
		return fmt.Errorf("JWT_EXPIRE_HOURS must be greater than 0")
	}
	if c.AuthRateLimit <= 0 || c.APIRateLimit <= 0 {
		return fmt.Errorf("AUTH_RATE_LIMIT and API_RATE_LIMIT must be greater than 0")
	}
	if c.DBMaxOpenConns <= 0 || c.DBMaxIdleConns < 0 {
		return fmt.Errorf("DB_MAX_OPEN_CONNS must be > 0 and DB_MAX_IDLE_CONNS must be >= 0")
	}
	if c.DBConnectRetries < 0 || c.DBConnectRetryIntervalSec < 0 {
		return fmt.Errorf("DB_CONNECT_RETRIES and DB_CONNECT_RETRY_INTERVAL_SEC must be >= 0")
	}
	if c.RedisConnectRetries < 0 || c.RedisConnectRetryInterval < 0 {
		return fmt.Errorf("REDIS_CONNECT_RETRIES and REDIS_CONNECT_RETRY_INTERVAL_SEC must be >= 0")
	}
	if c.Env == "production" {
		if c.JWTSecret == "" || len(c.JWTSecret) < 32 || isWeakSecret(c.JWTSecret) {
			return fmt.Errorf("JWT_SECRET must be replaced with a value of at least 32 characters in production")
		}
		if strings.TrimSpace(c.CORSAllowedOrigins) == "*" {
			return fmt.Errorf("CORS_ALLOWED_ORIGINS must not be '*' in production")
		}
	}
	return nil
}

// AllowedOrigins parses CORS_ALLOWED_ORIGINS. Empty or '*' disables CORS.
func (c *Config) AllowedOrigins() []string {
	raw := strings.TrimSpace(c.CORSAllowedOrigins)
	if raw == "" || raw == "*" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// DSN builds the PostgreSQL connection string.
func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		c.Postgres.Host, c.Postgres.Port, c.Postgres.User, c.Postgres.Password, c.Postgres.Database, c.Postgres.SSLMode)
}

// RedisAddr returns the Redis address.
func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// JWTExpireDuration returns the access token lifetime.
func (c *Config) JWTExpireDuration() time.Duration {
	return time.Duration(c.JWTExpire) * time.Hour
}

func isWeakSecret(v string) bool {
	for _, weak := range knownWeakSecrets {
		if strings.EqualFold(v, weak) {
			return true
		}
	}
	return false
}
