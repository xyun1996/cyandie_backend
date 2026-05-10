# Architecture

本文档描述通用后端框架的目标架构。未来实现必须优先符合这里的设计，除非用户明确要求变更，并同步新增 ADR。

## Goals

- 提供可复用的通用后端基础能力，覆盖游戏、应用、社区、工具和实时互动业务。
- 支持多平台账号接入，例如 Steam、Epic、Apple、Google、Xbox、PlayStation。
- 支持实时互动能力，例如聊天、在线状态、好友、邀请。
- 支持竞争和运营能力，例如排行榜、赛季、成就、奖励、公告。
- 支持后台管理和审计。
- 支持对外客户端 SDK，使不同类型客户端能快速接入。
- 支持 gRPC/Protobuf 作为服务间通信和长期契约基础。
- 支持 TCP 长连接作为聊天等实时能力的主要通道。

## Non Goals for MVP

- 不做完整微服务架构。
- 不做复杂分布式事务。
- 不做所有平台的完整商业化能力。
- 不做完整 BI 分析平台。
- 不做大型客服系统。
- 不做反作弊引擎，只保留接口和基础校验。

## High Level Architecture

默认技术方向：

- 核心后端：Go。
- 管理后台、工具链和部分 SDK：TypeScript。
- 数据库：MySQL。
- 缓存和实时排名：Redis。
- 服务间通信：gRPC + Protobuf。
- 实时聊天：TCP 长连接 + Protobuf payload。

```text
Client / App / Game / Admin / Service
        |
        | HTTP/REST + TCP + gRPC
        v
Client SDK / Admin UI / Service SDK
        |
        v
API Gateway / Backend App
        |
        +-- Auth
        +-- Users
        +-- Platform Adapters
        +-- Chat
        +-- Leaderboards
        +-- Friends
        +-- Achievements
        +-- Inventory
        +-- Payments
        +-- Admin
        +-- Notifications
        |
        +-- MySQL
        +-- Redis
        +-- Queue
        +-- Object Storage
```

## Deployment Model

MVP 使用模块化单体：

- 一个后端应用进程。
- 一个 MySQL。
- 一个 Redis。
- 可选一个 worker 进程处理异步任务。
- 可选一个 TCP gateway 进程处理聊天长连接；MVP 可先与主进程同部署。
- 后台管理可以与 API 同仓库，也可以独立前端。

后续生产阶段可以拆分：

- TCP Gateway：处理聊天 TCP 长连接。
- Worker：处理奖励、同步、通知、订单回调。
- Admin API：后台管理独立权限域。
- Public API：客户端公开 API。
- gRPC API：服务间调用和后续服务拆分。

拆分前必须保证模块边界已经清晰。

## Core Request Flow

### Login with Platform

```text
Client
  -> Auth API
  -> AuthService
  -> PlatformRegistry
  -> Steam/Epic/Apple Adapter
  -> AccountIdentity
  -> UserSession
  -> JWT + Refresh Token
```

关键约束：

- 平台适配器只负责验证平台身份并返回标准化平台用户。
- AuthService 负责创建或绑定本地用户。
- SessionService 负责签发和刷新 Token。

### Submit Leaderboard Score

```text
Client
  -> Leaderboard API
  -> Auth Guard
  -> LeaderboardService
  -> Validation / Anti-cheat Hook
  -> Redis Sorted Set
  -> Score Snapshot / Event
```

关键约束：

- Redis 负责实时排名。
- SQL 负责配置、赛季、快照、审计。
- 防作弊逻辑通过 hook 或 policy 接入，不和 Redis 写入逻辑混在一起。

### Send Chat Message

```text
Client TCP Connection
  -> Chat TCP Gateway
  -> Auth Guard
  -> ChatService
  -> ModerationService
  -> Message Store
  -> Realtime Delivery
```

关键约束：

- TCP 网关只处理连接、鉴权、心跳、编解码和消息入口。
- ChatService 处理消息业务规则。
- 审核、禁言、黑名单作为独立策略。

## Module Layout

推荐目录：

```text
src/
  app/
  config/
  shared/
    errors/
    logging/
    validation/
    events/
  auth/
  users/
  platforms/
    core/
    steam/
    epic/
    apple/
    google/
  chat/
  leaderboards/
  friends/
  achievements/
  inventory/
  payments/
  admin/
  sdk/
  protocol/
    proto/
    generated/
  workers/
```

实际目录以项目已有结构为准，但模块职责不得混乱。

## Data Stores

### MySQL

用于：

- 用户。
- 平台身份。
- 会话。
- 聊天历史。
- 榜单配置。
- 榜单快照。
- 成就、背包、订单、审计日志。

### Redis

用于：

- 排行榜实时排名。
- Session 缓存。
- 在线状态。
- TCP 节点间广播。
- 速率限制。
- 短期幂等键。

### Queue

用于：

- 平台资料同步。
- 排行榜赛季结算。
- 奖励发放。
- 支付回调处理。
- 通知推送。
- 审计事件异步落库。

MVP 可以先使用 Redis Streams 或内存队列，但必须保留可替换边界。

## gRPC and Protobuf

gRPC/Protobuf 用于：

- 后续服务拆分时的服务间通信。
- Worker 调用核心服务。
- 内部管理工具调用后端能力。
- 生成跨语言 SDK 的稳定类型。

MVP 可以在模块化单体内先使用 Go interface，但重要模块必须让接口形状接近未来 gRPC service：

- AuthService。
- UserService。
- PlatformService。
- ChatService。
- LeaderboardService。
- AdminService。

Protobuf 文件建议放在 `protocol/proto/`，生成代码放在 `protocol/generated/` 或语言约定目录。

## API Principles

- 所有公开 API 使用 `/api/v1` 前缀。
- HTTP API 响应采用统一 envelope。
- 错误使用稳定错误码。
- 客户端 API 和后台 API 分离权限。
- 所有需要登录的接口必须经过 Auth Guard。
- 高风险接口必须有审计日志。
- gRPC 错误必须映射到统一业务错误码。
- TCP 消息错误必须复用统一业务错误码。

示例：

```json
{
  "ok": false,
  "error": {
    "code": "AUTH_INVALID_TOKEN",
    "message": "Invalid token",
    "requestId": "req_..."
  }
}
```

## Security Principles

- Refresh Token 必须可撤销。
- 密钥、平台凭证、Webhook secret 不进入源码。
- 平台回调必须校验签名或来源。
- 管理后台必须启用 RBAC。
- 高风险后台操作必须记录操作者、目标、原因、前后值。
- 用户封禁后应影响登录、聊天、排行榜提交等相关能力。

## Observability

必须逐步建立：

- 结构化日志。
- request id / trace id。
- API latency metrics。
- TCP connection metrics。
- Redis leaderboard operation metrics。
- 平台 API 调用成功率和耗时。
- 后台审计日志。
