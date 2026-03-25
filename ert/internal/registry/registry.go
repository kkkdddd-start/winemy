package registry

import (
	"context"
	"fmt"
	"sync"

	"github.com/yourname/ert/internal/model"
)

type Module interface {
	ID() int
	Name() string
	Priority() int
	Init(ctx context.Context, s *Storage) error
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
	storage *Storage
}

func New(s *Storage) *Registry {
	return &Registry{
		modules: make(map[int]Module),
		storage: s,
	}
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

func (r *Registry) GetData(ctx context.Context, moduleID int, query string) ([]map[string]interface{}, error) {
	m, err := r.Get(moduleID)
	if err != nil {
		return nil, err
	}

	switch m.ID() {
	case 2:
		return queryProcessData(ctx, r.storage, query)
	case 3:
		return queryNetworkData(ctx, r.storage, query)
	case 4:
		return queryRegistryData(ctx, r.storage, query)
	case 5:
		return queryServiceData(ctx, r.storage, query)
	default:
		return nil, fmt.Errorf("module %d does not support data query", moduleID)
	}
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

func queryProcessData(ctx context.Context, s *Storage, query string) ([]map[string]interface{}, error) {
	sql := "SELECT * FROM processes"
	if query != "" {
		sql += " WHERE " + query
	}
	return s.Query(ctx, sql)
}

func queryNetworkData(ctx context.Context, s *Storage, query string) ([]map[string]interface{}, error) {
	sql := "SELECT * FROM network_connections"
	if query != "" {
		sql += " WHERE " + query
	}
	return s.Query(ctx, sql)
}

func queryRegistryData(ctx context.Context, s *Storage, query string) ([]map[string]interface{}, error) {
	sql := "SELECT * FROM registry_keys"
	if query != "" {
		sql += " WHERE " + query
	}
	return s.Query(ctx, sql)
}

func queryServiceData(ctx context.Context, s *Storage, query string) ([]map[string]interface{}, error) {
	sql := "SELECT * FROM services"
	if query != "" {
		sql += " WHERE " + query
	}
	return s.Query(ctx, sql)
}
