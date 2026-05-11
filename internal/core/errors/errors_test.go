package errors

import (
	"net/http"
	"testing"
)

func TestNewAppError(t *testing.T) {
	err := New(ErrNotFound, "user not found")
	if err.Code != ErrNotFound {
		t.Errorf("expected code %s, got %s", ErrNotFound, err.Code)
	}
	if err.Message != "user not found" {
		t.Errorf("expected message 'user not found', got '%s'", err.Message)
	}
}

func TestAppError_HTTPStatus(t *testing.T) {
	tests := []struct {
		code     string
		expected int
	}{
		{ErrBadRequest, http.StatusBadRequest},
		{ErrUnauthorized, http.StatusUnauthorized},
		{ErrForbidden, http.StatusForbidden},
		{ErrNotFound, http.StatusNotFound},
		{ErrConflict, http.StatusConflict},
		{ErrInternal, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		err := New(tt.code, "test")
		if err.HTTPStatus() != tt.expected {
			t.Errorf("code %s: expected %d, got %d", tt.code, tt.expected, err.HTTPStatus())
		}
	}
}

func TestAppError_Error(t *testing.T) {
	err := New(ErrNotFound, "user not found")
	if err.Error() != "[NOT_FOUND] user not found" {
		t.Errorf("unexpected error string: %s", err.Error())
	}
}

func TestAppError_WithDetail(t *testing.T) {
	err := New(ErrBadRequest, "validation failed").WithDetail("field", "email")
	if err.Details["field"] != "email" {
		t.Errorf("expected detail field=email, got %v", err.Details["field"])
	}
}

func TestAppError_WithRequestID(t *testing.T) {
	err := New(ErrInternal, "oops").WithRequestID("req_123")
	if err.RequestID != "req_123" {
		t.Errorf("expected requestID req_123, got %s", err.RequestID)
	}
}
