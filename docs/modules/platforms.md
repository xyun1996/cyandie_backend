# Platform Adapters

平台接入必须通过统一适配器完成，避免 Steam、Epic、Apple、Google 等平台逻辑散落在业务代码里。

## Design Goals

- 平台差异在适配器内部消化。
- Auth、Payments、Achievements 等业务模块只依赖标准接口。
- 新增平台时不修改核心业务流程。
- 每个平台实现同一套 contract tests。

## Core Types

建议 Go 接口：

```go
type PlatformProvider string

const (
	PlatformSteam       PlatformProvider = "steam"
	PlatformEpic        PlatformProvider = "epic"
	PlatformApple       PlatformProvider = "apple"
	PlatformGoogle      PlatformProvider = "google"
	PlatformXbox        PlatformProvider = "xbox"
	PlatformPlayStation PlatformProvider = "playstation"
	PlatformNintendo    PlatformProvider = "nintendo"
)

type VerifyTokenInput struct {
	Token       string
	ClientID    string
	Nonce       string
	RedirectURI string
}

type PlatformUser struct {
	Provider       PlatformProvider
	PlatformUserID string
	DisplayName    string
	AvatarURL      string
	Email          string
	EmailVerified  bool
	Raw            []byte
}

type VerifyPurchaseInput struct {
	Receipt       string
	ProductID     string
	TransactionID string
}

type PurchaseResult struct {
	Provider       PlatformProvider
	PlatformUserID string
	TransactionID  string
	ProductID      string
	PurchasedAt    time.Time
	Valid          bool
	Raw            []byte
}

type PlatformAdapter interface {
	Provider() PlatformProvider
	VerifyToken(ctx context.Context, input VerifyTokenInput) (*PlatformUser, error)
	GetUserProfile(ctx context.Context, platformUserID string) (*PlatformUser, error)
	VerifyPurchase(ctx context.Context, input VerifyPurchaseInput) (*PurchaseResult, error)
}
```
```

## Registry

业务模块不得直接实例化适配器，必须通过 registry：

```go
type PlatformRegistry interface {
	Get(provider PlatformProvider) (PlatformAdapter, error)
	Has(provider PlatformProvider) bool
	List() []PlatformProvider
}
```

## Auth Integration

平台登录流程：

```text
AuthController
  -> AuthService.loginWithPlatform(provider, token)
  -> PlatformRegistry.get(provider).verifyToken(...)
  -> AccountIdentityRepository.findByProviderUserId(...)
  -> UserService.createOrGet(...)
  -> SessionService.issue(...)
```

适配器返回 `PlatformUser` 后，不允许自己创建本地用户。

## Steam MVP

第一版 Steam 只要求：

- 验证 Steam OpenID 或 session ticket。
- 返回标准化 `PlatformUser`。
- 支持配置 app id / api key。
- 失败时返回统一错误码。

后续能力：

- 查询用户资料。
- 校验应用/游戏所有权。
- 成就同步。
- 统计同步。
- DLC 校验。

## Epic MVP

第一版 Epic 只要求：

- 验证 Epic access token。
- 返回标准化 `PlatformUser`。
- 支持配置 client id / client secret。
- 失败时返回统一错误码。

后续能力：

- 查询用户资料。
- EOS Connect。
- 成就同步。
- 商品和 entitlement 查询。

## Contract Tests

每个平台适配器必须通过同一套行为测试：

- provider 名称正确。
- 无效 token 会被拒绝。
- 有效 token 返回标准化 `PlatformUser`。
- 返回结果必须包含 `provider` 和 `platformUserId`。
- 适配器不创建本地用户。
- 平台 API 错误会转换成项目统一错误。

## Error Mapping

建议错误码：

- `PLATFORM_UNSUPPORTED_PROVIDER`
- `PLATFORM_INVALID_TOKEN`
- `PLATFORM_TOKEN_EXPIRED`
- `PLATFORM_API_UNAVAILABLE`
- `PLATFORM_RATE_LIMITED`
- `PLATFORM_INVALID_PURCHASE`
- `PLATFORM_CONFIG_MISSING`

## Configuration

示例环境变量：

```text
STEAM_APP_ID=
STEAM_WEB_API_KEY=
EPIC_CLIENT_ID=
EPIC_CLIENT_SECRET=
EPIC_DEPLOYMENT_ID=
```

平台密钥不得写入源码或测试快照。
