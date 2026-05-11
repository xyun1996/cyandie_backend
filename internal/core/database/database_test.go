package database

import (
	"testing"

	"github.com/cyandie/backend/internal/core/config"
)

func TestNew_InvalidDSN(t *testing.T) {
	cfg := config.DatabaseConfig{
		DSN: "postgres://invalid:invalid@localhost:99999/nonexistent",
	}
	_, err := New(cfg)
	if err == nil {
		t.Error("expected error for invalid DSN")
	}
}
