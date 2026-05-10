# Users

Users 模块负责本地用户实体、资料、状态和封禁。平台账号、Session、聊天关系等能力属于其他模块。

## Goals

- 维护稳定的本地用户 ID。
- 提供用户资料读写。
- 管理用户状态。
- 为其他模块提供用户查询和状态校验。

## MVP Scope

- 创建用户。
- 查询用户。
- 更新用户资料。
- 用户状态：active、banned、deleted。
- 封禁信息。
- 最后活跃时间。

## Responsibilities

- 生成和保存本地用户。
- 保存用户资料。
- 校验用户是否可执行某类操作。
- 向 Auth、Chat、Leaderboard、Admin 提供用户查询接口。

## Boundaries

- 不签发 Token。
- 不验证平台 token。
- 不直接处理聊天、排行榜或支付业务。
- 不直接写后台审计，但可以返回变更前后数据给 Admin 记录。

## Data

核心模型见 `docs/data-model.md`：

- `user`
- `user_profile`

## HTTP APIs

```text
GET /api/v1/me
PATCH /api/v1/me/profile
```

后台 API：

```text
GET /api/v1/admin/users
GET /api/v1/admin/users/:id
POST /api/v1/admin/users/:id/ban
POST /api/v1/admin/users/:id/unban
```

## gRPC

建议 service：

```text
UserService.GetUser
UserService.GetProfile
UserService.UpdateProfile
UserService.CheckUserStatus
UserService.BanUser
UserService.UnbanUser
```

## Tests

- 创建用户成功。
- 更新资料成功。
- 查询不存在用户返回 `USER_NOT_FOUND`。
- 封禁用户状态可被 Auth、Chat、Leaderboard 识别。
- 后台封禁和解封需要审计日志。

