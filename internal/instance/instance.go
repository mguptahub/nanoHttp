package instance

import (
	"fmt"
	"sync"

	"github.com/mguptahub/nanoHttp/internal/config"
	"github.com/mguptahub/nanoHttp/internal/server"
)

// Manager handles multiple server instances
type Manager struct {
	servers map[string]*server.Server
	mu      sync.RWMutex
}

// NewManager creates a new instance manager
func NewManager() *Manager {
	return &Manager{
		servers: make(map[string]*server.Server),
	}
}

// StartInstance starts a server instance
func (m *Manager) StartInstance(cfg config.InstanceConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.servers[cfg.Name]; exists {
		return fmt.Errorf("instance %s already exists", cfg.Name)
	}

	srv := server.NewServer(cfg)
	if err := srv.Start(); err != nil {
		return fmt.Errorf("failed to start instance %s: %v", cfg.Name, err)
	}

	m.servers[cfg.Name] = srv
	return nil
}

// StopInstance stops a server instance
func (m *Manager) StopInstance(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	srv, exists := m.servers[name]
	if !exists {
		return fmt.Errorf("instance %s does not exist", name)
	}

	if err := srv.Stop(); err != nil {
		return fmt.Errorf("failed to stop instance %s: %v", name, err)
	}

	delete(m.servers, name)
	return nil
}

// GetInstance returns a server instance by name
func (m *Manager) GetInstance(name string) (*server.Server, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	srv, exists := m.servers[name]
	return srv, exists
}

// ListInstances returns a list of all server instances
func (m *Manager) ListInstances() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instances := make([]string, 0, len(m.servers))
	for name := range m.servers {
		instances = append(instances, name)
	}
	return instances
}

// IsInstanceRunning checks if an instance is running
func (m *Manager) IsInstanceRunning(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	srv, exists := m.servers[name]
	if !exists {
		return false
	}
	return srv.IsRunning()
}
