package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestKeyManager_SignAndVerify(t *testing.T) {
	km := NewKeyManager([]JWTKey{
		{KID: "key-1", Secret: []byte("secret-one-1234567890123456")},
	})

	claims := &Claims{
		UserID:    "user-123",
		SessionID: "sess-456",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token, err := km.Sign(claims)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	parsed, err := km.Verify(token)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if parsed.UserID != "user-123" {
		t.Errorf("expected user-123, got %s", parsed.UserID)
	}
}

func TestKeyManager_Rotation(t *testing.T) {
	km := NewKeyManager([]JWTKey{
		{KID: "key-2", Secret: []byte("secret-two-1234567890123456")},
		{KID: "key-1", Secret: []byte("secret-one-1234567890123456")},
	})

	token, _ := km.Sign(&Claims{UserID: "user-1", RegisteredClaims: jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
	}})

	parsed, err := km.Verify(token)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if parsed.UserID != "user-1" {
		t.Errorf("expected user-1, got %s", parsed.UserID)
	}
}

func TestKeyManager_ExpiredToken(t *testing.T) {
	km := NewKeyManager([]JWTKey{
		{KID: "key-1", Secret: []byte("secret-one-1234567890123456")},
	})

	token, _ := km.Sign(&Claims{UserID: "user-1", RegisteredClaims: jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
	}})

	_, err := km.Verify(token)
	if err == nil {
		t.Error("expected error for expired token")
	}
}
