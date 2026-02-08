package registry

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
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

// execCredentialHelper invokes a Docker credential helper binary to retrieve
// credentials for the given registry URL.
//
// The Docker credential helper protocol works as follows:
//   - The binary is named docker-credential-<helperName> and must be on $PATH.
//   - To retrieve credentials: pipe the registry URL to stdin of
//     `docker-credential-<helper> get` and parse the JSON response:
//     {"ServerURL": "...", "Username": "...", "Secret": "..."}
//   - To store credentials: pipe JSON to stdin of `docker-credential-<helper> store`
//   - To erase credentials: pipe the registry URL to stdin of
//     `docker-credential-<helper> erase`
//   - To list credentials: invoke `docker-credential-<helper> list`
//
// See https://docs.docker.com/engine/reference/commandline/login/#credential-helpers
func execCredentialHelper(helperName string, registryURL string) (username, password string, err error) {
	// The binary is named docker-credential-<helperName> and must be on $PATH.
	binaryName := "docker-credential-" + helperName

	cmd := exec.Command(binaryName, "get")
	cmd.Stdin = bytes.NewBufferString(registryURL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// If the binary is not found, return ErrNotImplemented so the chain skips this store.
		if errors.Is(err, exec.ErrNotFound) {
			return "", "", ErrNotImplemented
		}
		// The helper returned an error (e.g. "credentials not found").
		return "", "", fmt.Errorf("credential helper %s: %w (%s)", binaryName, err, strings.TrimSpace(stderr.String()))
	}

	// Parse the JSON response: {"ServerURL": "...", "Username": "...", "Secret": "..."}
	var response struct {
		ServerURL string `json:"ServerURL"`
		Username  string `json:"Username"`
		Secret    string `json:"Secret"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return "", "", fmt.Errorf("credential helper %s: failed to parse response: %w", binaryName, err)
	}

	if response.Username == "" && response.Secret == "" {
		return "", "", ErrCredentialsNotFound
	}

	return response.Username, response.Secret, nil
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
