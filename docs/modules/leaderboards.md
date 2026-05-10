# Leaderboard

排行榜模块负责榜单配置、分数提交、排名查询、赛季和奖励结算入口。

## Goals

- 支持高频提交和低延迟查询。
- 支持多种分数策略。
- 支持全局榜、赛季榜，后续支持好友榜。
- SQL 保存可审计数据，Redis 提供实时排名。

## MVP Scope

- 创建和查询榜单配置。
- 提交分数。
- 查询 Top N。
- 查询自己的排名。
- 查询自己附近排名。
- 支持最高分策略。
- Redis Sorted Set 实时排名。
- SQL 记录提交事件。

## Later Scope

- 周榜/月榜。
- 赛季榜。
- 好友榜。
- 多维榜单。
- 奖励发放。
- 反作弊策略。
- 榜单快照。
- 历史榜单查询。

## Data Model

建议表：

```text
leaderboard
  id
  app_id
  code
  name
  sort_order
  score_policy
  reset_policy
  status
  created_at
  updated_at

leaderboard_season
  id
  leaderboard_id
  code
  starts_at
  ends_at
  status
  created_at
  updated_at

leaderboard_score_event
  id
  leaderboard_id
  season_id
  user_id
  score
  metadata
  submitted_at
  request_id

leaderboard_score_snapshot
  id
  leaderboard_id
  season_id
  user_id
  score
  rank
  snapshot_at
```

## Redis Keys

```text
leaderboard:{appId}:{boardCode}:{seasonCode}
leaderboard:user-meta:{appId}:{boardCode}:{seasonCode}
```

Sorted Set member 使用本地 `userId`，score 使用归一化后的排名分数。

## Score Policies

MVP 至少支持：

- `highest`：分数越高越好。
- `lowest`：分数越低越好，例如竞速耗时。
- `latest`：最新提交覆盖。

注意 Redis Sorted Set 默认按 score 从小到大排序，查询高分榜时需要使用 reverse range。

## API

### Submit Score

```http
POST /api/v1/leaderboards/:boardId/scores
```

Request:

```json
{
  "score": 1200,
  "metadata": {
    "level": 3,
    "character": "knight"
  }
}
```

Response:

```json
{
  "ok": true,
  "data": {
    "accepted": true,
    "score": 1200,
    "rank": 42
  }
}
```

### Query Top

```http
GET /api/v1/leaderboards/:boardId/top?limit=50&cursor=
```

### Query Me

```http
GET /api/v1/leaderboards/:boardId/me
```

### Query Around Me

```http
GET /api/v1/leaderboards/:boardId/around-me?range=5
```

## Validation Rules

- 用户必须登录。
- 被封禁用户不能提交。
- 榜单必须处于 active 状态。
- 分数必须在配置允许范围内。
- 提交频率必须受限。
- 同一个 request id 重复提交必须幂等。

## Tests

必须覆盖：

- 创建榜单配置。
- 提交第一条分数。
- 更高分覆盖旧分。
- 更低分不覆盖最高分榜。
- 查询 Top N。
- 查询用户排名。
- 查询用户附近排名。
- 被封禁用户提交失败。
- Redis 不可用时返回可诊断错误。
