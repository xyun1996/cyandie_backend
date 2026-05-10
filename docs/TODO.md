# Planning TODO

本文档记录当前后端框架计划在 review 后需要继续改进的点。这里先保存待办，不直接修改现有设计正文，避免把 review 记录和设计修订混在一起。

## P1

### 拆分平台适配器能力接口

问题：

- `docs/modules/platforms.md` 里 `PlatformAdapter` 当前同时包含登录验证、资料查询和购买校验。
- ADR 0002 又说明 MVP 先实现登录验证，`getUserProfile` 可选。
- 这会导致 Steam/Epic MVP 被迫写购买校验等空实现。

建议：

- 将平台能力拆成 capability 接口，例如：
  - `AuthAdapter`：平台登录和 token 验证。
  - `ProfileProvider`：平台用户资料查询。
  - `PurchaseVerifier`：平台购买或收据校验。
  - `AchievementSyncer`：平台成就同步。
- `PlatformRegistry` 支持按 provider 和 capability 查询。
- MVP 只要求 Steam/Epic 实现登录验证能力。

### 锁定 Go 技术栈细节

问题：

- 数据访问仍是 `sqlc/GORM/Ent` 三选一。
- HTTP 暴露方式仍是 `gRPC-Gateway 或独立 handler`。
- Protobuf 生成代码“可以提交，也可以构建生成”。
- 不同 agent 可能生成完全不同的项目骨架。

建议：

- 明确固定默认选择，例如：
  - 数据访问：`sqlc`。
  - 数据库迁移：`goose`。
  - HTTP router：`chi`。
  - Protobuf 管理：`buf`。
  - 生成代码：提交到仓库。
- 将这些选择写入 `AGENTS.md`、`docs/architecture.md` 和 ADR。

### 改成 Go 惯用项目目录

问题：

- `docs/architecture.md` 仍推荐 `src/app/config/...` 风格目录。
- 对 Go 项目不够自然，后续脚手架容易跑偏。

建议：

- 改成 Go idiomatic 布局，例如：
  - `cmd/`
  - `internal/`
  - `api/proto/`
  - `pkg/`
  - `web/admin/`
  - `sdk/`
  - `deployments/`
- 明确业务模块放在 `internal/<module>`。

## P2

### 明确多应用和租户策略

问题：

- 部分表已有 `app_id`，但部分关系表没有。
- `platform_account` 唯一约束当前是全局 `(provider, platform_user_id)`。
- 未来多应用接入时，可能不清楚同一个平台账号是否允许绑定到不同 app 下的不同本地用户。

建议：

- 明确 MVP 是否启用 `app_id`。
- 若保留多应用能力，所有业务表都应有一致的 `app_id` 或清楚说明继承路径。
- 明确平台账号唯一约束是全局唯一，还是 `(app_id, provider, platform_user_id)` 唯一。

### 固定 TCP 聊天协议参数

问题：

- TCP 协议当前仍保留一些实现期再决定的参数。
- 未固定大小端、最大 payload、心跳间隔、超时时间、magic/version 常量和未知消息策略。

建议：

- 固定为 network byte order。
- 固定最大 payload，例如 64 KiB。
- 固定心跳间隔和超时，例如 30s ping、90s timeout。
- 固定 `magic` 和初始 `version`。
- 明确未知消息类型返回错误后是否断开连接。

### 固定 Chat 数据模型

问题：

- `docs/modules/chat.md` 里频道可以用固定 conversation，也可以单独建 `chat_channel` 表。
- 这会影响 schema、API、权限判断和查询方式。

建议：

- MVP 明确增加 `chat_channel` 表。
- `chat_conversation` 专注私聊和群聊。
- 频道成员、频道配置和频道权限放在 channel 相关表里。

### 调整 Phase 1 和 Phase 2 边界

问题：

- Phase 1 已包含平台登录入口。
- Phase 2 才实现 Steam/Epic adapter。
- 边界略混，容易导致 Phase 1 需要 mock 或半成品平台登录。

建议：

- Phase 1 只做游客/本地 Session、用户模型和 adapter interface。
- Phase 2 再接 Steam/Epic，并完成平台登录闭环。

## P3

### 将 roadmap 加入必读顺序

问题：

- `docs/README.md` 当前“任何任务都先读”的列表没有包含 `docs/roadmap.md`。
- agent 做功能任务时可能只看模块文档，不看阶段边界。

建议：

- 将 `docs/roadmap.md` 加入必读顺序。
- 或至少标注“功能任务必须阅读 roadmap”。

