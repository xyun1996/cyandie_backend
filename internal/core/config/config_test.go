package config

import (
	"os"
	"testing"
)

func TestLoadFromYAML(t *testing.T) {
	yamlContent := `
server:
  http_addr: ":9090"
database:
  dsn: "postgres://test:test@localhost:5432/test?sslmode=disable"
cache:
  addr: "localhost:6379"
logger:
  level: "debug"
  format: "text"
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(yamlContent)
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Server.HTTPAddr != ":9090" {
		t.Errorf("expected :9090, got %s", cfg.Server.HTTPAddr)
	}
	if cfg.Database.DSN != "postgres://test:test@localhost:5432/test?sslmode=disable" {
		t.Errorf("unexpected database DSN: %s", cfg.Database.DSN)
	}
	if cfg.Cache.Addr != "localhost:6379" {
		t.Errorf("expected localhost:6379, got %s", cfg.Cache.Addr)
	}
	if cfg.Logger.Level != "debug" {
		t.Errorf("expected debug, got %s", cfg.Logger.Level)
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load with no file should use defaults: %v", err)
	}
	if cfg.Server.HTTPAddr != ":8080" {
		t.Errorf("expected default :8080, got %s", cfg.Server.HTTPAddr)
	}
	if cfg.Logger.Level != "info" {
		t.Errorf("expected default info, got %s", cfg.Logger.Level)
	}
}

func TestEnvOverride(t *testing.T) {
	os.Setenv("HTTP_ADDR", ":7070")
	defer os.Unsetenv("HTTP_ADDR")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Server.HTTPAddr != ":7070" {
		t.Errorf("expected :7070 from env, got %s", cfg.Server.HTTPAddr)
	}
}
