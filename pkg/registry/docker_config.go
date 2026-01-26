package registry

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// DockerConfig represents the Docker config.json structure
type DockerConfig struct {
	Auths       map[string]DockerAuth `json:"auths"`
	CredsStore  string                `json:"credsStore,omitempty"`
	CredHelpers map[string]string     `json:"credHelpers,omitempty"`
}

// DockerAuth represents auth entry in Docker config
type DockerAuth struct {
	Auth     string `json:"auth,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// LoadDockerConfig loads credentials from Docker config.json
func LoadDockerConfig() (*DockerConfig, error) {
	configPath := getDockerConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &DockerConfig{Auths: make(map[string]DockerAuth)}, nil
		}
		return nil, err
	}

	var config DockerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.Auths == nil {
		config.Auths = make(map[string]DockerAuth)
	}

	return &config, nil
}

// GetCredentials returns username and password for a registry
func (dc *DockerConfig) GetCredentials(registry string) (username, password string, found bool) {
	// Normalize registry URL
	registry = normalizeRegistry(registry)

	auth, ok := dc.Auths[registry]
	if !ok {
		// Try alternative forms
		alternatives := []string{
			"https://" + registry,
			"http://" + registry,
			registry + "/v1/",
			registry + "/v2/",
		}
		for _, alt := range alternatives {
			if auth, ok = dc.Auths[alt]; ok {
				break
			}
		}
	}

	if !ok {
		return "", "", false
	}

	// If auth field is set, decode it (base64 encoded "username:password")
	if auth.Auth != "" {
		decoded, err := base64.StdEncoding.DecodeString(auth.Auth)
		if err == nil {
			parts := strings.SplitN(string(decoded), ":", 2)
			if len(parts) == 2 {
				return parts[0], parts[1], true
			}
		}
	}

	// Otherwise use username/password fields directly
	if auth.Username != "" {
		return auth.Username, auth.Password, true
	}

	return "", "", false
}

// HasCredentials checks if credentials exist for a registry
func (dc *DockerConfig) HasCredentials(registry string) bool {
	_, _, found := dc.GetCredentials(registry)
	return found
}

func getDockerConfigPath() string {
	// Check DOCKER_CONFIG env var first
	if dockerConfig := os.Getenv("DOCKER_CONFIG"); dockerConfig != "" {
		return filepath.Join(dockerConfig, "config.json")
	}

	// Default to ~/.docker/config.json
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".docker", "config.json")
}

func normalizeRegistry(registry string) string {
	// Remove protocol prefix
	registry = strings.TrimPrefix(registry, "https://")
	registry = strings.TrimPrefix(registry, "http://")

	// Handle Docker Hub special case
	if registry == "docker.io" || registry == "registry-1.docker.io" {
		return "https://index.docker.io/v1/"
	}

	return registry
}
