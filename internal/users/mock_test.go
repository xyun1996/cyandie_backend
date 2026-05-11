package users

import (
	"context"
	"database/sql"

	"github.com/cyandie/backend/internal/db"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type mockUserQueries struct {
	user db.User
}

func (m *mockUserQueries) GetUserByID(_ context.Context, id uuid.UUID) (db.User, error) {
	m.user.ID = id
	m.user.Username = "testuser"
	m.user.Status = "active"
	m.user.Metadata = pqtype.NullRawMessage{Valid: true, RawMessage: []byte(`{}`)}
	return m.user, nil
}
func (m *mockUserQueries) GetUserByUsername(_ context.Context, _ string) (db.User, error) {
	return m.user, nil
}
func (m *mockUserQueries) CreateUser(_ context.Context, _ db.CreateUserParams) (db.User, error) {
	return db.User{}, nil
}
func (m *mockUserQueries) SearchUsers(_ context.Context, _ db.SearchUsersParams) ([]db.User, error) {
	return nil, nil
}
func (m *mockUserQueries) UpdateUserProfile(_ context.Context, _ db.UpdateUserProfileParams) (db.User, error) {
	return db.User{}, nil
}
func (m *mockUserQueries) UpdateUserStatus(_ context.Context, _ db.UpdateUserStatusParams) (db.User, error) {
	return db.User{}, nil
}
func (m *mockUserQueries) GetCredentialByTypeIdentifier(_ context.Context, _ db.GetCredentialByTypeIdentifierParams) (db.Credential, error) {
	return db.Credential{}, sql.ErrNoRows
}
func (m *mockUserQueries) CreateCredential(_ context.Context, _ db.CreateCredentialParams) (db.Credential, error) {
	return db.Credential{}, nil
}
func (m *mockUserQueries) GetCredentialsByUserID(_ context.Context, _ uuid.UUID) ([]db.Credential, error) {
	return nil, nil
}
func (m *mockUserQueries) CreateSession(_ context.Context, _ db.CreateSessionParams) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (m *mockUserQueries) GetSessionByID(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, sql.ErrNoRows
}
func (m *mockUserQueries) RevokeSession(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (m *mockUserQueries) RevokeSessionsByUserID(_ context.Context, _ uuid.UUID) ([]db.UserSession, error) {
	return nil, nil
}
