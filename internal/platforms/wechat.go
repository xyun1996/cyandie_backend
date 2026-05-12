package platforms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/cyandie/backend/internal/core/errors"
)

type WeChatConfig struct {
	AppID       string `yaml:"app_id" env:"WECHAT_APP_ID"`
	AppSecret   string `yaml:"app_secret" env:"WECHAT_APP_SECRET"`
	RedirectURI string `yaml:"redirect_uri" env:"WECHAT_REDIRECT_URI"`
}

type WeChatProvider struct {
	config WeChatConfig
	client *http.Client
}

func NewWeChatProvider(config WeChatConfig) *WeChatProvider {
	return &WeChatProvider{
		config: config,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *WeChatProvider) Name() string { return "wechat" }

func (w *WeChatProvider) GetAuthURL(state string) string {
	params := url.Values{}
	params.Set("appid", w.config.AppID)
	params.Set("redirect_uri", w.config.RedirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "snsapi_login")
	params.Set("state", state)
	return "https://open.weixin.qq.com/connect/qrconnect?" + params.Encode() + "#wechat_redirect"
}

func (w *WeChatProvider) ExchangeCode(_ context.Context, code string) (*OAuthToken, error) {
	tokenURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		w.config.AppID, w.config.AppSecret, code,
	)

	resp, err := w.client.Get(tokenURL)
	if err != nil {
		return nil, errors.New(errors.ErrPlatformAuthFailed, "wechat token exchange failed")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		AccessToken  string `json:"access_token"`
		OpenID       string `json:"openid"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		ErrCode      int    `json:"errcode"`
		ErrMsg       string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errors.New(errors.ErrPlatformAuthFailed, "wechat token response parse failed")
	}
	if result.ErrCode != 0 {
		return nil, errors.New(errors.ErrPlatformAuthFailed, fmt.Sprintf("wechat error: %d %s", result.ErrCode, result.ErrMsg))
	}

	return &OAuthToken{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		OpenID:       result.OpenID,
		ExpiresAt:    time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
	}, nil
}

func (w *WeChatProvider) GetUserInfo(_ context.Context, token *OAuthToken) (*PlatformUser, error) {
	userInfoURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s",
		token.AccessToken, token.OpenID,
	)

	resp, err := w.client.Get(userInfoURL)
	if err != nil {
		return nil, errors.New(errors.ErrPlatformAuthFailed, "wechat user info fetch failed")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		OpenID     string `json:"openid"`
		Nickname   string `json:"nickname"`
		HeadImgURL string `json:"headimgurl"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errors.New(errors.ErrPlatformAuthFailed, "wechat user info parse failed")
	}
	if result.ErrCode != 0 {
		return nil, errors.New(errors.ErrPlatformAuthFailed, fmt.Sprintf("wechat error: %d %s", result.ErrCode, result.ErrMsg))
	}

	return &PlatformUser{
		PlatformID:  result.OpenID,
		DisplayName: result.Nickname,
		AvatarURL:   result.HeadImgURL,
	}, nil
}
