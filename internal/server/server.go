package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/mguptahub/justHttp/internal/config"
)

// Instance represents a server instance configuration
type Instance struct {
	Name            string `json:"name"`
	Port            int    `json:"port"`
	WebFolder       string `json:"web_folder"`
	AllowDirListing bool   `json:"allow_dir_listing"`
	SSLCertFolder   string `json:"ssl_cert_folder,omitempty"`
	IsRunning       bool   `json:"is_running"`
}

// Manager manages multiple server instances
type Manager struct {
	configDir string
	servers   map[string]*Server
	mu        sync.RWMutex
}

// NewManager creates a new server manager
func NewManager(configDir string) *Manager {
	return &Manager{
		configDir: configDir,
		servers:   make(map[string]*Server),
	}
}

// AddInstance adds a new server instance
func (m *Manager) AddInstance(instance *Instance) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.servers[instance.Name]; exists {
		return fmt.Errorf("instance %s already exists", instance.Name)
	}

	cfg := config.InstanceConfig{
		Name:            instance.Name,
		Port:            instance.Port,
		WebFolder:       instance.WebFolder,
		AllowDirListing: instance.AllowDirListing,
		SSLCertFolder:   instance.SSLCertFolder,
	}

	server := NewServer(cfg)
	m.servers[instance.Name] = server

	// Save configuration
	return m.saveConfig()
}

// StartInstance starts a server instance
func (m *Manager) StartInstance(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	server, exists := m.servers[name]
	if !exists {
		return fmt.Errorf("instance %s not found", name)
	}

	if err := server.Start(); err != nil {
		return err
	}

	// Update configuration
	return m.saveConfig()
}

// StopInstance stops a server instance
func (m *Manager) StopInstance(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	server, exists := m.servers[name]
	if !exists {
		return fmt.Errorf("instance %s not found", name)
	}

	if err := server.Stop(); err != nil {
		return err
	}

	// Update configuration
	return m.saveConfig()
}

// DeleteInstance deletes a server instance
func (m *Manager) DeleteInstance(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	server, exists := m.servers[name]
	if !exists {
		return fmt.Errorf("instance %s not found", name)
	}

	if server.IsRunning() {
		if err := server.Stop(); err != nil {
			return err
		}
	}

	delete(m.servers, name)

	// Save configuration
	return m.saveConfig()
}

// ListInstances returns all server instances
func (m *Manager) ListInstances() []*Instance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instances := make([]*Instance, 0, len(m.servers))
	for name, server := range m.servers {
		instance := &Instance{
			Name:            name,
			Port:            server.config.Port,
			WebFolder:       server.config.WebFolder,
			AllowDirListing: server.config.AllowDirListing,
			SSLCertFolder:   server.config.SSLCertFolder,
			IsRunning:       server.IsRunning(),
		}
		instances = append(instances, instance)
	}
	return instances
}

// saveConfig saves the current configuration to disk
func (m *Manager) saveConfig() error {
	configPath := filepath.Join(m.configDir, "config.json")
	config := struct {
		Instances map[string]*Instance `json:"instances"`
	}{
		Instances: make(map[string]*Instance),
	}

	for name, server := range m.servers {
		config.Instances[name] = &Instance{
			Name:            name,
			Port:            server.config.Port,
			WebFolder:       server.config.WebFolder,
			AllowDirListing: server.config.AllowDirListing,
			SSLCertFolder:   server.config.SSLCertFolder,
			IsRunning:       server.IsRunning(),
		}
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("error writing config: %v", err)
	}

	return nil
}

// Server represents an HTTP server instance
type Server struct {
	config     config.InstanceConfig
	server     *http.Server
	mu         sync.Mutex
	isRunning  bool
	cancelFunc context.CancelFunc
}

// NewServer creates a new server instance
func NewServer(cfg config.InstanceConfig) *Server {
	return &Server{
		config: cfg,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("server %s is already running", s.config.Name)
	}

	// Create a new context for the server
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel

	// Create the server
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: s.createHandler(),
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	// Start the server in a goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error starting server %s: %v\n", s.config.Name, err)
		}
	}()

	s.isRunning = true
	return nil
}

// Stop stops the HTTP server
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return fmt.Errorf("server %s is not running", s.config.Name)
	}

	if s.cancelFunc != nil {
		s.cancelFunc()
	}

	if err := s.server.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("error shutting down server %s: %v", s.config.Name, err)
	}

	s.isRunning = false
	return nil
}

// IsRunning returns whether the server is currently running
func (s *Server) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isRunning
}

// createHandler creates the HTTP handler for the server
func (s *Server) createHandler() http.Handler {
	mux := http.NewServeMux()

	// Serve static files
	fileServer := http.FileServer(http.Dir(s.config.WebFolder))
	if s.config.AllowDirListing {
		mux.Handle("/", fileServer)
	} else {
		mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				indexPath := filepath.Join(s.config.WebFolder, "index.html")
				if _, err := os.Stat(indexPath); err == nil {
					http.ServeFile(w, r, indexPath)
					return
				}
			}
			http.NotFound(w, r)
		}))
	}

	return mux
}
