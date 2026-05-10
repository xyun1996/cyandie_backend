# SDK

SDK 模块负责对外客户端 SDK、内部服务 SDK 和协议生成封装。SDK 应尽量薄，复杂业务逻辑留在服务端。

## Goals

- 降低客户端接入成本。
- 复用 HTTP、gRPC、TCP 协议定义。
- 提供稳定错误类型。
- 支持 Token 自动刷新。

## MVP Scope

- Go SDK。
- TypeScript SDK。
- HTTP 客户端封装。
- TCP 聊天客户端封装。
- gRPC 客户端封装。
- Token 自动刷新。
- 示例项目。

## Later Scope

- Unity C# SDK。
- Unreal SDK。
- CLI 工具。
- 自动生成文档。

## Responsibilities

- 封装 Auth API。
- 封装 Leaderboard API。
- 封装 Chat TCP client。
- 封装 Friends 基础 API。
- 封装错误码。
- 管理协议生成。

## Boundaries

- 不复制服务端业务逻辑。
- 不硬编码 TCP 消息常量，必须来自协议定义或生成代码。
- 不吞掉服务端错误码。
- 不在 SDK 中保存长期敏感密钥。

## Directory Suggestion

```text
sdk/
  go/
  ts/
  unity/
examples/
  go-client/
  ts-client/
```

## Tests

- SDK 可以完成游客登录。
- SDK 可以自动刷新 Token。
- SDK 可以提交排行榜分数。
- SDK 可以建立 TCP 聊天连接。
- SDK 错误类型保留服务端错误码。

