# ADR 0003: Use Redis Sorted Set for Leaderboard

## Status

Accepted

## Context

排行榜需要频繁提交分数和低延迟查询排名。SQL 可以完成排序查询，但在高频榜单场景下成本较高。

Redis Sorted Set 天然支持：

- 分数排序。
- Top N 查询。
- 用户排名查询。
- 用户附近排名查询。

## Decision

MVP 排行榜使用 Redis Sorted Set 作为实时排名来源，MySQL 保存：

- 榜单配置。
- 赛季配置。
- 分数提交事件。
- 快照和审计数据。

## Consequences

优点：

- 排名查询快。
- 实现简单。
- 可支持高频提交。

代价：

- 需要处理 Redis 和 SQL 的一致性。
- 需要快照或重放机制恢复 Redis 数据。
- 复杂多维榜单可能需要额外索引或派生结构。
