package leaderboard

import (
	"context"

	"github.com/cyandie/backend/internal/core"
	"github.com/cyandie/backend/internal/db"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	core.BaseModule
	handler *Handler
	service *LeaderboardService
}

func NewModule(queries db.Querier, redis redisClient) *Module {
	service := NewLeaderboardService(queries, redis)
	handler := NewHandler(service)
	return &Module{handler: handler, service: service}
}

func (m *Module) Name() string { return "leaderboard" }

func (m *Module) RegisterServices(reg *core.ServiceRegistry) {
	reg.Register("leaderboard", m.service)
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
