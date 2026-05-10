# Data Model

本文档描述核心数据模型。具体字段可随实现微调，但实体关系必须保持清晰。

## User and Identity

```text
user
  id
  app_id
  status
  created_at
  updated_at

user_profile
  id
  user_id
  display_name
  avatar_url
  locale
  timezone
  created_at
  updated_at

account_identity
  id
  user_id
  type
  identifier
  verified_at
  created_at
  updated_at

platform_account
  id
  user_id
  provider
  platform_user_id
  display_name
  avatar_url
  linked_at
  last_seen_at
  raw_profile

user_session
  id
  user_id
  refresh_token_hash
  device_id
  ip_address
  user_agent
  expires_at
  revoked_at
  created_at

login_event
  id
  user_id
  provider
  ip_address
  user_agent
  success
  failure_code
  created_at
```

约束：

- `(provider, platform_user_id)` 必须唯一。
- Refresh Token 只存 hash。
- `user.status` 至少支持 active、banned、deleted。

## Leaderboard

见 `docs/modules/leaderboards.md`。

## Chat

见 `docs/modules/chat.md`。

## Friends

```text
friend_request
  id
  requester_id
  addressee_id
  status
  message
  created_at
  updated_at

friend_relation
  id
  user_id
  friend_user_id
  created_at

block_relation
  id
  user_id
  blocked_user_id
  reason
  created_at

presence_status
  user_id
  status
  last_seen_at
  metadata
```

约束：

- 好友关系可以双向存两行，查询简单；也可以单行无序对，具体实现需统一。
- 黑名单必须影响好友申请和私聊。
- 在线状态优先放 Redis，数据库只保留最后在线时间。

## Achievements and Stats

```text
player_stat
  id
  user_id
  key
  value
  updated_at

achievement
  id
  app_id
  code
  name
  description
  target
  status
  created_at
  updated_at

player_achievement
  id
  user_id
  achievement_id
  progress
  unlocked_at
  updated_at
```

约束：

- 成就解锁必须幂等。
- 统计数据更新必须支持原子增量。

## Inventory and Wallet

```text
inventory_item
  id
  user_id
  item_code
  quantity
  metadata
  created_at
  updated_at

wallet_balance
  id
  user_id
  currency_code
  balance
  updated_at

asset_ledger
  id
  user_id
  asset_type
  asset_code
  delta
  reason
  reference_type
  reference_id
  created_at
```

约束：

- 所有资产变化必须写流水。
- 奖励发放必须支持幂等 reference。

## Payments

```text
order
  id
  user_id
  provider
  product_id
  transaction_id
  status
  amount
  currency
  purchased_at
  created_at
  updated_at

payment_event
  id
  order_id
  provider
  event_type
  payload
  signature_valid
  created_at
```

约束：

- `(provider, transaction_id)` 必须唯一。
- 回调和收据校验必须幂等。

## Admin and Audit

```text
admin_user
  id
  email
  display_name
  status
  created_at
  updated_at

admin_role
  id
  code
  name

admin_user_role
  admin_user_id
  role_id

audit_log
  id
  actor_type
  actor_id
  action
  target_type
  target_id
  reason
  before
  after
  request_id
  ip_address
  created_at
```

约束：

- 高风险后台操作必须写 `audit_log`。
- 审计日志不可被普通后台用户修改。
