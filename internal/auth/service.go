package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/cyandie/backend/internal/core/errors"
	"github.com/cyandie/backend/internal/db"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type AuthServiceDeps struct {
	Queries     db.Querier
	KeyManager  *KeyManager
	Sessions    *SessionStore
	OTPNotifier OTPNotifier
}

type AuthService struct {
	queries    db.Querier
	keyManager *KeyManager
	sessions   *SessionStore
	otp        OTPNotifier
}

func NewAuthService(deps AuthServiceDeps) *AuthService {
	return &AuthService{
		queries:    deps.Queries,
		keyManager: deps.KeyManager,
		sessions:   deps.Sessions,
		otp:        deps.OTPNotifier,
	}
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (string, error) {
	if req.Username == "" || req.Password == "" {
		return "", errors.New(errors.ErrBadRequest, "username and password are required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Username:    req.Username,
		Email:       sql.NullString{String: req.Email, Valid: req.Email != ""},
		DisplayName: sql.NullString{String: req.Username, Valid: true},
	})
	if err != nil {
		return "", errors.New(errors.ErrConflict, "username or email already exists")
	}

	_, err = s.queries.CreateCredential(ctx, db.CreateCredentialParams{
		UserID:     user.ID,
		Type:       "password",
		Identifier: req.Username,
		SecretHash: sql.NullString{String: string(hash), Valid: true},
		Verified:   true,
	})
	if err != nil {
		return "", fmt.Errorf("create credential: %w", err)
	}

	return user.ID.String(), nil
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*TokenPair, error) {
	if req.Username == "" || req.Password == "" {
		return nil, errors.New(errors.ErrInvalidCredentials, "username and password are required")
	}

	cred, err := s.queries.GetCredentialByTypeIdentifier(ctx, db.GetCredentialByTypeIdentifierParams{
		Type:       "password",
		Identifier: req.Username,
	})
	if err != nil {
		return nil, errors.New(errors.ErrInvalidCredentials, "invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(cred.SecretHash.String), []byte(req.Password)); err != nil {
		return nil, errors.New(errors.ErrInvalidCredentials, "invalid credentials")
	}

	user, err := s.queries.GetUserByID(ctx, cred.UserID)
	if err != nil {
		return nil, errors.New(errors.ErrInternal, "user not found")
	}

	if user.Status != "active" {
		return nil, errors.New(errors.ErrUserBanned, "account is not active")
	}

	return s.generateTokenPair(ctx, user.ID.String())
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	session, err := s.sessions.Get(refreshToken)
	if err != nil {
		return nil, errors.New(errors.ErrSessionRevoked, "invalid or expired refresh token")
	}

	if err := s.sessions.Revoke(session.ID); err != nil {
		return nil, errors.New(errors.ErrInternal, "failed to revoke old session")
	}

	return s.generateTokenPair(ctx, session.UserID)
}

func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	return s.sessions.Revoke(sessionID)
}

func (s *AuthService) ValidateToken(ctx context.Context, accessToken string) (*Claims, error) {
	claims, err := s.keyManager.Verify(accessToken)
	if err != nil {
		return nil, errors.New(errors.ErrTokenInvalid, "invalid token")
	}
	return claims, nil
}

// GenerateToken generates a JWT for a given user without creating a session.
func (s *AuthService) GenerateToken(userID string) (string, error) {
	return s.keyManager.GenerateAccessToken(userID, "")
}

func (s *AuthService) generateTokenPair(ctx context.Context, userID string) (*TokenPair, error) {
	sessionID := generateSessionID()
	accessToken, err := s.keyManager.GenerateAccessToken(userID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken := generateRefreshToken()
	refreshHash := hashToken(refreshToken)

	_, err = s.sessions.Create(userID, refreshHash, 7*24*60*60)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func generateRefreshToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return "ref_" + hex.EncodeToString(b)
}
