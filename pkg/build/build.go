package build

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"oras.land/oras-go/v2/registry/remote/auth"
)

// ---------------------------------------------------------------------------
// Builder
// ---------------------------------------------------------------------------

// BuilderOptions configures a Builder.
type BuilderOptions struct {
	// Tag is the value for the {{ .Tag }} template variable.
	Tag string

	// Push enables pushing to registries after building (default true).
	Push bool

	// DryRun shows what would be built/pushed without doing it.
	DryRun bool

	// Quiet suppresses progress output.
	Quiet bool

	// Insecure allows HTTP for push targets.
	Insecure bool

	// Platforms overrides platforms for image-type artifacts.
	Platforms []string

	// ArtifactFilter limits the build to a specific artifact by name or 0-based index string.
	ArtifactFilter string

	// CredentialFunc resolves auth credentials for a given registry URL.
	CredentialFunc func(registryURL string) auth.CredentialFunc

	// Output is the writer for progress messages. Defaults to os.Stderr.
	Output io.Writer
}

// Builder orchestrates building and pushing OCI artifacts from a .lazy config.
type Builder struct {
	config  *Config
	opts    BuilderOptions
	baseDir string // directory containing the .lazy file
}

// BuildResult describes the outcome of building a single artifact.
type BuildResult struct {
	// Name is the artifact name from the config.
	Name string `json:"name" yaml:"name"`

	// Type is the artifact type.
	Type string `json:"type" yaml:"type"`

	// Targets lists push results for each target/tag combination.
	Targets []TargetResult `json:"targets,omitempty" yaml:"targets,omitempty"`

	// Error is set if the build failed.
	Error string `json:"error,omitempty" yaml:"error,omitempty"`
}

// TargetResult describes the outcome of pushing to a single target tag.
type TargetResult struct {
	// Reference is the full pushed reference (registry/repo:tag).
	Reference string `json:"reference" yaml:"reference"`

	// Digest is the manifest digest after push.
	Digest string `json:"digest,omitempty" yaml:"digest,omitempty"`

	// Pushed is true if the artifact was actually pushed (false in dry-run or no-push mode).
	Pushed bool `json:"pushed" yaml:"pushed"`
}

// NewBuilder creates a Builder for the given config.
// The configPath is the absolute path to the .lazy file (used to resolve relative paths).
func NewBuilder(cfg *Config, configPath string, opts BuilderOptions) *Builder {
	if opts.Output == nil {
		opts.Output = os.Stderr
	}
	return &Builder{
		config:  cfg,
		opts:    opts,
		baseDir: filepath.Dir(configPath),
	}
}

// Build builds all artifacts (or a filtered subset) and returns results.
func (b *Builder) Build(ctx context.Context) ([]BuildResult, error) {
	artifacts := b.config.Artifacts

	// Filter if requested
	if b.opts.ArtifactFilter != "" {
		filtered, err := b.filterArtifacts(artifacts)
		if err != nil {
			return nil, err
		}
		artifacts = filtered
	}

	var results []BuildResult
	for i, a := range artifacts {
		result, err := b.buildOne(ctx, &a, i)
		if err != nil {
			results = append(results, BuildResult{
				Name:  a.Name,
				Type:  a.Type,
				Error: err.Error(),
			})
			continue
		}
		results = append(results, *result)
	}

	return results, nil
}

// BuildArtifact builds a single artifact by its 0-based index.
func (b *Builder) BuildArtifact(ctx context.Context, idx int) (*BuildResult, error) {
	if idx < 0 || idx >= len(b.config.Artifacts) {
		return nil, fmt.Errorf("artifact index %d out of range (0-%d)", idx, len(b.config.Artifacts)-1)
	}
	return b.buildOne(ctx, &b.config.Artifacts[idx], idx)
}

// buildOne builds a single artifact: resolve templates, dispatch to handler, push.
func (b *Builder) buildOne(ctx context.Context, artifact *Artifact, idx int) (*BuildResult, error) {
	name := artifact.Name
	if name == "" {
		name = fmt.Sprintf("artifact[%d]", idx)
	}

	b.logf("Building %s (type: %s)...\n", name, artifact.Type)

	// Resolve chart version for helm artifacts (needed for template vars)
	chartVersion := ""
	if artifact.Type == TypeHelm {
		chartPath := b.resolvePath(artifact.ChartPath)
		v, err := ReadChartVersion(chartPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read chart version: %w", err)
		}
		chartVersion = v
	}

	// Resolve template variables and render tags
	vars := ResolveTemplateVars(b.opts.Tag, chartVersion)
	renderedTargets, err := RenderTags(artifact.Targets, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render tags: %w", err)
	}

	if b.opts.DryRun {
		return b.dryRunResult(name, artifact, renderedTargets), nil
	}

	// Dispatch to type-specific builder
	var ociLayoutPath string
	switch artifact.Type {
	case TypeImage:
		ociLayoutPath, err = b.buildImage(ctx, artifact)
	case TypeHelm:
		ociLayoutPath, err = b.buildHelm(ctx, artifact, chartVersion)
	case TypeArtifact:
		ociLayoutPath, err = b.buildArtifact(ctx, artifact)
	case TypeDocker:
		ociLayoutPath, err = b.buildDocker(ctx, artifact)
	default:
		return nil, fmt.Errorf("unsupported artifact type: %s", artifact.Type)
	}
	if err != nil {
		return nil, err
	}

	// Clean up the temporary OCI layout when done
	if ociLayoutPath != "" {
		defer os.RemoveAll(ociLayoutPath)
	}

	// Push to all targets
	result := &BuildResult{
		Name: name,
		Type: artifact.Type,
	}

	if b.opts.Push {
		for _, target := range renderedTargets {
			for _, tag := range target.Tags {
				tr, err := b.pushToTarget(ctx, ociLayoutPath, tag, target)
				if err != nil {
					return nil, fmt.Errorf("push to %s:%s failed: %w", target.Registry, tag, err)
				}
				result.Targets = append(result.Targets, *tr)
			}
		}
	} else {
		// No push — just record what would have been pushed
		for _, target := range renderedTargets {
			for _, tag := range target.Tags {
				result.Targets = append(result.Targets, TargetResult{
					Reference: target.Registry + ":" + tag,
					Pushed:    false,
				})
			}
		}
	}

	return result, nil
}

// dryRunResult creates a BuildResult showing what would be built/pushed.
func (b *Builder) dryRunResult(name string, artifact *Artifact, targets []Target) *BuildResult {
	result := &BuildResult{
		Name: name,
		Type: artifact.Type,
	}

	for _, target := range targets {
		for _, tag := range target.Tags {
			ref := target.Registry + ":" + tag
			b.logf("  [dry-run] would push %s\n", ref)
			result.Targets = append(result.Targets, TargetResult{
				Reference: ref,
				Pushed:    false,
			})
		}
	}

	return result
}

// filterArtifacts returns only the artifacts matching the filter.
// The filter can be a name or a 0-based index string.
func (b *Builder) filterArtifacts(artifacts []Artifact) ([]Artifact, error) {
	filter := b.opts.ArtifactFilter

	// Try matching by name
	for _, a := range artifacts {
		if a.Name == filter {
			return []Artifact{a}, nil
		}
	}

	// Try matching by type
	var byType []Artifact
	for _, a := range artifacts {
		if a.Type == filter {
			byType = append(byType, a)
		}
	}
	if len(byType) > 0 {
		return byType, nil
	}

	// Try matching by index
	var idx int
	if _, err := fmt.Sscanf(filter, "%d", &idx); err == nil {
		if idx >= 0 && idx < len(artifacts) {
			return []Artifact{artifacts[idx]}, nil
		}
	}

	return nil, fmt.Errorf("no artifact matching filter %q", filter)
}

// resolvePath resolves a path relative to the .lazy file's directory.
func (b *Builder) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(b.baseDir, path)
}

// logf writes a formatted message to the output writer (unless quiet).
func (b *Builder) logf(format string, args ...any) {
	if !b.opts.Quiet {
		fmt.Fprintf(b.opts.Output, format, args...)
	}
}
