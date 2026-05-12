package platforms

import (
	"context"
	"database/sql"
	"testing"

	"github.com/cyandie/backend/internal/db"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type platformMockQueries struct {
	bindings map[string]db.PlatformBinding
	users    map[string]db.User
}

func newPlatformMockQueries() *platformMockQueries {
	return &platformMockQueries{
		bindings: make(map[string]db.PlatformBinding),
		users:    make(map[string]db.User),
	}
}

func (m *platformMockQueries) GetPlatformBinding(_ context.Context, _ db.GetPlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, sql.ErrNoRows
}
func (m *platformMockQueries) GetPlatformBindingsByUserID(_ context.Context, _ uuid.UUID) ([]db.PlatformBinding, error) {
	return nil, nil
}
func (m *platformMockQueries) CreatePlatformBinding(_ context.Context, params db.CreatePlatformBindingParams) (db.PlatformBinding, error) {
	b := db.PlatformBinding{
		ID:             uuid.New(),
		UserID:         params.UserID,
		Platform:       params.Platform,
		PlatformUserID: params.PlatformUserID,
		Metadata:       pqtype.NullRawMessage{RawMessage: []byte(`{}`), Valid: true},
	}
	m.bindings[params.Platform+":"+params.PlatformUserID] = b
	return b, nil
}
func (m *platformMockQueries) DeletePlatformBinding(_ context.Context, _ db.DeletePlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, nil
}
func (m *platformMockQueries) CountUserCredentials(_ context.Context, _ uuid.UUID) (int64, error) {
	return 1, nil
}
func (m *platformMockQueries) CreateUser(_ context.Context, params db.CreateUserParams) (db.User, error) {
	u := db.User{ID: uuid.New(), Username: params.Username, Status: "active", Metadata: pqtype.NullRawMessage{RawMessage: []byte(`{}`), Valid: true}}
	m.users[params.Username] = u
	return u, nil
}
func (m *platformMockQueries) GetUserByID(_ context.Context, _ uuid.UUID) (db.User, error) {
	return db.User{}, sql.ErrNoRows
}
func (m *platformMockQueries) GetUserByUsername(_ context.Context, _ string) (db.User, error) {
	return db.User{}, sql.ErrNoRows
}
func (m *platformMockQueries) SearchUsers(_ context.Context, _ db.SearchUsersParams) ([]db.User, error) {
	return nil, nil
}
func (m *platformMockQueries) UpdateUserProfile(_ context.Context, _ db.UpdateUserProfileParams) (db.User, error) {
	return db.User{}, nil
}
func (m *platformMockQueries) UpdateUserStatus(_ context.Context, _ db.UpdateUserStatusParams) (db.User, error) {
	return db.User{}, nil
}
func (m *platformMockQueries) GetCredentialByTypeIdentifier(_ context.Context, _ db.GetCredentialByTypeIdentifierParams) (db.Credential, error) {
	return db.Credential{}, sql.ErrNoRows
}
func (m *platformMockQueries) CreateCredential(_ context.Context, _ db.CreateCredentialParams) (db.Credential, error) {
	return db.Credential{}, nil
}
func (m *platformMockQueries) GetCredentialsByUserID(_ context.Context, _ uuid.UUID) ([]db.Credential, error) {
	return nil, nil
}
func (m *platformMockQueries) CreateSession(_ context.Context, _ db.CreateSessionParams) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (m *platformMockQueries) GetSessionByID(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, sql.ErrNoRows
}
func (m *platformMockQueries) RevokeSession(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (m *platformMockQueries) RevokeSessionsByUserID(_ context.Context, _ uuid.UUID) ([]db.UserSession, error) {
	return nil, nil
}

func newTestPlatformService() *PlatformService {
	reg := NewPlatformRegistry()
	reg.RegisterOAuth(&mockProvider{name: "wechat"})
	return NewPlatformService(newPlatformMockQueries(), reg)
}

func TestPlatformService_GetAuthURL(t *testing.T) {
	svc := newTestPlatformService()
	url, err := svc.GetAuthURL("wechat", "state123")
	if err != nil {
		t.Fatalf("GetAuthURL failed: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty auth URL")
	}
}

func TestPlatformService_GetAuthURL_UnsupportedPlatform(t *testing.T) {
	svc := newTestPlatformService()
	_, err := svc.GetAuthURL("google", "state123")
	if err == nil {
		t.Error("expected error for unsupported platform")
	}
}

func TestPlatformService_UnbindPlatform_NotBound(t *testing.T) {
	svc := newTestPlatformService()
	err := svc.UnbindPlatform(context.Background(), uuid.New().String(), "wechat")
	if err == nil {
		t.Error("expected error for unbound platform")
	}
}
