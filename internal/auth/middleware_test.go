package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthGuard_ValidToken(t *testing.T) {
	svc := newTestService()
	// Register and login to get a valid token
	svc.Register(context.Background(), RegisterRequest{Username: "testuser", Password: "pass123"})
	tokens, _ := svc.Login(context.Background(), LoginRequest{Username: "testuser", Password: "pass123"})

	r := chi.NewRouter()
	r.Use(AuthGuard(svc))
	r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestAuthGuard_NoToken(t *testing.T) {
	svc := newTestService()

	r := chi.NewRouter()
	r.Use(AuthGuard(svc))
	r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuthGuard_InvalidToken(t *testing.T) {
	svc := newTestService()

	r := chi.NewRouter()
	r.Use(AuthGuard(svc))
	r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAdmin_RejectsNonAdmin(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := RequireAdmin()
	next := mw(handler)

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	ctx := context.WithValue(req.Context(), RoleKey, "user")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	next.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireAdmin_AllowsAdmin(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := RequireAdmin()
	next := mw(handler)

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	ctx := context.WithValue(req.Context(), RoleKey, "admin")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	next.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
