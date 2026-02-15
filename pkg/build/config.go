package build

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
)

// ---------------------------------------------------------------------------
// Artifact types
// ---------------------------------------------------------------------------

const (
	// TypeImage builds a container image from a Dockerfile.
	TypeImage = "image"
	// TypeHelm packages a Helm chart directory as an OCI artifact.
	TypeHelm = "helm"
	// TypeArtifact packages generic files as an OCI artifact.
	TypeArtifact = "artifact"
	// TypeDocker pushes an existing Docker daemon image to a registry.
	TypeDocker = "docker"
)

// validTypes is the set of supported artifact types.
var validTypes = map[string]bool{
	TypeImage:    true,
	TypeHelm:     true,
	TypeArtifact: true,
	TypeDocker:   true,
}

// ---------------------------------------------------------------------------
// Config types
// ---------------------------------------------------------------------------

// Config represents a parsed .lazy configuration file.
type Config struct {
	// Version is the config schema version (must be 1).
	Version int `yaml:"version"`

	// Artifacts is the list of artifacts to build and push.
	Artifacts []Artifact `yaml:"artifacts"`
}

// Artifact describes a single OCI artifact to build and push.
type Artifact struct {
	// Type is the artifact type: "image", "helm", "artifact", or "docker".
	Type string `yaml:"type"`

	// Name is a human-readable identifier for this artifact.
	Name string `yaml:"name"`

	// Targets specifies where to push the built artifact.
	Targets []Target `yaml:"targets"`

	// --- type: image ---

	// Dockerfile is the path to the Dockerfile (default: "Dockerfile").
	Dockerfile string `yaml:"dockerfile,omitempty"`

	// Context is the build context directory (default: ".").
	Context string `yaml:"context,omitempty"`

	// Platforms lists target platforms for multi-arch builds (e.g., ["linux/amd64", "linux/arm64"]).
	Platforms []string `yaml:"platforms,omitempty"`

	// BuildArgs are --build-arg key=value pairs passed to docker buildx.
	BuildArgs map[string]string `yaml:"buildArgs,omitempty"`

	// --- type: helm ---

	// ChartPath is the path to the Helm chart directory.
	ChartPath string `yaml:"chartPath,omitempty"`

	// --- type: artifact ---

	// Files lists the files to include in the generic OCI artifact.
	Files []FileEntry `yaml:"files,omitempty"`

	// MediaType is the artifactType annotation for generic artifacts.
	MediaType string `yaml:"mediaType,omitempty"`

	// --- type: docker ---

	// Image is the Docker daemon image reference to push (e.g., "myapp:latest").
	Image string `yaml:"image,omitempty"`
}

// Target describes a registry push destination.
type Target struct {
	// Registry is the registry URL (e.g., "ghcr.io/owner", "docker.io/library").
	// Supports template variables, e.g., "{{ .Registry }}/path/to/repo".
	Registry string `yaml:"registry"`

	// Tags are the tags to push. Template variables are supported:
	//   {{ .Registry }}          — base registry URL (from LAZYOCI_REGISTRY env var)
	//   {{ .Tag }}               — value of --tag flag (or LAZYOCI_TAG env var)
	//   {{ .GitSHA }}            — current git commit SHA (short)
	//   {{ .GitBranch }}         — current git branch name
	//   {{ .ChartVersion }}      — version from Chart.yaml (helm only)
	//   {{ .Timestamp }}         — build timestamp (YYYYMMDDHHmmss)
	//   {{ .Version }}           — semver from git tag (v prefix stripped), e.g. "1.2.3"
	//   {{ .VersionMajor }}      — major component, e.g. "1"
	//   {{ .VersionMinor }}      — minor component, e.g. "2"
	//   {{ .VersionPatch }}      — patch component, e.g. "3"
	//   {{ .VersionPrerelease }} — prerelease identifier, e.g. "rc.1" (empty if none)
	//   {{ .VersionMajorMinor }} — "MAJOR.MINOR", e.g. "1.2"
	//   {{ .VersionRaw }}        — raw git tag string, e.g. "v1.2.3-rc.1"
	Tags []string `yaml:"tags"`
}

// FileEntry describes a file to include in a generic OCI artifact.
type FileEntry struct {
	// Path is the file path relative to the .lazy file location.
	Path string `yaml:"path"`

	// MediaType is the OCI media type for this file's layer.
	MediaType string `yaml:"mediaType"`
}

// ---------------------------------------------------------------------------
// Template variables
// ---------------------------------------------------------------------------

// TemplateVars holds the values available in tag and registry templates.
type TemplateVars struct {
	// Registry is the base registry URL from LAZYOCI_REGISTRY env var.
	// Used in registry field templates: {{ .Registry }}/path/to/repo
	Registry string

	// Tag is the value of the --tag flag (or LAZYOCI_TAG env var).
	Tag string

	// GitSHA is the current git commit SHA (short).
	GitSHA string

	// GitBranch is the current git branch name.
	GitBranch string

	// ChartVersion is the version from Chart.yaml (populated for helm artifacts).
	ChartVersion string

	// Timestamp is the build timestamp in YYYYMMDDHHmmss format.
	Timestamp string

	// Version is the clean semver from git tag (v prefix stripped), e.g., "1.2.3".
	// Source priority: LAZYOCI_VERSION env var > --tag flag (if semver) > git describe tag.
	Version string

	// VersionMajor is the major version component, e.g., "1".
	VersionMajor string

	// VersionMinor is the minor version component, e.g., "2".
	VersionMinor string

	// VersionPatch is the patch version component, e.g., "3".
	VersionPatch string

	// VersionPrerelease is the prerelease identifier, e.g., "rc.1" (empty if none).
	VersionPrerelease string

	// VersionMajorMinor is "MAJOR.MINOR", e.g., "1.2".
	VersionMajorMinor string

	// VersionRaw is the raw version string before parsing (e.g., "v1.2.3-rc.1").
	VersionRaw string
}

// ---------------------------------------------------------------------------
// Loading and parsing
// ---------------------------------------------------------------------------

// LoadConfig reads and parses a .lazy configuration file.
// The configPath should be the absolute or relative path to the .lazy file.
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config %s: %w", configPath, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config %s: %w", configPath, err)
	}

	return &cfg, nil
}

// Validate checks the config for errors.
func (c *Config) Validate() error {
	if c.Version != 1 {
		return fmt.Errorf("unsupported config version %d (expected 1)", c.Version)
	}

	if len(c.Artifacts) == 0 {
		return fmt.Errorf("config must define at least one artifact")
	}

	for i, a := range c.Artifacts {
		if err := a.validate(i); err != nil {
			return err
		}
	}

	return nil
}

// validate checks a single artifact definition for errors.
func (a *Artifact) validate(index int) error {
	prefix := fmt.Sprintf("artifact[%d]", index)
	if a.Name != "" {
		prefix = fmt.Sprintf("artifact %q", a.Name)
	}

	if a.Type == "" {
		return fmt.Errorf("%s: type is required", prefix)
	}
	if !validTypes[a.Type] {
		return fmt.Errorf("%s: unsupported type %q (must be one of: image, helm, artifact, docker)", prefix, a.Type)
	}

	if len(a.Targets) == 0 {
		return fmt.Errorf("%s: at least one target is required", prefix)
	}

	for i, t := range a.Targets {
		if t.Registry == "" {
			return fmt.Errorf("%s: target[%d].registry is required", prefix, i)
		}
		if len(t.Tags) == 0 {
			return fmt.Errorf("%s: target[%d].tags must have at least one tag", prefix, i)
		}
	}

	// Type-specific validation
	switch a.Type {
	case TypeImage:
		// Dockerfile and Context have defaults, so no required fields.
	case TypeHelm:
		if a.ChartPath == "" {
			return fmt.Errorf("%s: chartPath is required for helm artifacts", prefix)
		}
	case TypeArtifact:
		if len(a.Files) == 0 {
			return fmt.Errorf("%s: files is required for artifact type", prefix)
		}
		for i, f := range a.Files {
			if f.Path == "" {
				return fmt.Errorf("%s: files[%d].path is required", prefix, i)
			}
			if f.MediaType == "" {
				return fmt.Errorf("%s: files[%d].mediaType is required", prefix, i)
			}
		}
	case TypeDocker:
		if a.Image == "" {
			return fmt.Errorf("%s: image is required for docker artifacts", prefix)
		}
	}

	return nil
}

// ---------------------------------------------------------------------------
// Template rendering
// ---------------------------------------------------------------------------

// ResolveTemplateVars creates a TemplateVars populated with git info, version
// detection, and timestamp.
//
// The tag parameter comes from the --tag CLI flag (which should already have
// been resolved with LAZYOCI_TAG env var fallback by the caller).
//
// The chartVersion parameter is populated from Chart.yaml for helm artifacts
// (pass "" otherwise).
//
// Version resolution priority:
//  1. LAZYOCI_VERSION env var (if set and valid semver)
//  2. tag parameter (if valid semver)
//  3. git describe --tags --abbrev=0 (nearest annotated/lightweight tag)
func ResolveTemplateVars(tag, chartVersion string) *TemplateVars {
	vars := &TemplateVars{
		Registry:     os.Getenv("LAZYOCI_REGISTRY"),
		Tag:          tag,
		ChartVersion: chartVersion,
		Timestamp:    time.Now().UTC().Format("20060102150405"),
	}

	// Resolve git SHA
	if out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output(); err == nil {
		vars.GitSHA = strings.TrimSpace(string(out))
	}

	// Resolve git branch
	if out, err := exec.Command("git", "branch", "--show-current").Output(); err == nil {
		vars.GitBranch = strings.TrimSpace(string(out))
	}

	// Resolve version — try sources in priority order
	resolveVersion(vars, tag)

	return vars
}

// resolveVersion populates the Version* fields on TemplateVars.
// Priority: LAZYOCI_VERSION env > tag param (if semver) > git tag.
func resolveVersion(vars *TemplateVars, tag string) {
	// 1. LAZYOCI_VERSION env var — explicit override
	if envVersion := os.Getenv("LAZYOCI_VERSION"); envVersion != "" {
		if sv, ok := ParseSemver(envVersion); ok {
			applyVersion(vars, sv)
			return
		}
		// If it's set but not valid semver, still use it as raw
		vars.VersionRaw = envVersion
		vars.Version = envVersion
		return
	}

	// 2. --tag flag value — if it happens to be a semver
	if tag != "" {
		if sv, ok := ParseSemver(tag); ok {
			applyVersion(vars, sv)
			return
		}
	}

	// 3. Git tag — nearest tag on current commit or ancestors
	if out, err := exec.Command("git", "describe", "--tags", "--abbrev=0").Output(); err == nil {
		gitTag := strings.TrimSpace(string(out))
		if sv, ok := ParseSemver(gitTag); ok {
			applyVersion(vars, sv)
			return
		}
		// Git tag exists but isn't semver — use raw value
		vars.VersionRaw = gitTag
	}
}

// applyVersion populates all Version* fields from a parsed Semver.
func applyVersion(vars *TemplateVars, sv *Semver) {
	vars.Version = sv.Version()
	vars.VersionMajor = sv.Major
	vars.VersionMinor = sv.Minor
	vars.VersionPatch = sv.Patch
	vars.VersionPrerelease = sv.Prerelease
	vars.VersionMajorMinor = sv.MajorMinor()
	vars.VersionRaw = sv.Raw
}

// RenderTags applies template variables to all registry URLs and tags in all
// targets of an artifact.
// Returns a new slice of Targets with rendered values (original is not modified).
func RenderTags(targets []Target, vars *TemplateVars) ([]Target, error) {
	rendered := make([]Target, len(targets))
	for i, t := range targets {
		// Render registry URL through templates (supports {{ .Registry }})
		reg, err := renderTemplate(t.Registry, vars)
		if err != nil {
			return nil, fmt.Errorf("target[%d].registry: failed to render %q: %w", i, t.Registry, err)
		}
		rendered[i] = Target{
			Registry: reg,
			Tags:     make([]string, len(t.Tags)),
		}
		for j, tagTmpl := range t.Tags {
			result, err := renderTemplate(tagTmpl, vars)
			if err != nil {
				return nil, fmt.Errorf("target[%d].tags[%d]: failed to render %q: %w", i, j, tagTmpl, err)
			}
			rendered[i].Tags[j] = result
		}
	}
	return rendered, nil
}

// renderTemplate applies Go template variables to a tag string.
func renderTemplate(tmplStr string, vars *TemplateVars) (string, error) {
	// If no template delimiters, return as-is
	if !strings.Contains(tmplStr, "{{") {
		return tmplStr, nil
	}

	tmpl, err := template.New("tag").Option("missingkey=error").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// ResolveConfigPath finds the .lazy config file.
// If path is a directory, it looks for .lazy in that directory.
// If path is empty, it looks for .lazy in the current directory.
func ResolveConfigPath(path string) (string, error) {
	if path == "" {
		path = ".lazy"
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("config not found: %w", err)
	}

	if info.IsDir() {
		path = filepath.Join(path, ".lazy")
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("no .lazy file found in %s: %w", filepath.Dir(path), err)
		}
	}

	return filepath.Abs(path)
}

// BaseDir returns the directory containing the config file.
// All relative paths in the config are resolved from this directory.
func BaseDir(configPath string) string {
	return filepath.Dir(configPath)
}

// ReadChartVersion reads the version field from a Chart.yaml file.
func ReadChartVersion(chartPath string) (string, error) {
	data, err := os.ReadFile(filepath.Join(chartPath, "Chart.yaml"))
	if err != nil {
		return "", fmt.Errorf("failed to read Chart.yaml: %w", err)
	}

	var chart struct {
		Version string `yaml:"version"`
	}
	if err := yaml.Unmarshal(data, &chart); err != nil {
		return "", fmt.Errorf("failed to parse Chart.yaml: %w", err)
	}

	if chart.Version == "" {
		return "", fmt.Errorf("Chart.yaml does not contain a version field")
	}

	return chart.Version, nil
}
