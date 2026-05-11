//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestAuthRegisterAndLogin(t *testing.T) {
	// Register
	body, _ := json.Marshal(map[string]string{
		"username": "inttestuser",
		"password": "testpass123",
	})
	resp, err := http.Post("http://localhost:8080/api/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, string(b))
	}

	// Login
	body, _ = json.Marshal(map[string]string{
		"username": "inttestuser",
		"password": "testpass123",
	})
	resp, err = http.Post("http://localhost:8080/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(b))
	}

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	data, _ := result["data"].(map[string]any)
	if data["accessToken"] == nil {
		t.Error("expected accessToken in login response")
	}
}

func TestGetMe(t *testing.T) {
	// First login to get token
	body, _ := json.Marshal(map[string]string{
		"username": "inttestuser",
		"password": "testpass123",
	})
	resp, _ := http.Post("http://localhost:8080/api/v1/auth/login", "application/json", bytes.NewReader(body))
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	data, _ := result["data"].(map[string]any)
	token, _ := data["accessToken"].(string)

	if token == "" {
		t.Skip("no access token, skipping GetMe test")
	}

	// Get /me
	req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get me failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
