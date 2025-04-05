package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// InstanceConfig represents the configuration for a single HTTP server instance
type InstanceConfig struct {
	Name            string `json:"name"`
	Port            int    `json:"port"`
	WebFolder       string `json:"web_folder"`
	AllowDirListing bool   `json:"allow_dir_listing"`
	SSLCertFolder   string `json:"ssl_cert_folder,omitempty"`
	IsRunning       bool   `json:"is_running"`
	PID             int    `json:"pid,omitempty"`
}

// Config represents the main configuration file structure
type Config struct {
	Instances map[string]InstanceConfig `json:"instances"`
}

// DefaultConfig returns a new default configuration
func DefaultConfig() *Config {
	return &Config{
		Instances: make(map[string]InstanceConfig),
	}
}

// GetConfigDir returns the path to the configuration directory
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(homeDir, ".nanoHttp")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return configDir, nil
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}

// LoadConfig loads the configuration from the config file
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig saves the configuration to the config file
func SaveConfig(config *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// AddInstance adds a new instance to the configuration
func AddInstance(config *Config, instance InstanceConfig) error {
	if instance.Port == 0 {
		instance.Port = 8080
	}
	config.Instances[instance.Name] = instance
	return SaveConfig(config)
}

// DeleteInstance removes an instance from the configuration
func DeleteInstance(config *Config, name string) error {
	delete(config.Instances, name)
	return SaveConfig(config)
}

// UpdateInstance updates an existing instance in the configuration
func UpdateInstance(config *Config, instance InstanceConfig) error {
	if _, exists := config.Instances[instance.Name]; !exists {
		return nil
	}
	config.Instances[instance.Name] = instance
	return SaveConfig(config)
}

// GetInstance retrieves an instance from the configuration
func (c *Config) GetInstance(name string) (InstanceConfig, bool) {
	instance, exists := c.Instances[name]
	return instance, exists
}
