package core

import (
	"context"
	"testing"
)

type testModule struct {
	BaseModule
	name    string
	started bool
	stopped bool
}

func (m *testModule) Name() string { return m.name }

func (m *testModule) OnStart(ctx context.Context) error {
	m.started = true
	return nil
}

func (m *testModule) OnStop(ctx context.Context) error {
	m.stopped = true
	return nil
}

func TestApp_RegisterAndStart(t *testing.T) {
	app := NewApp()
	mod := &testModule{name: "test"}
	app.Register(mod)

	ctx := context.Background()
	err := app.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if !mod.started {
		t.Error("expected module to be started")
	}
}

func TestApp_ModuleOrder(t *testing.T) {
	app := NewApp()
	m1 := &testModule{name: "alpha"}
	m2 := &testModule{name: "beta"}
	app.Register(m1)
	app.Register(m2)

	names := app.ModuleNames()
	if names[0] != "alpha" || names[1] != "beta" {
		t.Errorf("unexpected module order: %v", names)
	}
}

func TestApp_StopReverseOrder(t *testing.T) {
	app := NewApp()
	m1 := &testModule{name: "first"}
	m2 := &testModule{name: "second"}
	app.Register(m1)
	app.Register(m2)

	ctx := context.Background()
	app.Start(ctx)
	app.Stop(ctx)

	if !m1.stopped || !m2.stopped {
		t.Error("expected all modules to be stopped")
	}
}
