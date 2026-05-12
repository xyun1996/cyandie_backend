package platforms

import (
	"strings"
	"testing"
)

func TestWeChatProvider_GetAuthURL(t *testing.T) {
	p := NewWeChatProvider(WeChatConfig{
		AppID: "wx123456", AppSecret: "secret", RedirectURI: "https://example.com/callback",
	})
	url := p.GetAuthURL("test-state")
	if !strings.Contains(url, "wx123456") {
		t.Error("expected app_id in URL")
	}
	if !strings.Contains(url, "test-state") {
		t.Error("expected state in URL")
	}
}

func TestWeChatProvider_Name(t *testing.T) {
	p := NewWeChatProvider(WeChatConfig{})
	if p.Name() != "wechat" {
		t.Errorf("expected wechat, got %s", p.Name())
	}
}
