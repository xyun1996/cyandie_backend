# Friends

Friends 模块负责好友、好友申请、黑名单和在线状态查询。它不负责聊天消息投递，但会影响 Chat 的私聊权限。

## Goals

- 支持常见社交关系。
- 支持黑名单。
- 提供在线状态查询。
- 为 Chat、Invite、Presence 等模块提供关系判断。

## MVP Scope

- 发送好友申请。
- 接受/拒绝好友申请。
- 好友列表。
- 删除好友。
- 黑名单。
- 在线状态查询。

## Later Scope

- 分组。
- 备注。
- 最近联系人。
- 平台好友导入。
- 推荐好友。

## Responsibilities

- 维护好友申请状态。
- 维护好友关系。
- 维护黑名单。
- 查询在线状态。
- 给 Chat 提供是否允许私聊的判断。

## Boundaries

- 不发送聊天消息。
- 不创建本地用户。
- 不签发 Token。
- 不直接调用平台 SDK；平台好友导入必须通过 `platforms` 模块。

## Data

核心模型见 `docs/data-model.md`：

- `friend_request`
- `friend_relation`
- `block_relation`
- `presence_status`

## HTTP APIs

```text
POST /api/v1/friends/requests
GET /api/v1/friends/requests
POST /api/v1/friends/requests/:id/accept
POST /api/v1/friends/requests/:id/reject
GET /api/v1/friends
DELETE /api/v1/friends/:userId
POST /api/v1/blocks/:userId
DELETE /api/v1/blocks/:userId
```

## gRPC

建议 service：

```text
FriendService.SendRequest
FriendService.AcceptRequest
FriendService.RejectRequest
FriendService.ListFriends
FriendService.RemoveFriend
FriendService.BlockUser
FriendService.UnblockUser
FriendService.CanDirectMessage
```

## Tests

- 发送好友申请成功。
- 不能重复发送同一好友申请。
- 接受好友申请后生成好友关系。
- 黑名单阻止好友申请。
- 黑名单阻止私聊。
- 删除好友后双方关系不可见。

