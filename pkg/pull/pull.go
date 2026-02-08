package pull

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mistergrinvalds/lazyoci/pkg/registry"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// PullOptions configures a pull operation.
type PullOptions struct {
	// Reference is the full image reference (e.g., "docker.io/library/nginx:latest").
	Reference string

	// Destination is the explicit directory for storage.
	// If empty, uses ArtifactBase with type-aware subdirectories.
	Destination string

	// ArtifactBase is the base directory for artifacts when Destination is empty.
	// Defaults to ~/.cache/lazyoci/artifacts if not specified.
	ArtifactBase string

	// Platform filters to a specific platform (e.g., linux/amd64).
	// If nil, defaults to the current OS/arch for images.
	Platform *ocispec.Platform

	// ToDocker loads the pulled image into the Docker daemon after pulling.
	ToDocker bool

	// Quiet suppresses progress output.
	Quiet bool

	// Insecure allows pulling over HTTP.
	Insecure bool

	// CredentialFunc provides authentication credentials for the registry.
	// If nil, anonymous auth is used.
	CredentialFunc auth.CredentialFunc
}

// PullResult contains information about a completed pull.
type PullResult struct {
	Reference      string                `json:"reference" yaml:"reference"`
	Digest         string                `json:"digest" yaml:"digest"`
	Size           int64                 `json:"size" yaml:"size"`
	Destination    string                `json:"destination" yaml:"destination"`
	Layers         int                   `json:"layers" yaml:"layers"`
	ArtifactType   registry.ArtifactType `json:"artifactType" yaml:"artifactType"`
	TypeDetail     string                `json:"typeDetail,omitempty" yaml:"typeDetail,omitempty"`
	LoadedToDocker bool                  `json:"loadedToDocker" yaml:"loadedToDocker"`
}

// Puller handles pulling OCI artifacts from registries.
type Puller struct {
	tracker *ProgressTracker
}

// NewPuller creates a new Puller with the given options.
func NewPuller(quiet bool) *Puller {
	return &Puller{
		tracker: NewProgressTracker(quiet),
	}
}

// Pull downloads an OCI artifact from a registry to local storage.
func (p *Puller) Pull(ctx context.Context, opts PullOptions) (*PullResult, error) {
	// Parse reference
	ref, err := ParseReference(opts.Reference)
	if err != nil {
		return nil, fmt.Errorf("invalid reference: %w", err)
	}

	// Create remote repository
	repo, err := NewRemoteRepository(ref, opts.Insecure, opts.CredentialFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to registry: %w", err)
	}

	// Detect artifact type by inspecting manifest
	artifactType, typeDetail, configMediaType := p.detectArtifactType(ctx, repo, ref.Tag)

	// Determine destination based on artifact type
	dest := opts.Destination
	if dest == "" {
		// Use ArtifactBase or default
		artifactBase := opts.ArtifactBase
		if artifactBase == "" {
			homeDir, _ := os.UserHomeDir()
			artifactBase = filepath.Join(homeDir, ".cache", "lazyoci", "artifacts")
		}
		// Use type-specific subdirectories
		typeDir := getTypeDirectory(artifactType)
		dest = filepath.Join(artifactBase, typeDir, ref.Registry, ref.Repository, ref.Tag)
	}

	// Route to type-specific pull based on artifact type
	// (For now, all types use OCI layout; type-specific extraction will be added in Phase 5)
	switch artifactType {
	case registry.ArtifactTypeHelmChart:
		return p.pullOCILayout(ctx, repo, ref, dest, opts, artifactType, typeDetail, configMediaType)
	case registry.ArtifactTypeSBOM:
		return p.pullOCILayout(ctx, repo, ref, dest, opts, artifactType, typeDetail, configMediaType)
	case registry.ArtifactTypeSignature:
		return p.pullOCILayout(ctx, repo, ref, dest, opts, artifactType, typeDetail, configMediaType)
	case registry.ArtifactTypeAttestation:
		return p.pullOCILayout(ctx, repo, ref, dest, opts, artifactType, typeDetail, configMediaType)
	case registry.ArtifactTypeWasm:
		return p.pullOCILayout(ctx, repo, ref, dest, opts, artifactType, typeDetail, configMediaType)
	default:
		// Images and unknown artifacts use standard OCI layout
		return p.pullOCILayout(ctx, repo, ref, dest, opts, artifactType, typeDetail, configMediaType)
	}
}

// detectArtifactType fetches the manifest and determines the artifact type.
func (p *Puller) detectArtifactType(ctx context.Context, repo *remote.Repository, tag string) (registry.ArtifactType, string, string) {
	// Resolve tag to get manifest descriptor
	desc, err := repo.Resolve(ctx, tag)
	if err != nil {
		return registry.ArtifactTypeUnknown, "", ""
	}

	// Fetch manifest to inspect config media type
	manifestReader, err := repo.Fetch(ctx, desc)
	if err != nil {
		// Fall back to manifest media type only
		return detectTypeFromMediaType(desc.MediaType), "", ""
	}
	defer manifestReader.Close()

	var manifest struct {
		Config struct {
			MediaType string `json:"mediaType"`
		} `json:"config"`
		Layers []struct {
			MediaType string `json:"mediaType"`
		} `json:"layers"`
	}

	if err := json.NewDecoder(manifestReader).Decode(&manifest); err != nil {
		return detectTypeFromMediaType(desc.MediaType), "", ""
	}

	// Collect all media types for detection
	var layerTypes []string
	for _, layer := range manifest.Layers {
		layerTypes = append(layerTypes, layer.MediaType)
	}

	artifactType, typeDetail := detectArtifactTypeFromMediaTypes(
		desc.MediaType,
		manifest.Config.MediaType,
		layerTypes,
	)

	return artifactType, typeDetail, manifest.Config.MediaType
}

// pullOCILayout performs the standard OCI layout pull.
func (p *Puller) pullOCILayout(
	ctx context.Context,
	repo *remote.Repository,
	ref *Reference,
	dest string,
	opts PullOptions,
	artifactType registry.ArtifactType,
	typeDetail string,
	configMediaType string,
) (*PullResult, error) {
	// Create destination directory
	if err := os.MkdirAll(dest, 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination: %w", err)
	}

	// Create local OCI store
	store, err := oci.New(dest)
	if err != nil {
		return nil, fmt.Errorf("failed to create OCI store: %w", err)
	}

	// Set up platform filter (only for images)
	platform := opts.Platform
	if platform == nil && artifactType == registry.ArtifactTypeImage {
		platform = &ocispec.Platform{
			OS:           runtime.GOOS,
			Architecture: runtime.GOARCH,
		}
	}

	// Track layers for result
	var layerCount int

	// Set up copy options with progress tracking
	copyOpts := oras.CopyOptions{
		CopyGraphOptions: oras.CopyGraphOptions{
			PreCopy: func(ctx context.Context, desc ocispec.Descriptor) error {
				// Only show progress for layers (blobs), not manifests/configs
				if isLayer(desc.MediaType) {
					p.tracker.StartLayer(desc.Digest.String(), desc.Size)
					layerCount++
				}
				return nil
			},
			PostCopy: func(ctx context.Context, desc ocispec.Descriptor) error {
				if isLayer(desc.MediaType) {
					p.tracker.FinishLayer(desc.Digest.String())
				}
				return nil
			},
			OnCopySkipped: func(ctx context.Context, desc ocispec.Descriptor) error {
				// Layer already exists locally
				if isLayer(desc.MediaType) {
					if !opts.Quiet {
						fmt.Printf("Layer %s already exists, skipping\n", shortDigest(desc.Digest.String()))
					}
					layerCount++
				}
				return nil
			},
		},
	}

	// Only apply platform filter for images with explicit platform
	if opts.Platform != nil && artifactType == registry.ArtifactTypeImage {
		copyOpts.WithTargetPlatform(platform)
	}

	// Perform the copy
	desc, err := oras.Copy(ctx, repo, ref.Tag, store, ref.Tag, copyOpts)
	if err != nil {
		return nil, fmt.Errorf("pull failed: %w", err)
	}

	p.tracker.Finish()

	result := &PullResult{
		Reference:    opts.Reference,
		Digest:       desc.Digest.String(),
		Size:         desc.Size,
		Destination:  dest,
		Layers:       layerCount,
		ArtifactType: artifactType,
		TypeDetail:   typeDetail,
	}

	// Load into Docker if requested (only for images)
	if opts.ToDocker {
		if artifactType != registry.ArtifactTypeImage {
			return result, fmt.Errorf("cannot load %s artifact into Docker (only images supported)", artifactType)
		}
		// Build a full reference with tag — Docker requires "repo:tag" format.
		dockerRef := ref.Registry + "/" + ref.Repository + ":" + ref.Tag
		if err := LoadToDocker(dest, dockerRef); err != nil {
			return result, fmt.Errorf("pulled but failed to load into Docker: %w", err)
		}
		result.LoadedToDocker = true
	}

	return result, nil
}

// getTypeDirectory returns the subdirectory name for an artifact type.
func getTypeDirectory(t registry.ArtifactType) string {
	switch t {
	case registry.ArtifactTypeImage:
		return "oci"
	case registry.ArtifactTypeHelmChart:
		return "helm"
	case registry.ArtifactTypeSBOM:
		return "sbom"
	case registry.ArtifactTypeSignature:
		return "sig"
	case registry.ArtifactTypeAttestation:
		return "att"
	case registry.ArtifactTypeWasm:
		return "wasm"
	default:
		return "oci"
	}
}

// detectTypeFromMediaType performs simple type detection from manifest media type only.
func detectTypeFromMediaType(mediaType string) registry.ArtifactType {
	mt := strings.ToLower(mediaType)
	switch {
	case strings.Contains(mt, "helm"):
		return registry.ArtifactTypeHelmChart
	case strings.Contains(mt, "sbom") || strings.Contains(mt, "spdx") || strings.Contains(mt, "cyclonedx"):
		return registry.ArtifactTypeSBOM
	case strings.Contains(mt, "signature") || strings.Contains(mt, "cosign") || strings.Contains(mt, "notary"):
		return registry.ArtifactTypeSignature
	case strings.Contains(mt, "attestation") || strings.Contains(mt, "in-toto") || strings.Contains(mt, "dsse"):
		return registry.ArtifactTypeAttestation
	case strings.Contains(mt, "wasm"):
		return registry.ArtifactTypeWasm
	default:
		return registry.ArtifactTypeImage
	}
}

// detectArtifactTypeFromMediaTypes performs comprehensive type detection.
func detectArtifactTypeFromMediaTypes(manifestType, configType string, layerTypes []string) (registry.ArtifactType, string) {
	// Normalize to lowercase
	manifest := strings.ToLower(manifestType)
	config := strings.ToLower(configType)

	allTypes := append([]string{manifest, config}, layerTypes...)

	// Helm Chart
	for _, mt := range allTypes {
		if strings.Contains(mt, "helm") {
			return registry.ArtifactTypeHelmChart, ""
		}
	}

	// SBOM
	for _, mt := range allTypes {
		if strings.Contains(mt, "spdx") {
			return registry.ArtifactTypeSBOM, "spdx"
		}
		if strings.Contains(mt, "cyclonedx") {
			return registry.ArtifactTypeSBOM, "cyclonedx"
		}
		if strings.Contains(mt, "sbom") {
			return registry.ArtifactTypeSBOM, ""
		}
	}

	// Signature
	for _, mt := range allTypes {
		if strings.Contains(mt, "cosign") {
			return registry.ArtifactTypeSignature, "cosign"
		}
		if strings.Contains(mt, "notary") && strings.Contains(mt, "signature") {
			return registry.ArtifactTypeSignature, "notary"
		}
		if strings.Contains(mt, "signature") {
			return registry.ArtifactTypeSignature, ""
		}
	}

	// Attestation
	for _, mt := range allTypes {
		if strings.Contains(mt, "in-toto") {
			return registry.ArtifactTypeAttestation, "in-toto"
		}
		if strings.Contains(mt, "dsse") {
			return registry.ArtifactTypeAttestation, "dsse"
		}
		if strings.Contains(mt, "attestation") {
			return registry.ArtifactTypeAttestation, ""
		}
	}

	// WebAssembly
	for _, mt := range allTypes {
		if strings.Contains(mt, "wasm") {
			return registry.ArtifactTypeWasm, ""
		}
	}

	// Container Image
	if strings.Contains(manifest, "image") || strings.Contains(manifest, "docker") {
		return registry.ArtifactTypeImage, ""
	}
	if strings.Contains(config, "image") || strings.Contains(config, "docker") {
		return registry.ArtifactTypeImage, ""
	}

	return registry.ArtifactTypeUnknown, ""
}

// Reference represents a parsed OCI reference.
type Reference struct {
	Registry   string
	Repository string
	Tag        string
}

// ParseReference parses an image reference like "docker.io/library/nginx:latest".
func ParseReference(ref string) (*Reference, error) {
	// Handle docker.io shorthand
	if !strings.Contains(ref, "/") {
		// Single name like "nginx" → docker.io/library/nginx
		ref = "docker.io/library/" + ref
	} else if !strings.Contains(strings.Split(ref, "/")[0], ".") &&
		!strings.Contains(strings.Split(ref, "/")[0], ":") &&
		!strings.Contains(strings.Split(ref, "/")[0], "localhost") {
		// No registry specified, e.g., "library/nginx" → docker.io/library/nginx
		ref = "docker.io/" + ref
	}

	// Split off tag
	tag := "latest"
	if idx := strings.LastIndex(ref, ":"); idx != -1 {
		// Make sure this is a tag, not a port
		afterColon := ref[idx+1:]
		if !strings.Contains(afterColon, "/") {
			tag = afterColon
			ref = ref[:idx]
		}
	}

	// Split registry and repository
	parts := strings.SplitN(ref, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid reference format: %s", ref)
	}

	return &Reference{
		Registry:   parts[0],
		Repository: parts[1],
		Tag:        tag,
	}, nil
}

// isLayer returns true if the media type represents a layer blob.
func isLayer(mediaType string) bool {
	return strings.Contains(mediaType, "layer") ||
		strings.Contains(mediaType, "blob") ||
		strings.HasPrefix(mediaType, "application/vnd.docker.image.rootfs") ||
		strings.HasPrefix(mediaType, "application/vnd.oci.image.layer")
}
