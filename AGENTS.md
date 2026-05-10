# Agent Instructions

本仓库目标是实现一个尽量通用的后端框架，面向游戏、应用、社区、工具型产品和实时互动业务，提供账号、平台 SDK 接入、聊天、排行榜、好友/关系链、成就/任务、支付校验、后台管理、服务间 RPC 和对外 SDK 等通用能力。

未来 agent 在本仓库工作时，必须先阅读本文档，再阅读 `docs/README.md` 和相关 `docs/` 文档。不要只根据聊天上下文直接实现。

## Product Direction

- 做一个模块化、可扩展、可自托管的通用后端框架。
- 优先覆盖强互动、强账号、强运营的业务场景，包括但不限于游戏。
- 第一阶段目标不是大而全，而是能让一个真实项目快速接入账号、平台身份、TCP 聊天、排行榜、后台管理和基础 RPC。
- 框架应支持多个应用/项目接入，但 MVP 可以先以单应用为主，多租户能力分阶段落地。

## Default Technology Stack

如果用户没有明确指定更细的技术选型，默认采用：

- 核心后端语言：Go
- 管理后台、工具链和部分客户端 SDK：TypeScript
- 数据库：MySQL
- 数据访问：优先使用项目既有方案；新项目可选 sqlc、GORM 或 Ent，但必须统一
- 缓存和排行榜：Redis
- 服务间通信：gRPC + Protobuf
- 客户端公开 API：HTTP/REST 起步，可通过 gRPC-Gateway 或独立 HTTP handler 暴露
- 实时聊天：TCP 长连接，消息体优先使用 Protobuf
- 队列：Redis Streams 起步，后续可替换 NATS、RabbitMQ 或 Kafka
- API 文档：OpenAPI + Protobuf 文档
- 测试：Go test + integration tests + Testcontainers
- 部署：Docker Compose 起步，后续支持 Kubernetes

如果项目已经采用其他技术栈，优先跟随现有代码，不要擅自迁移。

## Architecture Rules

- 使用模块化单体作为第一阶段架构。未经用户明确要求，不要拆成微服务，但模块接口要为后续 gRPC 拆分留边界。
- 所有平台接入必须通过统一 `PlatformAdapter` 接口，不允许业务模块直接调用 Steam、Epic、Apple、Google 等平台 API。
- 平台账号只是身份来源，不是本地用户本体。本地用户统一由 Auth/User 模块管理。
- 排行榜实时排名必须优先使用 Redis Sorted Set，SQL 数据库存储配置、快照和审计数据。
- 聊天必须通过 TCP 长连接提供实时能力，通过数据库保存必要的消息历史。
- 服务间 RPC 契约必须优先使用 Protobuf 描述，gRPC 服务边界要和模块边界一致。
- 协议设计必须遵循 `docs/protocols.md`。
- 管理后台的高风险操作必须写审计日志。
- 公开 API 必须有统一错误格式、鉴权规则和版本策略。
- 新功能必须补充测试。涉及 API、数据库 schema、平台接口或模块边界时，必须同步更新文档。

## Module Ownership

- `auth`：登录、Token、Session、账号绑定、鉴权。
- `users`：本地用户、资料、状态、封禁。
- `platforms`：Steam、Epic、Apple、Google 等平台适配器。
- `chat`：TCP 长连接、私聊、频道、群聊、离线消息、禁言、举报。
- `leaderboards`：榜单配置、分数提交、排名查询、赛季、奖励。
- `friends`：好友、黑名单、好友申请、在线状态。
- `achievements`：成就定义、进度、解锁、平台同步。
- `inventory`：背包、货币、虚拟资产。
- `payments`：订单、收据校验、平台购买验证。
- `admin`：后台接口、运营工具、审计日志。
- `sdk`：对外客户端 SDK 和协议封装。
- `protocol`：Protobuf、gRPC service 定义、TCP 消息协议。
- `shared`：通用错误、日志、配置、工具、基础类型。

## Boundaries

- `platforms` 不允许直接创建本地用户、不允许签发 JWT、不允许写排行榜或聊天数据。
- `leaderboards` 不允许调用平台 SDK、不允许修改用户身份、不允许发送聊天消息。
- `chat` 不允许修改用户身份、不允许直接操作平台账号、不允许直接发放奖励。
- `admin` 可以编排多个模块，但必须通过模块公开服务接口，并记录审计日志。
- `shared` 不应依赖业务模块。

更多细节见 `docs/module-boundaries.md`。

## Required Workflow

每次任务必须遵循：

1. 阅读 `AGENTS.md`。
2. 阅读 `docs/README.md`。
3. 阅读与任务相关的 `docs/` 文档。
4. 先查找已有模块、接口和测试模式。
5. 只实现当前任务范围，不做无关重构。
6. 新增或更新测试。
7. 运行项目可用的验证命令。
8. 如果改动了 API、数据模型、架构或模块边界，更新文档。
9. 最终回复说明改了什么、验证了什么、还有什么风险。

## Required Validation

如果项目中存在这些命令或等价命令，完成前必须运行：

- lint
- typecheck
- unit tests
- integration tests
- database migration check
- OpenAPI/schema generation or validation

如果无法运行，必须在最终回复中说明原因。

## Definition of Done

一个功能只有满足以下条件才算完成：

- 行为符合任务描述和文档约束。
- 有清晰的模块边界。
- 有必要的测试覆盖。
- 错误处理和日志符合项目规范。
- API 契约、数据库迁移、配置项和文档已经同步。
- 验证命令已运行或明确说明无法运行的原因。

详细完成标准见 `docs/definition-of-done.md`。

## Do Not

- 不要在没有用户确认的情况下切换主要技术栈。
- 不要在 MVP 阶段引入微服务、复杂事件总线或过度抽象。
- 不要绕过统一接口直接访问第三方平台。
- 不要把平台用户 ID 当作本地用户 ID。
- 不要让聊天、排行榜、支付等模块互相直接写对方的数据。
- 不要删除或重写用户已有代码，除非任务明确要求。
- 不要提交无法解释的“大重构”。
