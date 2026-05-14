package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	coreerrors "github.com/cyandie/backend/internal/core/errors"
	"github.com/cyandie/backend/internal/auth"
	"github.com/go-chi/chi/v5"
)

type AdminHandler struct {
	service *AdminService
}

func NewAdminHandler(service *AdminService) *AdminHandler {
	return &AdminHandler{service: service}
}

func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req AdminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		coreerrors.New(coreerrors.ErrBadRequest, "invalid request body").WriteHTTP(w)
		return
	}

	admin, err := h.service.Login(r.Context(), req)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, admin)
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)

	users, err := h.service.ListUsers(r.Context(), limit, offset)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, users)
}

func (h *AdminHandler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		coreerrors.New(coreerrors.ErrBadRequest, "user ID is required").WriteHTTP(w)
		return
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		coreerrors.New(coreerrors.ErrBadRequest, "invalid request body").WriteHTTP(w)
		return
	}

	user, err := h.service.UpdateUserStatus(r.Context(), userID, body.Status)
	if err != nil {
		writeAppError(w, err)
		return
	}

	operatorID := auth.UserIDFromContext(r.Context())
	_, _ = h.service.CreateAuditLog(r.Context(), operatorID, "update_user_status", "user", userID, "status: "+body.Status, r.RemoteAddr)

	writeJSON(w, http.StatusOK, user)
}

func (h *AdminHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)

	logs, err := h.service.ListAuditLogs(r.Context(), limit, offset)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, logs)
}

func parsePagination(r *http.Request) (int32, int32) {
	limit := int32(20)
	offset := int32(0)
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil && n > 0 {
			limit = int32(n)
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil && n >= 0 {
			offset = int32(n)
		}
	}
	return limit, offset
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