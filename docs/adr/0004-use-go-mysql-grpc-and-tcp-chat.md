# ADR 0004: Use Go, MySQL, gRPC, and TCP Chat

## Status

Accepted

## Context

项目目标是做一个尽量通用的后端框架，需要兼顾账号、平台接入、排行榜、聊天、后台管理、服务间通信和未来拆分能力。

核心后端需要较好的并发能力、部署简单性和长期维护性。聊天场景需要稳定长连接和可控的协议开销。服务间通信需要清晰契约，方便后续从模块化单体拆分为独立服务。

## Decision

默认技术方向：

- 核心后端使用 Go。
- 管理后台、工具链和部分 SDK 使用 TypeScript。
- 数据库使用 MySQL。
- 缓存和排行榜使用 Redis。
- 服务间通信使用 gRPC + Protobuf。
- 聊天实时通道使用 TCP 长连接，payload 优先使用 Protobuf。

## Consequences

优点：

- Go 适合高并发网络服务和长连接网关。
- MySQL 生态成熟，适合通用业务数据。
- Redis 适合在线状态、排行榜、限流和短期缓存。
- gRPC/Protobuf 能提供稳定跨语言契约。
- TCP 聊天协议可控，适合后续做轻量 SDK 和高频消息。

代价：

- TCP 协议需要自己处理包边界、心跳、重连和兼容性。
- HTTP、gRPC、TCP 三套协议需要统一错误码和鉴权上下文。
- TypeScript 与 Go 的 SDK/协议生成流程需要规范化。

