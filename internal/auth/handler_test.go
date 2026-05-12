package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestHandler_Register(t *testing.T) {
	r := chi.NewRouter()
	svc := newTestService()
	h := NewHandler(svc)
	h.RegisterRoutes(r)

	body, _ := json.Marshal(RegisterRequest{Username: "newuser", Password: "pass123"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_Login(t *testing.T) {
	r := chi.NewRouter()
	svc := newTestService()
	h := NewHandler(svc)
	h.RegisterRoutes(r)

	// Register first
	body, _ := json.Marshal(RegisterRequest{Username: "loginuser", Password: "pass123"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Login
	body, _ = json.Marshal(LoginRequest{Username: "loginuser", Password: "pass123"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data, _ := resp["data"].(map[string]any)
	if data["accessToken"] == nil {
		t.Error("expected accessToken in response")
	}
}

func TestHandler_Login_InvalidBody(t *testing.T) {
	r := chi.NewRouter()
	svc := newTestService()
	h := NewHandler(svc)
	h.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
