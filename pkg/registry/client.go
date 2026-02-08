package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
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
	credStore    CredentialStore
	registries   map[string]*remote.Registry
	cache        Cache
}

// NewClient creates a new registry client with the default credential chain:
// Docker credential helpers → Docker config.json → app config (plaintext) → anonymous.
func NewClient(cfg *config.Config) *Client {
	dockerCfg, _ := LoadDockerConfig()

	// Build the credential store chain (highest priority first).
	var stores []CredentialStore

	// 1. Docker credential helpers (if configured in docker config.json)
	if dockerCfg != nil {
		// Per-registry credential helpers take highest priority.
		// credHelpers is a map[registryURL]helperName that routes each
		// registry to a specific helper binary.
		if len(dockerCfg.CredHelpers) > 0 {
			stores = append(stores, NewDockerCredHelperRoutingStore(dockerCfg.CredHelpers))
		}
		// Default credential helper (credsStore) is tried next for any
		// registry not covered by credHelpers.
		if dockerCfg.CredsStore != "" {
			stores = append(stores, NewDockerCredentialHelperStore(dockerCfg.CredsStore))
		}
	}

	// 2. Docker config.json static auths
	stores = append(stores, NewDockerConfigStore(dockerCfg))

	// 3. App config (plaintext YAML — legacy fallback)
	stores = append(stores, NewPlaintextFileStore(cfg))

	return &Client{
		config:       cfg,
		dockerConfig: dockerCfg,
		credStore:    NewChainedStore(stores...),
		registries:   make(map[string]*remote.Registry),
	}
}

// NewClientWithCredentialStore creates a registry client with a custom credential store.
// Useful for testing or when callers want full control over credential resolution.
func NewClientWithCredentialStore(cfg *config.Config, credStore CredentialStore) *Client {
	dockerCfg, _ := LoadDockerConfig()
	return &Client{
		config:       cfg,
		dockerConfig: dockerCfg,
		credStore:    credStore,
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

// AddRegistry adds a new registry. If name is empty, it defaults to the URL.
func (c *Client) AddRegistry(name, url string) error {
	return c.config.AddRegistry(name, url)
}

// AddRegistryWithAuth adds a registry with authentication.
// If name is empty, it defaults to the URL.
func (c *Client) AddRegistryWithAuth(name, url, username, password string) error {
	return c.config.AddRegistryWithAuth(name, url, username, password)
}

// RemoveRegistry removes a registry
func (c *Client) RemoveRegistry(url string) error {
	// Clear cached client
	delete(c.registries, url)
	return c.config.RemoveRegistry(url)
}

// TestRegistry tests connectivity to a registry
func (c *Client) TestRegistry(url string) error {
	reg, err := c.getRegistry(url)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try to ping the registry
	return reg.Ping(ctx)
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

	// Enable plain HTTP for registries marked as insecure (e.g. local dev registries).
	for _, r := range c.config.Registries {
		if r.URL == registryURL && r.Insecure {
			reg.PlainHTTP = true
			break
		}
	}

	// Resolve credentials through the credential store chain.
	var cred auth.CredentialFunc

	if creds, err := c.credStore.Get(registryURL); err == nil {
		cred = auth.StaticCredential(actualURL, auth.Credential{
			Username:     creds.Username,
			Password:     creds.Password,
			RefreshToken: creds.RefreshToken,
			AccessToken:  creds.AccessToken,
		})
	} else {
		// No credentials found — fall back to anonymous auth.
		cred = auth.StaticCredential(actualURL, auth.Credential{})
	}

	reg.Client = &auth.Client{
		Client:     retry.DefaultClient,
		Credential: cred,
	}

	c.registries[actualURL] = reg
	return reg, nil
}

// CredentialFunc returns an oras auth.CredentialFunc that resolves credentials
// for the given registry URL through the credential store chain.
// This is useful for callers (e.g. the pull package) that need to authenticate
// with a registry but build their own oras client.
func (c *Client) CredentialFunc(registryURL string) auth.CredentialFunc {
	// Normalize for docker.io
	actualURL := registryURL
	if registryURL == "docker.io" {
		actualURL = "registry-1.docker.io"
	}

	if creds, err := c.credStore.Get(registryURL); err == nil {
		return auth.StaticCredential(actualURL, auth.Credential{
			Username:     creds.Username,
			Password:     creds.Password,
			RefreshToken: creds.RefreshToken,
			AccessToken:  creds.AccessToken,
		})
	}
	// No credentials — anonymous auth.
	return auth.StaticCredential(actualURL, auth.Credential{})
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

	// Collect all tags first (for sorting before pagination)
	var allTags []string
	filter := strings.ToLower(opts.Filter)

	err = repo.Tags(ctx, "", func(tags []string) error {
		for _, tag := range tags {
			// Apply filter if specified
			if filter != "" && !strings.Contains(strings.ToLower(tag), filter) {
				continue
			}
			allTags = append(allTags, tag)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	// Sort tags: "latest" first, then semver descending, then alphabetical
	sortTags(allTags)

	// Apply pagination
	start := opts.Offset
	if start > len(allTags) {
		start = len(allTags)
	}
	end := len(allTags)
	if opts.Limit > 0 && start+opts.Limit < end {
		end = start + opts.Limit
	}

	var artifacts []*Artifact
	for _, tag := range allTags[start:end] {
		artifacts = append(artifacts, &Artifact{
			Repository: repoPath,
			Tag:        tag,
			Type:       ArtifactTypeImage,
		})
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

// GetArtifactDetails resolves a tag to its full artifact details (digest, size, type).
// repoPath should be in the form "registry/namespace/repo" (e.g. "localhost:5050/test/hello").
func (c *Client) GetArtifactDetails(repoPath, tag string) (*Artifact, error) {
	parts := strings.SplitN(repoPath, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository path: %s", repoPath)
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
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repo, err := reg.Repository(ctx, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	return c.getArtifactDetails(ctx, repo, repoPath, tag)
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

// GetArtifactInfo resolves detailed artifact type information by fetching and analyzing the manifest.
// This performs a deeper inspection than GetArtifactDetails, looking at config media type
// and layer media types to accurately determine the artifact type.
func (c *Client) GetArtifactInfo(repoPath, tag string) (*ArtifactInfo, error) {
	parts := strings.SplitN(repoPath, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository path: %s", repoPath)
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
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repo, err := reg.Repository(ctx, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	// Resolve the tag to get the manifest descriptor
	desc, err := repo.Resolve(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve tag: %w", err)
	}

	info := &ArtifactInfo{
		MediaType: desc.MediaType,
		Digest:    desc.Digest.String(),
		Size:      desc.Size,
	}

	// Fetch the manifest to get more details
	manifestReader, err := repo.Fetch(ctx, desc)
	if err != nil {
		// If we can't fetch, return basic info based on manifest media type
		info.Type, info.TypeDetail = detectArtifactType(desc.MediaType, "", nil)
		return info, nil
	}
	defer manifestReader.Close()

	var manifestData map[string]interface{}
	if err := json.NewDecoder(manifestReader).Decode(&manifestData); err != nil {
		info.Type, info.TypeDetail = detectArtifactType(desc.MediaType, "", nil)
		return info, nil
	}

	// Extract config media type if present
	if config, ok := manifestData["config"].(map[string]interface{}); ok {
		if configMediaType, ok := config["mediaType"].(string); ok {
			info.ConfigMediaType = configMediaType
		}
	}

	// Count layers
	if layers, ok := manifestData["layers"].([]interface{}); ok {
		info.Layers = len(layers)

		// Get first layer media type for additional type hints
		if len(layers) > 0 {
			if firstLayer, ok := layers[0].(map[string]interface{}); ok {
				if layerMediaType, ok := firstLayer["mediaType"].(string); ok {
					// Pass layer media types for type detection
					info.Type, info.TypeDetail = detectArtifactType(desc.MediaType, info.ConfigMediaType, []string{layerMediaType})
				}
			}
		}
	}

	// If type not yet determined, detect from manifest and config media types
	if info.Type == "" {
		info.Type, info.TypeDetail = detectArtifactType(desc.MediaType, info.ConfigMediaType, nil)
	}

	// Extract annotations
	if annotations, ok := manifestData["annotations"].(map[string]interface{}); ok {
		info.Annotations = make(map[string]string)
		for k, v := range annotations {
			if s, ok := v.(string); ok {
				info.Annotations[k] = s
			}
		}
	}

	return info, nil
}

// detectArtifactType determines the artifact type from media types.
// It checks manifest media type, config media type, and layer media types.
// Returns the artifact type and an optional detail string (e.g., "spdx" for SBOM).
func detectArtifactType(manifestMediaType, configMediaType string, layerMediaTypes []string) (ArtifactType, string) {
	// Normalize to lowercase for matching
	manifest := strings.ToLower(manifestMediaType)
	config := strings.ToLower(configMediaType)

	// Check all media types for patterns
	allTypes := append([]string{manifest, config}, layerMediaTypes...)

	// Helm Chart detection
	// Config: application/vnd.cncf.helm.config.v1+json
	// Layer: application/vnd.cncf.helm.chart.content.v1.tar+gzip
	for _, mt := range allTypes {
		if strings.Contains(mt, "helm") {
			return ArtifactTypeHelmChart, ""
		}
	}

	// SBOM detection
	// SPDX: application/spdx+json, application/vnd.spdx+json
	// CycloneDX: application/vnd.cyclonedx+json, application/vnd.cyclonedx+xml
	for _, mt := range allTypes {
		if strings.Contains(mt, "spdx") {
			return ArtifactTypeSBOM, "spdx"
		}
		if strings.Contains(mt, "cyclonedx") {
			return ArtifactTypeSBOM, "cyclonedx"
		}
		if strings.Contains(mt, "sbom") {
			return ArtifactTypeSBOM, ""
		}
	}

	// Signature detection
	// Cosign: application/vnd.dev.cosign.simplesigning.v1+json
	// Notary: application/vnd.cncf.notary.signature
	for _, mt := range allTypes {
		if strings.Contains(mt, "cosign") {
			return ArtifactTypeSignature, "cosign"
		}
		if strings.Contains(mt, "notary") && strings.Contains(mt, "signature") {
			return ArtifactTypeSignature, "notary"
		}
		if strings.Contains(mt, "signature") {
			return ArtifactTypeSignature, ""
		}
	}

	// Attestation detection
	// In-toto: application/vnd.in-toto+json
	// DSSE: application/vnd.dsse.envelope.v1+json
	for _, mt := range allTypes {
		if strings.Contains(mt, "in-toto") {
			return ArtifactTypeAttestation, "in-toto"
		}
		if strings.Contains(mt, "dsse") {
			return ArtifactTypeAttestation, "dsse"
		}
		if strings.Contains(mt, "attestation") {
			return ArtifactTypeAttestation, ""
		}
	}

	// WebAssembly detection
	// application/vnd.wasm.content.layer.v1+wasm
	// application/wasm
	for _, mt := range allTypes {
		if strings.Contains(mt, "wasm") {
			return ArtifactTypeWasm, ""
		}
	}

	// Container Image detection
	// OCI: application/vnd.oci.image.*
	// Docker: application/vnd.docker.distribution.manifest.*
	if strings.Contains(manifest, "oci.image") ||
		strings.Contains(manifest, "docker.distribution.manifest") ||
		strings.Contains(manifest, "docker.container.image") {
		return ArtifactTypeImage, ""
	}

	// Config-based image detection
	if strings.Contains(config, "oci.image.config") ||
		strings.Contains(config, "docker.container.image") {
		return ArtifactTypeImage, ""
	}

	// If manifest looks like a standard image manifest, assume image
	if strings.Contains(manifest, "manifest") {
		return ArtifactTypeImage, ""
	}

	return ArtifactTypeUnknown, ""
}

// getArtifactType is a simple type detection based on manifest media type only.
// For more accurate detection, use GetArtifactInfo which inspects config and layers.
func getArtifactType(mediaType string) ArtifactType {
	artifactType, _ := detectArtifactType(mediaType, "", nil)
	return artifactType
}

// sortTags sorts tags with "latest" first, then by semver descending, then alphabetically.
// This approximates "most recent" ordering without requiring manifest lookups.
func sortTags(tags []string) {
	sort.Slice(tags, func(i, j int) bool {
		ti, tj := tags[i], tags[j]

		// "latest" always comes first
		if ti == "latest" {
			return true
		}
		if tj == "latest" {
			return false
		}

		// Try semver comparison
		vi, oki := parseVersion(ti)
		vj, okj := parseVersion(tj)

		if oki && okj {
			// Both are semver — sort descending (higher version first)
			if vi.major != vj.major {
				return vi.major > vj.major
			}
			if vi.minor != vj.minor {
				return vi.minor > vj.minor
			}
			if vi.patch != vj.patch {
				return vi.patch > vj.patch
			}
			// Same version numbers — compare prerelease (stable > prerelease)
			if vi.prerelease == "" && vj.prerelease != "" {
				return true
			}
			if vi.prerelease != "" && vj.prerelease == "" {
				return false
			}
			return vi.prerelease > vj.prerelease
		}

		// Semver tags come before non-semver
		if oki && !okj {
			return true
		}
		if !oki && okj {
			return false
		}

		// Neither is semver — sort alphabetically descending
		return ti > tj
	})
}

type semver struct {
	major, minor, patch int
	prerelease          string
}

// parseVersion extracts semver components from a tag like "v1.2.3", "1.2.3", "v1.2.3-alpine"
func parseVersion(tag string) (semver, bool) {
	s := strings.TrimPrefix(tag, "v")

	// Split off prerelease suffix (e.g., "1.2.3-alpine" -> "1.2.3", "alpine")
	var prerelease string
	if idx := strings.IndexAny(s, "-+"); idx != -1 {
		prerelease = s[idx+1:]
		s = s[:idx]
	}

	parts := strings.Split(s, ".")
	if len(parts) < 1 || len(parts) > 3 {
		return semver{}, false
	}

	var v semver
	var err error

	if len(parts) >= 1 {
		_, err = fmt.Sscanf(parts[0], "%d", &v.major)
		if err != nil {
			return semver{}, false
		}
	}
	if len(parts) >= 2 {
		_, err = fmt.Sscanf(parts[1], "%d", &v.minor)
		if err != nil {
			return semver{}, false
		}
	}
	if len(parts) >= 3 {
		_, err = fmt.Sscanf(parts[2], "%d", &v.patch)
		if err != nil {
			return semver{}, false
		}
	}

	v.prerelease = prerelease
	return v, true
}
