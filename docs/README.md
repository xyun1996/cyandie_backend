# Documentation Index

本文档是 `docs/` 的入口。未来 agent 在本仓库执行任务时，应先读 `AGENTS.md`，再按本文档选择需要阅读的设计文档。

## Reading Order

任何任务都先读：

1. `AGENTS.md`
2. `docs/README.md`
3. `docs/architecture.md`
4. `docs/module-boundaries.md`
5. `docs/definition-of-done.md`

涉及协议时继续读：

- `docs/protocols.md`
- `docs/api-contracts.md`

涉及具体模块时继续读对应模块文档：

- `docs/modules/auth.md`
- `docs/modules/users.md`
- `docs/modules/platforms.md`
- `docs/modules/chat.md`
- `docs/modules/leaderboards.md`
- `docs/modules/friends.md`
- `docs/modules/admin.md`
- `docs/modules/payments.md`
- `docs/modules/inventory.md`
- `docs/modules/achievements.md`
- `docs/modules/sdk.md`

涉及架构决策时读：

- `docs/adr/`

## Global Documents

- `docs/architecture.md`：总体架构、部署模型、数据存储、模块布局。
- `docs/roadmap.md`：阶段路线图和 MVP 范围。
- `docs/TODO.md`：架构计划 review 后的待办改进项。
- `docs/module-boundaries.md`：模块边界和禁止事项。
- `docs/protocols.md`：HTTP、gRPC、TCP 协议方向。
- `docs/api-contracts.md`：HTTP、gRPC、TCP 通用契约。
- `docs/data-model.md`：核心数据模型索引。
- `docs/definition-of-done.md`：完成标准。
- `docs/agent-task-template.md`：给 agent 派任务的模板。

## Module Documents

模块文档放在 `docs/modules/`。每个模块文档负责说明：

- 目标。
- MVP 范围。
- 后续范围。
- 职责。
- 模块边界。
- 数据模型。
- HTTP/gRPC/TCP 接口。
- 验收和测试要求。

## How to Add a Module

新增模块时必须：

1. 在 `docs/modules/` 新增模块文档。
2. 在 `docs/module-boundaries.md` 增加模块边界。
3. 如涉及数据模型，更新 `docs/data-model.md`。
4. 如涉及 HTTP/gRPC/TCP，更新 `docs/protocols.md` 或 `docs/api-contracts.md`。
5. 如涉及关键架构选择，新增 ADR。
6. 更新本索引。
