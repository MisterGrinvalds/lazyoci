package registry

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mistergrinvalds/lazyoci/pkg/config"
)

// Sentinel errors for credential operations.
var (
	// ErrNotImplemented indicates a credential store backend is stubbed and not yet functional.
	ErrNotImplemented = errors.New("not implemented")

	// ErrCredentialsNotFound indicates no credentials exist for the requested registry.
	ErrCredentialsNotFound = errors.New("credentials not found")
)

// Credentials holds authentication material for an OCI registry.
type Credentials struct {
	Username     string
	Password     string
	RefreshToken string // OAuth2 refresh token (for token-based auth flows)
	AccessToken  string // OAuth2 access token / identity token
}

// CredentialStore is the interface for credential backends.
// Implementations range from plaintext config files to OS keychains
// and Docker credential helpers.
type CredentialStore interface {
	// Get retrieves credentials for the given registry URL.
	// Returns ErrCredentialsNotFound if no credentials are stored.
	Get(registryURL string) (*Credentials, error)

	// Store persists credentials for the given registry URL.
	Store(registryURL string, creds *Credentials) error

	// Delete removes stored credentials for the given registry URL.
	Delete(registryURL string) error

	// List returns all registry URLs that have stored credentials.
	List() ([]string, error)
}

// ---------------------------------------------------------------------------
// ChainedStore iterates through multiple stores, returning the first match.
// ---------------------------------------------------------------------------

// ChainedStore tries each store in order and returns the first successful result.
type ChainedStore struct {
	stores []CredentialStore
}

// NewChainedStore creates a credential store that queries each provided store
// in order, returning the first successful credential lookup.
func NewChainedStore(stores ...CredentialStore) *ChainedStore {
	return &ChainedStore{stores: stores}
}

// Get returns credentials from the first store that has them.
func (cs *ChainedStore) Get(registryURL string) (*Credentials, error) {
	for _, s := range cs.stores {
		creds, err := s.Get(registryURL)
		if err == nil {
			return creds, nil
		}
		// Skip stores that don't have credentials or aren't implemented yet
		if errors.Is(err, ErrCredentialsNotFound) || errors.Is(err, ErrNotImplemented) {
			continue
		}
		// Propagate unexpected errors
		return nil, fmt.Errorf("credential store error: %w", err)
	}
	return nil, ErrCredentialsNotFound
}

// Store persists credentials in the first store in the chain.
func (cs *ChainedStore) Store(registryURL string, creds *Credentials) error {
	if len(cs.stores) == 0 {
		return ErrNotImplemented
	}
	return cs.stores[0].Store(registryURL, creds)
}

// Delete removes credentials from all stores in the chain.
func (cs *ChainedStore) Delete(registryURL string) error {
	var lastErr error
	for _, s := range cs.stores {
		if err := s.Delete(registryURL); err != nil && !errors.Is(err, ErrNotImplemented) {
			lastErr = err
		}
	}
	return lastErr
}

// List returns the union of registry URLs across all stores.
func (cs *ChainedStore) List() ([]string, error) {
	seen := make(map[string]bool)
	var urls []string
	for _, s := range cs.stores {
		storeURLs, err := s.List()
		if err != nil {
			if errors.Is(err, ErrNotImplemented) {
				continue
			}
			return nil, err
		}
		for _, u := range storeURLs {
			if !seen[u] {
				seen[u] = true
				urls = append(urls, u)
			}
		}
	}
	return urls, nil
}

// ---------------------------------------------------------------------------
// PlaintextFileStore — wraps the existing app config (working implementation).
// ---------------------------------------------------------------------------

// PlaintextFileStore reads and writes credentials from the lazyoci YAML config.
// WARNING: credentials are stored in plaintext. This is the legacy behaviour
// and should be replaced with a more secure backend.
type PlaintextFileStore struct {
	config *config.Config
}

// NewPlaintextFileStore creates a credential store backed by the app config.
func NewPlaintextFileStore(cfg *config.Config) *PlaintextFileStore {
	return &PlaintextFileStore{config: cfg}
}

func (s *PlaintextFileStore) Get(registryURL string) (*Credentials, error) {
	for _, r := range s.config.Registries {
		if r.URL == registryURL && r.Username != "" {
			return &Credentials{
				Username: r.Username,
				Password: r.Password,
			}, nil
		}
	}
	return nil, ErrCredentialsNotFound
}

func (s *PlaintextFileStore) Store(registryURL string, creds *Credentials) error {
	return s.config.AddRegistryWithAuth("", registryURL, creds.Username, creds.Password)
}

func (s *PlaintextFileStore) Delete(registryURL string) error {
	return s.config.RemoveRegistry(registryURL)
}

func (s *PlaintextFileStore) List() ([]string, error) {
	var urls []string
	for _, r := range s.config.Registries {
		if r.Username != "" {
			urls = append(urls, r.URL)
		}
	}
	return urls, nil
}

// ---------------------------------------------------------------------------
// DockerConfigStore — reads from ~/.docker/config.json auths (working).
// ---------------------------------------------------------------------------

// DockerConfigStore reads credentials from the Docker config.json auths map.
// It does not invoke Docker credential helpers (see DockerCredentialHelperStore).
type DockerConfigStore struct {
	dockerConfig *DockerConfig
}

// NewDockerConfigStore creates a read-only credential store backed by Docker config.json.
func NewDockerConfigStore(dc *DockerConfig) *DockerConfigStore {
	return &DockerConfigStore{dockerConfig: dc}
}

func (s *DockerConfigStore) Get(registryURL string) (*Credentials, error) {
	if s.dockerConfig == nil {
		return nil, ErrCredentialsNotFound
	}
	username, password, found := s.dockerConfig.GetCredentials(registryURL)
	if !found {
		return nil, ErrCredentialsNotFound
	}
	return &Credentials{Username: username, Password: password}, nil
}

func (s *DockerConfigStore) Store(_ string, _ *Credentials) error {
	// Docker config.json is managed by `docker login`; we don't write to it.
	return ErrNotImplemented
}

func (s *DockerConfigStore) Delete(_ string) error {
	return ErrNotImplemented
}

func (s *DockerConfigStore) List() ([]string, error) {
	if s.dockerConfig == nil {
		return nil, nil
	}
	var urls []string
	for url := range s.dockerConfig.Auths {
		urls = append(urls, url)
	}
	return urls, nil
}

// ---------------------------------------------------------------------------
// DockerCredHelperRoutingStore — routes per-registry credential helpers.
// ---------------------------------------------------------------------------

// DockerCredHelperRoutingStore routes credential lookups to per-registry
// Docker credential helpers as specified by the "credHelpers" field in
// Docker config.json. Each entry maps a registry URL to a helper name.
type DockerCredHelperRoutingStore struct {
	// helpers maps normalized registry URLs to their credential helper names.
	helpers map[string]string
}

// NewDockerCredHelperRoutingStore creates a routing store from the credHelpers
// map (key = registry URL, value = helper name).
func NewDockerCredHelperRoutingStore(credHelpers map[string]string) *DockerCredHelperRoutingStore {
	return &DockerCredHelperRoutingStore{helpers: credHelpers}
}

func (s *DockerCredHelperRoutingStore) Get(registryURL string) (*Credentials, error) {
	normalized := normalizeRegistryForLookup(registryURL)
	for helperRegistry, helperName := range s.helpers {
		if normalizeRegistryForLookup(helperRegistry) == normalized {
			store := NewDockerCredentialHelperStore(helperName)
			return store.Get(registryURL)
		}
	}
	return nil, ErrCredentialsNotFound
}

func (s *DockerCredHelperRoutingStore) Store(registryURL string, creds *Credentials) error {
	return ErrNotImplemented
}

func (s *DockerCredHelperRoutingStore) Delete(registryURL string) error {
	return ErrNotImplemented
}

func (s *DockerCredHelperRoutingStore) List() ([]string, error) {
	var urls []string
	for url := range s.helpers {
		urls = append(urls, url)
	}
	return urls, nil
}

// normalizeRegistryForLookup strips protocol prefixes for comparison purposes.
func normalizeRegistryForLookup(registry string) string {
	registry = strings.TrimPrefix(registry, "https://")
	registry = strings.TrimPrefix(registry, "http://")
	registry = strings.TrimSuffix(registry, "/")
	return registry
}

// ---------------------------------------------------------------------------
// DockerCredentialHelperStore — invokes docker-credential-* helpers.
// ---------------------------------------------------------------------------

// DockerCredentialHelperStore invokes external Docker credential helpers
// (e.g. docker-credential-osxkeychain, docker-credential-secretservice,
// docker-credential-wincred, docker-credential-pass).
//
// TODO: Implement by exec-ing `docker-credential-<helper> get` with the
// registry URL on stdin and parsing the JSON response {ServerURL, Username, Secret}.
// See https://docs.docker.com/engine/reference/commandline/login/#credential-helpers
type DockerCredentialHelperStore struct {
	// helperName is the credential helper suffix (e.g. "osxkeychain", "secretservice").
	helperName string
}

// NewDockerCredentialHelperStore creates a store that delegates to the named
// Docker credential helper binary.
func NewDockerCredentialHelperStore(helperName string) *DockerCredentialHelperStore {
	return &DockerCredentialHelperStore{helperName: helperName}
}

func (s *DockerCredentialHelperStore) Get(registryURL string) (*Credentials, error) {
	username, password, err := execCredentialHelper(s.helperName, registryURL)
	if err != nil {
		return nil, err
	}
	return &Credentials{Username: username, Password: password}, nil
}

func (s *DockerCredentialHelperStore) Store(_ string, _ *Credentials) error {
	// TODO: invoke `docker-credential-<helper> store` with JSON on stdin
	return ErrNotImplemented
}

func (s *DockerCredentialHelperStore) Delete(_ string) error {
	// TODO: invoke `docker-credential-<helper> erase` with registry URL on stdin
	return ErrNotImplemented
}

func (s *DockerCredentialHelperStore) List() ([]string, error) {
	// TODO: invoke `docker-credential-<helper> list` and parse JSON response
	return nil, ErrNotImplemented
}

// ---------------------------------------------------------------------------
// KeychainStore — stub for OS keychain integration.
// ---------------------------------------------------------------------------

// KeychainStore provides credential storage via the operating system's native
// keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service).
//
// TODO: Implement using github.com/zalando/go-keyring or similar.
type KeychainStore struct{}

// NewKeychainStore creates a credential store backed by the OS keychain.
func NewKeychainStore() *KeychainStore { return &KeychainStore{} }

func (s *KeychainStore) Get(_ string) (*Credentials, error)   { return nil, ErrNotImplemented }
func (s *KeychainStore) Store(_ string, _ *Credentials) error { return ErrNotImplemented }
func (s *KeychainStore) Delete(_ string) error                { return ErrNotImplemented }
func (s *KeychainStore) List() ([]string, error)              { return nil, ErrNotImplemented }

// ---------------------------------------------------------------------------
// EncryptedFileStore — stub for encrypted credential file.
// ---------------------------------------------------------------------------

// EncryptedFileStore stores credentials in an encrypted file on disk,
// using a passphrase or key derived from the user's environment.
//
// TODO: Implement using age, nacl/secretbox, or similar symmetric encryption.
type EncryptedFileStore struct{}

// NewEncryptedFileStore creates a credential store backed by an encrypted file.
func NewEncryptedFileStore() *EncryptedFileStore { return &EncryptedFileStore{} }

func (s *EncryptedFileStore) Get(_ string) (*Credentials, error)   { return nil, ErrNotImplemented }
func (s *EncryptedFileStore) Store(_ string, _ *Credentials) error { return ErrNotImplemented }
func (s *EncryptedFileStore) Delete(_ string) error                { return ErrNotImplemented }
func (s *EncryptedFileStore) List() ([]string, error)              { return nil, ErrNotImplemented }
