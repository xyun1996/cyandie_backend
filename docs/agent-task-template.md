# Agent Task Template

未来给 agent 下任务时，建议复制这个模板，减少跑偏。

```text
请先阅读 AGENTS.md 和相关 docs 文档，再实现。

任务：
实现 <功能名称>。

背景：
<为什么要做，和哪个阶段有关。>

范围：
- <必须做 1>
- <必须做 2>
- <必须做 3>

不做：
- <明确不做 1>
- <明确不做 2>

设计约束：
- 必须符合 docs/architecture.md。
- 必须符合 docs/module-boundaries.md。
- 如果是平台接入，必须符合 docs/modules/platforms.md。
- 如果是排行榜，必须符合 docs/modules/leaderboards.md。
- 如果是聊天，必须符合 docs/modules/chat.md。

接口：
- <API、gRPC service 或 TCP 消息>

数据：
- <需要新增或修改的数据模型>

验收标准：
- <可验证标准 1>
- <可验证标准 2>
- <可验证标准 3>

测试要求：
- <单元测试>
- <集成测试>
- <契约测试>

完成前运行：
- <lint>
- <typecheck>
- <test>
```

## Example: Leaderboard MVP

```text
请先阅读 AGENTS.md、docs/README.md、docs/architecture.md、docs/module-boundaries.md 和 docs/modules/leaderboards.md，再实现。

任务：
实现排行榜 MVP。

范围：
- 新增 leaderboard 模块。
- 支持榜单配置。
- 支持提交分数。
- 支持查询 Top N。
- 支持查询我的排名。
- 支持查询我附近排名。
- Redis Sorted Set 作为实时排名来源。
- SQL 保存榜单配置和分数提交事件。

不做：
- 不做赛季奖励。
- 不做好友榜。
- 不做复杂反作弊。

验收标准：
- 同一用户重复提交符合最高分策略。
- 被封禁用户不能提交分数。
- Redis 集成测试通过。
- API 集成测试通过。
- OpenAPI 文档更新。
```

## Example: Steam Adapter

```text
请先阅读 AGENTS.md、docs/README.md、docs/modules/platforms.md 和 docs/api-contracts.md，再实现。

任务：
实现 Steam 平台登录适配器。

范围：
- 新增 SteamAdapter。
- 实现 PlatformAdapter.verifyToken。
- 返回标准化 PlatformUser。
- 接入 PlatformRegistry。
- Auth 平台登录可以使用 steam provider。

不做：
- 不做 Steam 支付。
- 不做成就同步。
- 不做好友导入。

验收标准：
- SteamAdapter 通过 PlatformAdapter contract tests。
- 无效 token 返回 PLATFORM_INVALID_TOKEN。
- Auth 模块不直接调用 Steam SDK。
```
