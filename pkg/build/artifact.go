package build

import (
	"context"
	"fmt"
	"os"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"
)

// buildArtifact packages generic files as an OCI artifact.
// Returns the path to a temporary OCI layout directory.
func (b *Builder) buildArtifact(ctx context.Context, artifact *Artifact) (string, error) {
	b.logf("  Packaging generic artifact (%d files)...\n", len(artifact.Files))

	store := memory.New()

	// Push each file as a layer
	var layers []ocispec.Descriptor
	for _, f := range artifact.Files {
		filePath := b.resolvePath(f.Path)
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", f.Path, err)
		}

		desc, err := pushBlob(ctx, store, f.MediaType, data)
		if err != nil {
			return "", fmt.Errorf("failed to store file %s: %w", f.Path, err)
		}

		// Annotate with the original filename
		desc.Annotations = map[string]string{
			ocispec.AnnotationTitle: f.Path,
		}

		layers = append(layers, desc)
		b.logf("    Added %s (%s, %d bytes)\n", f.Path, f.MediaType, len(data))
	}

	// Determine artifact type annotation
	artifactType := artifact.MediaType
	if artifactType == "" {
		artifactType = "application/vnd.unknown.artifact.v1"
	}

	// Pack manifest
	packOpts := oras.PackManifestOptions{
		Layers: layers,
	}

	manifestDesc, err := oras.PackManifest(ctx, store, oras.PackManifestVersion1_1, artifactType, packOpts)
	if err != nil {
		return "", fmt.Errorf("failed to pack manifest: %w", err)
	}

	// Tag the manifest
	tag := "latest"
	if err := store.Tag(ctx, manifestDesc, tag); err != nil {
		return "", fmt.Errorf("failed to tag manifest: %w", err)
	}

	// Export to OCI layout on disk
	tmpDir, err := os.MkdirTemp("", "lazyoci-artifact-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	if err := exportToOCILayout(ctx, store, tag, tmpDir); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to export OCI layout: %w", err)
	}

	b.logf("  Artifact packaged (%d layers)\n", len(layers))

	return tmpDir, nil
}
