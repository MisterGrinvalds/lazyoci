package mirror

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindTgz(t *testing.T) {
	tests := []struct {
		name      string
		files     []string
		chartName string
		version   string
		wantBase  string
		wantErr   bool
	}{
		{
			name:      "exact match",
			files:     []string{"vault-0.28.0.tgz"},
			chartName: "vault",
			version:   "0.28.0",
			wantBase:  "vault-0.28.0.tgz",
		},
		{
			name:      "build metadata suffix",
			files:     []string{"vault-0.28.0+build.123.tgz"},
			chartName: "vault",
			version:   "0.28.0",
			wantBase:  "vault-0.28.0+build.123.tgz",
		},
		{
			name:      "broader glob fallback",
			files:     []string{"myapp-1.2.3.tgz"},
			chartName: "myapp",
			version:   "1.0.0",
			wantBase:  "myapp-1.2.3.tgz",
		},
		{
			name:      "no match",
			files:     []string{"other-chart-1.0.0.tgz"},
			chartName: "myapp",
			version:   "1.0.0",
			wantErr:   true,
		},
		{
			name:      "empty directory",
			files:     []string{},
			chartName: "vault",
			version:   "0.28.0",
			wantErr:   true,
		},
		{
			name:      "prefers exact over glob",
			files:     []string{"vault-0.28.0.tgz", "vault-0.28.0+meta.tgz"},
			chartName: "vault",
			version:   "0.28.0",
			wantBase:  "vault-0.28.0.tgz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, f := range tt.files {
				path := filepath.Join(dir, f)
				if err := os.WriteFile(path, []byte("fake-chart"), 0644); err != nil {
					t.Fatal(err)
				}
			}

			got, err := findTgz(dir, tt.chartName, tt.version)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got path %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			gotBase := filepath.Base(got)
			if gotBase != tt.wantBase {
				t.Errorf("findTgz() = %q, want %q", gotBase, tt.wantBase)
			}
		})
	}
}

func TestPullChart_UnknownType(t *testing.T) {
	upstream := Upstream{
		Type:  SourceType("git"),
		Chart: "test",
	}

	_, _, err := PullChart(t.Context(), upstream, "1.0.0")
	if err == nil {
		t.Fatal("expected error for unknown source type")
	}
	if !contains(err.Error(), "unknown source type") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "unknown source type")
	}
}
