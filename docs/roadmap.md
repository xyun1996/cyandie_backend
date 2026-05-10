# Roadmap

本文档是通用后端框架的阶段计划。未来 agent 执行任务时，应优先按照本路线图拆分，而不是一次性实现全部能力。

## Phase 0: Foundation

目标：建立可持续开发的地基，让后续 agent 有明确轨道。

### Deliverables

- 项目脚手架。
- Go module 基础结构。
- TypeScript workspace，用于管理后台、工具链和 SDK。
- 配置系统。
- 统一日志。
- 统一错误码。
- API 响应格式。
- 数据库连接和迁移。
- Redis 连接封装。
- Protobuf/gRPC 基础目录和生成流程。
- TCP 网关基础骨架。
- 基础测试框架。
- OpenAPI 文档生成。
- Docker Compose 本地开发环境。
- MySQL、Redis 本地依赖编排。
- `AGENTS.md` 和 `docs/` 文档。

### Acceptance Criteria

- 能本地启动 API 服务。
- 能连接 MySQL 和 Redis。
- 有健康检查接口。
- 有至少一个示例单元测试和集成测试。
- 有 lint/test/proto generation 命令。

## Phase 1: Auth and User Core

目标：完成本地用户、平台身份和会话的核心模型。

### Deliverables

- 本地用户模型。
- 用户资料模型。
- 账号身份模型。
- 平台账号绑定模型。
- Session 模型。
- JWT Access Token。
- Refresh Token。
- 游客登录。
- 平台登录入口。
- 用户封禁状态。
- 登录日志。

### Required APIs

- `POST /api/v1/auth/guest`
- `POST /api/v1/auth/platform/:provider`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/me`
- `PATCH /api/v1/me/profile`
- `POST /api/v1/me/platforms/:provider/link`
- `DELETE /api/v1/me/platforms/:provider`

### Acceptance Criteria

- 一个本地用户可以绑定多个平台身份。
- 一个平台身份不能绑定到多个本地用户，除非走账号合并流程。
- Refresh Token 可以撤销。
- 被封禁用户不能登录或刷新 Session。
- Auth 流程有单元测试和 API 集成测试。

## Phase 2: Platform Adapter MVP

目标：建立统一平台适配层，并实现 Steam、Epic 的最小登录验证。

### Deliverables

- `PlatformAdapter` 接口。
- `PlatformRegistry` 注册和查找适配器。
- 标准化 `PlatformUser` 类型。
- Steam 登录验证适配器。
- Epic 登录验证适配器。
- 平台配置加载。
- 平台 API 调用日志。
- 平台适配器 contract tests。

### Supported Providers in MVP

- Steam。
- Epic。

### Later Providers

- Apple。
- Google。
- Xbox。
- PlayStation。
- Nintendo。

### Acceptance Criteria

- Auth 模块只能通过 `PlatformRegistry` 调用平台能力。
- Steam/Epic 适配器通过同一套 contract tests。
- 平台 API 错误会转换成统一错误码。
- 平台凭证从环境变量或 secret 管理读取，不写入源码。

## Phase 3: Leaderboard MVP

目标：实现可上线使用的排行榜核心能力。

### Deliverables

- 榜单配置模型。
- 榜单分数提交。
- Top N 查询。
- 用户排名查询。
- 用户附近排名查询。
- Redis Sorted Set 排名写入。
- SQL 分数快照或事件记录。
- 分数更新策略：最高分、最低耗时、最后提交覆盖。
- 基础速率限制。
- 管理后台榜单配置接口。

### Required APIs

- `POST /api/v1/leaderboards/:boardId/scores`
- `GET /api/v1/leaderboards/:boardId/top`
- `GET /api/v1/leaderboards/:boardId/me`
- `GET /api/v1/leaderboards/:boardId/around-me`

### Acceptance Criteria

- Redis Sorted Set 是实时排名来源。
- 数据库保存榜单配置和提交记录。
- 同一个用户重复提交时符合榜单更新策略。
- 支持分页查询 Top N。
- 有 Redis 集成测试。

## Phase 4: Chat MVP

目标：实现基于 TCP 长连接的基础实时聊天，支持真实客户端接入。

### Deliverables

- TCP 连接鉴权。
- TCP 消息编解码。
- 心跳和断线处理。
- 私聊。
- 频道聊天。
- 消息持久化。
- 最近消息查询。
- 在线连接管理。
- 离线消息查询。
- 禁言。
- 黑名单基础能力。
- 敏感词 hook。
- 聊天举报模型。

### Required APIs and TCP Messages

- `GET /api/v1/chat/conversations`
- `GET /api/v1/chat/conversations/:id/messages`
- `POST /api/v1/chat/reports`
- TCP message: `CHAT_MESSAGE_SEND`
- TCP message: `CHAT_MESSAGE_CREATED`
- TCP message: `CHAT_CHANNEL_JOIN`
- TCP message: `CHAT_CHANNEL_LEAVE`
- TCP message: `PING`
- TCP message: `PONG`

### Acceptance Criteria

- 未登录用户不能建立聊天连接。
- 被禁言用户不能发送消息。
- 私聊消息只投递给相关用户。
- 频道消息只投递给频道成员。
- 消息历史可以分页查询。
- 有 TCP 集成测试。

## Phase 5: Admin MVP

目标：提供最小可用运营后台 API。

### Deliverables

- Admin 用户和角色。
- RBAC 权限。
- 用户搜索。
- 用户详情。
- 封禁/解封。
- 查看平台绑定。
- 查看登录日志。
- 查看聊天举报。
- 榜单配置管理。
- 审计日志。

### Acceptance Criteria

- 后台接口和客户端接口权限隔离。
- 所有高风险操作记录审计日志。
- 非授权角色不能访问后台接口。
- 用户封禁会影响 Auth、Chat、Leaderboard。

## Phase 6: Social and Presence

目标：补齐好友、黑名单、在线状态和邀请能力。

### Deliverables

- 好友申请。
- 好友列表。
- 删除好友。
- 黑名单。
- 在线状态。
- 最近联系人。
- 邀请进入房间/队伍的协议。
- 平台好友导入预留接口。

### Acceptance Criteria

- 黑名单会影响私聊和好友申请。
- 在线状态通过 Redis 维护。
- 用户断线后状态能过期。

## Phase 7: Achievements, Stats, Inventory

目标：提供常见业务后端能力，包括游戏中的统计、成就、任务和通用资产体系。

### Deliverables

- 玩家统计数据。
- 成就定义。
- 成就进度。
- 成就解锁事件。
- 背包物品。
- 虚拟货币。
- 奖励发放接口。
- 平台成就同步预留接口。

### Acceptance Criteria

- 成就解锁幂等。
- 奖励发放幂等。
- 背包和货币操作有审计或流水记录。

## Phase 8: Payments and Platform Commerce

目标：支持平台购买校验和订单管理。

### Deliverables

- 订单模型。
- 平台收据校验接口。
- Steam/Epic/Apple/Google 支付校验适配预留。
- Webhook 处理。
- 幂等键。
- 订单状态机。
- 后台订单查询。

### Acceptance Criteria

- 重复回调不会重复发货。
- 平台验签失败不会更新订单。
- 订单状态变化有日志。

## Phase 9: Client SDKs

目标：降低客户端接入成本。

### Deliverables

- Go SDK。
- TypeScript SDK。
- Unity C# SDK。
- TCP 客户端封装。
- gRPC 客户端封装。
- Token 自动刷新。
- 登录、排行榜、聊天、好友基础 API。
- SDK 示例项目。

### Acceptance Criteria

- SDK 示例能完成登录、提交分数、查询排行榜、发送聊天消息。
- SDK 错误类型和服务端错误码一致。

## Phase 10: Production Hardening

目标：从可用走向可上线。

### Deliverables

- 压测脚本。
- 性能指标。
- 速率限制。
- 防刷基础策略。
- 灰度配置。
- 监控告警。
- 数据备份策略。
- 数据迁移策略。
- 多应用/多租户支持。

### Acceptance Criteria

- 有明确容量目标。
- 关键 API 有压测结果。
- 关键失败场景有告警。
- 多租户数据隔离通过测试。

## Recommended MVP Scope

第一版建议只做：

1. Foundation。
2. Auth/User。
3. Platform Adapter with Steam/Epic。
4. Leaderboard MVP。
5. Chat MVP。
6. Admin MVP。
7. Go/TypeScript SDK 的最小封装。

不要在第一版同时做成就、背包、支付、全平台同步和复杂分析系统。
