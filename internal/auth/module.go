package auth

import (
	"context"
	"net/http"

	"github.com/cyandie/backend/internal/core"
	"github.com/cyandie/backend/internal/db"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	core.BaseModule
	handler *Handler
	svc     *AuthService
}

func NewModule(queries db.Querier, km *KeyManager, sessions *SessionStore) *Module {
	svc := NewAuthService(AuthServiceDeps{
		Queries:     queries,
		KeyManager:  km,
		Sessions:    sessions,
		OTPNotifier: LogNotifier{},
	})
	handler := NewHandler(svc)
	return &Module{
		handler: handler,
		svc:     svc,
	}
}

func (m *Module) Name() string { return "auth" }

// Service returns the auth service for middleware wiring.
func (m *Module) Service() *AuthService { return m.svc }

// RequireAdmin returns middleware that rejects requests from non-admin users.
func (m *Module) RequireAdmin() func(http.Handler) http.Handler {
	return RequireAdmin()
}

func (m *Module) Dependencies() []string { return []string{"users"} }

func (m *Module) RegisterServices(reg *core.ServiceRegistry) {
	reg.Register("auth", m.svc)
}

func (m *Module) RegisterRoutes(router chi.Router) {
	m.handler.RegisterRoutes(router)
}

func (m *Module) OnStart(_ context.Context) error { return nil }
func (m *Module) OnStop(_ context.Context) error  { return nil }
