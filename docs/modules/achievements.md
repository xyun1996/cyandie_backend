# Achievements

Achievements 模块负责成就、任务式进度、玩家统计和解锁事件。

## Goals

- 支持通用成就系统。
- 支持玩家统计数据驱动进度。
- 支持后续同步 Steam/Epic/Apple/Google 等平台成就。
- 支持解锁奖励。

## MVP Scope

- 成就定义。
- 玩家统计数据。
- 成就进度。
- 成就解锁。
- 解锁幂等。

## Later Scope

- 周期任务。
- 每日任务。
- 成就分组。
- 平台成就同步。
- 自动奖励。

## Responsibilities

- 维护成就定义。
- 更新玩家统计。
- 计算进度。
- 解锁成就。
- 触发奖励或事件。

## Boundaries

- 不直接调用平台 SDK，同步平台成就必须通过 PlatformAdapter。
- 不直接写背包或货币，奖励必须通过 Inventory。
- 不签发 Token。

## Data

核心模型见 `docs/data-model.md`：

- `player_stat`
- `achievement`
- `player_achievement`

## gRPC

建议 service：

```text
AchievementService.UpdateStat
AchievementService.GetStats
AchievementService.ListAchievements
AchievementService.GetPlayerAchievements
AchievementService.UnlockAchievement
```

## Tests

- 更新统计成功。
- 成就进度计算正确。
- 成就解锁幂等。
- 已解锁成就不会重复发奖励。
- 平台同步失败不影响本地解锁状态。

