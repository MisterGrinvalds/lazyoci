package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mistergrinvalds/lazyoci/pkg/config"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

// Cache interface for the registry client
type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, data []byte, ttl time.Duration) error
}

// Client handles communication with OCI registries
type Client struct {
	config       *config.Config
	dockerConfig *DockerConfig
	registries   map[string]*remote.Registry
	cache        Cache
}

// NewClient creates a new registry client
func NewClient(cfg *config.Config) *Client {
	// Load Docker config for credentials
	dockerCfg, _ := LoadDockerConfig()

	return &Client{
		config:       cfg,
		dockerConfig: dockerCfg,
		registries:   make(map[string]*remote.Registry),
	}
}

// SetCache sets the cache for the client
func (c *Client) SetCache(cache Cache) {
	c.cache = cache
}

// getCached retrieves cached data
func (c *Client) getCached(key string, target interface{}) bool {
	if c.cache == nil {
		return false
	}
	data, ok := c.cache.Get(key)
	if !ok {
		return false
	}
	return json.Unmarshal(data, target) == nil
}

// setCache stores data in cache
func (c *Client) setCache(key string, data interface{}, ttl time.Duration) {
	if c.cache == nil {
		return
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return
	}
	c.cache.Set(key, encoded, ttl)
}

// GetRegistries returns the list of configured registries
func (c *Client) GetRegistries() []config.Registry {
	return c.config.Registries
}

// getRegistry returns a registry client for the given URL, creating one if necessary
func (c *Client) getRegistry(registryURL string) (*remote.Registry, error) {
	// Normalize registry URL
	actualURL := registryURL
	if registryURL == "docker.io" {
		actualURL = "registry-1.docker.io"
	}

	if reg, ok := c.registries[actualURL]; ok {
		return reg, nil
	}

	reg, err := remote.NewRegistry(actualURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create registry client: %w", err)
	}

	// Configure authentication - check multiple sources
	var cred auth.CredentialFunc

	// 1. Check app config first
	for _, r := range c.config.Registries {
		if r.URL == registryURL && r.Username != "" {
			cred = auth.StaticCredential(actualURL, auth.Credential{
				Username: r.Username,
				Password: r.Password,
			})
			break
		}
	}

	// 2. Check Docker config.json
	if cred == nil && c.dockerConfig != nil {
		if username, password, found := c.dockerConfig.GetCredentials(registryURL); found {
			cred = auth.StaticCredential(actualURL, auth.Credential{
				Username: username,
				Password: password,
			})
		}
	}

	// 3. Use anonymous auth if no credentials configured
	if cred == nil {
		cred = auth.StaticCredential(actualURL, auth.Credential{})
	}

	reg.Client = &auth.Client{
		Client:     retry.DefaultClient,
		Credential: cred,
	}

	c.registries[actualURL] = reg
	return reg, nil
}

// ListNamespaces lists namespaces (organizations/users) in a registry
// Note: Not all registries support listing namespaces
func (c *Client) ListNamespaces(registryURL string) ([]string, error) {
	reg, err := c.getRegistry(registryURL)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	var namespaces []string
	seen := make(map[string]bool)

	err = reg.Repositories(ctx, "", func(repos []string) error {
		for _, repo := range repos {
			parts := strings.SplitN(repo, "/", 2)
			if len(parts) > 1 {
				ns := parts[0]
				if !seen[ns] {
					seen[ns] = true
					namespaces = append(namespaces, ns)
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	return namespaces, nil
}

// ListRepositories lists repositories in a namespace
func (c *Client) ListRepositories(registryURL, namespace string) ([]string, error) {
	reg, err := c.getRegistry(registryURL)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	var repos []string
	prefix := namespace + "/"

	err = reg.Repositories(ctx, "", func(allRepos []string) error {
		for _, repo := range allRepos {
			if strings.HasPrefix(repo, prefix) {
				// Remove the namespace prefix
				name := strings.TrimPrefix(repo, prefix)
				// Handle nested repos
				if !strings.Contains(name, "/") {
					repos = append(repos, name)
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	return repos, nil
}

// ListArtifactsOptions configures artifact listing
type ListArtifactsOptions struct {
	Limit  int    // Max artifacts to return (0 = all)
	Offset int    // Skip first N artifacts
	Filter string // Filter tags containing this string
}

// ListArtifacts lists artifacts (tags) in a repository with options
func (c *Client) ListArtifacts(repoPath string) ([]*Artifact, error) {
	return c.ListArtifactsWithOptions(repoPath, ListArtifactsOptions{Limit: 20})
}

// ListArtifactsWithOptions lists artifacts with pagination and filtering
func (c *Client) ListArtifactsWithOptions(repoPath string, opts ListArtifactsOptions) ([]*Artifact, error) {
	parts := strings.SplitN(repoPath, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository path: %s", repoPath)
	}

	registryURL := parts[0]
	repoName := parts[1]

	// Handle Docker Hub special cases
	if registryURL == "docker.io" {
		registryURL = "registry-1.docker.io"
		// Official images need "library/" prefix
		if !strings.Contains(repoName, "/") {
			repoName = "library/" + repoName
		}
	}

	reg, err := c.getRegistry(registryURL)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repo, err := reg.Repository(ctx, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	var artifacts []*Artifact
	skipped := 0
	collected := 0
	filter := strings.ToLower(opts.Filter)

	err = repo.Tags(ctx, "", func(tags []string) error {
		for _, tag := range tags {
			// Apply filter if specified
			if filter != "" && !strings.Contains(strings.ToLower(tag), filter) {
				continue
			}

			// Handle offset
			if skipped < opts.Offset {
				skipped++
				continue
			}

			// Check limit
			if opts.Limit > 0 && collected >= opts.Limit {
				return nil // Stop iteration
			}

			artifact := &Artifact{
				Repository: repoPath,
				Tag:        tag,
				Type:       ArtifactTypeImage,
			}
			artifacts = append(artifacts, artifact)
			collected++
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	return artifacts, nil
}

// CountArtifacts returns the total number of tags in a repository
func (c *Client) CountArtifacts(repoPath string, filter string) (int, error) {
	parts := strings.SplitN(repoPath, "/", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid repository path: %s", repoPath)
	}

	registryURL := parts[0]
	repoName := parts[1]

	if registryURL == "docker.io" {
		registryURL = "registry-1.docker.io"
		if !strings.Contains(repoName, "/") {
			repoName = "library/" + repoName
		}
	}

	reg, err := c.getRegistry(registryURL)
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repo, err := reg.Repository(ctx, repoName)
	if err != nil {
		return 0, err
	}

	count := 0
	filterLower := strings.ToLower(filter)

	err = repo.Tags(ctx, "", func(tags []string) error {
		for _, tag := range tags {
			if filter == "" || strings.Contains(strings.ToLower(tag), filterLower) {
				count++
			}
		}
		return nil
	})

	return count, err
}

func (c *Client) getArtifactDetails(ctx context.Context, repo registry.Repository, repoPath, tag string) (*Artifact, error) {
	desc, err := repo.Resolve(ctx, tag)
	if err != nil {
		return nil, err
	}

	artifact := &Artifact{
		Repository: repoPath,
		Tag:        tag,
		Digest:     desc.Digest.String(),
		Size:       desc.Size,
		Type:       getArtifactType(desc.MediaType),
	}

	return artifact, nil
}

func getArtifactType(mediaType string) ArtifactType {
	switch {
	case strings.Contains(mediaType, "helm"):
		return ArtifactTypeHelmChart
	case strings.Contains(mediaType, "sbom"):
		return ArtifactTypeSBOM
	case strings.Contains(mediaType, "signature"):
		return ArtifactTypeSignature
	case strings.Contains(mediaType, "attestation"):
		return ArtifactTypeAttestation
	default:
		return ArtifactTypeImage
	}
}
