package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Ensure time is used for SearchCacheTTL
var _ = time.Second

// SearchResult represents a repository from search
type SearchResult struct {
	Name            string `json:"repo_name"`
	Description     string `json:"short_description"`
	StarCount       int    `json:"star_count"`
	PullCount       int64  `json:"pull_count"`
	IsOfficial      bool   `json:"is_official"`
	IsAutomated     bool   `json:"is_automated"`
	RegistryURL     string `json:"registry_url"`
	LastUpdated     string `json:"last_updated,omitempty"`
}

// SearchResponse represents the Docker Hub search API response
type SearchResponse struct {
	Count    int             `json:"count"`
	Next     string          `json:"next"`
	Previous string          `json:"previous"`
	Results  []*SearchResult `json:"results"`
}

// SearchOptions configures a search request
type SearchOptions struct {
	Query    string
	PageSize int
	Page     int
}

// SearchCacheTTL is how long search results are cached
const SearchCacheTTL = 5 * time.Minute

// Search searches for repositories across registries
func (c *Client) Search(registryURL, query string) ([]*SearchResult, error) {
	// Check cache first
	cacheKey := "search:" + registryURL + ":" + query
	var cached []*SearchResult
	if c.getCached(cacheKey, &cached) {
		return cached, nil
	}

	var results []*SearchResult
	var err error

	switch registryURL {
	case "docker.io":
		results, err = c.searchDockerHub(query)
	case "quay.io":
		results, err = c.searchQuay(query)
	case "ghcr.io":
		results, err = c.searchGHCR(query)
	default:
		// For other registries, try OCI catalog with filter
		results, err = c.searchOCI(registryURL, query)
	}

	if err != nil {
		return nil, err
	}

	// Cache the results
	c.setCache(cacheKey, results, SearchCacheTTL)

	return results, nil
}

func (c *Client) searchDockerHub(query string) ([]*SearchResult, error) {
	baseURL := "https://hub.docker.com/v2/search/repositories/"
	params := url.Values{}
	params.Set("query", query)
	params.Set("page_size", "25")

	req, err := http.NewRequest("GET", baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed with status: %d", resp.StatusCode)
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Set registry URL for all results
	for _, r := range searchResp.Results {
		r.RegistryURL = "docker.io"
	}

	return searchResp.Results, nil
}

// QuaySearchResponse represents Quay.io search response
type QuaySearchResponse struct {
	Results []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Popularity  float64 `json:"popularity"`
		IsPublic    bool   `json:"is_public"`
		Namespace   struct {
			Name string `json:"name"`
		} `json:"namespace"`
	} `json:"results"`
}

func (c *Client) searchQuay(query string) ([]*SearchResult, error) {
	baseURL := "https://quay.io/api/v1/find/repositories"
	params := url.Values{}
	params.Set("query", query)

	req, err := http.NewRequest("GET", baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed with status: %d", resp.StatusCode)
	}

	var quayResp QuaySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&quayResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var results []*SearchResult
	for _, r := range quayResp.Results {
		results = append(results, &SearchResult{
			Name:        r.Namespace.Name + "/" + r.Name,
			Description: r.Description,
			StarCount:   int(r.Popularity),
			RegistryURL: "quay.io",
		})
	}

	return results, nil
}

func (c *Client) searchGHCR(query string) ([]*SearchResult, error) {
	// GHCR doesn't have a public search API
	// Return empty with a note that user should enter full path
	return nil, fmt.Errorf("GHCR requires full repository path (e.g., owner/repo)")
}

func (c *Client) searchOCI(registryURL, query string) ([]*SearchResult, error) {
	// For generic OCI registries, try to list and filter
	namespaces, err := c.ListNamespaces(registryURL)
	if err != nil {
		return nil, err
	}

	var results []*SearchResult
	for _, ns := range namespaces {
		if containsIgnoreCase(ns, query) {
			results = append(results, &SearchResult{
				Name:        ns,
				RegistryURL: registryURL,
			})
		}
	}

	return results, nil
}

func containsIgnoreCase(s, substr string) bool {
	// Simple case-insensitive contains
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalFoldSlice(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func equalFoldSlice(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// FormatPullCount formats large numbers for display
func FormatPullCount(count int64) string {
	switch {
	case count >= 1_000_000_000:
		return fmt.Sprintf("%.1fB", float64(count)/1_000_000_000)
	case count >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(count)/1_000_000)
	case count >= 1_000:
		return fmt.Sprintf("%.1fK", float64(count)/1_000)
	default:
		return fmt.Sprintf("%d", count)
	}
}
