# Payments

Payments 模块负责订单、平台收据校验、Webhook、订单状态机和发货编排。

## Goals

- 支持多平台购买校验。
- 保证订单状态清晰。
- 保证发货幂等。
- 后台可查询和审计订单。

## MVP Scope

- 订单模型。
- 订单状态机。
- 平台收据校验入口。
- Webhook 处理入口。
- 幂等键。
- 发货编排接口。

## Later Scope

- Steam 支付。
- Epic 购买和 entitlement。
- Apple IAP。
- Google Play Billing。
- 退款和撤销。
- 对账。

## Responsibilities

- 创建和查询订单。
- 调用 `platforms` 验证购买。
- 调用 `inventory` 发货。
- 处理平台 Webhook。
- 记录支付事件。

## Boundaries

- 不直接写背包或货币余额，必须调用 Inventory。
- 不签发 Token。
- 不直接调用平台 SDK，必须通过 PlatformAdapter。
- 不重复发货。

## Data

核心模型见 `docs/data-model.md`：

- `order`
- `payment_event`

## HTTP APIs

```text
POST /api/v1/payments/verify
GET /api/v1/payments/orders/:id
POST /api/v1/platform-webhooks/:provider
```

后台 API：

```text
GET /api/v1/admin/orders
GET /api/v1/admin/orders/:id
```

## gRPC

建议 service：

```text
PaymentService.VerifyReceipt
PaymentService.HandleWebhook
PaymentService.GetOrder
PaymentService.ListOrders
```

## Tests

- 有效收据创建订单。
- 无效收据返回 `PAYMENT_INVALID_RECEIPT`。
- 重复回调不会重复发货。
- 平台验签失败不会更新订单。
- 订单状态变化有支付事件记录。

