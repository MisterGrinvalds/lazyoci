package cache

import (
	"strings"
	"testing"
	"time"
)

func TestSanitizeKey(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(string) bool // validator instead of exact match for hash cases
	}{
		{
			name:  "simple key unchanged",
			input: "my-cache-key_123",
			check: func(s string) bool { return s == "my-cache-key_123" },
		},
		{
			name:  "slashes replaced",
			input: "docker.io/library/nginx",
			check: func(s string) bool { return s == "docker_io_library_nginx" },
		},
		{
			name:  "colons replaced",
			input: "search:docker.io:nginx",
			check: func(s string) bool { return s == "search_docker_io_nginx" },
		},
		{
			name:  "only alphanumeric dash underscore kept",
			input: "a!b@c#d$e%f",
			check: func(s string) bool { return s == "a_b_c_d_e_f" },
		},
		{
			name:  "long key hashed",
			input: strings.Repeat("a", 101),
			check: func(s string) bool {
				// Should be a hex hash, 32 chars (16 bytes hex)
				return len(s) == 32
			},
		},
		{
			name:  "exactly 100 chars not hashed",
			input: strings.Repeat("a", 100),
			check: func(s string) bool { return len(s) == 100 },
		},
		{
			name:  "empty key",
			input: "",
			check: func(s string) bool { return s == "" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeKey(tt.input)
			if !tt.check(got) {
				t.Errorf("sanitizeKey(%q) = %q, failed check", tt.input, got)
			}
		})
	}
}

func TestCacheSetGet(t *testing.T) {
	c := New(t.TempDir())

	key := "test-key"
	data := []byte("test-data")

	// Set
	if err := c.Set(key, data, 5*time.Minute); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Get
	got, ok := c.Get(key)
	if !ok {
		t.Fatal("Get() returned not found")
	}
	if string(got) != string(data) {
		t.Errorf("Get() = %q, want %q", got, data)
	}
}

func TestCacheExpiry(t *testing.T) {
	c := New(t.TempDir())

	key := "expiring-key"
	data := []byte("will-expire")

	// Set with very short TTL
	if err := c.Set(key, data, 1*time.Millisecond); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Wait for expiry
	time.Sleep(10 * time.Millisecond)

	// Should not be found
	_, ok := c.Get(key)
	if ok {
		t.Error("Get() found expired entry")
	}
}

func TestCacheMiss(t *testing.T) {
	c := New(t.TempDir())

	_, ok := c.Get("nonexistent")
	if ok {
		t.Error("Get() found nonexistent key")
	}
}

func TestCacheDelete(t *testing.T) {
	c := New(t.TempDir())

	key := "delete-me"
	if err := c.Set(key, []byte("data"), 5*time.Minute); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Verify exists
	if _, ok := c.Get(key); !ok {
		t.Fatal("key not found after Set")
	}

	// Delete
	if err := c.Delete(key); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Should be gone
	if _, ok := c.Get(key); ok {
		t.Error("key still found after Delete")
	}
}

func TestCacheClear(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)

	// Set multiple keys
	for _, key := range []string{"key1", "key2", "key3"} {
		if err := c.Set(key, []byte("data"), 5*time.Minute); err != nil {
			t.Fatalf("Set(%q) error = %v", key, err)
		}
	}

	// Clear
	if err := c.Clear(); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	// All should be gone
	for _, key := range []string{"key1", "key2", "key3"} {
		if _, ok := c.Get(key); ok {
			t.Errorf("key %q still found after Clear", key)
		}
	}
}

func TestCacheOverwrite(t *testing.T) {
	c := New(t.TempDir())

	key := "overwrite-me"
	if err := c.Set(key, []byte("first"), 5*time.Minute); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if err := c.Set(key, []byte("second"), 5*time.Minute); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	got, ok := c.Get(key)
	if !ok {
		t.Fatal("Get() returned not found")
	}
	if string(got) != "second" {
		t.Errorf("Get() = %q, want %q", got, "second")
	}
}

func TestSearchCacheKey(t *testing.T) {
	got := SearchCacheKey("docker.io", "nginx")
	want := "search:docker.io:nginx"
	if got != want {
		t.Errorf("SearchCacheKey() = %q, want %q", got, want)
	}
}

func TestArtifactsCacheKey(t *testing.T) {
	got := ArtifactsCacheKey("docker.io/library/nginx")
	want := "artifacts:docker.io/library/nginx"
	if got != want {
		t.Errorf("ArtifactsCacheKey() = %q, want %q", got, want)
	}
}
