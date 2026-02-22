package mirror

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// SourceType identifies how a chart is pulled from upstream.
type SourceType string

const (
	// SourceRepo is a traditional Helm chart repository (index.yaml based).
	SourceRepo SourceType = "repo"
	// SourceOCI is an OCI registry hosting chart artifacts.
	SourceOCI SourceType = "oci"
	// SourceLocal is a chart directory on disk (e.g. a git submodule).
	SourceLocal SourceType = "local"
)

// Config is the top-level mirror configuration, typically loaded from a
// mirror.yaml file.
type Config struct {
	// Target is the destination registry for all mirrored artifacts.
	Target TargetConfig `yaml:"target"`
	// Upstreams maps chart keys to their upstream source definitions.
	Upstreams map[string]Upstream `yaml:"upstreams"`
}

// TargetConfig describes the destination OCI registry.
type TargetConfig struct {
	// URL is the registry host and optional path prefix
	// (e.g. "registry.digitalocean.com/greenforests").
	URL string `yaml:"url"`
	// Insecure allows plain HTTP connections.
	Insecure bool `yaml:"insecure,omitempty"`
	// ChartsPrefix is a path segment inserted between the registry URL and
	// the chart name so that chart OCI artifacts live in their own namespace
	// (e.g. "charts" → url/charts/<name>:<ver>).
	ChartsPrefix string `yaml:"charts-prefix,omitempty"`
}

// ChartOCIBase returns the OCI base path for chart artifacts.
// If ChartsPrefix is set the result is "url/prefix", otherwise just "url".
func (t TargetConfig) ChartOCIBase() string {
	if t.ChartsPrefix != "" {
		return t.URL + "/" + t.ChartsPrefix
	}
	return t.URL
}

// Upstream describes where a single chart can be pulled from.
type Upstream struct {
	// Type is the source type: "repo", "oci", or "local".
	Type SourceType `yaml:"type"`

	// Repo is the Helm repository URL (type=repo only).
	// Example: "https://helm.releases.hashicorp.com"
	Repo string `yaml:"repo,omitempty"`

	// Registry is the OCI registry URL including oci:// prefix (type=oci only).
	// Example: "oci://registry-1.docker.io/bitnamicharts"
	Registry string `yaml:"registry,omitempty"`

	// Path is the local filesystem path to a chart directory (type=local only).
	// Resolved relative to the config file's directory.
	Path string `yaml:"path,omitempty"`

	// Chart is the chart name as published upstream.
	Chart string `yaml:"chart"`

	// Versions is the explicit list of chart versions to mirror.
	Versions []string `yaml:"versions"`
}

// Validate checks that the configuration is internally consistent.
func (c *Config) Validate() error {
	if c.Target.URL == "" {
		return fmt.Errorf("target.url is required")
	}
	if len(c.Upstreams) == 0 {
		return fmt.Errorf("at least one upstream is required")
	}
	for key, u := range c.Upstreams {
		if u.Chart == "" {
			return fmt.Errorf("upstream %q: chart name is required", key)
		}
		switch u.Type {
		case SourceRepo:
			if u.Repo == "" {
				return fmt.Errorf("upstream %q: repo URL is required for type=repo", key)
			}
		case SourceOCI:
			if u.Registry == "" {
				return fmt.Errorf("upstream %q: registry URL is required for type=oci", key)
			}
		case SourceLocal:
			if u.Path == "" {
				return fmt.Errorf("upstream %q: path is required for type=local", key)
			}
		default:
			return fmt.Errorf("upstream %q: unknown source type %q", key, u.Type)
		}
	}
	return nil
}

// LoadConfig reads and parses a mirror configuration file.
// Relative paths in the config (e.g. Upstream.Path for local charts)
// are resolved relative to the directory containing the config file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// Resolve relative paths for local upstreams.
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolving config path: %w", err)
	}
	configDir := filepath.Dir(absPath)
	for key, u := range cfg.Upstreams {
		if u.Type == SourceLocal && u.Path != "" && !filepath.IsAbs(u.Path) {
			u.Path = filepath.Join(configDir, u.Path)
			cfg.Upstreams[key] = u
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}
