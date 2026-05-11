package core

import (
	"context"

	"github.com/go-chi/chi/v5"
)

type Module interface {
	Name() string
	Dependencies() []string
	RegisterServices(reg *ServiceRegistry)
	RegisterRoutes(router chi.Router)
	OnStart(ctx context.Context) error
	OnStop(ctx context.Context) error
}

// BaseModule provides default implementations for modules that don't need all hooks.
type BaseModule struct{}

func (BaseModule) Dependencies() []string            { return nil }
func (BaseModule) RegisterServices(*ServiceRegistry) {}
func (BaseModule) RegisterRoutes(chi.Router)         {}
func (BaseModule) OnStart(context.Context) error      { return nil }
func (BaseModule) OnStop(context.Context) error       { return nil }
