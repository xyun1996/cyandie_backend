package platforms

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cyandie/backend/internal/core/errors"
	"github.com/cyandie/backend/internal/db"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type PlatformService struct {
	queries  db.Querier
	registry *PlatformRegistry
}

func NewPlatformService(queries db.Querier, registry *PlatformRegistry) *PlatformService {
	return &PlatformService{queries: queries, registry: registry}
}

func (s *PlatformService) GetAuthURL(platformName, state string) (string, error) {
	provider, ok := s.registry.GetOAuth(platformName)
	if !ok {
		return "", errors.New(errors.ErrPlatformNotSupported, "platform not supported: "+platformName)
	}
	return provider.GetAuthURL(state), nil
}

func (s *PlatformService) HandleCallback(ctx context.Context, platformName, code string) (*CallbackResult, error) {
	provider, ok := s.registry.GetOAuth(platformName)
	if !ok {
		return nil, errors.New(errors.ErrPlatformNotSupported, "platform not supported: "+platformName)
	}

	token, err := provider.ExchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}

	platformUser, err := provider.GetUserInfo(ctx, token)
	if err != nil {
		return nil, err
	}

	// Check if this platform identity is already bound
	binding, err := s.queries.GetPlatformBinding(ctx, db.GetPlatformBindingParams{
		Platform:       platformName,
		PlatformUserID: platformUser.PlatformID,
	})
	if err == nil {
		// Existing user — return user ID
		return &CallbackResult{
			UserID:    binding.UserID.String(),
			IsNewUser: false,
		}, nil
	}

	// New user — create user + binding
	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Username:    fmt.Sprintf("wx_%s", platformUser.PlatformID[:8]),
		DisplayName: sql.NullString{String: platformUser.DisplayName, Valid: true},
	})
	if err != nil {
		return nil, errors.New(errors.ErrInternal, "failed to create user")
	}

	_, err = s.queries.CreatePlatformBinding(ctx, db.CreatePlatformBindingParams{
		UserID:         user.ID,
		Platform:       platformName,
		PlatformUserID: platformUser.PlatformID,
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
		ExpiresAt:      sql.NullTime{Time: token.ExpiresAt, Valid: !token.ExpiresAt.IsZero()},
		Metadata:       pqtype.NullRawMessage{RawMessage: []byte(`{}`), Valid: true},
	})
	if err != nil {
		return nil, errors.New(errors.ErrInternal, "failed to create platform binding")
	}

	return &CallbackResult{
		UserID:    user.ID.String(),
		IsNewUser: true,
	}, nil
}

func (s *PlatformService) BindPlatform(ctx context.Context, userID, platformName, code string) error {
	provider, ok := s.registry.GetOAuth(platformName)
	if !ok {
		return errors.New(errors.ErrPlatformNotSupported, "platform not supported: "+platformName)
	}

	token, err := provider.ExchangeCode(ctx, code)
	if err != nil {
		return err
	}

	platformUser, err := provider.GetUserInfo(ctx, token)
	if err != nil {
		return err
	}

	uid, _ := uuid.Parse(userID)

	// Check if already bound to another user
	existing, err := s.queries.GetPlatformBinding(ctx, db.GetPlatformBindingParams{
		Platform:       platformName,
		PlatformUserID: platformUser.PlatformID,
	})
	if err == nil && existing.UserID != uid {
		return errors.New(errors.ErrPlatformBindingExists, "platform identity already bound to another user")
	}

	_, err = s.queries.CreatePlatformBinding(ctx, db.CreatePlatformBindingParams{
		UserID:         uid,
		Platform:       platformName,
		PlatformUserID: platformUser.PlatformID,
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
		ExpiresAt:      sql.NullTime{Time: token.ExpiresAt, Valid: !token.ExpiresAt.IsZero()},
		Metadata:       pqtype.NullRawMessage{RawMessage: []byte(`{}`), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("create binding: %w", err)
	}

	return nil
}

func (s *PlatformService) UnbindPlatform(ctx context.Context, userID, platformName string) error {
	uid, _ := uuid.Parse(userID)

	bindings, err := s.queries.GetPlatformBindingsByUserID(ctx, uid)
	if err != nil {
		return errors.New(errors.ErrPlatformNotBound, "no platform bindings found")
	}

	found := false
	for _, b := range bindings {
		if b.Platform == platformName {
			found = true
			break
		}
	}
	if !found {
		return errors.New(errors.ErrPlatformNotBound, "platform not bound")
	}

	// Check if user would have any way to log in after unbinding
	credCount, _ := s.queries.CountUserCredentials(ctx, uid)
	otherBindings := 0
	for _, b := range bindings {
		if b.Platform != platformName {
			otherBindings++
		}
	}
	if credCount == 0 && otherBindings == 0 {
		return errors.New(errors.ErrLastCredential, "cannot unbind last credential, set a password first")
	}

	_, err = s.queries.DeletePlatformBinding(ctx, db.DeletePlatformBindingParams{
		UserID:   uid,
		Platform: platformName,
	})
	if err != nil {
		return fmt.Errorf("delete binding: %w", err)
	}

	return nil
}

func (s *PlatformService) ListBindings(ctx context.Context, userID string) ([]db.PlatformBinding, error) {
	uid, _ := uuid.Parse(userID)
	return s.queries.GetPlatformBindingsByUserID(ctx, uid)
}

type CallbackResult struct {
	UserID    string `json:"userId"`
	IsNewUser bool   `json:"isNewUser"`
}
