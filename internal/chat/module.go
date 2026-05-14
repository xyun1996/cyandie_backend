package chat

import (
	"context"
	"time"

	"github.com/cyandie/backend/internal/core"
	"github.com/cyandie/backend/internal/db"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	core.BaseModule
	handler  *Handler
	service  *ChatService
	server   *TCPServer
	notifier *ChatPresenceNotifier
}

func NewModule(queries db.Querier, tcpAddr string, blockChecker BlockChecker, presenceSetter PresenceSetter, tokenValidator TokenValidator) *Module {
	server := NewTCPServer(tcpAddr, 30*time.Second, 90*time.Second)
	service := NewChatService(queries, server, blockChecker, presenceSetter, tokenValidator)
	handler := NewHandler(service)
	notifier := NewChatPresenceNotifier(server)
	return &Module{handler: handler, service: service, server: server, notifier: notifier}
}

func (m *Module) Name() string { return "chat" }

func (m *Module) RegisterServices(reg *core.ServiceRegistry) {
	reg.Register("chat", m.service)
}

func (m *Module) RegisterRoutes(router chi.Router) {
	m.handler.RegisterRoutes(router)
}

func (m *Module) OnStart(_ context.Context) error { return m.service.Start() }
func (m *Module) OnStop(_ context.Context) error  { return m.service.Stop() }

// PresenceNotifier returns the presence notifier for other modules to use.
func (m *Module) PresenceNotifier() PresenceNotifier { return m.notifier }

// SetBlockChecker allows late wiring of the block checker.
func (m *Module) SetBlockChecker(checker BlockChecker) {
	m.service.setBlockChecker(checker)
}

// SetPresenceSetter allows late wiring of the presence setter.
func (m *Module) SetPresenceSetter(setter PresenceSetter) {
	m.service.presenceSetter = setter
}

// SetTokenValidator allows late wiring of the token validator.
func (m *Module) SetTokenValidator(validator TokenValidator) {
	m.service.tokenValidator = validator
}