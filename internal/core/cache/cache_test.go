package cache

import (
	"testing"

	"github.com/cyandie/backend/internal/core/config"
)

func TestNew_InvalidAddr(t *testing.T) {
	cfg := config.CacheConfig{
		Addr: "localhost:99999",
	}
	_, err := New(cfg)
	if err == nil {
		t.Error("expected error for invalid Redis addr")
	}
}
