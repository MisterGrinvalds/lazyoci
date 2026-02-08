package registry

import (
	"testing"

	"github.com/mistergrinvalds/lazyoci/pkg/config"
)

func TestChainedStore_SkipsNotImplemented(t *testing.T) {
	// A chain with a stub helper followed by a working store should skip the stub.
	cfg := &config.Config{
		Registries: []config.Registry{
			{URL: "ghcr.io", Username: "user", Password: "pass"},
		},
	}
	chain := NewChainedStore(
		NewDockerCredentialHelperStore("nonexistent-helper-that-does-not-exist"),
		NewPlaintextFileStore(cfg),
	)

	creds, err := chain.Get("ghcr.io")
	if err != nil {
		t.Fatalf("ChainedStore.Get() error = %v, want nil", err)
	}
	if creds.Username != "user" || creds.Password != "pass" {
		t.Errorf("ChainedStore.Get() = %q:%q, want user:pass", creds.Username, creds.Password)
	}
}

func TestChainedStore_NoCredentials(t *testing.T) {
	cfg := &config.Config{}
	chain := NewChainedStore(
		NewDockerConfigStore(nil),
		NewPlaintextFileStore(cfg),
	)

	_, err := chain.Get("unknown.registry.io")
	if err != ErrCredentialsNotFound {
		t.Errorf("ChainedStore.Get() error = %v, want ErrCredentialsNotFound", err)
	}
}

func TestPlaintextFileStore_Get(t *testing.T) {
	cfg := &config.Config{
		Registries: []config.Registry{
			{URL: "ghcr.io", Username: "alice", Password: "secret"},
			{URL: "quay.io"}, // no credentials
		},
	}
	store := NewPlaintextFileStore(cfg)

	creds, err := store.Get("ghcr.io")
	if err != nil {
		t.Fatalf("Get(ghcr.io) error = %v", err)
	}
	if creds.Username != "alice" || creds.Password != "secret" {
		t.Errorf("Get(ghcr.io) = %q:%q, want alice:secret", creds.Username, creds.Password)
	}

	_, err = store.Get("quay.io")
	if err != ErrCredentialsNotFound {
		t.Errorf("Get(quay.io) error = %v, want ErrCredentialsNotFound", err)
	}

	_, err = store.Get("unknown.io")
	if err != ErrCredentialsNotFound {
		t.Errorf("Get(unknown.io) error = %v, want ErrCredentialsNotFound", err)
	}
}

func TestDockerConfigStore_Get(t *testing.T) {
	dc := &DockerConfig{
		Auths: map[string]DockerAuth{
			"ghcr.io": {Username: "user", Password: "pass"},
		},
	}
	store := NewDockerConfigStore(dc)

	creds, err := store.Get("ghcr.io")
	if err != nil {
		t.Fatalf("Get(ghcr.io) error = %v", err)
	}
	if creds.Username != "user" || creds.Password != "pass" {
		t.Errorf("Get(ghcr.io) = %q:%q, want user:pass", creds.Username, creds.Password)
	}

	_, err = store.Get("quay.io")
	if err != ErrCredentialsNotFound {
		t.Errorf("Get(quay.io) error = %v, want ErrCredentialsNotFound", err)
	}
}

func TestDockerConfigStore_NilConfig(t *testing.T) {
	store := NewDockerConfigStore(nil)
	_, err := store.Get("ghcr.io")
	if err != ErrCredentialsNotFound {
		t.Errorf("Get() with nil config error = %v, want ErrCredentialsNotFound", err)
	}
}

func TestDockerConfigStore_EmptyAuths(t *testing.T) {
	// Simulates Docker Desktop scenario: auths has entries but no credentials
	// (credentials are in credsStore)
	dc := &DockerConfig{
		Auths: map[string]DockerAuth{
			"registry.digitalocean.com": {}, // empty — creds are in credsStore
		},
		CredsStore: "desktop",
	}
	store := NewDockerConfigStore(dc)

	_, err := store.Get("registry.digitalocean.com")
	if err != ErrCredentialsNotFound {
		t.Errorf("Get() with empty auth entry error = %v, want ErrCredentialsNotFound", err)
	}
}

func TestDockerCredentialHelperStore_NonexistentBinary(t *testing.T) {
	// A helper that doesn't exist on $PATH should return ErrNotImplemented
	// so the chain can skip it gracefully.
	store := NewDockerCredentialHelperStore("nonexistent-helper-xyzzy-12345")
	_, err := store.Get("ghcr.io")
	if err != ErrNotImplemented {
		t.Errorf("Get() with nonexistent binary error = %v, want ErrNotImplemented", err)
	}
}

func TestDockerCredHelperRoutingStore_Routes(t *testing.T) {
	// Without a real helper binary, this should fall through to the helper's
	// error handling. We test the routing logic by verifying that a request
	// for an unmatched registry returns ErrCredentialsNotFound.
	store := NewDockerCredHelperRoutingStore(map[string]string{
		"ghcr.io": "nonexistent-helper-xyzzy",
	})

	// Unmatched registry
	_, err := store.Get("quay.io")
	if err != ErrCredentialsNotFound {
		t.Errorf("Get(quay.io) error = %v, want ErrCredentialsNotFound", err)
	}

	// Matched registry (helper doesn't exist, so we get ErrNotImplemented from the exec)
	_, err = store.Get("ghcr.io")
	if err == ErrCredentialsNotFound {
		t.Error("Get(ghcr.io) returned ErrCredentialsNotFound, expected helper-related error")
	}
}

func TestDockerCredHelperRoutingStore_List(t *testing.T) {
	store := NewDockerCredHelperRoutingStore(map[string]string{
		"ghcr.io": "helper-a",
		"quay.io": "helper-b",
	})

	urls, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(urls) != 2 {
		t.Errorf("List() returned %d URLs, want 2", len(urls))
	}
}

func TestNormalizeRegistryForLookup(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"ghcr.io", "ghcr.io"},
		{"https://ghcr.io", "ghcr.io"},
		{"http://ghcr.io", "ghcr.io"},
		{"ghcr.io/", "ghcr.io"},
		{"https://ghcr.io/", "ghcr.io"},
		{"registry.digitalocean.com", "registry.digitalocean.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeRegistryForLookup(tt.input)
			if got != tt.want {
				t.Errorf("normalizeRegistryForLookup(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
