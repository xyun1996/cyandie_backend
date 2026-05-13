# Phase 6: Social Presence Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend Friends module with blacklist, delete friend, recent contacts, and real-time online presence push via the existing Chat TCP connection.

**Architecture:** Extend existing Friends and Chat modules in-place. Add `block_relations` table, new sqlc queries, new protobuf message types in `chat.proto`, a `PresenceNotifier` interface in the chat package, and block-check integration in both Chat and Friends services.

**Tech Stack:** Go, chi v5, sqlc, PostgreSQL, Redis, Protobuf (buf), goose migrations

---

## File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `migrations/008_block_relations.sql` | Create | Block relations table migration |
| `queries/friends.sql` | Modify | Add block + delete-friend-by-users queries |
| `api/proto/chat/v1/chat.proto` | Modify | Add presence/invite message types and payloads |
| `internal/core/errors/codes.go` | Modify | Add `ErrNotImplemented` code |
| `internal/friends/service.go` | Modify | Add Block, Unblock, IsBlocked, ListBlockedUsers, RemoveFriend, ListRecentContacts, ImportPlatformFriends |
| `internal/friends/handler.go` | Modify | Add block/unblock/blocked/delete-friend/recent-contacts HTTP endpoints |
| `internal/friends/service_test.go` | Modify | Add tests for block, unblock, remove friend, is blocked |
| `internal/chat/presence.go` | Create | PresenceNotifier interface + ChatPresenceNotifier implementation |
| `internal/chat/server.go` | Modify | Add SendToUser method |
| `internal/chat/service.go` | Modify | Add block check in handleSendMessage, invite room handler |
| `internal/chat/module.go` | Modify | Expose PresenceNotifier, accept FriendsService reference |
| `internal/friends/module.go` | Modify | Accept PresenceNotifier, wire into service |
| `cmd/server/main.go` | Modify | Wire PresenceNotifier between Chat and Friends modules |
| `internal/auth/mock_test.go` | Modify | Add new Querier method stubs after sqlc regen |
| `internal/users/mock_test.go` | Modify | Add new Querier method stubs after sqlc regen |
| `internal/leaderboard/service_test.go` | Modify | Add new Querier method stubs after sqlc regen |
| `internal/platforms/service_test.go` | Modify | Add new Querier method stubs after sqlc regen |
| `internal/admin/service_test.go` | Modify | Add new Querier method stubs after sqlc regen |
| `internal/friends/service_test.go` | Modify | Add new Querier method stubs after sqlc regen |
| `internal/chat/service_test.go` | Modify | Add new Querier method stubs after sqlc regen (if exists) |

---

### Task 1: Add ErrNotImplemented error code

**Files:**
- Modify: `internal/core/errors/codes.go`

- [ ] **Step 1: Add the error code**

In `internal/core/errors/codes.go`, add after the last `Err` constant:

```go
ErrNotImplemented = "NOT_IMPLEMENTED"
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/core/errors/...`
Expected: success, no errors

- [ ] **Step 3: Commit**

```bash
git add internal/core/errors/codes.go
git commit -m "feat: add ErrNotImplemented error code for Phase 6"
```

---

### Task 2: Create block_relations migration

**Files:**
- Create: `migrations/008_block_relations.sql`

- [ ] **Step 1: Write the migration file**

Create `migrations/008_block_relations.sql`:

```sql
-- +goose Up
CREATE TABLE block_relations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blocker_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_block_unique ON block_relations(blocker_id, blocked_id);
CREATE INDEX idx_block_blocked ON block_relations(blocked_id);

-- +goose Down
DROP TABLE IF EXISTS block_relations;
```

- [ ] **Step 2: Run the migration**

Run: `goose -dir migrations postgres "host=localhost port=5432 user=cyandie password=cyandie dbname=cyandie sslmode=disable" up`
Expected: migration 008 applied successfully

- [ ] **Step 3: Commit**

```bash
git add migrations/008_block_relations.sql
git commit -m "feat: add block_relations migration"
```

---

### Task 3: Add sqlc queries for block + delete-friend-by-users

**Files:**
- Modify: `queries/friends.sql`

- [ ] **Step 1: Add new queries**

Append to `queries/friends.sql`:

```sql
-- name: CreateBlockRelation :one
INSERT INTO block_relations (blocker_id, blocked_id, reason) VALUES ($1, $2, $3) RETURNING *;

-- name: DeleteBlockRelation :one
DELETE FROM block_relations WHERE blocker_id = $1 AND blocked_id = $2 RETURNING *;

-- name: ListBlockedUsers :many
SELECT * FROM block_relations WHERE blocker_id = $1 ORDER BY created_at DESC;

-- name: IsBlockedBy :one
SELECT id FROM block_relations WHERE blocker_id = $1 AND blocked_id = $2;

-- name: DeleteFriendshipByUsers :one
DELETE FROM friendships WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1) RETURNING *;
```

- [ ] **Step 2: Regenerate sqlc code**

Run: `sqlc generate`
Expected: no errors, new Go types and methods generated

- [ ] **Step 3: Verify generated code compiles**

Run: `go build ./internal/db/...`
Expected: success

- [ ] **Step 4: Commit**

```bash
git add queries/friends.sql internal/db/
git commit -m "feat: add block + delete-friend-by-users sqlc queries"
```

---

### Task 4: Fix all mock files for new Querier methods

**Files:**
- Modify: `internal/auth/mock_test.go`
- Modify: `internal/users/mock_test.go`
- Modify: `internal/leaderboard/service_test.go`
- Modify: `internal/platforms/service_test.go`
- Modify: `internal/admin/service_test.go`
- Modify: `internal/friends/service_test.go`

- [ ] **Step 1: Check which new methods were added to Querier**

Run: `go build ./...`
Expected: compilation errors listing missing methods in mock types

- [ ] **Step 2: Add stub implementations to each mock file**

For each mock file that fails compilation, add stub methods matching the new Querier interface. The new methods from sqlc regen will be:

```go
func (m *mockXxxQueries) CreateBlockRelation(_ context.Context, _ db.CreateBlockRelationParams) (db.BlockRelation, error) {
    return db.BlockRelation{}, nil
}
func (m *mockXxxQueries) DeleteBlockRelation(_ context.Context, _ db.DeleteBlockRelationParams) (db.BlockRelation, error) {
    return db.BlockRelation{}, nil
}
func (m *mockXxxQueries) ListBlockedUsers(_ context.Context, _ uuid.UUID) ([]db.BlockRelation, error) {
    return nil, nil
}
func (m *mockXxxQueries) IsBlockedBy(_ context.Context, _ db.IsBlockedByParams) (uuid.UUID, error) {
    return uuid.UUID{}, sql.ErrNoRows
}
func (m *mockXxxQueries) DeleteFriendshipByUsers(_ context.Context, _ db.DeleteFriendshipByUsersParams) (db.Friendship, error) {
    return db.Friendship{}, nil
}
```

Note: The exact parameter types depend on sqlc output. Read `internal/db/querier.go` after regen to confirm.

- [ ] **Step 3: Verify all tests compile**

Run: `go build ./...`
Expected: success, no errors

- [ ] **Step 4: Commit**

```bash
git add internal/auth/mock_test.go internal/users/mock_test.go internal/leaderboard/service_test.go internal/platforms/service_test.go internal/admin/service_test.go internal/friends/service_test.go
git commit -m "fix: add new Querier method stubs to all mock files"
```

---

### Task 5: Extend chat.proto with presence and invite messages

**Files:**
- Modify: `api/proto/chat/v1/chat.proto`

- [ ] **Step 1: Add new MessageType enum values**

In the `MessageType` enum in `api/proto/chat/v1/chat.proto`, add after the existing values:

```protobuf
PRESENCE_ONLINE  = 320;
PRESENCE_OFFLINE = 321;
FRIEND_REMOVED   = 322;
BLOCK_NOTIFY     = 323;
INVITE_ROOM      = 336;
INVITE_ROOM_ACK  = 337;
```

- [ ] **Step 2: Add new message definitions**

Add after the existing message definitions:

```protobuf
message PresenceOnline { string user_id = 1; string username = 2; string status = 3; }
message PresenceOffline { string user_id = 1; }
message FriendRemoved { string user_id = 1; }
message BlockNotify { string blocker_id = 1; }
message InviteRoomRequest { string room_id = 1; string target_user_id = 2; }
message InviteRoomNotify { string room_id = 1; string inviter_id = 2; string inviter_username = 3; }
message InviteRoomAccept { string room_id = 1; }
message InviteRoomReject { string room_id = 1; }
```

- [ ] **Step 3: Add new oneof payload fields in ChatEnvelope**

In the `ChatEnvelope` message, add to the `oneof payload`:

```protobuf
PresenceOnline presence_online = 40;
PresenceOffline presence_offline = 41;
FriendRemoved friend_removed = 42;
BlockNotify block_notify = 43;
InviteRoomRequest invite_room_request = 60;
InviteRoomNotify invite_room_notify = 61;
InviteRoomAccept invite_room_accept = 62;
InviteRoomReject invite_room_reject = 63;
```

- [ ] **Step 4: Regenerate protobuf Go code**

Run: `buf generate`
Expected: no errors, updated Go files in `gen/proto/chat/v1/`

- [ ] **Step 5: Verify generated code compiles**

Run: `go build ./...`
Expected: success

- [ ] **Step 6: Commit**

```bash
git add api/proto/chat/v1/chat.proto gen/proto/
git commit -m "feat: extend chat.proto with presence and invite messages"
```

---

### Task 6: Create PresenceNotifier interface and implementation

**Files:**
- Create: `internal/chat/presence.go`

- [ ] **Step 1: Write the PresenceNotifier interface and implementation**

Create `internal/chat/presence.go`:

```go
package chat

import (
	"context"

	chatv1 "github.com/cyandie/backend/gen/proto/chat/v1"
)

// PresenceNotifier pushes real-time social events to users over TCP.
type PresenceNotifier interface {
	NotifyOnline(ctx context.Context, userID, username string, friendIDs []string)
	NotifyOffline(ctx context.Context, userID string, friendIDs []string)
	NotifyFriendRemoved(ctx context.Context, targetUserID, removedByID string)
	NotifyBlocked(ctx context.Context, blockedUserID, blockerID string)
}

// ChatPresenceNotifier implements PresenceNotifier using the TCP server.
type ChatPresenceNotifier struct {
	srv *TCPServer
}

func NewChatPresenceNotifier(srv *TCPServer) *ChatPresenceNotifier {
	return &ChatPresenceNotifier{srv: srv}
}

func (n *ChatPresenceNotifier) NotifyOnline(_ context.Context, userID, username string, friendIDs []string) {
	frame := &chatv1.ChatEnvelope{
		Type: chatv1.MessageType_PRESENCE_ONLINE,
		Payload: &chatv1.ChatEnvelope_PresenceOnline{
			PresenceOnline: &chatv1.PresenceOnline{
				UserId:   userID,
				Username: username,
			},
		},
	}
	for _, fid := range friendIDs {
		n.srv.SendToUser(fid, frame)
	}
}

func (n *ChatPresenceNotifier) NotifyOffline(_ context.Context, userID string, friendIDs []string) {
	frame := &chatv1.ChatEnvelope{
		Type: chatv1.MessageType_PRESENCE_OFFLINE,
		Payload: &chatv1.ChatEnvelope_PresenceOffline{
			PresenceOffline: &chatv1.PresenceOffline{
				UserId: userID,
			},
		},
	}
	for _, fid := range friendIDs {
		n.srv.SendToUser(fid, frame)
	}
}

func (n *ChatPresenceNotifier) NotifyFriendRemoved(_ context.Context, targetUserID, removedByID string) {
	frame := &chatv1.ChatEnvelope{
		Type: chatv1.MessageType_FRIEND_REMOVED,
		Payload: &chatv1.ChatEnvelope_FriendRemoved{
			FriendRemoved: &chatv1.FriendRemoved{
				UserId: removedByID,
			},
		},
	}
	n.srv.SendToUser(targetUserID, frame)
}

func (n *ChatPresenceNotifier) NotifyBlocked(_ context.Context, blockedUserID, blockerID string) {
	frame := &chatv1.ChatEnvelope{
		Type: chatv1.MessageType_BLOCK_NOTIFY,
		Payload: &chatv1.ChatEnvelope_BlockNotify{
			BlockNotify: &chatv1.BlockNotify{
				BlockerId: blockerID,
			},
		},
	}
	n.srv.SendToUser(blockedUserID, frame)
}
```

- [ ] **Step 2: Add SendToUser method to TCPServer**

In `internal/chat/server.go`, add the `SendToUser` method to `TCPServer`:

```go
// SendToUser sends a protobuf frame to a specific user if they are connected.
func (s *TCPServer) SendToUser(userID string, msg *chatv1.ChatEnvelope) {
	s.mu.RLock()
	conn, ok := s.conns[userID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return
	}
	frame := EncodeTLV(uint16(msg.Type), data)
	conn.Write(frame)
}
```

Note: Read the existing `server.go` to confirm the exact field names for `s.conns`, `s.mu`, and the `EncodeTLV` function signature. Adjust accordingly.

- [ ] **Step 3: Verify it compiles**

Run: `go build ./internal/chat/...`
Expected: success

- [ ] **Step 4: Commit**

```bash
git add internal/chat/presence.go internal/chat/server.go
git commit -m "feat: add PresenceNotifier interface and SendToUser"
```

---

### Task 7: Add block/unblock/is-blocked service methods

**Files:**
- Modify: `internal/friends/service.go`

- [ ] **Step 1: Add block-related methods to FriendsService**

Add these methods to `internal/friends/service.go`. Read the existing file first to confirm the struct fields and constructor signature.

```go
// Block creates a block relation, removes friendship if exists, rejects pending requests.
func (s *FriendsService) Block(ctx context.Context, blockerID, blockedID, reason string) error {
	bid, err := uuid.Parse(blockerID)
	if err != nil {
		return coreerrors.New(coreerrors.ErrBadRequest, "invalid blocker id")
	}
	blid, err := uuid.Parse(blockedID)
	if err != nil {
		return coreerrors.New(coreerrors.ErrBadRequest, "invalid blocked id")
	}
	if blockerID == blockedID {
		return coreerrors.New(coreerrors.ErrBadRequest, "cannot block yourself")
	}

	_, err = s.q.CreateBlockRelation(ctx, db.CreateBlockRelationParams{
		BlockerID: bid,
		BlockedID: blid,
		Reason:    sql.NullString{String: reason, Valid: reason != ""},
	})
	if err != nil {
		return coreerrors.Wrap(err, coreerrors.ErrInternal, "create block relation")
	}

	// Delete friendship if exists (ignore error if not found)
	_, _ = s.q.DeleteFriendshipByUsers(ctx, db.DeleteFriendshipByUsersParams{
		UserID:   bid,
		FriendID: blid,
	})

	// Reject pending friend requests in both directions (ignore error if not found)
	pending, _ := s.q.ListPendingRequests(ctx, bid)
	for _, f := range pending {
		if f.Status == "pending" && (f.FriendID == blid || f.UserID == blid) {
			_, _ = s.q.UpdateFriendshipStatus(ctx, db.UpdateFriendshipStatusParams{
				ID:     f.ID,
				Status: "rejected",
			})
		}
	}

	// Update Redis cache
	if s.rdb != nil {
		s.rdb.SAdd(ctx, fmt.Sprintf("friends:blocked:%s", blockerID), blockedID)
	}

	// Notify blocked user via presence notifier
	if s.notifier != nil {
		s.notifier.NotifyBlocked(ctx, blockedID, blockerID)
	}

	return nil
}

// Unblock removes a block relation.
func (s *FriendsService) Unblock(ctx context.Context, blockerID, blockedID string) error {
	bid, err := uuid.Parse(blockerID)
	if err != nil {
		return coreerrors.New(coreerrors.ErrBadRequest, "invalid blocker id")
	}
	blid, err := uuid.Parse(blockedID)
	if err != nil {
		return coreerrors.New(coreerrors.ErrBadRequest, "invalid blocked id")
	}

	_, err = s.q.DeleteBlockRelation(ctx, db.DeleteBlockRelationParams{
		BlockerID: bid,
		BlockedID: blid,
	})
	if err != nil {
		return coreerrors.Wrap(err, coreerrors.ErrInternal, "delete block relation")
	}

	// Remove from Redis cache
	if s.rdb != nil {
		s.rdb.SRem(ctx, fmt.Sprintf("friends:blocked:%s", blockerID), blockedID)
	}

	return nil
}

// IsBlocked checks if targetUserID has blocked byUserID (i.e., byUserID cannot interact with targetUserID).
func (s *FriendsService) IsBlocked(ctx context.Context, targetUserID, byUserID string) (bool, error) {
	// Check Redis first
	if s.rdb != nil {
		isMember, err := s.rdb.SIsMember(ctx, fmt.Sprintf("friends:blocked:%s", targetUserID), byUserID).Result()
		if err == nil {
			return isMember, nil
		}
	}

	// Fallback to DB
	tid, err := uuid.Parse(targetUserID)
	if err != nil {
		return false, coreerrors.New(coreerrors.ErrBadRequest, "invalid target user id")
	}
	bid, err := uuid.Parse(byUserID)
	if err != nil {
		return false, coreerrors.New(coreerrors.ErrBadRequest, "invalid by user id")
	}

	_, err = s.q.IsBlockedBy(ctx, db.IsBlockedByParams{
		BlockerID: tid,
		BlockedID: bid,
	})
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, coreerrors.Wrap(err, coreerrors.ErrInternal, "check block relation")
	}
	return true, nil
}

// ListBlockedUsers returns all users blocked by the given user.
func (s *FriendsService) ListBlockedUsers(ctx context.Context, userID string) ([]db.BlockRelation, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, coreerrors.New(coreerrors.ErrBadRequest, "invalid user id")
	}
	return s.q.ListBlockedUsers(ctx, uid)
}
```

- [ ] **Step 2: Update FriendsService struct to include notifier field**

In the `FriendsService` struct, add:

```go
notifier chat.PresenceNotifier
```

Update the `NewFriendsService` constructor to accept and store the notifier:

```go
func NewFriendsService(q db.Querier, rdb *redis.Client, notifier chat.PresenceNotifier) *FriendsService {
	return &FriendsService{q: q, rdb: rdb, notifier: notifier}
}
```

Note: Read the existing `service.go` to confirm the current struct fields and constructor. The `notifier` may be nil initially and checked before use.

- [ ] **Step 3: Add required imports**

Ensure these imports are present:

```go
"database/sql"
"fmt"

"github.com/cyandie/backend/internal/chat"
```

- [ ] **Step 4: Verify it compiles**

Run: `go build ./internal/friends/...`
Expected: may fail because module.go still calls old constructor — will fix in Task 12

- [ ] **Step 5: Commit**

```bash
git add internal/friends/service.go
git commit -m "feat: add block/unblock/is-blocked/list-blocked service methods"
```

---

### Task 8: Add RemoveFriend and ListRecentContacts service methods

**Files:**
- Modify: `internal/friends/service.go`

- [ ] **Step 1: Add RemoveFriend method**

```go
// RemoveFriend deletes the friendship between two users and notifies the other party.
func (s *FriendsService) RemoveFriend(ctx context.Context, userID, friendID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return coreerrors.New(coreerrors.ErrBadRequest, "invalid user id")
	}
	fid, err := uuid.Parse(friendID)
	if err != nil {
		return coreerrors.New(coreerrors.ErrBadRequest, "invalid friend id")
	}

	_, err = s.q.DeleteFriendshipByUsers(ctx, db.DeleteFriendshipByUsersParams{
		UserID:   uid,
		FriendID: fid,
	})
	if err != nil {
		return coreerrors.Wrap(err, coreerrors.ErrInternal, "delete friendship")
	}

	// Notify the removed friend
	if s.notifier != nil {
		s.notifier.NotifyFriendRemoved(ctx, friendID, userID)
	}

	return nil
}
```

- [ ] **Step 2: Add ListRecentContacts method**

```go
// ListRecentContacts returns recent contacts sorted by last interaction time.
func (s *FriendsService) ListRecentContacts(ctx context.Context, userID string, limit int) ([]string, error) {
	if s.rdb == nil {
		return nil, coreerrors.New(coreerrors.ErrInternal, "redis not available")
	}

	key := fmt.Sprintf("friends:recent:%s", userID)
	now := float64(time.Now().Unix())

	// Remove entries older than 30 days
	cutoff := now - 30*24*3600
	s.rdb.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%f", cutoff))

	// Get top N recent contacts
	results, err := s.rdb.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{
		Max:   fmt.Sprintf("%f", now),
		Min:   "0",
		Count: int64(limit),
	}).Result()
	if err != nil {
		return nil, coreerrors.Wrap(err, coreerrors.ErrInternal, "get recent contacts")
	}

	return results, nil
}
```

- [ ] **Step 3: Add required imports**

Ensure `time` is imported.

- [ ] **Step 4: Verify it compiles**

Run: `go build ./internal/friends/...`
Expected: may still fail due to module.go — will fix in Task 12

- [ ] **Step 5: Commit**

```bash
git add internal/friends/service.go
git commit -m "feat: add RemoveFriend and ListRecentContacts service methods"
```

---

### Task 9: Add ImportPlatformFriends stub

**Files:**
- Modify: `internal/friends/service.go`

- [ ] **Step 1: Add the stub method**

```go
// PlatformFriendInfo represents a friend from an external platform.
type PlatformFriendInfo struct {
	PlatformUserID string
	Username       string
}

// ImportPlatformFriends is a stub for future platform friend import.
func (s *FriendsService) ImportPlatformFriends(_ context.Context, _, _ string, _ []PlatformFriendInfo) error {
	return coreerrors.New(coreerrors.ErrNotImplemented, "platform friend import is not yet available")
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/friends/service.go
git commit -m "feat: add ImportPlatformFriends stub"
```

---

### Task 10: Add block check in Friends SendRequest

**Files:**
- Modify: `internal/friends/service.go`

- [ ] **Step 1: Add IsBlocked check at the start of SendRequest**

At the beginning of the existing `SendRequest` method, after the self-check, add:

```go
blocked, err := s.IsBlocked(ctx, toUserID, fromUserID)
if err != nil {
	return nil, err
}
if blocked {
	return nil, coreerrors.New(coreerrors.ErrForbidden, "you are blocked by this user")
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/friends/service.go
git commit -m "feat: add block check in SendRequest"
```

---

### Task 11: Add HTTP handler endpoints for block, delete friend, recent contacts

**Files:**
- Modify: `internal/friends/handler.go`

- [ ] **Step 1: Add block/unblock/blocked handler methods**

Read the existing `handler.go` to confirm the handler struct and `writeJSON`/`writeAppError` patterns.

```go
func (h *FriendsHandler) Block(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	var req struct {
		UserID string `json:"user_id"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, coreerrors.New(coreerrors.ErrBadRequest, "invalid request body"))
		return
	}
	if err := h.svc.Block(r.Context(), userID, req.UserID, req.Reason); err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FriendsHandler) Unblock(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	blockedUserID := chi.URLParam(r, "userID")
	if err := h.svc.Unblock(r.Context(), userID, blockedUserID); err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FriendsHandler) ListBlocked(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	blocked, err := h.svc.ListBlockedUsers(r.Context(), userID)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, blocked)
}

func (h *FriendsHandler) RemoveFriend(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	friendUserID := chi.URLParam(r, "userID")
	if err := h.svc.RemoveFriend(r.Context(), userID, friendUserID); err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FriendsHandler) ListRecentContacts(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	contacts, err := h.svc.ListRecentContacts(r.Context(), userID, limit)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, contacts)
}
```

- [ ] **Step 2: Add required imports**

Ensure these are imported:

```go
"strconv"

"github.com/go-chi/chi/v5"
```

- [ ] **Step 3: Update Routes() to include new endpoints**

In the `Routes()` method, add these routes:

```go
r.Post("/block", h.Block)
r.Delete("/block/{userID}", h.Unblock)
r.Get("/blocked", h.ListBlocked)
r.Delete("/{userID}", h.RemoveFriend)
r.Get("/recent", h.ListRecentContacts)
```

Note: The `/{userID}` route for RemoveFriend must be placed after `/{id}/accept` and `/{id}/reject` to avoid conflicts. Read the existing Routes() to confirm the order.

- [ ] **Step 4: Verify it compiles**

Run: `go build ./internal/friends/...`
Expected: may still fail due to module.go — will fix in Task 12

- [ ] **Step 5: Commit**

```bash
git add internal/friends/handler.go
git commit -m "feat: add block/unblock/delete-friend/recent-contacts HTTP endpoints"
```

---

### Task 12: Update Friends module to accept PresenceNotifier

**Files:**
- Modify: `internal/friends/module.go`

- [ ] **Step 1: Update the module to accept and pass notifier**

Read the existing `module.go` to confirm the current structure. Update the `NewModule` function to accept a `chat.PresenceNotifier` and pass it to `NewFriendsService`:

```go
func NewModule(q db.Querier, rdb *redis.Client, notifier chat.PresenceNotifier) *Module {
	// ...
	svc := NewFriendsService(q, rdb, notifier)
	// ...
}
```

Add the import for the chat package:

```go
"github.com/cyandie/backend/internal/chat"
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/friends/...`
Expected: success

- [ ] **Step 3: Commit**

```bash
git add internal/friends/module.go
git commit -m "feat: update Friends module to accept PresenceNotifier"
```

---

### Task 13: Add block check in Chat handleSendMessage

**Files:**
- Modify: `internal/chat/service.go`

- [ ] **Step 1: Add FriendsService interface and field**

In `internal/chat/service.go`, add an interface for the block check (to avoid circular import):

```go
// BlockChecker checks if a user is blocked by another user.
type BlockChecker interface {
	IsBlocked(ctx context.Context, targetUserID, byUserID string) (bool, error)
}
```

Add a `blockChecker BlockChecker` field to the `ChatService` struct.

Update the `NewChatService` constructor to accept the block checker:

```go
func NewChatService(q db.Querier, srv *TCPServer, blockChecker BlockChecker) *ChatService {
	return &ChatService{q: q, srv: srv, blockChecker: blockChecker}
}
```

- [ ] **Step 2: Add block check in handleSendMessage**

In the `handleSendMessage` method, before sending a private message, add:

```go
if msg.RoomId == "" && msg.TargetUserId != "" {
	blocked, err := s.blockChecker.IsBlocked(ctx, msg.TargetUserId, senderID)
	if err != nil {
		s.sendError(conn, "INTERNAL", "block check failed")
		return
	}
	if blocked {
		s.sendError(conn, "FORBIDDEN", "you are blocked by this user")
		return
	}
}
```

Note: Read the existing `handleSendMessage` to confirm the exact field names and the private message detection logic. Adjust accordingly.

- [ ] **Step 3: Verify it compiles**

Run: `go build ./internal/chat/...`
Expected: may fail because module.go still calls old constructor — will fix in Task 14

- [ ] **Step 4: Commit**

```bash
git add internal/chat/service.go
git commit -m "feat: add block check in Chat handleSendMessage"
```

---

### Task 14: Update Chat module to accept BlockChecker and expose PresenceNotifier

**Files:**
- Modify: `internal/chat/module.go`

- [ ] **Step 1: Update the module**

Read the existing `module.go`. Update `NewModule` to accept a `BlockChecker` and pass it to `NewChatService`. Also expose a method to get the `PresenceNotifier`:

```go
func NewModule(q db.Querier, tcpAddr string, blockChecker BlockChecker) *Module {
	// ...
	svc := NewChatService(q, srv, blockChecker)
	notifier := NewChatPresenceNotifier(srv)
	// ...
}

func (m *Module) PresenceNotifier() chat.PresenceNotifier {
	return m.notifier
}
```

Add a `notifier` field to the Module struct.

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/chat/...`
Expected: success

- [ ] **Step 3: Commit**

```bash
git add internal/chat/module.go
git commit -m "feat: update Chat module to accept BlockChecker and expose PresenceNotifier"
```

---

### Task 15: Wire everything together in main.go

**Files:**
- Modify: `cmd/server/main.go`

- [ ] **Step 1: Update module creation order**

The key change: Chat module must be created first (with nil block checker initially), then Friends module gets the PresenceNotifier from Chat, then Chat gets the Friends service as block checker.

Read the existing `main.go` to confirm the current structure. The wiring will be:

```go
// Create chat module first (with nil block checker temporarily)
chatModule := chat.NewModule(queries, ":9091", nil)
app.Register(chatModule)

// Create friends module with PresenceNotifier from chat
friendsModule := friends.NewModule(queries, rdb.Client, chatModule.PresenceNotifier())
app.Register(friendsModule)

// Wire block checker back to chat
chatModule.SetBlockChecker(friendsModule.BlockChecker())
```

This requires:
1. Adding a `SetBlockChecker` method to the Chat module
2. Adding a `BlockChecker()` method to the Friends module that returns the service as a `chat.BlockChecker`

- [ ] **Step 2: Add SetBlockChecker to Chat module**

In `internal/chat/module.go`, add:

```go
func (m *Module) SetBlockChecker(checker BlockChecker) {
	m.svc.setBlockChecker(checker)
}
```

In `internal/chat/service.go`, add:

```go
func (s *ChatService) setBlockChecker(checker BlockChecker) {
	s.blockChecker = checker
}
```

- [ ] **Step 3: Add BlockChecker() to Friends module**

In `internal/friends/module.go`, add:

```go
func (m *Module) BlockChecker() chat.BlockChecker {
	return m.svc
}
```

This works because `FriendsService` already has the `IsBlocked` method that satisfies the `chat.BlockChecker` interface.

- [ ] **Step 4: Update main.go**

Update the module creation section:

```go
chatModule := chat.NewModule(queries, ":9091", nil)
app.Register(chatModule)

friendsModule := friends.NewModule(queries, rdb.Client, chatModule.PresenceNotifier())
app.Register(friendsModule)

chatModule.SetBlockChecker(friendsModule.BlockChecker())
```

Remove the old `chatModule` and `friendsModule` creation lines.

- [ ] **Step 5: Verify it compiles**

Run: `go build ./cmd/server/...`
Expected: success

- [ ] **Step 6: Commit**

```bash
git add cmd/server/main.go internal/chat/module.go internal/chat/service.go internal/friends/module.go
git commit -m "feat: wire PresenceNotifier and BlockChecker between Chat and Friends"
```

---

### Task 16: Add invite room handler in Chat service

**Files:**
- Modify: `internal/chat/service.go`

- [ ] **Step 1: Add handleInviteRoom method**

```go
func (s *ChatService) handleInviteRoom(ctx context.Context, conn net.Conn, userID string, envelope *chatv1.ChatEnvelope) {
	req := envelope.GetInviteRoomRequest()
	if req == nil {
		s.sendError(conn, "BAD_REQUEST", "invalid invite room request")
		return
	}

	// Verify inviter is in the room
	members, err := s.q.GetRoomMembers(ctx, uuid.MustParse(req.RoomId))
	if err != nil {
		s.sendError(conn, "NOT_FOUND", "room not found")
		return
	}
	inRoom := false
	for _, m := range members {
		if m.UserID.String() == userID {
			inRoom = true
			break
		}
	}
	if !inRoom {
		s.sendError(conn, "FORBIDDEN", "you are not in this room")
		return
	}

	// Forward invite to target user
	notify := &chatv1.ChatEnvelope{
		Type: chatv1.MessageType_INVITE_ROOM,
		Payload: &chatv1.ChatEnvelope_InviteRoomNotify{
			InviteRoomNotify: &chatv1.InviteRoomNotify{
				RoomId:         req.RoomId,
				InviterId:      userID,
				InviterUsername: "", // Could look up from DB if needed
			},
		},
	}
	s.srv.SendToUser(req.TargetUserId, notify)
}
```

- [ ] **Step 2: Register invite room handler in the message dispatcher**

In the message handling switch/dispatch in `service.go`, add a case for `INVITE_ROOM`:

```go
case chatv1.MessageType_INVITE_ROOM:
	s.handleInviteRoom(ctx, conn, userID, envelope)
```

Note: Read the existing message dispatch logic to confirm the exact pattern.

- [ ] **Step 3: Verify it compiles**

Run: `go build ./internal/chat/...`
Expected: success

- [ ] **Step 4: Commit**

```bash
git add internal/chat/service.go
git commit -m "feat: add invite room handler in Chat service"
```

---

### Task 17: Update online presence push in Friends service

**Files:**
- Modify: `internal/friends/service.go`

- [ ] **Step 1: Update SetOnline to detect state change and push**

The existing `SetOnline` method sets a Redis key. Update it to detect first-online and push to friends:

```go
func (s *FriendsService) SetOnline(ctx context.Context, userID, username string) error {
	if s.rdb == nil {
		return nil
	}
	key := fmt.Sprintf("friends:online:%s", userID)
	// SETNX returns true if the key was set (first online)
	set, err := s.rdb.SetNX(ctx, key, username, 5*time.Minute).Result()
	if err != nil {
		return coreerrors.Wrap(err, coreerrors.ErrInternal, "set online")
	}
	if set && s.notifier != nil {
		// Just came online — notify friends
		friends, _ := s.ListFriends(ctx, userID)
		friendIDs := make([]string, 0, len(friends))
		for _, f := range friends {
			friendIDs = append(friendIDs, f.FriendID.String())
		}
		s.notifier.NotifyOnline(ctx, userID, username, friendIDs)
	}
	// Refresh TTL
	s.rdb.Expire(ctx, key, 5*time.Minute)
	return nil
}
```

- [ ] **Step 2: Update SetOffline similarly**

```go
func (s *FriendsService) SetOffline(ctx context.Context, userID string) error {
	if s.rdb == nil {
		return nil
	}
	key := fmt.Sprintf("friends:online:%s", userID)
	// Check if currently online before removing
	exists, err := s.rdb.Exists(ctx, key).Result()
	if err != nil {
		return coreerrors.Wrap(err, coreerrors.ErrInternal, "check online")
	}
	s.rdb.Del(ctx, key)
	if exists > 0 && s.notifier != nil {
		friends, _ := s.ListFriends(ctx, userID)
		friendIDs := make([]string, 0, len(friends))
		for _, f := range friends {
			friendIDs = append(friendIDs, f.FriendID.String())
		}
		s.notifier.NotifyOffline(ctx, userID, friendIDs)
	}
	return nil
}
```

- [ ] **Step 3: Verify it compiles**

Run: `go build ./internal/friends/...`
Expected: success

- [ ] **Step 4: Commit**

```bash
git add internal/friends/service.go
git commit -m "feat: add online/offline presence push to friends"
```

---

### Task 18: Write tests for block functionality

**Files:**
- Modify: `internal/friends/service_test.go`

- [ ] **Step 1: Add block-related fields to mockFriendsQueries**

Add fields for block query results:

```go
type mockFriendsQueries struct {
	// existing fields...
	blockRelation    db.BlockRelation
	blockErr         error
	blockedList      []db.BlockRelation
	blockedListErr   error
	isBlockedID      uuid.UUID
	isBlockedErr     error
	deleteByUsers    db.Friendship
	deleteByUsersErr error
}
```

- [ ] **Step 2: Add new mock method implementations**

```go
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
	return m.deleteByUsers, m.deleteByUsersErr
}
```

- [ ] **Step 3: Write test for Block**

```go
func TestFriendsService_Block(t *testing.T) {
	blocker := uuid.New()
	blocked := uuid.New()
	q := &mockFriendsQueries{
		blockRelation: db.BlockRelation{BlockerID: blocker, BlockedID: blocked},
		friendErr:     sql.ErrNoRows,
	}
	svc := NewFriendsService(q, nil, nil)

	err := svc.Block(context.Background(), blocker.String(), blocked.String(), "spam")
	if err != nil {
		t.Fatalf("Block failed: %v", err)
	}
}
```

- [ ] **Step 4: Write test for Block self**

```go
func TestFriendsService_Block_Self(t *testing.T) {
	svc := NewFriendsService(&mockFriendsQueries{}, nil, nil)
	uid := uuid.New().String()

	err := svc.Block(context.Background(), uid, uid, "")
	if err == nil {
		t.Error("expected error for self-block")
	}
}
```

- [ ] **Step 5: Write test for Unblock**

```go
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
```

- [ ] **Step 6: Write test for IsBlocked**

```go
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
```

- [ ] **Step 7: Write test for RemoveFriend**

```go
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
```

- [ ] **Step 8: Run tests**

Run: `go test ./internal/friends/... -v`
Expected: all tests pass

- [ ] **Step 9: Commit**

```bash
git add internal/friends/service_test.go
git commit -m "test: add block/unblock/is-blocked/remove-friend tests"
```

---

### Task 19: Full build and test verification

**Files:**
- None (verification only)

- [ ] **Step 1: Run full build**

Run: `go build ./...`
Expected: success

- [ ] **Step 2: Run all tests**

Run: `go test ./... -v`
Expected: all tests pass

- [ ] **Step 3: Run the server and test health endpoint**

```bash
# Start server
go run ./cmd/server/
# In another terminal:
curl http://localhost:8080/health
```
Expected: health response

- [ ] **Step 4: Commit any remaining fixes**

If any fixes were needed, commit them.

---

### Task 20: Final commit with Phase 6 tag

**Files:**
- None

- [ ] **Step 1: Create git tag**

```bash
git tag -a phase6 -m "Phase 6: Social Presence - block, delete friend, presence push, invite protocol"
```

- [ ] **Step 2: Verify clean state**

Run: `git status`
Expected: clean working tree
