package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/mguptahub/nanoHttp/internal/config"
)

// Instance represents a server instance configuration
type Instance struct {
	Name            string `json:"name"`
	Port            int    `json:"port"`
	WebFolder       string `json:"web_folder"`
	AllowDirListing bool   `json:"allow_dir_listing"`
	IsRunning       bool   `json:"is_running"`
	PID             int    `json:"pid,omitempty"`
}

// Manager manages multiple server instances
type Manager struct {
	configDir string
	servers   map[string]*Server
	mu        sync.RWMutex
}

// NewManager creates a new server manager
func NewManager(configDir string) *Manager {
	manager := &Manager{
		configDir: configDir,
		servers:   make(map[string]*Server),
	}

	// Load configuration from disk
	config, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return manager
	}

	// Create server instances from config
	for name, instance := range config.Instances {
		server := NewServer(instance)
		manager.servers[name] = server
	}

	return manager
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

	// Get the executable path
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %v", err)
	}

	// Start the server in a new process
	cmd := exec.Command(executable, "serve", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting server process: %v", err)
	}

	// Write PID file
	if err := os.WriteFile(server.getPIDFilePath(), []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644); err != nil {
		return fmt.Errorf("error writing PID file: %v", err)
	}

	server.pid = cmd.Process.Pid
	server.isRunning = true

	// Load and update configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	instance := cfg.Instances[name]
	instance.IsRunning = true
	instance.PID = cmd.Process.Pid
	cfg.Instances[name] = instance

	// Save configuration
	return config.SaveConfig(cfg)
}

// StopInstance stops a server instance
func (m *Manager) StopInstance(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	server, exists := m.servers[name]
	if !exists {
		return fmt.Errorf("instance %s not found", name)
	}

	// Check if running using PID file
	if !server.checkIfRunning() {
		return fmt.Errorf("server %s is not running", name)
	}

	// Read PID from file
	pidFile := server.getPIDFilePath()
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("error reading PID file: %v", err)
	}

	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return fmt.Errorf("error parsing PID: %v", err)
	}

	// Find and kill the process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("error finding process: %v", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("error stopping process: %v", err)
	}

	// Remove PID file
	if err := server.removePIDFile(); err != nil {
		return fmt.Errorf("error removing PID file: %v", err)
	}

	server.isRunning = false

	// Load and update configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	instance := cfg.Instances[name]
	instance.IsRunning = false
	instance.PID = 0
	cfg.Instances[name] = instance

	// Save configuration
	return config.SaveConfig(cfg)
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

	// Load configuration from disk
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return nil
	}

	instances := make([]*Instance, 0, len(cfg.Instances))
	for name, instance := range cfg.Instances {
		instances = append(instances, &Instance{
			Name:            name,
			Port:            instance.Port,
			WebFolder:       instance.WebFolder,
			AllowDirListing: instance.AllowDirListing,
			IsRunning:       instance.IsRunning,
			PID:             instance.PID,
		})
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
	pid        int
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

	// Check if already running using PID file
	if s.checkIfRunning() {
		return fmt.Errorf("server %s is already running", s.config.Name)
	}

	// Get the executable path
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %v", err)
	}

	// Start the server in a new process
	cmd := exec.Command(executable, "serve", s.config.Name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting server process: %v", err)
	}

	// Write PID file
	if err := os.WriteFile(s.getPIDFilePath(), []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644); err != nil {
		return fmt.Errorf("error writing PID file: %v", err)
	}

	s.pid = cmd.Process.Pid
	s.isRunning = true

	return nil
}

// Stop stops the HTTP server
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if running using PID file
	if !s.checkIfRunning() {
		return fmt.Errorf("server %s is not running", s.config.Name)
	}

	// Read PID from file
	pidFile := s.getPIDFilePath()
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("error reading PID file: %v", err)
	}

	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return fmt.Errorf("error parsing PID: %v", err)
	}

	// Find and kill the process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("error finding process: %v", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("error stopping process: %v", err)
	}

	// Remove PID file
	if err := s.removePIDFile(); err != nil {
		return fmt.Errorf("error removing PID file: %v", err)
	}

	s.isRunning = false
	return nil
}

// IsRunning returns whether the server is currently running
func (s *Server) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.checkIfRunning()
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

func (s *Server) getPIDFilePath() string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("nanohttp_%s.pid", s.config.Name))
}

func (s *Server) writePIDFile() error {
	pidFile := s.getPIDFilePath()
	return os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
}

func (s *Server) removePIDFile() error {
	pidFile := s.getPIDFilePath()
	return os.Remove(pidFile)
}

func (s *Server) checkIfRunning() bool {
	pidFile := s.getPIDFilePath()
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return false
	}

	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return false
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix-like systems, FindProcess always succeeds, so we need to check if the process is actually running
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// GetServer returns a server instance by name
func (m *Manager) GetServer(name string) (*Server, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	server, exists := m.servers[name]
	if !exists {
		return nil, fmt.Errorf("instance %s not found", name)
	}

	return server, nil
}

// GetConfig returns the server configuration
func (s *Server) GetConfig() config.InstanceConfig {
	return s.config
}

// CreateHandler creates and returns the HTTP handler for the server
func (s *Server) CreateHandler() http.Handler {
	return s.createHandler()
}
