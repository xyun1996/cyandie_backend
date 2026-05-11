package auth

import (
	"context"
	"testing"
)

func newTestService() *AuthService {
	return NewAuthService(AuthServiceDeps{
		Queries:     newMockQueries(),
		KeyManager:  NewKeyManager([]JWTKey{{KID: "k1", Secret: []byte("test-secret-key-1234567890123456")}}),
		Sessions:    NewSessionStore(newMockRedisClient()),
		OTPNotifier: LogNotifier{},
	})
}

func TestAuthService_Register(t *testing.T) {
	svc := newTestService()
	userID, err := svc.Register(context.Background(), RegisterRequest{
		Username: "testuser",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if userID == "" {
		t.Error("expected non-empty user ID")
	}
}

func TestAuthService_Register_DuplicateUsername(t *testing.T) {
	svc := newTestService()
	svc.Register(context.Background(), RegisterRequest{Username: "testuser", Password: "password123"})
	_, err := svc.Register(context.Background(), RegisterRequest{Username: "testuser", Password: "password456"})
	if err == nil {
		t.Error("expected error for duplicate username")
	}
}

func TestAuthService_Login(t *testing.T) {
	svc := newTestService()
	svc.Register(context.Background(), RegisterRequest{Username: "testuser", Password: "password123"})
	tokens, err := svc.Login(context.Background(), LoginRequest{Username: "testuser", Password: "password123"})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if tokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if tokens.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	svc := newTestService()
	svc.Register(context.Background(), RegisterRequest{Username: "testuser", Password: "password123"})
	_, err := svc.Login(context.Background(), LoginRequest{Username: "testuser", Password: "wrong"})
	if err == nil {
		t.Error("expected error for wrong password")
	}
}
