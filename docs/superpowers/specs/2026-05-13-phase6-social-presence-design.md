# Phase 6: Social Presence Design

## Scope

Phase 6 补齐好友模块缺失能力，并实现实时在线状态推送。MVP 范围：

- 黑名单（Block/Unblock，双向阻断好友申请和私聊）
- 删除好友
- 黑名单影响私聊（Chat 模块调用 Friends 检查）
- 在线状态实时推送（复用 Chat TCP 连接 + Protobuf）
- 最近联系人
- 邀请进入房间协议
- 平台好友导入预留接口

**不在 MVP 范围**：分组、备注、推荐好友、好友申请消息。

## Approach

方案 A：在现有 Friends 模块上扩展，不新建模块。Chat 模块新增 presence 消息类型和 PresenceNotifier 接口。Friends 服务通过 notifier 推送状态变化。

## Data Model

### block_relation（新增迁移）

```sql
CREATE TABLE block_relations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blocker_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_block_unique ON block_relations(blocker_id, blocked_id);
CREATE INDEX idx_block_blocked ON block_relations(blocked_id);
```

### Redis Keys

| Key | Type | Purpose | TTL |
|-----|------|---------|-----|
| `friends:blocked:{user_id}` | Set | 被该用户拉黑的所有 user_id 列表 | 1h |
| `friends:recent:{user_id}` | Sorted Set | 最近联系人，score=最后互动时间戳 | 30d auto-filter |
| `friends:online:{user_id}` | String | 在线状态心跳（已实现） | 5min |

## HTTP APIs

### 黑名单

```
POST   /api/v1/friends/block          body: {"user_id": "..."}
DELETE /api/v1/friends/block/{user_id}
GET    /api/v1/friends/blocked
```

拉黑时：
1. 创建 block_relation
2. 如果存在好友关系，删除双方 friendship
3. 如果存在待处理好友申请（任何方向），拒绝
4. 写入 Redis 缓存 `friends:blocked:{blocker_id}` 添加 blocked_id
5. 如果被拉黑用户在线，通过 TCP 推送 `BLOCK_NOTIFY`

取消拉黑时：
1. 删除 block_relation
2. 从 Redis 缓存移除
3. 不自动恢复好友关系

### 删除好友

```
DELETE /api/v1/friends/{user_id}
```

删除时：
1. 删除双方 friendship 行
2. 如果对方在线，推送 `FRIEND_REMOVED`

### 最近联系人

```
GET /api/v1/friends/recent?limit=20
```

返回最近联系人列表，含在线状态。score < now-30d 的条目在查询时过滤并删除。

## Protobuf Messages

在 `api/proto/chat/v1/chat.proto` 中扩展：

### MessageType 新增

```protobuf
PRESENCE_ONLINE  = 320;
PRESENCE_OFFLINE = 321;
FRIEND_REMOVED   = 322;
BLOCK_NOTIFY     = 323;
INVITE_ROOM      = 336;
INVITE_ROOM_ACK  = 337;
```

### Message 定义

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

### ChatEnvelope.payload 新增

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

## Online Presence Push

### 推送流程

1. 用户 heartbeat 时调用 `FriendsService.SetOnline(ctx, userID)`
2. 检查是否刚上线：Redis `SETNX` `friends:online:{user_id}`，返回 NX=true 表示刚上线
3. 如果刚上线 → 查好友列表 → 遍历好友 → 查好友的 TCP 连接 → 发送 `PRESENCE_ONLINE`
4. 用户下线时同理推送 `PRESENCE_OFFLINE`
5. heartbeat 超时（5min TTL）后 Redis key 自动过期，下次有人查询时发现 offline

### PresenceNotifier 接口

```go
// internal/chat/presence.go
type PresenceNotifier interface {
    NotifyOnline(ctx context.Context, userID, username string, friendIDs []string)
    NotifyOffline(ctx context.Context, userID string, friendIDs []string)
    NotifyFriendRemoved(ctx context.Context, targetUserID, removedByID string)
    NotifyBlocked(ctx context.Context, blockedUserID, blockerID string)
}
```

Chat 模块实现该接口，通过 TCPServer 的 `SendToUser` 方法定向推送。

Friends 模块在状态变化时调用 notifier。

## Block Impact on Chat

### Chat 私聊阻断

Chat 模块在 `handleSendMessage` 中新增检查：

```go
// 在发送私聊消息前
blocked, err := friendsService.IsBlocked(ctx, senderID, receiverID)
if blocked {
    sendError(conn, "FORBIDDEN", "you are blocked by this user")
    return
}
```

`IsBlocked` 实现：
1. 查 Redis `friends:blocked:{receiver_id}` 是否包含 senderID
2. miss 时查 DB `block_relations WHERE blocker_id=receiver AND blocked_id=sender`
3. 回填 Redis 缓存

### Friends 好友申请阻断

`SendRequest` 中新增检查：

```go
blocked, err := s.IsBlocked(ctx, toUserID, fromUserID)
if blocked {
    return nil, errors.New(errors.ErrForbidden, "you are blocked by this user")
}
```

## Invite Protocol

### 流程

1. 用户 A 发送 `INVITE_ROOM` TCP 消息：`InviteRoomRequest { room_id, target_user_id }`
2. 服务端检查：A 是否在房间中、B 是否是好友、B 是否在线
3. 通过 TCP 转发给 B：`InviteRoomNotify { room_id, inviter_id, inviter_username }`
4. B 回复 `InviteRoomAccept` 或 `InviteRoomReject`
5. Accept 时服务端自动将 B 加入房间，返回 `JOIN_ROOM_OK`

## Platform Friends Import Stub

```go
func (s *FriendsService) ImportPlatformFriends(ctx context.Context, userID, provider string, platformFriends []PlatformFriendInfo) error {
    return errors.New(errors.ErrNotImplemented, "platform friend import is not yet available")
}
```

仅定义接口形状，不做实现。

## sqlc 新增查询

### migrations/006_block_relations.sql

```sql
-- +goose Up
CREATE TABLE block_relations (...);

-- +goose Down
DROP TABLE IF EXISTS block_relations;
```

### queries/friends.sql 新增

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

## Tests

- 拉黑用户成功
- 拉黑时自动删除好友关系
- 拉黑时自动拒绝好友申请
- 取消拉黑成功
- 被拉黑用户不能发送好友申请（返回 403）
- 被拉黑用户不能私聊（Chat 检查返回 403）
- 删除好友成功
- 删除好友后对方收到 FRIEND_REMOVED 推送
- 最近联系人列表返回正确
- 上线时好友收到 PRESENCE_ONLINE 推送
- 下线时好友收到 PRESENCE_OFFLINE 推送