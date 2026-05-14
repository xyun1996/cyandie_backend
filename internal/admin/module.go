package admin

import (
	"context"

	"github.com/cyandie/backend/internal/auth"
	"github.com/cyandie/backend/internal/core"
	"github.com/cyandie/backend/internal/db"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	core.BaseModule
	handler *AdminHandler
	service *AdminService
	authSvc *auth.AuthService
}

func NewModule(queries db.Querier, authSvc *auth.AuthService) *Module {
	service := NewAdminService(queries, authSvc)
	handler := NewAdminHandler(service)
	return &Module{handler: handler, service: service, authSvc: authSvc}
}

func (m *Module) Name() string { return "admin" }

func (m *Module) RegisterServices(reg *core.ServiceRegistry) {
	reg.Register("admin", m.service)
}

func (m *Module) RegisterRoutes(router chi.Router) {
	mux := chi.NewRouter()
	mux.Post("/login", m.handler.Login)
	mux.Group(func(r chi.Router) {
		r.Use(auth.AuthGuard(m.authSvc))
		r.Use(auth.RequireAdmin())
		r.Get("/users", m.handler.ListUsers)
		r.Put("/users/{id}/status", m.handler.UpdateUserStatus)
		r.Get("/audit-logs", m.handler.ListAuditLogs)
	})
	router.Mount("/api/v1/admin", mux)
}

func (m *Module) OnStart(_ context.Context) error { return nil }
func (m *Module) OnStop(_ context.Context) error  { return nil }