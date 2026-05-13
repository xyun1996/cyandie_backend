package friends

import (
	"context"

	"github.com/cyandie/backend/internal/core"
	"github.com/cyandie/backend/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

type Module struct {
	core.BaseModule
	handler *FriendsHandler
	service *FriendsService
}

func NewModule(queries db.Querier, rdb *redis.Client) *Module {
	service := NewFriendsService(queries, rdb)
	handler := NewFriendsHandler(service)
	return &Module{handler: handler, service: service}
}

func (m *Module) Name() string { return "friends" }

func (m *Module) RegisterServices(reg *core.ServiceRegistry) {
	reg.Register("friends", m.service)
}

func (m *Module) RegisterRoutes(router chi.Router) {
	router.Mount("/api/v1/friends", m.handler.Routes())
}

func (m *Module) OnStart(_ context.Context) error { return nil }
func (m *Module) OnStop(_ context.Context) error  { return nil }