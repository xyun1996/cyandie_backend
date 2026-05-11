package core

import "fmt"

type ServiceRegistry struct {
	services map[string]any
}

func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]any),
	}
}

func (r *ServiceRegistry) Register(name string, service any) {
	r.services[name] = service
}

func (r *ServiceRegistry) Resolve(name string) (any, bool) {
	svc, ok := r.services[name]
	return svc, ok
}

func (r *ServiceRegistry) MustResolve(name string) any {
	svc, ok := r.services[name]
	if !ok {
		panic(fmt.Sprintf("service %q not registered", name))
	}
	return svc
}
