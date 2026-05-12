package platforms

import (
	"context"
	"testing"
)

type mockProvider struct {
	name string
}

func (m *mockProvider) Name() string                                                  { return m.name }
func (m *mockProvider) GetAuthURL(state string) string                                { return "https://example.com/auth?state=" + state }
func (m *mockProvider) ExchangeCode(_ context.Context, _ string) (*OAuthToken, error) { return &OAuthToken{AccessToken: "tok"}, nil }
func (m *mockProvider) GetUserInfo(_ context.Context, _ *OAuthToken) (*PlatformUser, error) {
	return &PlatformUser{PlatformID: "id1", DisplayName: "Test User"}, nil
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	reg := NewPlatformRegistry()
	reg.RegisterOAuth(&mockProvider{name: "test"})
	p, ok := reg.GetOAuth("test")
	if !ok {
		t.Error("expected to find registered provider")
	}
	if p.Name() != "test" {
		t.Errorf("expected test, got %s", p.Name())
	}
}

func TestRegistry_GetNotFound(t *testing.T) {
	reg := NewPlatformRegistry()
	_, ok := reg.GetOAuth("nonexistent")
	if ok {
		t.Error("expected not found")
	}
}

func TestRegistry_ListPlatforms(t *testing.T) {
	reg := NewPlatformRegistry()
	reg.RegisterOAuth(&mockProvider{name: "wechat"})
	reg.RegisterOAuth(&mockProvider{name: "discord"})
	if len(reg.ListPlatforms()) != 2 {
		t.Errorf("expected 2 platforms")
	}
}
