package chat

import (
	"context"

	"github.com/cyandie/backend/internal/core"
	"github.com/cyandie/backend/internal/db"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	core.BaseModule
	handler *Handler
	service *ChatService
	server  *TCPServer
}

func NewModule(queries db.Querier, tcpAddr string) *Module {
	server := NewTCPServer(tcpAddr)
	service := NewChatService(queries, server)
	handler := NewHandler(service)
	return &Module{handler: handler, service: service, server: server}
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
