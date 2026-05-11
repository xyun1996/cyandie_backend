//go:build integration

package integration

import (
	"io"
	"net/http"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/healthz")
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) == "" {
		t.Error("expected non-empty body")
	}
}
