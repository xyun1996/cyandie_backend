package friends

import (
	"context"

	"github.com/cyandie/backend/internal/core"
	"github.com/cyandie/backend/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

// PresenceSetter updates online/offline status for presence.
type PresenceSetter interface {
	SetOnline(ctx context.Context, userID, username string) error
	SetOffline(ctx context.Context, userID string) error
}

type Module struct {
	core.BaseModule
	handler *FriendsHandler
	service *FriendsService
}

// BlockChecker checks if a user is blocked by another.
// Defined locally to avoid circular import with chat package.
type BlockChecker interface {
	IsBlocked(ctx context.Context, targetUserID, byUserID string) (bool, error)
}

func NewModule(queries db.Querier, rdb *redis.Client, notifier PresenceNotifier) *Module {
	service := NewFriendsService(queries, rdb, notifier)
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

// BlockChecker returns the service as a BlockChecker for the chat module.
func (m *Module) BlockChecker() BlockChecker { return m.service }

// PresenceSetter returns the service as a PresenceSetter for the chat module.
func (m *Module) PresenceSetter() PresenceSetter { return m.service }

func (m *Module) OnStart(_ context.Context) error { return nil }
func (m *Module) OnStop(_ context.Context) error  { return nil }