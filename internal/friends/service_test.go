package friends

import (
	"context"
	"database/sql"
	"testing"

	"github.com/cyandie/backend/internal/db"
	"github.com/google/uuid"
)

type mockFriendsQueries struct {
	friendship       db.Friendship
	friendErr        error
	createFriendship db.Friendship
	createFriendErr  error
	friendships      []db.Friendship
	friendsErr       error
	deleteErr        error

	// Block-related mock fields
	blockRelation    db.BlockRelation
	blockErr         error
	blockedList      []db.BlockRelation
	blockedListErr   error
	isBlockedID      uuid.UUID
	isBlockedErr     error
	deleteByUsers    db.Friendship
	deleteByUsersErr error

	// GetFriendshipByUsers mock fields
	friendshipByUsers    db.Friendship
	friendshipByUsersErr error

	// DeleteFriendshipByUsers call tracking
	deleteByUsersCalls int32
}

func (m *mockFriendsQueries) CreateFriendship(_ context.Context, _ db.CreateFriendshipParams) (db.Friendship, error) {
	return m.createFriendship, m.createFriendErr
}
func (m *mockFriendsQueries) DeleteFriendship(_ context.Context, _ uuid.UUID) (db.Friendship, error) {
	return db.Friendship{}, m.deleteErr
}
func (m *mockFriendsQueries) GetFriendship(_ context.Context, _ uuid.UUID) (db.Friendship, error) {
	return m.friendship, m.friendErr
}
func (m *mockFriendsQueries) GetFriendshipByUsers(_ context.Context, _ db.GetFriendshipByUsersParams) (db.Friendship, error) {
	return m.friendshipByUsers, m.friendshipByUsersErr
}
func (m *mockFriendsQueries) ListFriends(_ context.Context, _ uuid.UUID) ([]db.Friendship, error) {
	return m.friendships, m.friendsErr
}
func (m *mockFriendsQueries) ListPendingRequests(_ context.Context, _ uuid.UUID) ([]db.Friendship, error) {
	return m.friendships, m.friendsErr
}
func (m *mockFriendsQueries) UpdateFriendshipStatus(_ context.Context, _ db.UpdateFriendshipStatusParams) (db.Friendship, error) {
	return m.friendship, m.friendErr
}

// Stub remaining Querier methods
func (m *mockFriendsQueries) CreateUser(_ context.Context, _ db.CreateUserParams) (db.User, error) { return db.User{}, nil }
func (m *mockFriendsQueries) GetUserByID(_ context.Context, _ uuid.UUID) (db.User, error)          { return db.User{}, sql.ErrNoRows }
func (m *mockFriendsQueries) GetUserByUsername(_ context.Context, _ string) (db.User, error)        { return db.User{}, sql.ErrNoRows }
func (m *mockFriendsQueries) UpdateUserProfile(_ context.Context, _ db.UpdateUserProfileParams) (db.User, error) {
	return db.User{}, nil
}
func (m *mockFriendsQueries) UpdateUserStatus(_ context.Context, _ db.UpdateUserStatusParams) (db.User, error) {
	return db.User{}, nil
}
func (m *mockFriendsQueries) SearchUsers(_ context.Context, _ db.SearchUsersParams) ([]db.User, error) { return nil, nil }
func (m *mockFriendsQueries) GetCredentialByTypeIdentifier(_ context.Context, _ db.GetCredentialByTypeIdentifierParams) (db.Credential, error) {
	return db.Credential{}, sql.ErrNoRows
}
func (m *mockFriendsQueries) CreateCredential(_ context.Context, _ db.CreateCredentialParams) (db.Credential, error) {
	return db.Credential{}, nil
}
func (m *mockFriendsQueries) GetCredentialsByUserID(_ context.Context, _ uuid.UUID) ([]db.Credential, error) {
	return nil, nil
}
func (m *mockFriendsQueries) CountUserCredentials(_ context.Context, _ uuid.UUID) (int64, error) { return 0, nil }
func (m *mockFriendsQueries) CreateSession(_ context.Context, _ db.CreateSessionParams) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (m *mockFriendsQueries) GetSessionByID(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, sql.ErrNoRows
}
func (m *mockFriendsQueries) RevokeSession(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (m *mockFriendsQueries) RevokeSessionsByUserID(_ context.Context, _ uuid.UUID) ([]db.UserSession, error) {
	return nil, nil
}
func (m *mockFriendsQueries) CreatePlatformBinding(_ context.Context, _ db.CreatePlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, nil
}
func (m *mockFriendsQueries) GetPlatformBinding(_ context.Context, _ db.GetPlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, sql.ErrNoRows
}
func (m *mockFriendsQueries) GetPlatformBindingsByUserID(_ context.Context, _ uuid.UUID) ([]db.PlatformBinding, error) {
	return nil, nil
}
func (m *mockFriendsQueries) DeletePlatformBinding(_ context.Context, _ db.DeletePlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, nil
}
func (m *mockFriendsQueries) CreateLeaderboardConfig(_ context.Context, _ db.CreateLeaderboardConfigParams) (db.LeaderboardConfig, error) {
	return db.LeaderboardConfig{}, nil
}
func (m *mockFriendsQueries) GetLeaderboardConfig(_ context.Context, _ string) (db.LeaderboardConfig, error) {
	return db.LeaderboardConfig{}, sql.ErrNoRows
}
func (m *mockFriendsQueries) ListLeaderboardConfigs(_ context.Context) ([]db.LeaderboardConfig, error) { return nil, nil }
func (m *mockFriendsQueries) CreateScore(_ context.Context, _ db.CreateScoreParams) (db.LeaderboardScore, error) {
	return db.LeaderboardScore{}, nil
}
func (m *mockFriendsQueries) AddRoomMember(_ context.Context, _ db.AddRoomMemberParams) (db.ChatRoomMember, error) {
	return db.ChatRoomMember{}, nil
}
func (m *mockFriendsQueries) CreateChatMessage(_ context.Context, _ db.CreateChatMessageParams) (db.ChatMessage, error) {
	return db.ChatMessage{}, nil
}
func (m *mockFriendsQueries) CreateChatRoom(_ context.Context, _ db.CreateChatRoomParams) (db.ChatRoom, error) {
	return db.ChatRoom{}, nil
}
func (m *mockFriendsQueries) GetChatMessages(_ context.Context, _ db.GetChatMessagesParams) ([]db.ChatMessage, error) {
	return nil, nil
}
func (m *mockFriendsQueries) GetChatRoom(_ context.Context, _ uuid.UUID) (db.ChatRoom, error) {
	return db.ChatRoom{}, sql.ErrNoRows
}
func (m *mockFriendsQueries) GetRoomMembers(_ context.Context, _ uuid.UUID) ([]db.ChatRoomMember, error) {
	return nil, nil
}
func (m *mockFriendsQueries) RemoveRoomMember(_ context.Context, _ db.RemoveRoomMemberParams) (db.ChatRoomMember, error) {
	return db.ChatRoomMember{}, nil
}
func (m *mockFriendsQueries) CreateAdminUser(_ context.Context, _ db.CreateAdminUserParams) (db.AdminUser, error) {
	return db.AdminUser{}, nil
}
func (m *mockFriendsQueries) CreateAuditLog(_ context.Context, _ db.CreateAuditLogParams) (db.AuditLog, error) {
	return db.AuditLog{}, nil
}
func (m *mockFriendsQueries) GetAdminByUsername(_ context.Context, _ string) (db.AdminUser, error) {
	return db.AdminUser{}, sql.ErrNoRows
}
func (m *mockFriendsQueries) ListAuditLogs(_ context.Context, _ db.ListAuditLogsParams) ([]db.AuditLog, error) {
	return nil, nil
}
func (m *mockFriendsQueries) CreateBlockRelation(_ context.Context, _ db.CreateBlockRelationParams) (db.BlockRelation, error) {
	return m.blockRelation, m.blockErr
}
func (m *mockFriendsQueries) DeleteBlockRelation(_ context.Context, _ db.DeleteBlockRelationParams) (db.BlockRelation, error) {
	return m.blockRelation, m.blockErr
}
func (m *mockFriendsQueries) ListBlockedUsers(_ context.Context, _ uuid.UUID) ([]db.BlockRelation, error) {
	return m.blockedList, m.blockedListErr
}
func (m *mockFriendsQueries) IsBlockedBy(_ context.Context, _ db.IsBlockedByParams) (uuid.UUID, error) {
	return m.isBlockedID, m.isBlockedErr
}
func (m *mockFriendsQueries) DeleteFriendshipByUsers(_ context.Context, _ db.DeleteFriendshipByUsersParams) (db.Friendship, error) {
	m.deleteByUsersCalls++
	return m.deleteByUsers, m.deleteByUsersErr
}
func (m *mockFriendsQueries) ListRoomsByUser(_ context.Context, _ uuid.UUID) ([]db.ChatRoom, error) { return nil, nil }

func TestFriendsService_SendRequest(t *testing.T) {
	from := uuid.New()
	to := uuid.New()
	q := &mockFriendsQueries{
		friendshipByUsersErr: sql.ErrNoRows,
		isBlockedErr:        sql.ErrNoRows,
		createFriendship:    db.Friendship{ID: uuid.New(), UserID: from, FriendID: to, Status: "pending"},
	}
	svc := NewFriendsService(q, nil, nil)

	f, err := svc.SendRequest(context.Background(), from.String(), to.String())
	if err != nil {
		t.Fatalf("SendRequest failed: %v", err)
	}
	if f.Status != "pending" {
		t.Errorf("expected pending, got %s", f.Status)
	}
}

func TestFriendsService_SendRequest_Self(t *testing.T) {
	svc := NewFriendsService(&mockFriendsQueries{}, nil, nil)
	uid := uuid.New().String()

	_, err := svc.SendRequest(context.Background(), uid, uid)
	if err == nil {
		t.Error("expected error for self-friend request")
	}
}

func TestFriendsService_AcceptRequest(t *testing.T) {
	fid := uuid.New()
	q := &mockFriendsQueries{
		friendship: db.Friendship{ID: fid, Status: "accepted"},
	}
	svc := NewFriendsService(q, nil, nil)

	f, err := svc.AcceptRequest(context.Background(), fid.String())
	if err != nil {
		t.Fatalf("AcceptRequest failed: %v", err)
	}
	if f.Status != "accepted" {
		t.Errorf("expected accepted, got %s", f.Status)
	}
}

func TestFriendsService_AcceptRequest_NotFound(t *testing.T) {
	q := &mockFriendsQueries{friendErr: sql.ErrNoRows}
	svc := NewFriendsService(q, nil, nil)

	_, err := svc.AcceptRequest(context.Background(), uuid.New().String())
	if err == nil {
		t.Error("expected error for nonexistent request")
	}
}

func TestFriendsService_RejectRequest(t *testing.T) {
	q := &mockFriendsQueries{}
	svc := NewFriendsService(q, nil, nil)

	err := svc.RejectRequest(context.Background(), uuid.New().String())
	if err != nil {
		t.Fatalf("RejectRequest failed: %v", err)
	}
}

func TestFriendsService_ListFriends(t *testing.T) {
	uid := uuid.New()
	q := &mockFriendsQueries{
		friendships: []db.Friendship{
			{ID: uuid.New(), UserID: uid, FriendID: uuid.New(), Status: "accepted"},
		},
	}
	svc := NewFriendsService(q, nil, nil)

	friends, err := svc.ListFriends(context.Background(), uid.String())
	if err != nil {
		t.Fatalf("ListFriends failed: %v", err)
	}
	if len(friends) != 1 {
		t.Errorf("expected 1 friend, got %d", len(friends))
	}
}

func TestFriendsService_Block(t *testing.T) {
	blocker := uuid.New()
	blocked := uuid.New()
	q := &mockFriendsQueries{
		blockRelation:        db.BlockRelation{BlockerID: blocker, BlockedID: blocked},
		friendshipByUsersErr: sql.ErrNoRows,
		isBlockedErr:         sql.ErrNoRows,
	}
	svc := NewFriendsService(q, nil, nil)

	err := svc.Block(context.Background(), blocker.String(), blocked.String(), "spam")
	if err != nil {
		t.Fatalf("Block failed: %v", err)
	}
}

func TestFriendsService_Block_Self(t *testing.T) {
	svc := NewFriendsService(&mockFriendsQueries{}, nil, nil)
	uid := uuid.New().String()

	err := svc.Block(context.Background(), uid, uid, "")
	if err == nil {
		t.Error("expected error for self-block")
	}
}

func TestFriendsService_Block_DeletesPendingFriendship(t *testing.T) {
	blocker := uuid.New()
	blocked := uuid.New()
	q := &mockFriendsQueries{
		blockRelation: db.BlockRelation{BlockerID: blocker, BlockedID: blocked},
		friendshipByUsers: db.Friendship{
			ID:     uuid.New(),
			UserID: blocked,
			FriendID: blocker,
			Status: "pending",
		},
		friendshipByUsersErr: nil,
		isBlockedErr:        sql.ErrNoRows,
	}
	svc := NewFriendsService(q, nil, nil)

	err := svc.Block(context.Background(), blocker.String(), blocked.String(), "spam")
	if err != nil {
		t.Fatalf("Block failed: %v", err)
	}
	if q.deleteByUsersCalls != 1 {
		t.Errorf("expected deleteByUsersCalls=1, got %d", q.deleteByUsersCalls)
	}
}

func TestFriendsService_Unblock(t *testing.T) {
	blocker := uuid.New()
	blocked := uuid.New()
	q := &mockFriendsQueries{
		blockRelation: db.BlockRelation{BlockerID: blocker, BlockedID: blocked},
	}
	svc := NewFriendsService(q, nil, nil)

	err := svc.Unblock(context.Background(), blocker.String(), blocked.String())
	if err != nil {
		t.Fatalf("Unblock failed: %v", err)
	}
}

func TestFriendsService_IsBlocked(t *testing.T) {
	target := uuid.New()
	by := uuid.New()
	q := &mockFriendsQueries{
		isBlockedID: uuid.New(),
	}
	svc := NewFriendsService(q, nil, nil)

	blocked, err := svc.IsBlocked(context.Background(), target.String(), by.String())
	if err != nil {
		t.Fatalf("IsBlocked failed: %v", err)
	}
	if !blocked {
		t.Error("expected blocked=true")
	}
}

func TestFriendsService_IsBlocked_NotBlocked(t *testing.T) {
	target := uuid.New()
	by := uuid.New()
	q := &mockFriendsQueries{
		isBlockedErr: sql.ErrNoRows,
	}
	svc := NewFriendsService(q, nil, nil)

	blocked, err := svc.IsBlocked(context.Background(), target.String(), by.String())
	if err != nil {
		t.Fatalf("IsBlocked failed: %v", err)
	}
	if blocked {
		t.Error("expected blocked=false")
	}
}

func TestFriendsService_RemoveFriend(t *testing.T) {
	user := uuid.New()
	friend := uuid.New()
	q := &mockFriendsQueries{
		deleteByUsers: db.Friendship{UserID: user, FriendID: friend},
	}
	svc := NewFriendsService(q, nil, nil)

	err := svc.RemoveFriend(context.Background(), user.String(), friend.String())
	if err != nil {
		t.Fatalf("RemoveFriend failed: %v", err)
	}
}
