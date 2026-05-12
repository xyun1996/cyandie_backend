package errors

import (
	"encoding/json"
	"net/http"
)

type AppError struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details,omitempty"`
	RequestID string         `json:"requestId,omitempty"`
}

func New(code string, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: make(map[string]any),
	}
}

func (e *AppError) Error() string {
	return "[" + e.Code + "] " + e.Message
}

func (e *AppError) HTTPStatus() int {
	switch e.Code {
	case ErrBadRequest, ErrInvalidCredentials, ErrOTPInvalid, ErrOTPExpired:
		return http.StatusBadRequest
	case ErrUnauthorized, ErrTokenExpired, ErrTokenInvalid, ErrSessionRevoked:
		return http.StatusUnauthorized
	case ErrForbidden, ErrUserBanned:
		return http.StatusForbidden
	case ErrNotFound, ErrPlatformNotSupported:
		return http.StatusNotFound
	case ErrConflict, ErrUserExists, ErrPlatformBindingExists:
		return http.StatusConflict
	case ErrTimeout:
		return http.StatusServiceUnavailable
	case ErrRateLimited:
		return http.StatusTooManyRequests
	case ErrPlatformNotBound:
		return http.StatusNotFound
	case ErrLastCredential:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func (e *AppError) WithDetail(key string, value any) *AppError {
	e.Details[key] = value
	return e
}

func (e *AppError) WithRequestID(id string) *AppError {
	e.RequestID = id
	return e
}

func (e *AppError) WriteHTTP(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.HTTPStatus())
	type response struct {
		OK    bool     `json:"ok"`
		Error AppError `json:"error"`
	}
	resp := response{OK: false, Error: *e}
	data, _ := json.Marshal(resp)
	w.Write(data)
}
