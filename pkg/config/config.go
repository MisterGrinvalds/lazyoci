package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// artifactDirOverride is set by CLI flag and takes highest priority
var artifactDirOverride string

// SetArtifactDirOverride sets the CLI override for artifact directory.
// This should be called by the CLI flag processing.
func SetArtifactDirOverride(path string) {
	artifactDirOverride = path
}

// GetArtifactDirOverride returns the current CLI override (for testing).
func GetArtifactDirOverride() string {
	return artifactDirOverride
}

// Config holds the application configuration
type Config struct {
	// Registries is a list of configured OCI registries
	Registries []Registry `yaml:"registries"`

	// CacheDir is the directory for caching metadata
	CacheDir string `yaml:"cacheDir"`

	// ArtifactDir is the directory for storing pulled artifacts
	ArtifactDir string `yaml:"artifactDir,omitempty"`

	// DefaultRegistry is the registry to show on startup
	DefaultRegistry string `yaml:"defaultRegistry"`

	// Theme is the name of the active color theme
	Theme string `yaml:"theme,omitempty"`

	// Mode controls dark/light mode: "auto", "dark", or "light"
	Mode string `yaml:"mode,omitempty"`
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
			{Name: "Docker Hub", URL: "docker.io"},
			{Name: "Quay.io", URL: "quay.io"},
			{Name: "GitHub Packages", URL: "ghcr.io"},
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

	return os.WriteFile(configPath, data, 0600)
}

// AddRegistry adds a new registry to the configuration.
// If name is empty, it defaults to the URL.
func (c *Config) AddRegistry(name, url string) error {
	// Check if already exists
	for _, r := range c.Registries {
		if r.URL == url {
			return nil // Already exists
		}
	}

	if name == "" {
		name = url
	}

	c.Registries = append(c.Registries, Registry{
		Name: name,
		URL:  url,
	})

	return c.Save()
}

// AddRegistryWithAuth adds a registry with authentication.
// If name is empty, it defaults to the URL.
func (c *Config) AddRegistryWithAuth(name, url, username, password string) error {
	// Remove if exists
	c.RemoveRegistry(url)

	if name == "" {
		name = url
	}

	c.Registries = append(c.Registries, Registry{
		Name:     name,
		URL:      url,
		Username: username,
		Password: password,
	})

	return c.Save()
}

// RemoveRegistry removes a registry from the configuration
func (c *Config) RemoveRegistry(url string) error {
	for i, r := range c.Registries {
		if r.URL == url {
			c.Registries = append(c.Registries[:i], c.Registries[i+1:]...)
			return c.Save()
		}
	}
	return nil
}

// GetRegistry returns a registry by URL
func (c *Config) GetRegistry(url string) *Registry {
	for i := range c.Registries {
		if c.Registries[i].URL == url {
			return &c.Registries[i]
		}
	}
	return nil
}

func getConfigPath() string {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "lazyoci", "config.yaml")
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "lazyoci", "config.yaml")
}

// GetConfigPath returns the path to the config file (exported for CLI).
func GetConfigPath() string {
	return getConfigPath()
}

// GetArtifactDir returns the artifact directory with priority resolution:
// 1. CLI flag (--artifact-dir) via SetArtifactDirOverride
// 2. Environment variable ($LAZYOCI_ARTIFACT_DIR)
// 3. Config file (artifactDir in config.yaml)
// 4. Default (~/.cache/lazyoci/artifacts)
func (c *Config) GetArtifactDir() string {
	// 1. Check for CLI override (highest priority)
	if artifactDirOverride != "" {
		return ExpandPath(artifactDirOverride)
	}

	// 2. Check environment variable
	if envDir := os.Getenv("LAZYOCI_ARTIFACT_DIR"); envDir != "" {
		return ExpandPath(envDir)
	}

	// 3. Check config file
	if c.ArtifactDir != "" {
		return ExpandPath(c.ArtifactDir)
	}

	// 4. Default
	return DefaultArtifactDir()
}

// DefaultArtifactDir returns the default artifact directory.
func DefaultArtifactDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".cache", "lazyoci", "artifacts")
}

// SetArtifactDir sets the artifact directory in config after validation.
// Pass createDir=true to create the directory if it doesn't exist.
func (c *Config) SetArtifactDir(path string, createDir bool) error {
	expanded := ExpandPath(path)

	// Validate path
	if err := ValidatePath(expanded, createDir); err != nil {
		return err
	}

	c.ArtifactDir = path // Store original (possibly with ~)
	return c.Save()
}

// ExpandPath expands ~ to the user's home directory.
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(homeDir, path[2:])
	}
	if path == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return homeDir
	}
	return path
}

// ValidatePath checks if a path is valid for use as artifact directory.
// If createDir is true and path doesn't exist, it will be created.
func ValidatePath(path string, createDir bool) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		if createDir {
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			return nil
		}
		return fmt.Errorf("directory does not exist: %s", path)
	}
	if err != nil {
		return fmt.Errorf("failed to stat path: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	// Check if writable by creating a temp file
	testFile := filepath.Join(path, ".lazyoci-write-test")
	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("directory is not writable: %s", path)
	}
	f.Close()
	os.Remove(testFile)

	return nil
}

// PathExists checks if a path exists.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsArtifactDirDefault returns true if using the default artifact directory.
func (c *Config) IsArtifactDirDefault() bool {
	return c.GetArtifactDir() == DefaultArtifactDir()
}

// GetTheme returns the configured theme name, defaulting to "default".
func (c *Config) GetTheme() string {
	if c.Theme == "" {
		return "default"
	}
	return c.Theme
}

// SetTheme sets the theme name and saves the config.
func (c *Config) SetTheme(name string) error {
	c.Theme = name
	return c.Save()
}

// GetMode returns the configured mode, defaulting to "auto".
func (c *Config) GetMode() string {
	if c.Mode == "" {
		return "auto"
	}
	return c.Mode
}

// SetMode sets the dark/light mode and saves the config.
func (c *Config) SetMode(mode string) error {
	c.Mode = mode
	return c.Save()
}
