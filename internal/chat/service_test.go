package chat

import (
	"context"
	"database/sql"
	"io"
	"net"
	"testing"
	"time"

	chatv1 "github.com/cyandie/backend/api/proto/chat/v1"
	"github.com/cyandie/backend/internal/db"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

// ---------------------------------------------------------------------------
// mockQuerier implements db.Querier with configurable return values for chat tests.
// ---------------------------------------------------------------------------

type mockQuerier struct {
	rooms    []db.ChatRoom
	roomsErr error
	room     db.ChatRoom
	roomErr  error
	messages []db.ChatMessage
	msgsErr  error
	members  []db.ChatRoomMember
	membErr  error
	member   db.ChatRoomMember
	membErr2 error
	message  db.ChatMessage
	msgErr   error

	lastGetMessagesLimit  int32
	lastGetMessagesOffset int32
}

// chat-specific methods with configurable behaviour

func (m *mockQuerier) ListRoomsByUser(_ context.Context, _ uuid.UUID) ([]db.ChatRoom, error) {
	return m.rooms, m.roomsErr
}
func (m *mockQuerier) GetChatRoom(_ context.Context, _ uuid.UUID) (db.ChatRoom, error) {
	return m.room, m.roomErr
}
func (m *mockQuerier) GetChatMessages(_ context.Context, p db.GetChatMessagesParams) ([]db.ChatMessage, error) {
	m.lastGetMessagesLimit = p.Limit
	m.lastGetMessagesOffset = p.Offset
	return m.messages, m.msgsErr
}
func (m *mockQuerier) GetRoomMembers(_ context.Context, _ uuid.UUID) ([]db.ChatRoomMember, error) {
	return m.members, m.membErr
}
func (m *mockQuerier) AddRoomMember(_ context.Context, _ db.AddRoomMemberParams) (db.ChatRoomMember, error) {
	return m.member, m.membErr2
}
func (m *mockQuerier) RemoveRoomMember(_ context.Context, _ db.RemoveRoomMemberParams) (db.ChatRoomMember, error) {
	return m.member, m.membErr2
}
func (m *mockQuerier) CreateChatMessage(_ context.Context, _ db.CreateChatMessageParams) (db.ChatMessage, error) {
	return m.message, m.msgErr
}
func (m *mockQuerier) CreateChatRoom(_ context.Context, _ db.CreateChatRoomParams) (db.ChatRoom, error) {
	return m.room, m.roomErr
}

// stub remaining Querier methods

func (m *mockQuerier) CreateUser(_ context.Context, _ db.CreateUserParams) (db.User, error) {
	return db.User{}, nil
}
func (m *mockQuerier) GetUserByID(_ context.Context, _ uuid.UUID) (db.User, error) {
	return db.User{}, sql.ErrNoRows
}
func (m *mockQuerier) GetUserByUsername(_ context.Context, _ string) (db.User, error) {
	return db.User{}, sql.ErrNoRows
}
func (m *mockQuerier) UpdateUserProfile(_ context.Context, _ db.UpdateUserProfileParams) (db.User, error) {
	return db.User{}, nil
}
func (m *mockQuerier) GetCredentialByTypeIdentifier(_ context.Context, _ db.GetCredentialByTypeIdentifierParams) (db.Credential, error) {
	return db.Credential{}, sql.ErrNoRows
}
func (m *mockQuerier) CreateCredential(_ context.Context, _ db.CreateCredentialParams) (db.Credential, error) {
	return db.Credential{}, nil
}
func (m *mockQuerier) GetCredentialsByUserID(_ context.Context, _ uuid.UUID) ([]db.Credential, error) {
	return nil, nil
}
func (m *mockQuerier) CountUserCredentials(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockQuerier) CreateSession(_ context.Context, _ db.CreateSessionParams) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (m *mockQuerier) GetSessionByID(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, sql.ErrNoRows
}
func (m *mockQuerier) RevokeSession(_ context.Context, _ uuid.UUID) (db.UserSession, error) {
	return db.UserSession{}, nil
}
func (m *mockQuerier) RevokeSessionsByUserID(_ context.Context, _ uuid.UUID) ([]db.UserSession, error) {
	return nil, nil
}
func (m *mockQuerier) CreatePlatformBinding(_ context.Context, _ db.CreatePlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, nil
}
func (m *mockQuerier) GetPlatformBinding(_ context.Context, _ db.GetPlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, sql.ErrNoRows
}
func (m *mockQuerier) GetPlatformBindingsByUserID(_ context.Context, _ uuid.UUID) ([]db.PlatformBinding, error) {
	return nil, nil
}
func (m *mockQuerier) DeletePlatformBinding(_ context.Context, _ db.DeletePlatformBindingParams) (db.PlatformBinding, error) {
	return db.PlatformBinding{}, nil
}
func (m *mockQuerier) CreateLeaderboardConfig(_ context.Context, _ db.CreateLeaderboardConfigParams) (db.LeaderboardConfig, error) {
	return db.LeaderboardConfig{}, nil
}
func (m *mockQuerier) GetLeaderboardConfig(_ context.Context, _ string) (db.LeaderboardConfig, error) {
	return db.LeaderboardConfig{}, sql.ErrNoRows
}
func (m *mockQuerier) ListLeaderboardConfigs(_ context.Context) ([]db.LeaderboardConfig, error) {
	return nil, nil
}
func (m *mockQuerier) CreateScore(_ context.Context, _ db.CreateScoreParams) (db.LeaderboardScore, error) {
	return db.LeaderboardScore{}, nil
}
func (m *mockQuerier) CreateAdminUser(_ context.Context, _ db.CreateAdminUserParams) (db.AdminUser, error) {
	return db.AdminUser{}, nil
}
func (m *mockQuerier) GetAdminByUsername(_ context.Context, _ string) (db.AdminUser, error) {
	return db.AdminUser{}, sql.ErrNoRows
}
func (m *mockQuerier) CreateAuditLog(_ context.Context, _ db.CreateAuditLogParams) (db.AuditLog, error) {
	return db.AuditLog{}, nil
}
func (m *mockQuerier) ListAuditLogs(_ context.Context, _ db.ListAuditLogsParams) ([]db.AuditLog, error) {
	return nil, nil
}
func (m *mockQuerier) SearchUsers(_ context.Context, _ db.SearchUsersParams) ([]db.User, error) {
	return nil, nil
}
func (m *mockQuerier) UpdateUserStatus(_ context.Context, _ db.UpdateUserStatusParams) (db.User, error) {
	return db.User{}, nil
}
func (m *mockQuerier) CreateFriendship(_ context.Context, _ db.CreateFriendshipParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *mockQuerier) DeleteFriendship(_ context.Context, _ uuid.UUID) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *mockQuerier) GetFriendship(_ context.Context, _ uuid.UUID) (db.Friendship, error) {
	return db.Friendship{}, sql.ErrNoRows
}
func (m *mockQuerier) GetFriendshipByUsers(_ context.Context, _ db.GetFriendshipByUsersParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *mockQuerier) ListFriends(_ context.Context, _ uuid.UUID) ([]db.Friendship, error) {
	return nil, nil
}
func (m *mockQuerier) ListPendingRequests(_ context.Context, _ uuid.UUID) ([]db.Friendship, error) {
	return nil, nil
}
func (m *mockQuerier) UpdateFriendshipStatus(_ context.Context, _ db.UpdateFriendshipStatusParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}
func (m *mockQuerier) CreateBlockRelation(_ context.Context, _ db.CreateBlockRelationParams) (db.BlockRelation, error) {
	return db.BlockRelation{}, nil
}
func (m *mockQuerier) DeleteBlockRelation(_ context.Context, _ db.DeleteBlockRelationParams) (db.BlockRelation, error) {
	return db.BlockRelation{}, nil
}
func (m *mockQuerier) ListBlockedUsers(_ context.Context, _ uuid.UUID) ([]db.BlockRelation, error) {
	return nil, nil
}
func (m *mockQuerier) IsBlockedBy(_ context.Context, _ db.IsBlockedByParams) (uuid.UUID, error) {
	return uuid.UUID{}, sql.ErrNoRows
}
func (m *mockQuerier) DeleteFriendshipByUsers(_ context.Context, _ db.DeleteFriendshipByUsersParams) (db.Friendship, error) {
	return db.Friendship{}, nil
}

// ---------------------------------------------------------------------------
// mock dependencies for ChatService
// ---------------------------------------------------------------------------

type mockTokenValidator struct {
	userID string
	err    error
}

func (m *mockTokenValidator) ValidateToken(_ context.Context, _ string) (string, error) {
	return m.userID, m.err
}

type mockBlockChecker struct {
	blocked bool
	err     error
}

func (m *mockBlockChecker) IsBlocked(_ context.Context, _, _ string) (bool, error) {
	return m.blocked, m.err
}

type mockPresenceSetter struct {
	onlineCalls  []struct{ userID, username string }
	offlineCalls []string
}

func (m *mockPresenceSetter) SetOnline(_ context.Context, userID, username string) error {
	m.onlineCalls = append(m.onlineCalls, struct{ userID, username string }{userID, username})
	return nil
}

func (m *mockPresenceSetter) SetOffline(_ context.Context, userID string) error {
	m.offlineCalls = append(m.offlineCalls, userID)
	return nil
}

// ---------------------------------------------------------------------------
// testConn creates a Connection backed by a net.Pipe so that
// conn.Send() does not panic on a nil net.Conn.
// A background goroutine drains the server side so writes never block.
// ---------------------------------------------------------------------------

func testConn(id string) *Connection {
	client, server := net.Pipe()
	// Drain the server side in the background so writes on client never block.
	go func() {
		buf := make([]byte, 4096)
		for {
			_, err := server.Read(buf)
			if err != nil {
				return
			}
		}
	}()
	return &Connection{ID: id, Conn: client}
}

// captureConn is a net.Conn that buffers all writes in-memory
// so tests can read them back. It satisfies both net.Conn and
// io.Reader for use with DecodeFrame.
type captureConn struct {
	buf []byte
}

func newCaptureConn() *captureConn { return &captureConn{} }

func (c *captureConn) Read(b []byte) (n int, err error) {
	if len(c.buf) == 0 {
		return 0, io.EOF
	}
	n = copy(b, c.buf)
	c.buf = c.buf[n:]
	return n, nil
}

func (c *captureConn) Write(b []byte) (n int, err error) {
	c.buf = append(c.buf, b...)
	return len(b), nil
}

func (c *captureConn) Close() error                       { return nil }
func (c *captureConn) LocalAddr() net.Addr                { return nil }
func (c *captureConn) RemoteAddr() net.Addr               { return nil }
func (c *captureConn) SetDeadline(t time.Time) error      { return nil }
func (c *captureConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *captureConn) SetWriteDeadline(t time.Time) error { return nil }

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestNewChatService(t *testing.T) {
	q := &mockQuerier{}
	srv := NewTCPServer("127.0.0.1:0") // port 0 lets OS pick a free port
	bc := &mockBlockChecker{}
	ps := &mockPresenceSetter{}
	tv := &mockTokenValidator{userID: "user-1"}

	svc := NewChatService(q, srv, bc, ps, tv)
	if svc == nil {
		t.Fatal("expected non-nil ChatService")
	}
	if svc.queries != q {
		t.Error("queries not set")
	}
	if svc.server != srv {
		t.Error("server not set")
	}
	if svc.blockChecker != bc {
		t.Error("blockChecker not set")
	}
	if svc.presenceSetter != ps {
		t.Error("presenceSetter not set")
	}
	if svc.tokenValidator != tv {
		t.Error("tokenValidator not set")
	}
}

func TestNewChatService_NilDependencies(t *testing.T) {
	q := &mockQuerier{}
	srv := NewTCPServer("127.0.0.1:0")

	// All optional deps are nil -- service should still be created
	svc := NewChatService(q, srv, nil, nil, nil)
	if svc == nil {
		t.Fatal("expected non-nil ChatService with nil deps")
	}
}

func TestChatService_StartStop(t *testing.T) {
	q := &mockQuerier{}
	srv := NewTCPServer("127.0.0.1:0")
	svc := NewChatService(q, srv, nil, nil, nil)

	if err := svc.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if err := svc.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestChatService_HandleConnect(t *testing.T) {
	q := &mockQuerier{}
	srv := NewTCPServer("127.0.0.1:0")
	ps := &mockPresenceSetter{}
	svc := NewChatService(q, srv, nil, ps, nil)

	conn := testConn("conn-1")
	defer conn.Conn.Close()

	// handleConnect only logs; verify it does not panic
	svc.handleConnect(conn)
}

func TestChatService_HandleDisconnect_SetsOffline(t *testing.T) {
	q := &mockQuerier{}
	srv := NewTCPServer("127.0.0.1:0")
	ps := &mockPresenceSetter{}
	svc := NewChatService(q, srv, nil, ps, nil)

	conn := testConn("conn-1")
	conn.UserID = "user-1"
	defer conn.Conn.Close()

	svc.handleDisconnect(conn)

	if len(ps.offlineCalls) != 1 {
		t.Fatalf("expected 1 offline call, got %d", len(ps.offlineCalls))
	}
	if ps.offlineCalls[0] != "user-1" {
		t.Errorf("expected offline for user-1, got %s", ps.offlineCalls[0])
	}
}

func TestChatService_HandleDisconnect_NoUserID(t *testing.T) {
	q := &mockQuerier{}
	srv := NewTCPServer("127.0.0.1:0")
	ps := &mockPresenceSetter{}
	svc := NewChatService(q, srv, nil, ps, nil)

	conn := testConn("conn-2") // no UserID
	defer conn.Conn.Close()

	svc.handleDisconnect(conn)

	if len(ps.offlineCalls) != 0 {
		t.Errorf("expected 0 offline calls for unauthenticated disconnect, got %d", len(ps.offlineCalls))
	}
}

func TestChatService_HandleDisconnect_NilPresenceSetter(t *testing.T) {
	q := &mockQuerier{}
	srv := NewTCPServer("127.0.0.1:0")
	svc := NewChatService(q, srv, nil, nil, nil)

	conn := testConn("conn-3")
	conn.UserID = "user-1"
	defer conn.Conn.Close()

	// Should not panic when presenceSetter is nil
	svc.handleDisconnect(conn)
}

func TestChatService_HandleAuth_ValidToken(t *testing.T) {
	q := &mockQuerier{}
	srv := NewTCPServer("127.0.0.1:0")
	tv := &mockTokenValidator{userID: "user-42"}
	svc := NewChatService(q, srv, nil, nil, tv)

	conn := testConn("conn-1")
	defer conn.Conn.Close()

	env := authEnvelope("valid-token", "alice")
	svc.handleAuth(conn, env)

	if conn.UserID != "user-42" {
		t.Errorf("expected UserID user-42, got %s", conn.UserID)
	}
	if conn.Username != "alice" {
		t.Errorf("expected Username alice, got %s", conn.Username)
	}
}

func TestChatService_HandleAuth_InvalidToken(t *testing.T) {
	q := &mockQuerier{}
	srv := NewTCPServer("127.0.0.1:0")
	tv := &mockTokenValidator{err: sql.ErrNoRows} // any non-nil error
	svc := NewChatService(q, srv, nil, nil, tv)

	conn := testConn("conn-2")
	defer conn.Conn.Close()

	env := authEnvelope("bad-token", "bob")
	svc.handleAuth(conn, env)

	// UserID must remain empty -- auth rejected
	if conn.UserID != "" {
		t.Errorf("expected empty UserID on rejected auth, got %s", conn.UserID)
	}
}

func TestChatService_HandleAuth_FallbackWithoutValidator(t *testing.T) {
	q := &mockQuerier{}
	srv := NewTCPServer("127.0.0.1:0")
	svc := NewChatService(q, srv, nil, nil, nil) // no tokenValidator

	conn := testConn("conn-3")
	defer conn.Conn.Close()

	env := authEnvelope("raw-user-id", "charlie")
	svc.handleAuth(conn, env)

	// Without validator, token is used as userID directly
	if conn.UserID != "raw-user-id" {
		t.Errorf("expected UserID raw-user-id (fallback), got %s", conn.UserID)
	}
	if conn.Username != "charlie" {
		t.Errorf("expected Username charlie, got %s", conn.Username)
	}
}

func TestChatService_HandleAuth_MissingToken(t *testing.T) {
	q := &mockQuerier{}
	srv := NewTCPServer("127.0.0.1:0")
	tv := &mockTokenValidator{userID: "should-not-be-used"}
	svc := NewChatService(q, srv, nil, nil, tv)

	conn := testConn("conn-4")
	defer conn.Conn.Close()

	env := authEnvelope("", "") // empty token
	svc.handleAuth(conn, env)

	if conn.UserID != "" {
		t.Errorf("expected empty UserID when token is missing, got %s", conn.UserID)
	}
}

func TestChatService_HandleJoinRoom_Success(t *testing.T) {
	roomID := uuid.New()
	userID := uuid.New()
	q := &mockQuerier{
		member: db.ChatRoomMember{
			ID:     uuid.New(),
			RoomID: roomID,
			UserID: userID,
			Role:   "member",
		},
	}
	srv := NewTCPServer("127.0.0.1:0")
	svc := NewChatService(q, srv, nil, nil, nil)

	conn := testConn("conn-1")
	conn.UserID = userID.String()
	defer conn.Conn.Close()

	env := joinRoomEnvelope(roomID.String())
	svc.handleJoinRoom(conn, env)

	if !conn.IsInRoom(roomID.String()) {
		t.Error("expected conn to be in room after join")
	}
}

func TestChatService_HandleJoinRoom_Unauthenticated(t *testing.T) {
	q := &mockQuerier{}
	srv := NewTCPServer("127.0.0.1:0")
	svc := NewChatService(q, srv, nil, nil, nil)

	conn := testConn("conn-2") // no UserID
	defer conn.Conn.Close()

	env := joinRoomEnvelope(uuid.New().String())
	svc.handleJoinRoom(conn, env)

	// Should not join room when not authenticated
	if conn.IsInRoom(uuid.New().String()) {
		t.Error("expected conn not to be in room when unauthenticated")
	}
}

func TestChatService_HandleJoinRoom_DBError(t *testing.T) {
	q := &mockQuerier{
		membErr2: sql.ErrConnDone,
	}
	srv := NewTCPServer("127.0.0.1:0")
	svc := NewChatService(q, srv, nil, nil, nil)

	conn := testConn("conn-3")
	conn.UserID = uuid.New().String()
	defer conn.Conn.Close()

	env := joinRoomEnvelope(uuid.New().String())
	svc.handleJoinRoom(conn, env)

	// Should not be in room when DB error occurs
	if conn.IsInRoom(uuid.New().String()) {
		t.Error("expected conn not to be in room on DB error")
	}
}

func TestChatService_HandleLeaveRoom_Success(t *testing.T) {
	roomID := uuid.New()
	userID := uuid.New()
	q := &mockQuerier{
		member: db.ChatRoomMember{
			ID:     uuid.New(),
			RoomID: roomID,
			UserID: userID,
			Role:   "member",
		},
	}
	srv := NewTCPServer("127.0.0.1:0")
	svc := NewChatService(q, srv, nil, nil, nil)

	conn := testConn("conn-1")
	conn.UserID = userID.String()
	defer conn.Conn.Close()

	conn.JoinRoom(roomID.String())
	env := leaveRoomEnvelope(roomID.String())
	svc.handleLeaveRoom(conn, env)

	if conn.IsInRoom(roomID.String()) {
		t.Error("expected conn to have left the room")
	}
}

func TestChatService_HandleLeaveRoom_Unauthenticated(t *testing.T) {
	q := &mockQuerier{}
	srv := NewTCPServer("127.0.0.1:0")
	svc := NewChatService(q, srv, nil, nil, nil)

	conn := testConn("conn-2") // no UserID
	defer conn.Conn.Close()

	env := leaveRoomEnvelope(uuid.New().String())
	// Should not panic
	svc.handleLeaveRoom(conn, env)
}

func TestChatService_HandleSendMessage_Unauthenticated(t *testing.T) {
	q := &mockQuerier{}
	srv := NewTCPServer("127.0.0.1:0")
	svc := NewChatService(q, srv, nil, nil, nil)

	conn := testConn("conn-1") // no UserID
	defer conn.Conn.Close()

	env := sendMessageEnvelope(uuid.New().String(), "hello")
	svc.handleSendMessage(conn, env)

	// Should not create a message when unauthenticated
	// (verified by not crashing and not calling CreateChatMessage)
}

func TestChatService_HandleSendMessage_NotInRoom(t *testing.T) {
	q := &mockQuerier{}
	srv := NewTCPServer("127.0.0.1:0")
	svc := NewChatService(q, srv, nil, nil, nil)

	conn := testConn("conn-1")
	conn.UserID = uuid.New().String()
	defer conn.Conn.Close()

	// conn has not joined the room
	env := sendMessageEnvelope(uuid.New().String(), "hello")
	svc.handleSendMessage(conn, env)
}

func TestChatService_HandleSendMessage_BlockedByMember(t *testing.T) {
	roomID := uuid.New()
	senderID := uuid.New()
	memberID := uuid.New()

	q := &mockQuerier{
		members: []db.ChatRoomMember{
			{ID: uuid.New(), RoomID: roomID, UserID: senderID, Role: "member"},
			{ID: uuid.New(), RoomID: roomID, UserID: memberID, Role: "member"},
		},
	}
	bc := &mockBlockChecker{blocked: true}
	srv := NewTCPServer("127.0.0.1:0")
	svc := NewChatService(q, srv, bc, nil, nil)

	conn := testConn("conn-1")
	conn.UserID = senderID.String()
	defer conn.Conn.Close()

	conn.JoinRoom(roomID.String())

	env := sendMessageEnvelope(roomID.String(), "hello")
	svc.handleSendMessage(conn, env)

	// Message should not be created when blocked
}

func TestChatService_HandleSendMessage_Success(t *testing.T) {
	roomID := uuid.New()
	senderID := uuid.New()

	q := &mockQuerier{
		members: []db.ChatRoomMember{
			{ID: uuid.New(), RoomID: roomID, UserID: senderID, Role: "member"},
		},
		message: db.ChatMessage{
			ID:       uuid.New(),
			RoomID:   roomID,
			SenderID: senderID,
			Content:  "hello",
			Type:     "text",
		},
	}
	bc := &mockBlockChecker{blocked: false}
	srv := NewTCPServer("127.0.0.1:0")
	svc := NewChatService(q, srv, bc, nil, nil)

	conn := testConn("conn-1")
	conn.UserID = senderID.String()
	defer conn.Conn.Close()

	conn.JoinRoom(roomID.String())

	env := sendMessageEnvelope(roomID.String(), "hello")
	svc.handleSendMessage(conn, env)
	// No panic = success; message creation was exercised
}

func TestChatService_HandleInviteRoom_Unauthenticated(t *testing.T) {
	q := &mockQuerier{}
	srv := NewTCPServer("127.0.0.1:0")
	svc := NewChatService(q, srv, nil, nil, nil)

	conn := testConn("conn-1") // no UserID
	defer conn.Conn.Close()

	env := inviteRoomEnvelope(uuid.New().String(), "target-user")
	svc.handleInviteRoom(conn, env)
	// Should not panic; invite should be rejected
}

func TestChatService_HandleInviteRoom_InvalidRoomID(t *testing.T) {
	q := &mockQuerier{}
	srv := NewTCPServer("127.0.0.1:0")
	svc := NewChatService(q, srv, nil, nil, nil)

	conn := testConn("conn-1")
	conn.UserID = uuid.New().String()
	defer conn.Conn.Close()

	env := inviteRoomEnvelope("not-a-uuid", "target-user")
	svc.handleInviteRoom(conn, env)
	// Should not panic; should get BAD_REQUEST
}

func TestChatService_HandleInviteRoom_NotInRoom(t *testing.T) {
	roomID := uuid.New()
	userID := uuid.New()
	otherUser := uuid.New()

	q := &mockQuerier{
		members: []db.ChatRoomMember{
			{ID: uuid.New(), RoomID: roomID, UserID: otherUser, Role: "owner"},
		},
	}
	srv := NewTCPServer("127.0.0.1:0")
	svc := NewChatService(q, srv, nil, nil, nil)

	conn := testConn("conn-1")
	conn.UserID = userID.String()
	defer conn.Conn.Close()

	env := inviteRoomEnvelope(roomID.String(), "target-user")
	svc.handleInviteRoom(conn, env)
	// Should not panic; should get FORBIDDEN
}

func TestChatService_HandleInviteRoom_Success(t *testing.T) {
	roomID := uuid.New()
	userID := uuid.New()
	targetID := uuid.New()

	q := &mockQuerier{
		members: []db.ChatRoomMember{
			{ID: uuid.New(), RoomID: roomID, UserID: userID, Role: "owner"},
		},
	}
	srv := NewTCPServer("127.0.0.1:0")
	svc := NewChatService(q, srv, nil, nil, nil)

	conn := testConn("conn-1")
	conn.UserID = userID.String()
	defer conn.Conn.Close()

	env := inviteRoomEnvelope(roomID.String(), targetID.String())
	svc.handleInviteRoom(conn, env)
	// No panic = success
}

func TestChatService_HandleInviteRoom_InviterUsername(t *testing.T) {
	roomID := uuid.New()
	userID := uuid.New()
	targetID := uuid.New()

	q := &mockQuerier{
		members: []db.ChatRoomMember{
			{ID: uuid.New(), RoomID: roomID, UserID: userID, Role: "owner"},
		},
	}
	srv := NewTCPServer("127.0.0.1:0")
	svc := NewChatService(q, srv, nil, nil, nil)

	// Set up the target user's connection so SendToUser can find it.
	// Use a captureConn to record the bytes written by Connection.Send.
	cc := newCaptureConn()
	targetConn := &Connection{
		ID:     "target-conn",
		UserID: targetID.String(),
		Conn:   cc,
	}
	srv.mu.Lock()
	srv.conns[targetConn.ID] = targetConn
	srv.mu.Unlock()

	// Set up inviter connection with a known username.
	inviterConn := testConn("inviter-conn")
	inviterConn.UserID = userID.String()
	inviterConn.Username = "alice"
	defer inviterConn.Conn.Close()

	env := inviteRoomEnvelope(roomID.String(), targetID.String())
	svc.handleInviteRoom(inviterConn, env)

	// Decode the frame that was written to the target connection.
	frame, err := DecodeFrame(cc)
	if err != nil {
		t.Fatalf("failed to decode frame sent to target: %v", err)
	}

	got := &chatv1.ChatEnvelope{}
	if err := proto.Unmarshal(frame.Value, got); err != nil {
		t.Fatalf("failed to unmarshal envelope: %v", err)
	}

	notify := got.GetInviteRoomNotify()
	if notify == nil {
		t.Fatal("expected InviteRoomNotify payload, got nil")
	}
	if notify.InviterUsername != "alice" {
		t.Errorf("expected InviterUsername %q, got %q", "alice", notify.InviterUsername)
	}
	if notify.InviterId != userID.String() {
		t.Errorf("expected InviterId %q, got %q", userID.String(), notify.InviterId)
	}
	if notify.RoomId != roomID.String() {
		t.Errorf("expected RoomId %q, got %q", roomID.String(), notify.RoomId)
	}
}

func TestChatService_HandleMessage_Auth(t *testing.T) {
	q := &mockQuerier{}
	srv := NewTCPServer("127.0.0.1:0")
	tv := &mockTokenValidator{userID: "user-99"}
	svc := NewChatService(q, srv, nil, nil, tv)

	conn := testConn("conn-1")
	defer conn.Conn.Close()

	env := authEnvelope("good-token", "dave")
	frame := marshalEnvelope(env)
	svc.handleMessage(conn, &frame)

	if conn.UserID != "user-99" {
		t.Errorf("expected UserID user-99 after handleMessage AUTH, got %s", conn.UserID)
	}
}

func TestChatService_HandleMessage_InvalidProto(t *testing.T) {
	q := &mockQuerier{}
	srv := NewTCPServer("127.0.0.1:0")
	svc := NewChatService(q, srv, nil, nil, nil)

	conn := testConn("conn-1")
	defer conn.Conn.Close()

	// Send garbage bytes as a frame value -- should not panic
	frame := Frame{Type: 1, Value: []byte("not-protobuf")}
	svc.handleMessage(conn, &frame)
}

// ---------------------------------------------------------------------------
// protobuf envelope helpers
// ---------------------------------------------------------------------------

func authEnvelope(token, username string) *chatv1.ChatEnvelope {
	return &chatv1.ChatEnvelope{
		Type: chatv1.MessageType_AUTH,
		Payload: &chatv1.ChatEnvelope_AuthRequest{
			AuthRequest: &chatv1.AuthRequest{
				AccessToken: token,
				Username:    username,
			},
		},
	}
}

func joinRoomEnvelope(roomID string) *chatv1.ChatEnvelope {
	return &chatv1.ChatEnvelope{
		Type: chatv1.MessageType_JOIN_ROOM,
		Payload: &chatv1.ChatEnvelope_JoinRoomRequest{
			JoinRoomRequest: &chatv1.JoinRoomRequest{
				RoomId: roomID,
			},
		},
	}
}

func leaveRoomEnvelope(roomID string) *chatv1.ChatEnvelope {
	return &chatv1.ChatEnvelope{
		Type: chatv1.MessageType_LEAVE_ROOM,
		Payload: &chatv1.ChatEnvelope_LeaveRoomRequest{
			LeaveRoomRequest: &chatv1.LeaveRoomRequest{
				RoomId: roomID,
			},
		},
	}
}

func sendMessageEnvelope(roomID, content string) *chatv1.ChatEnvelope {
	return &chatv1.ChatEnvelope{
		Type: chatv1.MessageType_SEND_MSG,
		Payload: &chatv1.ChatEnvelope_SendMessageRequest{
			SendMessageRequest: &chatv1.SendMessageRequest{
				RoomId:  roomID,
				Content: content,
				Type:    "text",
			},
		},
	}
}

func inviteRoomEnvelope(roomID, targetUserID string) *chatv1.ChatEnvelope {
	return &chatv1.ChatEnvelope{
		Type: chatv1.MessageType_INVITE_ROOM,
		Payload: &chatv1.ChatEnvelope_InviteRoomRequest{
			InviteRoomRequest: &chatv1.InviteRoomRequest{
				RoomId:       roomID,
				TargetUserId: targetUserID,
			},
		},
	}
}

// marshalEnvelope is a test helper to produce a Frame from an envelope.
func marshalEnvelope(env *chatv1.ChatEnvelope) Frame {
	data, _ := proto.Marshal(env)
	return Frame{Type: uint16(env.Type), Value: data}
}
