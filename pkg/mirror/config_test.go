package mirror

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "mirror.yaml")
	content := `
target:
  url: registry.example.com/ns
  insecure: false
  charts-prefix: charts

upstreams:
  vault:
    type: repo
    repo: https://helm.releases.hashicorp.com
    chart: vault
    versions:
      - "0.28.0"
  keycloak:
    type: oci
    registry: oci://registry-1.docker.io/bitnamicharts
    chart: keycloak
    versions:
      - "24.0.1"
  myapp:
    type: local
    path: ./charts/myapp
    chart: myapp
    versions:
      - "1.0.0"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Target
	if cfg.Target.URL != "registry.example.com/ns" {
		t.Errorf("target.url = %q, want %q", cfg.Target.URL, "registry.example.com/ns")
	}
	if cfg.Target.ChartsPrefix != "charts" {
		t.Errorf("target.charts-prefix = %q, want %q", cfg.Target.ChartsPrefix, "charts")
	}
	if cfg.Target.ChartOCIBase() != "registry.example.com/ns/charts" {
		t.Errorf("ChartOCIBase() = %q, want %q", cfg.Target.ChartOCIBase(), "registry.example.com/ns/charts")
	}

	// Upstreams
	if len(cfg.Upstreams) != 3 {
		t.Fatalf("got %d upstreams, want 3", len(cfg.Upstreams))
	}

	vault := cfg.Upstreams["vault"]
	if vault.Type != SourceRepo {
		t.Errorf("vault.type = %q, want %q", vault.Type, SourceRepo)
	}
	if vault.Repo != "https://helm.releases.hashicorp.com" {
		t.Errorf("vault.repo = %q", vault.Repo)
	}
	if len(vault.Versions) != 1 || vault.Versions[0] != "0.28.0" {
		t.Errorf("vault.versions = %v", vault.Versions)
	}

	keycloak := cfg.Upstreams["keycloak"]
	if keycloak.Type != SourceOCI {
		t.Errorf("keycloak.type = %q, want %q", keycloak.Type, SourceOCI)
	}

	myapp := cfg.Upstreams["myapp"]
	if myapp.Type != SourceLocal {
		t.Errorf("myapp.type = %q, want %q", myapp.Type, SourceLocal)
	}
	// Path should be resolved relative to the config dir.
	wantPath := filepath.Join(dir, "charts/myapp")
	if myapp.Path != wantPath {
		t.Errorf("myapp.path = %q, want %q", myapp.Path, wantPath)
	}
}

func TestLoadConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "missing target url",
			content: "target:\n  insecure: false\nupstreams:\n  x:\n    type: repo\n    repo: https://x\n    chart: x\n",
			wantErr: "target.url is required",
		},
		{
			name:    "no upstreams",
			content: "target:\n  url: foo\nupstreams: {}\n",
			wantErr: "at least one upstream is required",
		},
		{
			name:    "missing chart name",
			content: "target:\n  url: foo\nupstreams:\n  x:\n    type: repo\n    repo: https://x\n",
			wantErr: "chart name is required",
		},
		{
			name:    "repo without url",
			content: "target:\n  url: foo\nupstreams:\n  x:\n    type: repo\n    chart: x\n",
			wantErr: "repo URL is required",
		},
		{
			name:    "oci without registry",
			content: "target:\n  url: foo\nupstreams:\n  x:\n    type: oci\n    chart: x\n",
			wantErr: "registry URL is required",
		},
		{
			name:    "local without path",
			content: "target:\n  url: foo\nupstreams:\n  x:\n    type: local\n    chart: x\n",
			wantErr: "path is required",
		},
		{
			name:    "unknown type",
			content: "target:\n  url: foo\nupstreams:\n  x:\n    type: git\n    chart: x\n",
			wantErr: "unknown source type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			cfgPath := filepath.Join(dir, "mirror.yaml")
			if err := os.WriteFile(cfgPath, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			_, err := LoadConfig(cfgPath)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestTargetConfig_ChartOCIBase(t *testing.T) {
	tests := []struct {
		name   string
		target TargetConfig
		want   string
	}{
		{
			name:   "with prefix",
			target: TargetConfig{URL: "registry.example.com/ns", ChartsPrefix: "charts"},
			want:   "registry.example.com/ns/charts",
		},
		{
			name:   "without prefix",
			target: TargetConfig{URL: "registry.example.com/ns"},
			want:   "registry.example.com/ns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.target.ChartOCIBase()
			if got != tt.want {
				t.Errorf("ChartOCIBase() = %q, want %q", got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
