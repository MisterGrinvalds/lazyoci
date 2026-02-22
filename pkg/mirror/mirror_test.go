package mirror

import (
	"bytes"
	"testing"

	"github.com/mistergrinvalds/lazyoci/pkg/config"
)

// testAppConfig returns a minimal app config for tests that create a Mirrorer.
// Without this, the credential chain panics on nil config in CI where no
// Docker credentials exist to short-circuit the lookup.
func testAppConfig() *config.Config {
	return &config.Config{}
}

func TestNew_DefaultConcurrency(t *testing.T) {
	tests := []struct {
		name        string
		concurrency int
		want        int
	}{
		{name: "zero defaults to 4", concurrency: 0, want: 4},
		{name: "negative defaults to 4", concurrency: -1, want: 4},
		{name: "explicit value preserved", concurrency: 8, want: 8},
		{name: "one preserved", concurrency: 1, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(Options{Concurrency: tt.concurrency})
			if m.opts.Concurrency != tt.want {
				t.Errorf("Concurrency = %d, want %d", m.opts.Concurrency, tt.want)
			}
		})
	}
}

func TestMirrorOne_UnknownChart(t *testing.T) {
	cfg := &Config{
		Target:    TargetConfig{URL: "registry.example.com"},
		Upstreams: map[string]Upstream{},
	}

	m := New(Options{Config: cfg})
	_, err := m.MirrorOne(t.Context(), "nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for unknown chart key")
	}
	if !contains(err.Error(), "not found in config") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "not found in config")
	}
}

func TestMirrorOne_NoVersions(t *testing.T) {
	cfg := &Config{
		Target: TargetConfig{URL: "registry.example.com"},
		Upstreams: map[string]Upstream{
			"myapp": {
				Type:     SourceLocal,
				Chart:    "myapp",
				Path:     "/tmp/charts/myapp",
				Versions: []string{},
			},
		},
	}

	m := New(Options{Config: cfg})
	_, err := m.MirrorOne(t.Context(), "myapp", nil)
	if err == nil {
		t.Fatal("expected error for no versions")
	}
	if !contains(err.Error(), "no versions specified") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "no versions specified")
	}
}

func TestMirrorOne_VersionOverride(t *testing.T) {
	cfg := &Config{
		Target: TargetConfig{URL: "registry.example.com"},
		Upstreams: map[string]Upstream{
			"myapp": {
				Type:     SourceLocal,
				Chart:    "myapp",
				Path:     "/tmp/charts/myapp",
				Versions: []string{"1.0.0"},
			},
		},
	}

	m := New(Options{
		Config:     cfg,
		AppConfig:  testAppConfig(),
		DryRun:     true,
		ChartsOnly: true, // skip image extraction to avoid needing helm binary
		Log:        &bytes.Buffer{},
	})

	result, err := m.MirrorOne(t.Context(), "myapp", []string{"2.0.0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Charts) != 1 {
		t.Fatalf("got %d charts, want 1", len(result.Charts))
	}
	if len(result.Charts[0].Versions) != 1 {
		t.Fatalf("got %d versions, want 1", len(result.Charts[0].Versions))
	}
	if result.Charts[0].Versions[0].Version != "2.0.0" {
		t.Errorf("version = %q, want %q", result.Charts[0].Versions[0].Version, "2.0.0")
	}
	if !result.DryRun {
		t.Error("expected DryRun=true in result")
	}
}

func TestMirrorAll_SkipsEmptyVersions(t *testing.T) {
	cfg := &Config{
		Target: TargetConfig{URL: "registry.example.com"},
		Upstreams: map[string]Upstream{
			"empty": {
				Type:     SourceLocal,
				Chart:    "empty",
				Path:     "/tmp/charts/empty",
				Versions: []string{},
			},
		},
	}

	var log bytes.Buffer
	m := New(Options{
		Config:    cfg,
		AppConfig: testAppConfig(),
		DryRun:    true,
		Log:       &log,
	})

	result, err := m.MirrorAll(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Charts) != 0 {
		t.Errorf("got %d charts, want 0 (empty versions should be skipped)", len(result.Charts))
	}
	if !contains(log.String(), "skipping") {
		t.Errorf("log should mention skipping, got: %s", log.String())
	}
}

func TestLogf_NilWriter(t *testing.T) {
	m := New(Options{Log: nil})
	// Should not panic when Log is nil.
	m.logf("this should not panic: %s", "test")
}
