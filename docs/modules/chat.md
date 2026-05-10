# Chat

聊天模块负责基于 TCP 长连接的实时消息、私聊、频道、群聊、离线消息、禁言、举报和基础审核。

## Goals

- 支持游戏、应用、社区和工具场景内的实时聊天。
- 先保证可靠、可审计、可扩展。
- TCP 长连接负责实时投递，HTTP/gRPC 负责历史查询和管理。

## MVP Scope

- TCP 连接鉴权。
- TCP 消息编解码。
- 心跳和断线重连。
- 私聊。
- 频道聊天。
- 消息持久化。
- 最近消息查询。
- 离线消息查询。
- 禁言。
- 黑名单影响私聊。
- 举报消息。
- 基础敏感词 hook。

## Later Scope

- 群聊。
- 富文本和表情。
- 附件。
- 语音状态同步。
- 多语言敏感词。
- AI 审核。
- 大频道分片。
- 跨节点广播优化。

## Data Model

建议表：

```text
chat_conversation
  id
  app_id
  type
  created_at
  updated_at

chat_participant
  id
  conversation_id
  user_id
  role
  joined_at
  last_read_message_id

chat_message
  id
  conversation_id
  sender_id
  type
  content
  metadata
  status
  created_at

chat_mute
  id
  user_id
  scope
  reason
  starts_at
  ends_at
  created_by

chat_report
  id
  reporter_id
  message_id
  reason
  status
  created_at
```

频道可以用固定 conversation，也可以单独建 `chat_channel` 表，具体看实现复杂度。

## TCP Protocol

MVP 聊天协议建议采用：

- 传输层：TCP。
- 包格式：固定长度 header + Protobuf payload。
- 编码：大端或小端必须在协议文档中固定，推荐 network byte order。
- 鉴权：连接建立后先发送 auth 消息，认证成功前只能处理 auth、ping、pong。
- 心跳：客户端定期发送 `PING`，服务端返回 `PONG`。
- 错误：使用统一业务错误码。

推荐基础包头：

```text
magic        uint16
version      uint16
messageType  uint32
requestId    uint64
payloadLen   uint32
```

具体实现可以调整，但必须有版本号、消息类型、请求 ID 和 payload 长度。

## TCP Messages

### Client to Server

```text
AUTH
PING
CHAT_MESSAGE_SEND
CHAT_CHANNEL_JOIN
CHAT_CHANNEL_LEAVE
CHAT_TYPING_START
CHAT_TYPING_STOP
```

### Server to Client

```text
AUTH_OK
AUTH_REJECTED
PONG
CHAT_MESSAGE_CREATED
CHAT_MESSAGE_REJECTED
CHAT_CHANNEL_JOINED
CHAT_CHANNEL_LEFT
CHAT_MUTE_UPDATED
ERROR
```

## Message Send Flow

```text
TCP Gateway receives packet
  -> Decode header and Protobuf payload
  -> Auth context
  -> ChatService.validateSender
  -> MutePolicy
  -> BlockPolicy
  -> ModerationPolicy
  -> Persist message
  -> Deliver to online recipients
  -> Mark offline recipients for later sync
```

## API

```http
GET /api/v1/chat/conversations
GET /api/v1/chat/conversations/:id/messages
POST /api/v1/chat/reports
```

后台 API 或 gRPC：

```http
GET /api/v1/admin/chat/reports
POST /api/v1/admin/chat/mutes
DELETE /api/v1/admin/chat/mutes/:id
```

## Validation Rules

- 未登录用户不能连接。
- 被封禁用户不能连接或发送消息。
- 被禁言用户不能发送消息。
- 黑名单阻止私聊。
- 用户只能读取自己参与的会话。
- 频道加入必须符合频道规则。

## Storage Policy

MVP 默认：

- 私聊保存完整历史。
- 频道消息保存最近一段时间或固定条数，具体由配置决定。
- 举报相关消息必须保留到处理完成后再按策略清理。

## Tests

必须覆盖：

- 未登录连接失败。
- 登录用户连接成功。
- 私聊消息只投递给双方。
- 频道消息只投递给频道成员。
- 心跳超时后连接会被关闭。
- 无效包长或未知消息类型会被拒绝。
- 禁言用户发送失败。
- 黑名单阻止私聊。
- 消息历史分页查询。
- 举报消息创建成功。
