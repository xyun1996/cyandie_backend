package users

import (
	"context"

	"github.com/cyandie/backend/internal/core"
	"github.com/cyandie/backend/internal/db"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	core.BaseModule
	handler *Handler
	svc     *UserService
}

func NewModule(queries db.Querier) *Module {
	svc := NewUserService(queries)
	handler := NewHandler(svc)
	return &Module{
		handler: handler,
		svc:     svc,
	}
}

func (m *Module) Name() string { return "users" }

func (m *Module) RegisterServices(reg *core.ServiceRegistry) {
	reg.Register("users", m.svc)
}

func (m *Module) RegisterRoutes(router chi.Router) {
	m.handler.RegisterRoutes(router)
}

func (m *Module) OnStart(_ context.Context) error { return nil }
func (m *Module) OnStop(_ context.Context) error  { return nil }
