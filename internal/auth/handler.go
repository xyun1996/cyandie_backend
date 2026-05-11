package auth

import (
	"encoding/json"
	"net/http"

	coreerrors "github.com/cyandie/backend/internal/core/errors"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *AuthService
}

func NewHandler(svc *AuthService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Post("/api/v1/auth/register", h.register)
	router.Post("/api/v1/auth/login", h.login)
	router.Post("/api/v1/auth/refresh", h.refresh)
	router.Post("/api/v1/auth/logout", h.logout)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		coreerrors.New(coreerrors.ErrBadRequest, "invalid request body").WriteHTTP(w)
		return
	}

	userID, err := h.svc.Register(r.Context(), req)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"userId": userID})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		coreerrors.New(coreerrors.ErrBadRequest, "invalid request body").WriteHTTP(w)
		return
	}

	tokens, err := h.svc.Login(r.Context(), req)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, tokens)
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		coreerrors.New(coreerrors.ErrBadRequest, "invalid request body").WriteHTTP(w)
		return
	}

	tokens, err := h.svc.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, tokens)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		coreerrors.New(coreerrors.ErrBadRequest, "missing session id").WriteHTTP(w)
		return
	}

	if err := h.svc.Logout(r.Context(), sessionID); err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": data})
}

func writeAppError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*coreerrors.AppError); ok {
		appErr.WithRequestID(w.Header().Get("X-Request-ID")).WriteHTTP(w)
		return
	}
	coreerrors.New(coreerrors.ErrInternal, "internal error").WriteHTTP(w)
}
