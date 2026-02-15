package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	content := `version: 1
artifacts:
  - type: image
    name: myapp
    dockerfile: Dockerfile
    context: "."
    platforms:
      - linux/amd64
      - linux/arm64
    buildArgs:
      GO_VERSION: "1.22"
    targets:
      - registry: ghcr.io/owner/myapp
        tags:
          - "{{ .Tag }}"
          - "{{ .GitSHA }}"
  - type: helm
    name: mychart
    chartPath: charts/mychart
    targets:
      - registry: ghcr.io/owner/charts/mychart
        tags:
          - "{{ .ChartVersion }}"
`
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".lazy")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.Version != 1 {
		t.Errorf("Version = %d, want 1", cfg.Version)
	}
	if len(cfg.Artifacts) != 2 {
		t.Fatalf("len(Artifacts) = %d, want 2", len(cfg.Artifacts))
	}

	// Check image artifact
	img := cfg.Artifacts[0]
	if img.Type != TypeImage {
		t.Errorf("Artifacts[0].Type = %q, want %q", img.Type, TypeImage)
	}
	if img.Name != "myapp" {
		t.Errorf("Artifacts[0].Name = %q, want %q", img.Name, "myapp")
	}
	if img.Dockerfile != "Dockerfile" {
		t.Errorf("Artifacts[0].Dockerfile = %q, want %q", img.Dockerfile, "Dockerfile")
	}
	if len(img.Platforms) != 2 {
		t.Errorf("len(Artifacts[0].Platforms) = %d, want 2", len(img.Platforms))
	}
	if img.BuildArgs["GO_VERSION"] != "1.22" {
		t.Errorf("BuildArgs[GO_VERSION] = %q, want %q", img.BuildArgs["GO_VERSION"], "1.22")
	}
	if len(img.Targets) != 1 {
		t.Fatalf("len(Artifacts[0].Targets) = %d, want 1", len(img.Targets))
	}
	if img.Targets[0].Registry != "ghcr.io/owner/myapp" {
		t.Errorf("Target.Registry = %q", img.Targets[0].Registry)
	}
	if len(img.Targets[0].Tags) != 2 {
		t.Errorf("len(Target.Tags) = %d, want 2", len(img.Targets[0].Tags))
	}

	// Check helm artifact
	helm := cfg.Artifacts[1]
	if helm.Type != TypeHelm {
		t.Errorf("Artifacts[1].Type = %q, want %q", helm.Type, TypeHelm)
	}
	if helm.ChartPath != "charts/mychart" {
		t.Errorf("Artifacts[1].ChartPath = %q", helm.ChartPath)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr string
	}{
		{
			name:    "bad version",
			config:  Config{Version: 2, Artifacts: []Artifact{{Type: TypeImage, Targets: []Target{{Registry: "r", Tags: []string{"t"}}}}}},
			wantErr: "unsupported config version",
		},
		{
			name:    "no artifacts",
			config:  Config{Version: 1},
			wantErr: "at least one artifact",
		},
		{
			name: "missing type",
			config: Config{Version: 1, Artifacts: []Artifact{{
				Targets: []Target{{Registry: "r", Tags: []string{"t"}}},
			}}},
			wantErr: "type is required",
		},
		{
			name: "invalid type",
			config: Config{Version: 1, Artifacts: []Artifact{{
				Type:    "bogus",
				Targets: []Target{{Registry: "r", Tags: []string{"t"}}},
			}}},
			wantErr: "unsupported type",
		},
		{
			name: "no targets",
			config: Config{Version: 1, Artifacts: []Artifact{{
				Type: TypeImage,
			}}},
			wantErr: "at least one target",
		},
		{
			name: "empty registry",
			config: Config{Version: 1, Artifacts: []Artifact{{
				Type:    TypeImage,
				Targets: []Target{{Tags: []string{"t"}}},
			}}},
			wantErr: "registry is required",
		},
		{
			name: "no tags",
			config: Config{Version: 1, Artifacts: []Artifact{{
				Type:    TypeImage,
				Targets: []Target{{Registry: "r"}},
			}}},
			wantErr: "at least one tag",
		},
		{
			name: "helm missing chartPath",
			config: Config{Version: 1, Artifacts: []Artifact{{
				Type:    TypeHelm,
				Targets: []Target{{Registry: "r", Tags: []string{"t"}}},
			}}},
			wantErr: "chartPath is required",
		},
		{
			name: "artifact missing files",
			config: Config{Version: 1, Artifacts: []Artifact{{
				Type:    TypeArtifact,
				Targets: []Target{{Registry: "r", Tags: []string{"t"}}},
			}}},
			wantErr: "files is required",
		},
		{
			name: "artifact file missing path",
			config: Config{Version: 1, Artifacts: []Artifact{{
				Type:    TypeArtifact,
				Files:   []FileEntry{{MediaType: "application/json"}},
				Targets: []Target{{Registry: "r", Tags: []string{"t"}}},
			}}},
			wantErr: "files[0].path is required",
		},
		{
			name: "artifact file missing mediaType",
			config: Config{Version: 1, Artifacts: []Artifact{{
				Type:    TypeArtifact,
				Files:   []FileEntry{{Path: "data.json"}},
				Targets: []Target{{Registry: "r", Tags: []string{"t"}}},
			}}},
			wantErr: "files[0].mediaType is required",
		},
		{
			name: "docker missing image",
			config: Config{Version: 1, Artifacts: []Artifact{{
				Type:    TypeDocker,
				Targets: []Target{{Registry: "r", Tags: []string{"t"}}},
			}}},
			wantErr: "image is required",
		},
		{
			name: "valid image",
			config: Config{Version: 1, Artifacts: []Artifact{{
				Type:    TypeImage,
				Targets: []Target{{Registry: "ghcr.io/owner/app", Tags: []string{"latest"}}},
			}}},
		},
		{
			name: "valid helm",
			config: Config{Version: 1, Artifacts: []Artifact{{
				Type:      TypeHelm,
				ChartPath: "charts/myapp",
				Targets:   []Target{{Registry: "ghcr.io/owner/charts/myapp", Tags: []string{"0.1.0"}}},
			}}},
		},
		{
			name: "valid artifact",
			config: Config{Version: 1, Artifacts: []Artifact{{
				Type:    TypeArtifact,
				Files:   []FileEntry{{Path: "data.json", MediaType: "application/json"}},
				Targets: []Target{{Registry: "ghcr.io/owner/data", Tags: []string{"v1"}}},
			}}},
		},
		{
			name: "valid docker",
			config: Config{Version: 1, Artifacts: []Artifact{{
				Type:    TypeDocker,
				Image:   "myapp:latest",
				Targets: []Target{{Registry: "ghcr.io/owner/myapp", Tags: []string{"latest"}}},
			}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Validate() error = %q, want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestRenderTags(t *testing.T) {
	vars := &TemplateVars{
		Tag:          "v1.2.3",
		GitSHA:       "abc1234",
		GitBranch:    "main",
		ChartVersion: "0.5.0",
		Timestamp:    "20260208120000",
	}

	targets := []Target{
		{
			Registry: "ghcr.io/owner/app",
			Tags:     []string{"{{ .Tag }}", "{{ .GitSHA }}", "latest"},
		},
		{
			Registry: "docker.io/owner/app",
			Tags:     []string{"{{ .Tag }}-{{ .GitBranch }}", "{{ .Timestamp }}"},
		},
	}

	rendered, err := RenderTags(targets, vars)
	if err != nil {
		t.Fatalf("RenderTags() error = %v", err)
	}

	// First target
	if rendered[0].Registry != "ghcr.io/owner/app" {
		t.Errorf("rendered[0].Registry = %q", rendered[0].Registry)
	}
	wantTags0 := []string{"v1.2.3", "abc1234", "latest"}
	for i, want := range wantTags0 {
		if rendered[0].Tags[i] != want {
			t.Errorf("rendered[0].Tags[%d] = %q, want %q", i, rendered[0].Tags[i], want)
		}
	}

	// Second target
	wantTags1 := []string{"v1.2.3-main", "20260208120000"}
	for i, want := range wantTags1 {
		if rendered[1].Tags[i] != want {
			t.Errorf("rendered[1].Tags[%d] = %q, want %q", i, rendered[1].Tags[i], want)
		}
	}
}

func TestRenderTagsWithVersionVars(t *testing.T) {
	vars := &TemplateVars{
		Tag:               "v1.2.3",
		Version:           "1.2.3",
		VersionMajor:      "1",
		VersionMinor:      "2",
		VersionPatch:      "3",
		VersionPrerelease: "",
		VersionMajorMinor: "1.2",
		VersionRaw:        "v1.2.3",
	}

	targets := []Target{
		{
			Registry: "ghcr.io/owner/app",
			Tags: []string{
				"{{ .Version }}",
				"{{ .VersionMajor }}.{{ .VersionMinor }}",
				"{{ .VersionMajorMinor }}",
				"{{ .VersionRaw }}",
			},
		},
	}

	rendered, err := RenderTags(targets, vars)
	if err != nil {
		t.Fatalf("RenderTags() error = %v", err)
	}

	want := []string{"1.2.3", "1.2", "1.2", "v1.2.3"}
	for i, w := range want {
		if rendered[0].Tags[i] != w {
			t.Errorf("rendered[0].Tags[%d] = %q, want %q", i, rendered[0].Tags[i], w)
		}
	}
}

func TestRenderTagsWithPrerelease(t *testing.T) {
	vars := &TemplateVars{
		Tag:               "v2.0.0-rc.1",
		Version:           "2.0.0",
		VersionMajor:      "2",
		VersionMinor:      "0",
		VersionPatch:      "0",
		VersionPrerelease: "rc.1",
		VersionMajorMinor: "2.0",
		VersionRaw:        "v2.0.0-rc.1",
	}

	targets := []Target{
		{
			Registry: "ghcr.io/owner/app",
			Tags: []string{
				"{{ .Version }}-{{ .VersionPrerelease }}",
				"{{ .Version }}",
			},
		},
	}

	rendered, err := RenderTags(targets, vars)
	if err != nil {
		t.Fatalf("RenderTags() error = %v", err)
	}

	want := []string{"2.0.0-rc.1", "2.0.0"}
	for i, w := range want {
		if rendered[0].Tags[i] != w {
			t.Errorf("rendered[0].Tags[%d] = %q, want %q", i, rendered[0].Tags[i], w)
		}
	}
}

func TestResolveVersionFromEnvVar(t *testing.T) {
	// LAZYOCI_VERSION env var should take highest priority
	t.Setenv("LAZYOCI_VERSION", "v3.4.5-beta.1")

	vars := &TemplateVars{}
	resolveVersion(vars, "v1.0.0") // tag arg should be ignored when env is set

	if vars.Version != "3.4.5" {
		t.Errorf("Version = %q, want %q", vars.Version, "3.4.5")
	}
	if vars.VersionMajor != "3" {
		t.Errorf("VersionMajor = %q, want %q", vars.VersionMajor, "3")
	}
	if vars.VersionMinor != "4" {
		t.Errorf("VersionMinor = %q, want %q", vars.VersionMinor, "4")
	}
	if vars.VersionPatch != "5" {
		t.Errorf("VersionPatch = %q, want %q", vars.VersionPatch, "5")
	}
	if vars.VersionPrerelease != "beta.1" {
		t.Errorf("VersionPrerelease = %q, want %q", vars.VersionPrerelease, "beta.1")
	}
	if vars.VersionMajorMinor != "3.4" {
		t.Errorf("VersionMajorMinor = %q, want %q", vars.VersionMajorMinor, "3.4")
	}
	if vars.VersionRaw != "v3.4.5-beta.1" {
		t.Errorf("VersionRaw = %q, want %q", vars.VersionRaw, "v3.4.5-beta.1")
	}
}

func TestResolveVersionFromEnvVarNonSemver(t *testing.T) {
	// If LAZYOCI_VERSION is set but not valid semver, use it as raw/version
	t.Setenv("LAZYOCI_VERSION", "nightly-20260214")

	vars := &TemplateVars{}
	resolveVersion(vars, "")

	if vars.Version != "nightly-20260214" {
		t.Errorf("Version = %q, want %q", vars.Version, "nightly-20260214")
	}
	if vars.VersionRaw != "nightly-20260214" {
		t.Errorf("VersionRaw = %q, want %q", vars.VersionRaw, "nightly-20260214")
	}
	// Component fields should be empty for non-semver
	if vars.VersionMajor != "" {
		t.Errorf("VersionMajor = %q, want empty", vars.VersionMajor)
	}
}

func TestResolveVersionFromTag(t *testing.T) {
	// No env var set — should fall back to tag parameter
	t.Setenv("LAZYOCI_VERSION", "") // explicitly clear

	vars := &TemplateVars{}
	resolveVersion(vars, "v2.1.0")

	if vars.Version != "2.1.0" {
		t.Errorf("Version = %q, want %q", vars.Version, "2.1.0")
	}
	if vars.VersionMajor != "2" {
		t.Errorf("VersionMajor = %q, want %q", vars.VersionMajor, "2")
	}
	if vars.VersionMinor != "1" {
		t.Errorf("VersionMinor = %q, want %q", vars.VersionMinor, "1")
	}
	if vars.VersionPatch != "0" {
		t.Errorf("VersionPatch = %q, want %q", vars.VersionPatch, "0")
	}
}

func TestResolveVersionNonSemverTagFallsToGit(t *testing.T) {
	// Non-semver tag should not populate version components — falls through to git
	// In test env, git describe may or may not find a tag; we just verify the tag
	// didn't populate the semver fields
	t.Setenv("LAZYOCI_VERSION", "")

	vars := &TemplateVars{}
	resolveVersion(vars, "latest")

	// "latest" is not semver, so version fields should come from git (or be empty)
	// We can't predict git state, but VersionMajor should NOT be set from "latest"
	if vars.VersionMajor == "latest" {
		t.Error("VersionMajor should not be set from non-semver tag")
	}
}

func TestApplyVersion(t *testing.T) {
	sv, ok := ParseSemver("v1.2.3-rc.1+build.42")
	if !ok {
		t.Fatal("ParseSemver failed")
	}

	vars := &TemplateVars{}
	applyVersion(vars, sv)

	if vars.Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", vars.Version, "1.2.3")
	}
	if vars.VersionMajor != "1" {
		t.Errorf("VersionMajor = %q, want %q", vars.VersionMajor, "1")
	}
	if vars.VersionMinor != "2" {
		t.Errorf("VersionMinor = %q, want %q", vars.VersionMinor, "2")
	}
	if vars.VersionPatch != "3" {
		t.Errorf("VersionPatch = %q, want %q", vars.VersionPatch, "3")
	}
	if vars.VersionPrerelease != "rc.1" {
		t.Errorf("VersionPrerelease = %q, want %q", vars.VersionPrerelease, "rc.1")
	}
	if vars.VersionMajorMinor != "1.2" {
		t.Errorf("VersionMajorMinor = %q, want %q", vars.VersionMajorMinor, "1.2")
	}
	if vars.VersionRaw != "v1.2.3-rc.1+build.42" {
		t.Errorf("VersionRaw = %q, want %q", vars.VersionRaw, "v1.2.3-rc.1+build.42")
	}
}

func TestRenderTagsWithRegistryTemplate(t *testing.T) {
	vars := &TemplateVars{
		Registry: "registry.digitalocean.com/greenforests",
		Tag:      "v1.0.0",
		Version:  "1.0.0",
	}

	targets := []Target{
		{
			Registry: "{{ .Registry }}/examples/hello-server",
			Tags:     []string{"{{ .Version }}", "latest"},
		},
	}

	rendered, err := RenderTags(targets, vars)
	if err != nil {
		t.Fatalf("RenderTags() error = %v", err)
	}

	wantRegistry := "registry.digitalocean.com/greenforests/examples/hello-server"
	if rendered[0].Registry != wantRegistry {
		t.Errorf("rendered[0].Registry = %q, want %q", rendered[0].Registry, wantRegistry)
	}
	if rendered[0].Tags[0] != "1.0.0" {
		t.Errorf("rendered[0].Tags[0] = %q, want %q", rendered[0].Tags[0], "1.0.0")
	}
	if rendered[0].Tags[1] != "latest" {
		t.Errorf("rendered[0].Tags[1] = %q, want %q", rendered[0].Tags[1], "latest")
	}
}

func TestRenderTagsHardcodedRegistryStillWorks(t *testing.T) {
	// When Registry is empty, hardcoded registry URLs should pass through unchanged
	vars := &TemplateVars{
		Tag: "v1.0.0",
	}

	targets := []Target{
		{
			Registry: "ghcr.io/owner/myapp",
			Tags:     []string{"{{ .Tag }}", "latest"},
		},
	}

	rendered, err := RenderTags(targets, vars)
	if err != nil {
		t.Fatalf("RenderTags() error = %v", err)
	}

	if rendered[0].Registry != "ghcr.io/owner/myapp" {
		t.Errorf("rendered[0].Registry = %q, want %q", rendered[0].Registry, "ghcr.io/owner/myapp")
	}
}

func TestResolveTemplateVarsRegistryFromEnv(t *testing.T) {
	t.Setenv("LAZYOCI_REGISTRY", "localhost:5050")
	t.Setenv("LAZYOCI_VERSION", "")

	vars := ResolveTemplateVars("v1.0.0", "")
	if vars.Registry != "localhost:5050" {
		t.Errorf("Registry = %q, want %q", vars.Registry, "localhost:5050")
	}
}

func TestResolveTemplateVarsRegistryEmpty(t *testing.T) {
	t.Setenv("LAZYOCI_REGISTRY", "")

	vars := ResolveTemplateVars("v1.0.0", "")
	if vars.Registry != "" {
		t.Errorf("Registry = %q, want empty", vars.Registry)
	}
}

func TestRenderTagsError(t *testing.T) {
	vars := &TemplateVars{Tag: "v1"}

	// Missing key should error with missingkey=error
	targets := []Target{
		{Registry: "r", Tags: []string{"{{ .NoSuchField }}"}},
	}

	_, err := RenderTags(targets, vars)
	if err == nil {
		t.Fatal("expected error for undefined template variable")
	}
}

func TestResolveConfigPath(t *testing.T) {
	dir := t.TempDir()

	// Create a .lazy file
	lazyPath := filepath.Join(dir, ".lazy")
	if err := os.WriteFile(lazyPath, []byte("version: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test with directory
	resolved, err := ResolveConfigPath(dir)
	if err != nil {
		t.Fatalf("ResolveConfigPath(dir) error = %v", err)
	}
	if filepath.Base(resolved) != ".lazy" {
		t.Errorf("expected .lazy, got %s", filepath.Base(resolved))
	}

	// Test with file path directly
	resolved2, err := ResolveConfigPath(lazyPath)
	if err != nil {
		t.Fatalf("ResolveConfigPath(file) error = %v", err)
	}
	if resolved != resolved2 {
		t.Errorf("paths differ: %s vs %s", resolved, resolved2)
	}

	// Test with nonexistent path
	_, err = ResolveConfigPath(filepath.Join(dir, "nonexistent"))
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestReadChartVersion(t *testing.T) {
	dir := t.TempDir()
	chartYaml := `apiVersion: v2
name: mychart
version: 1.2.3
description: A test chart
`
	if err := os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte(chartYaml), 0644); err != nil {
		t.Fatal(err)
	}

	version, err := ReadChartVersion(dir)
	if err != nil {
		t.Fatalf("ReadChartVersion() error = %v", err)
	}
	if version != "1.2.3" {
		t.Errorf("version = %q, want %q", version, "1.2.3")
	}

	// Missing version field
	dir2 := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir2, "Chart.yaml"), []byte("apiVersion: v2\nname: x\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err = ReadChartVersion(dir2)
	if err == nil {
		t.Fatal("expected error for missing version")
	}

	// Missing Chart.yaml
	_, err = ReadChartVersion(t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing Chart.yaml")
	}
}
