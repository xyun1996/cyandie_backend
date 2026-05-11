package users

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cyandie/backend/internal/core/errors"
	"github.com/cyandie/backend/internal/db"
	"github.com/google/uuid"
)

type User struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	AvatarURL   string `json:"avatarUrl,omitempty"`
	Status      string `json:"status"`
}

type UpdateProfileRequest struct {
	DisplayName *string `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl"`
}

type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type UserService struct {
	queries db.Querier
}

func NewUserService(queries db.Querier) *UserService {
	return &UserService{queries: queries}
}

func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New(errors.ErrBadRequest, "invalid user id")
	}
	u, err := s.queries.GetUserByID(ctx, uid)
	if err != nil {
		return nil, errors.New(errors.ErrNotFound, "user not found")
	}
	return toUser(u), nil
}

func (s *UserService) UpdateProfile(ctx context.Context, id string, req UpdateProfileRequest) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return errors.New(errors.ErrBadRequest, "invalid user id")
	}
	_, err = s.queries.UpdateUserProfile(ctx, db.UpdateUserProfileParams{
		ID:          uid,
		DisplayName: sql.NullString{String: stringValue(req.DisplayName), Valid: req.DisplayName != nil},
		AvatarUrl:   sql.NullString{String: stringValue(req.AvatarURL), Valid: req.AvatarURL != nil},
	})
	if err != nil {
		return fmt.Errorf("update profile: %w", err)
	}
	return nil
}

func (s *UserService) SearchUsers(ctx context.Context, query string, page Pagination) ([]*User, error) {
	users, err := s.queries.SearchUsers(ctx, db.SearchUsersParams{
		Column1: query,
		Limit:   int32(page.Limit),
		Offset:  int32(page.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("search users: %w", err)
	}
	result := make([]*User, len(users))
	for i, u := range users {
		result[i] = toUser(u)
	}
	return result, nil
}

func toUser(u db.User) *User {
	return &User{
		ID:          u.ID.String(),
		Username:    u.Username,
		Email:       u.Email.String,
		DisplayName: u.DisplayName.String,
		AvatarURL:   u.AvatarUrl.String,
		Status:      u.Status,
	}
}

func stringValue(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
