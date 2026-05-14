package leaderboard

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestHandler_GetRanking(t *testing.T) {
	svc := newTestLBService()
	h := NewHandler(svc)
	r := chi.NewRouter()
	h.RegisterPublicRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/test-board?limit=10", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_SubmitScore(t *testing.T) {
	svc := newTestLBService()
	h := NewHandler(svc)
	r := chi.NewRouter()
	h.RegisterProtectedRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/test-board/submit", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Should get 401 since no auth context
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}
