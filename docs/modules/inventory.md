# Inventory

Inventory 模块负责背包、货币、奖励发放和资产流水。它是所有资产变化的唯一入口。

## Goals

- 管理用户虚拟资产。
- 保证资产变化可追溯。
- 支持奖励发放幂等。
- 支持支付、排行榜、成就、任务等模块发放奖励。

## MVP Scope

- 物品数量。
- 货币余额。
- 资产流水。
- 奖励发放接口。
- 幂等 reference。

## Responsibilities

- 增减物品。
- 增减货币。
- 写资产流水。
- 校验余额。
- 提供幂等发放。

## Boundaries

- 不判断平台购买是否有效。
- 不处理支付 Webhook。
- 不直接决定排行榜奖励资格。
- 不签发 Token。

## Data

核心模型见 `docs/data-model.md`：

- `inventory_item`
- `wallet_balance`
- `asset_ledger`

## gRPC

建议 service：

```text
InventoryService.GrantReward
InventoryService.AddItem
InventoryService.ConsumeItem
InventoryService.AddCurrency
InventoryService.ConsumeCurrency
InventoryService.GetInventory
InventoryService.GetWallet
```

## Tests

- 发放物品成功。
- 发放货币成功。
- 余额不足消费失败。
- 同一幂等 reference 不会重复发放。
- 所有资产变化写流水。

