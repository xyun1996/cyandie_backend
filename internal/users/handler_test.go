package users

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cyandie/backend/internal/auth"
	"github.com/go-chi/chi/v5"
)

func TestHandler_GetMe(t *testing.T) {
	svc := NewUserService(&mockUserQueries{})
	h := NewHandler(svc)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), auth.UserIDKey, "550e8400-e29b-41d4-a716-446655440000")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	h.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
