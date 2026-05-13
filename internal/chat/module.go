package chat

import (
	"context"

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

func NewModule(queries db.Querier, tcpAddr string, blockChecker BlockChecker) *Module {
	server := NewTCPServer(tcpAddr)
	service := NewChatService(queries, server, blockChecker)
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