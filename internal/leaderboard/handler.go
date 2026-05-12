package leaderboard

import (
	"encoding/json"
	"net/http"
	"strconv"

	coreerrors "github.com/cyandie/backend/internal/core/errors"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *LeaderboardService
}

func NewHandler(svc *LeaderboardService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Get("/api/v1/leaderboard/{code}", h.getRanking)
	router.Post("/api/v1/leaderboard/{code}/submit", h.submitScore)
	router.Get("/api/v1/leaderboard/{code}/me", h.getMyRank)
}

func (h *Handler) getRanking(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	entries, err := h.svc.GetRanking(r.Context(), code, limit, offset)
	if err != nil {
		writeLBError(w, err)
		return
	}
	writeLBJSON(w, http.StatusOK, entries)
}

func (h *Handler) submitScore(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		coreerrors.New(coreerrors.ErrUnauthorized, "authentication required").WriteHTTP(w)
		return
	}

	code := chi.URLParam(r, "code")
	var req struct {
		Score float64 `json:"score"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		coreerrors.New(coreerrors.ErrBadRequest, "invalid request body").WriteHTTP(w)
		return
	}

	if err := h.svc.SubmitScore(r.Context(), code, userID, req.Score); err != nil {
		writeLBError(w, err)
		return
	}
	writeLBJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) getMyRank(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		coreerrors.New(coreerrors.ErrUnauthorized, "authentication required").WriteHTTP(w)
		return
	}

	code := chi.URLParam(r, "code")
	entry, err := h.svc.GetUserRank(r.Context(), code, userID)
	if err != nil {
		writeLBError(w, err)
		return
	}
	writeLBJSON(w, http.StatusOK, entry)
}

func writeLBJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": data})
}

func writeLBError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*coreerrors.AppError); ok {
		appErr.WriteHTTP(w)
		return
	}
	coreerrors.New(coreerrors.ErrInternal, "internal error").WriteHTTP(w)
}
