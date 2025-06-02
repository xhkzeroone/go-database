package db

import (
	"fmt"
	"log"
	"sync"
)

type Manager struct {
	mu        sync.RWMutex
	instances map[string]*DataSource
}

// NewManager creates a new Manager
func NewManager() *Manager {
	return &Manager{
		instances: make(map[string]*DataSource),
	}
}

// Register a new database instance with optional overrides
func (m *Manager) Register(name string, cfg *Config, opts ...Option) error {
	db, err := Open(cfg, opts...)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.instances[name] = db
	return nil
}

// Get retrieves a registered database by name
func (m *Manager) Get(name string) (*DataSource, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	db, ok := m.instances[name]
	if !ok {
		return nil, fmt.Errorf("database instance not found: %s", name)
	}
	return db, nil
}

// CloseAll closes all database connections
func (m *Manager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, db := range m.instances {
		if err := db.Close(); err != nil {
			log.Printf("failed to close db [%s]: %v", name, err)
		}
	}
	return nil
}
