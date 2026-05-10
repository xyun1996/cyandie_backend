# API Contracts

本文档定义 API 约定。实现时必须保持一致。

## Versioning

- 客户端公开 API 使用 `/api/v1`。
- 后台 API 使用 `/api/v1/admin`。
- TCP 消息和 gRPC service 也应有版本兼容策略，破坏性变更需要新增消息类型、字段或版本。

## Response Envelope

成功：

```json
{
  "ok": true,
  "data": {}
}
```

失败：

```json
{
  "ok": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "requestId": "req_..."
  }
}
```

分页：

```json
{
  "ok": true,
  "data": [],
  "page": {
    "limit": 50,
    "nextCursor": "..."
  }
}
```

## Error Code Families

```text
AUTH_*
USER_*
PLATFORM_*
CHAT_*
LEADERBOARD_*
FRIEND_*
PAYMENT_*
ADMIN_*
VALIDATION_*
RATE_LIMIT_*
INTERNAL_*
```

## Authentication

客户端 API：

```http
Authorization: Bearer <access_token>
```

后台 API：

```http
Authorization: Bearer <admin_access_token>
```

Refresh Token 不应作为普通 API 鉴权凭证。

## Idempotency

需要幂等的接口：

- 支付回调。
- 收据校验。
- 奖励发放。
- 排行榜分数提交，可选。
- 账号绑定，可选。

推荐请求头：

```http
Idempotency-Key: <unique-key>
```

## Rate Limiting

MVP 至少对以下接口限流：

- 登录。
- 刷新 Token。
- 聊天发送。
- 排行榜提交。
- 平台账号绑定。

## Client APIs MVP

Auth：

```text
POST /api/v1/auth/guest
POST /api/v1/auth/platform/:provider
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
GET  /api/v1/me
PATCH /api/v1/me/profile
```

Platform binding：

```text
POST   /api/v1/me/platforms/:provider/link
DELETE /api/v1/me/platforms/:provider
GET    /api/v1/me/platforms
```

Leaderboard：

```text
POST /api/v1/leaderboards/:boardId/scores
GET  /api/v1/leaderboards/:boardId/top
GET  /api/v1/leaderboards/:boardId/me
GET  /api/v1/leaderboards/:boardId/around-me
```

Chat：

```text
GET  /api/v1/chat/conversations
GET  /api/v1/chat/conversations/:id/messages
POST /api/v1/chat/reports
```

Friends：

```text
POST   /api/v1/friends/requests
GET    /api/v1/friends/requests
POST   /api/v1/friends/requests/:id/accept
POST   /api/v1/friends/requests/:id/reject
GET    /api/v1/friends
DELETE /api/v1/friends/:userId
POST   /api/v1/blocks/:userId
DELETE /api/v1/blocks/:userId
```

## Admin APIs MVP

```text
GET  /api/v1/admin/users
GET  /api/v1/admin/users/:id
POST /api/v1/admin/users/:id/ban
POST /api/v1/admin/users/:id/unban
GET  /api/v1/admin/audit-logs
GET  /api/v1/admin/chat/reports
POST /api/v1/admin/chat/mutes
GET  /api/v1/admin/leaderboards
POST /api/v1/admin/leaderboards
PATCH /api/v1/admin/leaderboards/:id
```

## gRPC Contract

gRPC 主要用于服务间通信、内部工具和后续服务拆分。

约束：

- Protobuf 文件放在 `protocol/proto/`。
- 字段编号一旦发布不得复用。
- 删除字段必须使用 `reserved`。
- gRPC 错误必须映射到统一业务错误码。
- 公开给其他服务的接口必须有契约测试或集成测试。

建议服务：

```text
AuthService
UserService
PlatformService
ChatService
LeaderboardService
AdminService
```

## TCP Chat Contract

连接：

```text
tcp://host:port
```

认证流程：

```text
Client connects
Client sends AUTH payload with access token
Server replies AUTH_OK or AUTH_REJECTED
```

包格式建议：

```text
magic        uint16
version      uint16
messageType  uint32
requestId    uint64
payloadLen   uint32
payload       protobuf bytes
```

消息 payload 使用 Protobuf。客户端发包示例概念：

```json
{
  "messageType": "CHAT_MESSAGE_SEND",
  "requestId": "req_...",
  "payload": {}
}
```

错误消息：

```json
{
  "messageType": "ERROR",
  "requestId": "req_...",
  "error": {
    "code": "CHAT_MUTED",
    "message": "User is muted"
  }
}
```
