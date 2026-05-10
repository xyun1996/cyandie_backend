# Admin

Admin 模块负责后台管理、RBAC、运营操作和审计日志。它可以编排其他模块，但不能绕过模块边界随意改核心数据。

## Goals

- 提供安全的后台操作入口。
- 支持用户、聊天、排行榜、订单等运营管理。
- 所有高风险操作可追溯。
- 权限清晰，避免后台越权。

## MVP Scope

- 后台账号。
- 后台角色。
- RBAC。
- 用户搜索。
- 用户封禁/解封。
- 查看平台绑定。
- 查看登录日志。
- 聊天举报处理。
- 聊天禁言。
- 榜单配置管理。
- 审计日志查询。

## Responsibilities

- 校验后台身份和权限。
- 调用 Users、Chat、Leaderboard、Payments 等模块的公开服务。
- 写审计日志。
- 暴露后台 HTTP API。

## Boundaries

- 不绕过业务模块直接修改核心数据。
- 高风险操作必须写审计日志。
- 后台权限不能复用普通用户权限。
- 审计日志不允许普通后台用户修改。

## Data

核心模型见 `docs/data-model.md`：

- `admin_user`
- `admin_role`
- `admin_user_role`
- `audit_log`

## HTTP APIs

```text
GET /api/v1/admin/users
GET /api/v1/admin/users/:id
POST /api/v1/admin/users/:id/ban
POST /api/v1/admin/users/:id/unban
GET /api/v1/admin/audit-logs
GET /api/v1/admin/chat/reports
POST /api/v1/admin/chat/mutes
GET /api/v1/admin/leaderboards
POST /api/v1/admin/leaderboards
PATCH /api/v1/admin/leaderboards/:id
```

## gRPC

Admin gRPC 默认只开放给内部工具，不作为公网接口。

建议 service：

```text
AdminService.CheckPermission
AdminService.WriteAuditLog
AdminService.ListAuditLogs
```

## Tests

- 未登录后台账号不能访问后台接口。
- 无权限角色不能执行高风险操作。
- 封禁用户写审计日志。
- 禁言用户写审计日志。
- 榜单配置变更写审计日志。

