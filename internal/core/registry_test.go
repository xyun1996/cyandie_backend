package core

import (
	"testing"
)

type MockService struct {
	Name string
}

func TestRegistry_RegisterAndResolve(t *testing.T) {
	reg := NewServiceRegistry()
	svc := &MockService{Name: "test"}
	reg.Register("mock", svc)

	resolved, ok := reg.Resolve("mock")
	if !ok {
		t.Error("expected to resolve registered service")
	}
	if resolved.(*MockService).Name != "test" {
		t.Error("resolved service has wrong value")
	}
}

func TestRegistry_ResolveNotFound(t *testing.T) {
	reg := NewServiceRegistry()
	_, ok := reg.Resolve("nonexistent")
	if ok {
		t.Error("expected not found for unregistered service")
	}
}

func TestRegistry_MustResolve(t *testing.T) {
	reg := NewServiceRegistry()
	svc := &MockService{Name: "test"}
	reg.Register("mock", svc)

	resolved := reg.MustResolve("mock").(*MockService)
	if resolved.Name != "test" {
		t.Error("resolved service has wrong value")
	}
}

func TestRegistry_MustResolvePanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic for unregistered service")
		}
	}()
	reg := NewServiceRegistry()
	reg.MustResolve("nonexistent")
}
