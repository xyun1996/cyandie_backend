# Auth

Auth 模块负责登录、Token、Session、账号绑定和账号合并。它是所有客户端能力的入口，但不直接承载用户资料、聊天、排行榜或资产逻辑。

## Goals

- 支持游客、本地账号和平台账号登录。
- 支持 Steam、Epic 等平台身份绑定。
- 统一签发和刷新 Access Token。
- Refresh Token 可撤销。
- 支持设备和登录日志。
- 为 HTTP、gRPC、TCP 复用同一套鉴权上下文。

## MVP Scope

- 游客登录。
- 平台登录入口。
- JWT Access Token。
- Refresh Token。
- Session 创建、刷新、撤销。
- 平台账号绑定和解绑。
- 用户封禁检查。
- 登录日志。

## Non Goals for MVP

- 不做复杂账号合并 UI。
- 不做企业 SSO。
- 不做完整风控系统。
- 不做多因素认证。

## Responsibilities

- 校验登录请求。
- 调用 `platforms` 模块验证平台身份。
- 调用 `users` 模块创建或读取本地用户。
- 维护 Session。
- 签发 Token。
- 记录登录事件。

## Boundaries

- 不直接调用 Steam、Epic 等平台 SDK。
- 不写聊天、排行榜、背包或支付数据。
- 不把平台用户 ID 当成本地用户 ID。
- 不把 Refresh Token 明文写入数据库。

## Data

核心模型见 `docs/data-model.md`：

- `user`
- `account_identity`
- `platform_account`
- `user_session`
- `login_event`

## HTTP APIs

```text
POST /api/v1/auth/guest
POST /api/v1/auth/platform/:provider
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
POST /api/v1/me/platforms/:provider/link
DELETE /api/v1/me/platforms/:provider
GET /api/v1/me/platforms
```

## gRPC

建议 service：

```text
AuthService.LoginGuest
AuthService.LoginWithPlatform
AuthService.Refresh
AuthService.Logout
AuthService.VerifyAccessToken
AuthService.LinkPlatform
AuthService.UnlinkPlatform
```

## Validation

- 被封禁用户不能登录或刷新 Token。
- 无效平台 token 不能创建用户。
- 一个平台账号不能绑定多个本地用户。
- Refresh Token 必须可以撤销。
- Token 错误必须映射到 `AUTH_*` 错误码。

## Tests

- 游客登录成功。
- 平台登录成功。
- 无效平台 token 登录失败。
- Refresh Token 刷新成功。
- 撤销后的 Refresh Token 不能继续使用。
- 被封禁用户不能登录。
- 同一平台账号重复登录返回同一用户。

