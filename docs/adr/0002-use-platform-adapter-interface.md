# ADR 0002: Use Platform Adapter Interface

## Status

Accepted

## Context

本项目需要接入 Steam、Epic、Apple、Google、Xbox、PlayStation 等平台。各平台的登录、用户资料、购买校验、成就、统计接口差异很大。

如果业务模块直接调用平台 SDK，会造成：

- 业务代码被平台细节污染。
- 新增平台时要改多个模块。
- 测试困难。
- agent 容易重复实现相似逻辑。

## Decision

所有平台能力必须通过 `PlatformAdapter` 接口暴露。业务模块通过 `PlatformRegistry` 获取适配器。

MVP 先实现登录验证：

- `verifyToken`
- `getUserProfile` 可选
- `verifyPurchase` 预留

## Consequences

优点：

- 平台差异被隔离。
- 新增平台成本低。
- 可对所有平台运行同一套 contract tests。
- Auth、Payments、Achievements 不需要知道具体平台 SDK。

代价：

- 初期需要设计标准化类型。
- 某些平台特殊能力需要通过扩展接口或 capability 声明表达。

