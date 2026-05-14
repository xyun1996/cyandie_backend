package platforms

import (
	"context"

	"github.com/cyandie/backend/internal/core"
	"github.com/cyandie/backend/internal/db"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	core.BaseModule
	handler  *Handler
	service  *PlatformService
	registry *PlatformRegistry
}

func NewModule(queries db.Querier, registry *PlatformRegistry) *Module {
	service := NewPlatformService(queries, registry)
	handler := NewHandler(service)
	return &Module{
		handler:  handler,
		service:  service,
		registry: registry,
	}
}

func (m *Module) Name() string { return "platforms" }

func (m *Module) Dependencies() []string { return []string{"auth", "users"} }

func (m *Module) RegisterServices(reg *core.ServiceRegistry) {
	reg.Register("platforms", m.service)
	reg.Register("platform-registry", m.registry)
}

func (m *Module) RegisterRoutes(router chi.Router) {
	m.handler.RegisterRoutes(router)
}

// RegisterPublicRoutes registers routes that don't require authentication.
func (m *Module) RegisterPublicRoutes(router chi.Router) {
	m.handler.RegisterPublicRoutes(router)
}

// RegisterProtectedRoutes registers routes that require authentication.
func (m *Module) RegisterProtectedRoutes(router chi.Router) {
	m.handler.RegisterProtectedRoutes(router)
}

func (m *Module) OnStart(_ context.Context) error { return nil }
func (m *Module) OnStop(_ context.Context) error  { return nil }
