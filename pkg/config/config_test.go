package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestExpandPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "tilde slash expands",
			input: "~/Documents/artifacts",
			want:  filepath.Join(homeDir, "Documents/artifacts"),
		},
		{
			name:  "tilde alone expands",
			input: "~",
			want:  homeDir,
		},
		{
			name:  "absolute path unchanged",
			input: "/tmp/artifacts",
			want:  "/tmp/artifacts",
		},
		{
			name:  "relative path unchanged",
			input: "relative/path",
			want:  "relative/path",
		},
		{
			name:  "empty string unchanged",
			input: "",
			want:  "",
		},
		{
			name:  "tilde in middle unchanged",
			input: "/some/~/path",
			want:  "/some/~/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandPath(tt.input)
			if got != tt.want {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Should have 3 default registries
	if len(cfg.Registries) != 3 {
		t.Errorf("len(Registries) = %d, want 3", len(cfg.Registries))
	}

	// Check default registries
	names := make(map[string]bool)
	for _, r := range cfg.Registries {
		names[r.URL] = true
	}
	for _, url := range []string{"docker.io", "quay.io", "ghcr.io"} {
		if !names[url] {
			t.Errorf("missing default registry: %s", url)
		}
	}

	// Default registry should be docker.io
	if cfg.DefaultRegistry != "docker.io" {
		t.Errorf("DefaultRegistry = %q, want %q", cfg.DefaultRegistry, "docker.io")
	}

	// CacheDir should be set
	if cfg.CacheDir == "" {
		t.Error("CacheDir is empty")
	}

	// ArtifactDir should be empty (use default)
	if cfg.ArtifactDir != "" {
		t.Errorf("ArtifactDir = %q, want empty", cfg.ArtifactDir)
	}
}

func TestConfigSaveLoad(t *testing.T) {
	// Create a temp directory for config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create a config
	cfg := &Config{
		Registries: []Registry{
			{Name: "Test", URL: "test.io"},
			{Name: "Local", URL: "localhost:5050", Insecure: true},
		},
		CacheDir:        filepath.Join(tmpDir, "cache"),
		ArtifactDir:     "~/artifacts",
		DefaultRegistry: "test.io",
	}

	// Save manually (since Save() uses getConfigPath)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("write error: %v", err)
	}

	// Read back and verify
	readData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	var loaded Config
	if err := yaml.Unmarshal(readData, &loaded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if len(loaded.Registries) != 2 {
		t.Errorf("len(Registries) = %d, want 2", len(loaded.Registries))
	}
	if loaded.DefaultRegistry != "test.io" {
		t.Errorf("DefaultRegistry = %q, want %q", loaded.DefaultRegistry, "test.io")
	}
	if loaded.ArtifactDir != "~/artifacts" {
		t.Errorf("ArtifactDir = %q, want %q", loaded.ArtifactDir, "~/artifacts")
	}

	// Verify insecure flag round-trips
	var foundLocal bool
	for _, r := range loaded.Registries {
		if r.URL == "localhost:5050" {
			foundLocal = true
			if !r.Insecure {
				t.Error("expected Insecure=true for localhost:5050")
			}
		}
	}
	if !foundLocal {
		t.Error("localhost:5050 registry not found after round-trip")
	}
}

func TestGetArtifactDirPriority(t *testing.T) {
	// Save and restore global state
	origOverride := GetArtifactDirOverride()
	origEnv := os.Getenv("LAZYOCI_ARTIFACT_DIR")
	defer func() {
		SetArtifactDirOverride(origOverride)
		os.Setenv("LAZYOCI_ARTIFACT_DIR", origEnv)
	}()

	t.Run("default when nothing set", func(t *testing.T) {
		SetArtifactDirOverride("")
		os.Unsetenv("LAZYOCI_ARTIFACT_DIR")
		cfg := &Config{}

		got := cfg.GetArtifactDir()
		want := DefaultArtifactDir()
		if got != want {
			t.Errorf("GetArtifactDir() = %q, want default %q", got, want)
		}
	})

	t.Run("config file value used", func(t *testing.T) {
		SetArtifactDirOverride("")
		os.Unsetenv("LAZYOCI_ARTIFACT_DIR")
		cfg := &Config{ArtifactDir: "/opt/artifacts"}

		got := cfg.GetArtifactDir()
		if got != "/opt/artifacts" {
			t.Errorf("GetArtifactDir() = %q, want %q", got, "/opt/artifacts")
		}
	})

	t.Run("env var overrides config", func(t *testing.T) {
		SetArtifactDirOverride("")
		os.Setenv("LAZYOCI_ARTIFACT_DIR", "/env/artifacts")
		cfg := &Config{ArtifactDir: "/opt/artifacts"}

		got := cfg.GetArtifactDir()
		if got != "/env/artifacts" {
			t.Errorf("GetArtifactDir() = %q, want %q", got, "/env/artifacts")
		}
	})

	t.Run("CLI flag overrides env and config", func(t *testing.T) {
		SetArtifactDirOverride("/cli/artifacts")
		os.Setenv("LAZYOCI_ARTIFACT_DIR", "/env/artifacts")
		cfg := &Config{ArtifactDir: "/opt/artifacts"}

		got := cfg.GetArtifactDir()
		if got != "/cli/artifacts" {
			t.Errorf("GetArtifactDir() = %q, want %q", got, "/cli/artifacts")
		}
	})

	t.Run("tilde expansion in config", func(t *testing.T) {
		SetArtifactDirOverride("")
		os.Unsetenv("LAZYOCI_ARTIFACT_DIR")
		cfg := &Config{ArtifactDir: "~/my-artifacts"}

		got := cfg.GetArtifactDir()
		homeDir, _ := os.UserHomeDir()
		want := filepath.Join(homeDir, "my-artifacts")
		if got != want {
			t.Errorf("GetArtifactDir() = %q, want %q", got, want)
		}
	})
}

func TestRegistryCRUD(t *testing.T) {
	cfg := &Config{
		Registries: []Registry{
			{Name: "Docker Hub", URL: "docker.io"},
		},
	}

	// GetRegistry - found
	reg := cfg.GetRegistry("docker.io")
	if reg == nil {
		t.Fatal("GetRegistry(docker.io) returned nil")
	}
	if reg.Name != "Docker Hub" {
		t.Errorf("Name = %q, want %q", reg.Name, "Docker Hub")
	}

	// GetRegistry - not found
	reg = cfg.GetRegistry("nonexistent.io")
	if reg != nil {
		t.Errorf("GetRegistry(nonexistent.io) = %v, want nil", reg)
	}
}

func TestIsArtifactDirDefault(t *testing.T) {
	origOverride := GetArtifactDirOverride()
	origEnv := os.Getenv("LAZYOCI_ARTIFACT_DIR")
	defer func() {
		SetArtifactDirOverride(origOverride)
		os.Setenv("LAZYOCI_ARTIFACT_DIR", origEnv)
	}()

	SetArtifactDirOverride("")
	os.Unsetenv("LAZYOCI_ARTIFACT_DIR")

	t.Run("default returns true", func(t *testing.T) {
		cfg := &Config{}
		if !cfg.IsArtifactDirDefault() {
			t.Error("IsArtifactDirDefault() = false, want true")
		}
	})

	t.Run("custom returns false", func(t *testing.T) {
		cfg := &Config{ArtifactDir: "/custom/path"}
		if cfg.IsArtifactDirDefault() {
			t.Error("IsArtifactDirDefault() = true, want false")
		}
	})
}

func TestValidatePath(t *testing.T) {
	t.Run("nonexistent without create", func(t *testing.T) {
		err := ValidatePath("/nonexistent/path/abcdef12345", false)
		if err == nil {
			t.Error("expected error for nonexistent path")
		}
	})

	t.Run("create directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		newDir := filepath.Join(tmpDir, "new-dir")

		err := ValidatePath(newDir, true)
		if err != nil {
			t.Fatalf("ValidatePath() error = %v", err)
		}

		info, err := os.Stat(newDir)
		if err != nil {
			t.Fatalf("directory not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("path is not a directory")
		}
	})

	t.Run("file instead of directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "file.txt")
		os.WriteFile(filePath, []byte("test"), 0644)

		err := ValidatePath(filePath, false)
		if err == nil {
			t.Error("expected error for file path")
		}
	})

	t.Run("existing writable directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := ValidatePath(tmpDir, false)
		if err != nil {
			t.Errorf("ValidatePath() error = %v", err)
		}
	})
}

func TestPathExists(t *testing.T) {
	tmpDir := t.TempDir()

	if !PathExists(tmpDir) {
		t.Error("PathExists() = false for existing dir")
	}

	if PathExists(filepath.Join(tmpDir, "nonexistent")) {
		t.Error("PathExists() = true for nonexistent path")
	}
}
