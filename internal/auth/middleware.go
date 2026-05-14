package auth

import (
	"context"
	"net/http"
	"strings"

	coreerrors "github.com/cyandie/backend/internal/core/errors"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	SessionIDKey contextKey = "session_id"
	RoleKey      contextKey = "role"
)

func AuthGuard(svc *AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				coreerrors.New(coreerrors.ErrUnauthorized, "missing authorization header").WriteHTTP(w)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				coreerrors.New(coreerrors.ErrUnauthorized, "invalid authorization format").WriteHTTP(w)
				return
			}

			claims, err := svc.ValidateToken(r.Context(), parts[1])
			if err != nil {
				coreerrors.New(coreerrors.ErrTokenInvalid, "invalid token").WriteHTTP(w)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, SessionIDKey, claims.SessionID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) string {
	val, _ := ctx.Value(UserIDKey).(string)
	return val
}

func SessionIDFromContext(ctx context.Context) string {
	val, _ := ctx.Value(SessionIDKey).(string)
	return val
}

func RoleFromContext(ctx context.Context) string {
	val, _ := ctx.Value(RoleKey).(string)
	return val
}

func RequireAdmin() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := r.Context().Value(RoleKey).(string)
			if role != "admin" {
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
