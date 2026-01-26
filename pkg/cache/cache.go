package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Cache provides local caching for registry metadata
type Cache struct {
	dir string
	mu  sync.RWMutex
}

// CacheEntry represents a cached item
type CacheEntry struct {
	Data      []byte    `json:"data"`
	ExpiresAt time.Time `json:"expires_at"`
}

// DefaultTTL is the default cache TTL
const DefaultTTL = 5 * time.Minute

// New creates a new cache instance
func New(dir string) *Cache {
	if err := os.MkdirAll(dir, 0755); err != nil {
		// Fall back to temp directory
		dir = os.TempDir()
	}

	return &Cache{
		dir: dir,
	}
}

// Get retrieves an item from the cache
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	path := c.keyPath(key)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		// Entry expired
		go c.Delete(key)
		return nil, false
	}

	return entry.Data, true
}

// Set stores an item in the cache
func (c *Cache) Set(key string, data []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl),
	}

	encoded, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	path := c.keyPath(key)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return os.WriteFile(path, encoded, 0644)
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return os.Remove(c.keyPath(key))
}

// Clear removes all cached items
func (c *Cache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return os.RemoveAll(c.dir)
}

func (c *Cache) keyPath(key string) string {
	// Use a hash-like structure to avoid filesystem issues with special characters
	return filepath.Join(c.dir, sanitizeKey(key)+".json")
}

func sanitizeKey(key string) string {
	// Use hash for long keys to avoid filesystem issues
	if len(key) > 100 {
		hash := sha256.Sum256([]byte(key))
		return hex.EncodeToString(hash[:16])
	}

	// Replace problematic characters
	result := make([]byte, 0, len(key))
	for i := 0; i < len(key); i++ {
		c := key[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			result = append(result, c)
		} else {
			result = append(result, '_')
		}
	}
	return string(result)
}

// SearchCacheKey generates a cache key for search queries
func SearchCacheKey(registry, query string) string {
	return "search:" + registry + ":" + query
}

// ArtifactsCacheKey generates a cache key for repository artifacts
func ArtifactsCacheKey(repoPath string) string {
	return "artifacts:" + repoPath
}
