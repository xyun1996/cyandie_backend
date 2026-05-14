package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTKey struct {
	KID    string `yaml:"kid"`
	Secret []byte `yaml:"secret"`
}

type Claims struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	Role      string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

type KeyManager struct {
	keys []JWTKey
}

func NewKeyManager(keys []JWTKey) *KeyManager {
	return &KeyManager{keys: keys}
}

func (km *KeyManager) Sign(claims *Claims) (string, error) {
	if len(km.keys) == 0 {
		return "", fmt.Errorf("no JWT keys configured")
	}
	key := km.keys[0]
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = key.KID
	return token.SignedString(key.Secret)
}

func (km *KeyManager) Verify(tokenStr string) (*Claims, error) {
	var lastErr error
	for _, key := range km.keys {
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			return key.Secret, nil
		})
		if err != nil {
			lastErr = err
			continue
		}
		if token.Valid {
			return claims, nil
		}
	}
	return nil, fmt.Errorf("token verification failed: %w", lastErr)
}

func (km *KeyManager) IsEmpty() bool {
	return len(km.keys) == 0
}

func (km *KeyManager) GenerateAccessToken(userID, sessionID, role string) (string, error) {
	claims := &Claims{
		UserID:    userID,
		SessionID: sessionID,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return km.Sign(claims)
}
