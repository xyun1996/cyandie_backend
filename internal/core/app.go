package core

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

type App struct {
	modules  []Module
	registry *ServiceRegistry
	logger   *slog.Logger
}

func NewApp() *App {
	return &App{
		registry: NewServiceRegistry(),
		logger:   slog.Default(),
	}
}

func (a *App) Register(module Module) {
	a.modules = append(a.modules, module)
}

func (a *App) ModuleNames() []string {
	names := make([]string, len(a.modules))
	for i, m := range a.modules {
		names[i] = m.Name()
	}
	return names
}

func (a *App) Registry() *ServiceRegistry {
	return a.registry
}

func (a *App) SetLogger(l *slog.Logger) {
	a.logger = l
}

func (a *App) Start(ctx context.Context) error {
	// Register services
	for _, m := range a.modules {
		m.RegisterServices(a.registry)
	}

	// Start modules in order
	for _, m := range a.modules {
		a.logger.Info("starting module", "module", m.Name())
		if err := m.OnStart(ctx); err != nil {
			return fmt.Errorf("start module %s: %w", m.Name(), err)
		}
	}

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	// Stop modules in reverse order
	for i := len(a.modules) - 1; i >= 0; i-- {
		m := a.modules[i]
		a.logger.Info("stopping module", "module", m.Name())
		if err := m.OnStop(ctx); err != nil {
			a.logger.Error("stop module failed", "module", m.Name(), "error", err)
		}
	}
	return nil
}

func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := a.Start(ctx); err != nil {
		return err
	}

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	a.logger.Info("received signal, shutting down", "signal", sig)

	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	defer shutdownCancel()

	return a.Stop(shutdownCtx)
}
