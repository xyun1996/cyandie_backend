package auth

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cyandie/backend/internal/db"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type mockQueries struct {
	users       map[string]db.User
	credentials map[string]db.Credential
	passwords   map[string]string // username -> plaintext password for testing
}

func newMockQueries() *mockQueries {
	return &mockQueries{
		users:       make(map[string]db.User),
		credentials: make(map[string]db.Credential),
		passwords:   make(map[string]string),
	}
}

func (m *mockQueries) GetUserByID(_ context.Context, id uuid.UUID) (db.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return db.User{}, sql.ErrNoRows
}

func (m *mockQueries) GetUserByUsername(_ context.Context, username string) (db.User, error) {
	if u, ok := m.users[username]; ok {
		return u, nil
	}
	return db.User{}, sql.ErrNoRows
}

func (m *mockQueries) CreateUser(_ context.Context, params db.CreateUserParams) (db.User, error) {
	if _, exists := m.users[params.Username]; exists {
		return db.User{}, fmt.Errorf("duplicate username")
	}
	user := db.User{
		ID:          uuid.New(),
		Username:    params.Username,
		Email:       params.Email,
		DisplayName: params.DisplayName,
		Status:      "active",
		Metadata:    pqtype.NullRawMessage{Valid: true, RawMessage: []byte(`{}`)},
	}
	m.users[params.Username] = user
	return user, nil
}

func (m *mockQueries) GetCredentialByTypeIdentifier(_ context.Context, params db.GetCredentialByTypeIdentifierParams) (db.Credential, error) {
	key := params.Type + ":" + params.Identifier
	if c, ok := m.credentials[key]; ok {
		return c, nil
	}
	return db.Credential{}, sql.ErrNoRows
}

func (m *mockQueries) CreateCredential(_ context.Context, params db.CreateCredentialParams) (db.Credential, error) {
	key := params.Type + ":" + params.Identifier
	cred := db.Credential{
		ID:         uuid.New(),
		UserID:     params.UserID,
		Type:       params.Type,
		Identifier: params.Identifier,
		SecretHash: params.SecretHash,
		Verified:   params.Verified,
	}
	m.credentials[key] = cred
	return cred, nil
}

func (m *mockQueries) SearchUsers(_ context.Context, _ db.SearchUsersParams) ([]db.User, error) {
	return nil, nil
}
func (m *mockQueries) UpdateUserProfile(_ context.Context, _ db.UpdateUserProfileParams) (db.User, error) {
	return db.User{}, nil
}
func (m *mockQueries) UpdateUserStatus(_ context.Context, _ db.UpdateUserStatusParams) (db.User, error) {
	return db.User{}, nil
}
func (m *mockQueries) GetCredentialsByUserID(_ context.Context, _ uuid.UUID) ([]db.Credential, error) {
	return nil, nil
}
func (m *mockQueries) CreateSession(_ context.Context, _ db.CreateSessionParams) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (m *mockQueries) GetSessionByID(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, sql.ErrNoRows
}
func (m *mockQueries) RevokeSession(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (m *mockQueries) RevokeSessionsByUserID(_ context.Context, _ uuid.UUID) ([]db.UserSession, error) {
	return nil, nil
}
