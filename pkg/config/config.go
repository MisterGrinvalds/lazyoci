package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	// Registries is a list of configured OCI registries
	Registries []Registry `yaml:"registries"`

	// CacheDir is the directory for caching metadata
	CacheDir string `yaml:"cacheDir"`

	// DefaultRegistry is the registry to show on startup
	DefaultRegistry string `yaml:"defaultRegistry"`
}

// Registry represents an OCI registry configuration
type Registry struct {
	Name     string `yaml:"name"`
	URL      string `yaml:"url"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Insecure bool   `yaml:"insecure,omitempty"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		Registries: []Registry{
			{Name: "docker.io", URL: "docker.io"},
			{Name: "quay.io", URL: "quay.io"},
			{Name: "ghcr.io", URL: "ghcr.io"},
		},
		CacheDir:        filepath.Join(homeDir, ".cache", "lazyoci"),
		DefaultRegistry: "docker.io",
	}
}

// Load loads configuration from file or returns defaults
func Load() (*Config, error) {
	configPath := getConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save saves the configuration to file
func (c *Config) Save() error {
	configPath := getConfigPath()

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func getConfigPath() string {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "lazyoci", "config.yaml")
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "lazyoci", "config.yaml")
}
