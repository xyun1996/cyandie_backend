package users

import (
	"encoding/json"
	"net/http"
	"strconv"

	coreerrors "github.com/cyandie/backend/internal/core/errors"
	"github.com/cyandie/backend/internal/auth"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *UserService
}

func NewHandler(svc *UserService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Get("/api/v1/users/me", h.getMe)
	router.Put("/api/v1/users/me", h.updateMe)
	router.Get("/api/v1/users/{id}", h.getUser)
	router.Get("/api/v1/users/search", h.searchUsers)
}

func (h *Handler) getMe(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	user, err := h.svc.GetUser(r.Context(), userID)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) updateMe(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		coreerrors.New(coreerrors.ErrBadRequest, "invalid request body").WriteHTTP(w)
		return
	}
	if err := h.svc.UpdateProfile(r.Context(), userID, req); err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := h.svc.GetUser(r.Context(), id)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) searchUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	users, err := h.svc.SearchUsers(r.Context(), q, Pagination{Limit: limit, Offset: offset})
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": data})
}

func writeAppError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*coreerrors.AppError); ok {
		appErr.WriteHTTP(w)
		return
	}
	coreerrors.New(coreerrors.ErrInternal, "internal error").WriteHTTP(w)
}
