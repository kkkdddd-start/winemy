package registry

import (
	"context"
	"fmt"
	"sync"
)

type Module interface {
	ID() int
	Name() string
	Priority() int
	Init(ctx context.Context, s Storage) error
	Collect(ctx context.Context) error
	Stop() error
}

type Storage interface {
	Write(ctx context.Context, table string, data interface{}) error
	Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error)
}

type Registry struct {
	modules map[int]Module
	mu      sync.RWMutex
	storage Storage
}

func New(s Storage) *Registry {
	r := &Registry{
		modules: make(map[int]Module),
	}
	if s != nil {
		r.storage = s
	} else {
		r.storage = &storageAdapter{}
	}
	return r
}

func (r *Registry) SetStorage(s Storage) {
	if s != nil {
		r.storage = s
	}
}

type storageAdapter struct{}

func (a *storageAdapter) Write(ctx context.Context, table string, data interface{}) error {
	return nil
}

func (a *storageAdapter) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	return nil, nil
}

func (r *Registry) Register(m Module) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.modules[m.ID()]; exists {
		return fmt.Errorf("module %d already registered", m.ID())
	}

	r.modules[m.ID()] = m
	return nil
}

func (r *Registry) Get(id int) (Module, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m, ok := r.modules[id]
	if !ok {
		return nil, fmt.Errorf("module %d not found", id)
	}
	return m, nil
}

func (r *Registry) List() []Module {
	r.mu.RLock()
	defer r.mu.RUnlock()

	modules := make([]Module, 0, len(r.modules))
	for _, m := range r.modules {
		modules = append(modules, m)
	}
	return modules
}

func (r *Registry) Init(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, m := range r.modules {
		if err := m.Init(ctx, r.storage); err != nil {
			return fmt.Errorf("failed to init module %d (%s): %w", m.ID(), m.Name(), err)
		}
	}
	return nil
}

func (r *Registry) Collect(ctx context.Context, moduleID int) error {
	m, err := r.Get(moduleID)
	if err != nil {
		return err
	}
	return m.Collect(ctx)
}

func (r *Registry) Stop(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, m := range r.modules {
		if err := m.Stop(); err != nil {
			return fmt.Errorf("failed to stop module %d (%s): %w", m.ID(), m.Name(), err)
		}
	}
	return nil
}

func (r *Registry) CollectModule(ctx context.Context, moduleID int) error {
	m, ok := r.modules[moduleID]
	if !ok {
		return fmt.Errorf("module %d not found", moduleID)
	}
	return m.Collect(ctx)
}

func (r *Registry) GetModuleData(ctx context.Context, moduleID int, query string) ([]map[string]interface{}, error) {
	m, ok := r.modules[moduleID]
	if !ok {
		return nil, fmt.Errorf("module %d not found", moduleID)
	}

	if module, ok := m.(interface {
		GetData() ([]map[string]interface{}, error)
	}); ok {
		return module.GetData()
	}

	return nil, fmt.Errorf("module %d does not support GetModuleData", moduleID)
}
