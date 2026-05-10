# Protocols

本文档定义项目的通信协议方向。核心后端以 Go 为主，协议层必须服务于长期通用性和后续可拆分性。

## Protocol Goals

- HTTP/REST 面向客户端、管理后台和简单外部集成。
- gRPC/Protobuf 面向服务间通信、内部工具和跨语言 SDK。
- TCP 长连接面向聊天、在线状态和高频实时消息。
- 所有协议复用统一鉴权、错误码、日志和 request id。

## HTTP

用途：

- 客户端公开 API。
- 管理后台 API。
- SDK 易接入接口。
- 文件上传、配置查询、历史消息查询等低频请求。

约束：

- 使用 `/api/v1` 版本前缀。
- 响应格式遵循 `docs/api-contracts.md`。
- 鉴权使用 `Authorization: Bearer <access_token>`。
- 管理后台接口使用独立权限域。

## gRPC

用途：

- 服务间 RPC。
- 后台 worker 调用核心服务。
- 内部工具调用后端能力。
- 后续模块拆成独立服务时复用契约。

目录建议：

```text
protocol/
  proto/
    auth/v1/auth.proto
    user/v1/user.proto
    platform/v1/platform.proto
    chat/v1/chat.proto
    leaderboard/v1/leaderboard.proto
  generated/
```

约束：

- Protobuf 字段编号发布后不得复用。
- 删除字段必须 `reserved`。
- gRPC service 以模块边界为单位，不要做万能 service。
- gRPC 错误必须映射统一业务错误码。
- 重要 service 需要契约测试或集成测试。
- 外部公开 gRPC 前必须明确鉴权和版本策略。

建议 service：

```text
AuthService
UserService
PlatformService
ChatService
LeaderboardService
FriendService
AdminService
```

## TCP Chat Protocol

用途：

- 聊天实时消息。
- 在线状态同步。
- 高频轻量事件。

MVP 包格式建议：

```text
magic        uint16
version      uint16
messageType  uint32
requestId    uint64
payloadLen   uint32
payload       protobuf bytes
```

连接流程：

```text
1. Client opens TCP connection.
2. Client sends AUTH message with access token.
3. Server validates token and user status.
4. Server replies AUTH_OK or AUTH_REJECTED.
5. Client and server exchange chat messages.
6. Client sends PING periodically.
7. Server replies PONG.
8. Server closes connection after heartbeat timeout or auth revocation.
```

基础消息类型：

```text
AUTH
AUTH_OK
AUTH_REJECTED
PING
PONG
ERROR
CHAT_MESSAGE_SEND
CHAT_MESSAGE_CREATED
CHAT_MESSAGE_REJECTED
CHAT_CHANNEL_JOIN
CHAT_CHANNEL_JOINED
CHAT_CHANNEL_LEAVE
CHAT_CHANNEL_LEFT
CHAT_TYPING_START
CHAT_TYPING_STOP
CHAT_MUTE_UPDATED
```

约束：

- 未认证连接只能处理 `AUTH`、`PING`、`PONG`。
- payload 长度必须有上限。
- 未知 message type 必须返回错误或断开连接。
- TCP gateway 只负责连接、鉴权、心跳、编解码和投递入口。
- 聊天业务规则必须放在 ChatService。
- 错误码复用统一错误码。

## Protocol Ownership

- `protocol/proto` 保存源协议。
- 生成代码可以提交，也可以由构建生成，但策略必须统一。
- Go 服务端使用生成后的 Go 类型。
- TypeScript SDK 可由 OpenAPI 和 Protobuf 共同生成或手写薄封装。
- TCP 客户端 SDK 必须复用同一套消息定义，不能私自复制常量。

