package admin

import (
	"context"

	"github.com/cyandie/backend/internal/core"
	"github.com/cyandie/backend/internal/db"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	core.BaseModule
	handler *AdminHandler
	service *AdminService
}

func NewModule(queries db.Querier) *Module {
	service := NewAdminService(queries)
	handler := NewAdminHandler(service)
	return &Module{handler: handler, service: service}
}

func (m *Module) Name() string { return "admin" }

func (m *Module) RegisterServices(reg *core.ServiceRegistry) {
	reg.Register("admin", m.service)
}

func (m *Module) RegisterRoutes(router chi.Router) {
	mux := chi.NewRouter()
	mux.Post("/login", m.handler.Login)
	mux.Get("/users", m.handler.ListUsers)
	mux.Put("/users/{id}/status", m.handler.UpdateUserStatus)
	mux.Get("/audit-logs", m.handler.ListAuditLogs)
	router.Mount("/api/v1/admin", mux)
}

func (m *Module) OnStart(_ context.Context) error { return nil }
func (m *Module) OnStop(_ context.Context) error  { return nil }