package platforms

type PlatformRegistry struct {
	oauthProviders map[string]OAuthProvider
}

func NewPlatformRegistry() *PlatformRegistry {
	return &PlatformRegistry{
		oauthProviders: make(map[string]OAuthProvider),
	}
}

func (r *PlatformRegistry) RegisterOAuth(provider OAuthProvider) {
	r.oauthProviders[provider.Name()] = provider
}

func (r *PlatformRegistry) GetOAuth(name string) (OAuthProvider, bool) {
	p, ok := r.oauthProviders[name]
	return p, ok
}

func (r *PlatformRegistry) ListPlatforms() []string {
	names := make([]string, 0, len(r.oauthProviders))
	for name := range r.oauthProviders {
		names = append(names, name)
	}
	return names
}
