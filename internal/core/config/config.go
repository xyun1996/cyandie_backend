package config

import (
	"fmt"
	"os"

	"github.com/cyandie/backend/internal/core/logger"
	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	HTTPAddr string `yaml:"http_addr" env:"HTTP_ADDR"`
	GRPCAddr string `yaml:"grpc_addr" env:"GRPC_ADDR"`
}

type DatabaseConfig struct {
	DSN             string `yaml:"dsn" env:"DATABASE_DSN"`
	MaxOpenConns    int    `yaml:"max_open_conns" env:"DATABASE_MAX_OPEN_CONNS"`
	MaxIdleConns    int    `yaml:"max_idle_conns" env:"DATABASE_MAX_IDLE_CONNS"`
	ConnMaxLifetime string `yaml:"conn_max_lifetime" env:"DATABASE_CONN_MAX_LIFETIME"`
}

type JWTKeyConfig struct {
	KID    string `yaml:"kid"`
	Secret string `yaml:"secret" env:"JWT_SECRET"`
}

type AuthConfig struct {
	JWTKeys []JWTKeyConfig `yaml:"jwt_keys"`
}

type CacheConfig struct {
	Addr     string `yaml:"addr" env:"CACHE_ADDR"`
	Password string `yaml:"password" env:"CACHE_PASSWORD"`
	DB       int    `yaml:"db" env:"CACHE_DB"`
}

type RateLimitRule struct {
	Limit  int    `yaml:"limit"`
	Window string `yaml:"window"`
}

type RateLimitConfig struct {
	Auth  RateLimitRule `yaml:"auth"`
	Write RateLimitRule `yaml:"write"`
	Read  RateLimitRule `yaml:"read"`
}

type TimeoutConfig struct {
	Default string            `yaml:"default"`
	Routes  map[string]string `yaml:"routes"`
}

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	Cache     CacheConfig     `yaml:"cache"`
	Logger    logger.Config   `yaml:"logger"`
	Auth      AuthConfig      `yaml:"auth"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	Timeout   TimeoutConfig   `yaml:"timeout"`
}

func defaults() Config {
	return Config{
		Server: ServerConfig{
			HTTPAddr: ":8080",
			GRPCAddr: ":9090",
		},
		Database: DatabaseConfig{
			DSN:          "postgres://cyandie:cyandie@localhost:5432/cyandie?sslmode=disable",
			MaxOpenConns: 25,
			MaxIdleConns: 5,
		},
		Cache: CacheConfig{
			Addr: "localhost:6379",
		},
		Logger: logger.Config{
			Level:  "info",
			Format: "json",
		},
		RateLimit: RateLimitConfig{
			Auth:  RateLimitRule{Limit: 10, Window: "1m"},
			Write: RateLimitRule{Limit: 30, Window: "1m"},
			Read:  RateLimitRule{Limit: 60, Window: "1m"},
		},
		Timeout: TimeoutConfig{
			Default: "30s",
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := defaults()

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse config file: %w", err)
		}
	}

	applyEnv(&cfg)

	return &cfg, nil
}

func applyEnv(cfg *Config) {
	envMap := map[*string]string{
		&cfg.Server.HTTPAddr: "HTTP_ADDR",
		&cfg.Server.GRPCAddr: "GRPC_ADDR",
		&cfg.Database.DSN:    "DATABASE_DSN",
		&cfg.Cache.Addr:      "CACHE_ADDR",
		&cfg.Cache.Password:  "CACHE_PASSWORD",
		&cfg.Logger.Level:    "LOG_LEVEL",
		&cfg.Logger.Format:   "LOG_FORMAT",
	}
	for ptr, key := range envMap {
		if v := os.Getenv(key); v != "" {
			*ptr = v
		}
	}

	// For each JWT key, allow overriding the secret via env
	for i := range cfg.Auth.JWTKeys {
		if v := os.Getenv("JWT_SECRET_" + cfg.Auth.JWTKeys[i].KID); v != "" {
			cfg.Auth.JWTKeys[i].Secret = v
		}
	}
}
