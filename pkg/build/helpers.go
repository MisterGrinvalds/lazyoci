package build

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mistergrinvalds/lazyoci/pkg/ociutil"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"gopkg.in/yaml.v3"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/oci"
)

// pushBlob pushes a blob to a content storage and returns its descriptor.
func pushBlob(ctx context.Context, store content.Storage, mediaType string, data []byte) (ocispec.Descriptor, error) {
	desc := content.NewDescriptorFromBytes(mediaType, data)
	if err := store.Push(ctx, desc, newBlobReader(data)); err != nil {
		return ocispec.Descriptor{}, err
	}
	return desc, nil
}

// newBlobReader wraps a byte slice as an io.Reader for content.Storage.Push.
func newBlobReader(data []byte) *blobReader {
	return &blobReader{data: data, pos: 0}
}

type blobReader struct {
	data []byte
	pos  int
}

func (r *blobReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// exportToOCILayout copies a manifest and its content from a memory store to an OCI layout on disk.
// The srcTag is the tag assigned to the manifest in the source store — oras memory
// store requires resolving by tag, not by raw digest string.
func exportToOCILayout(ctx context.Context, src content.ReadOnlyStorage, srcTag, destDir string) error {
	store, err := oci.New(destDir)
	if err != nil {
		return fmt.Errorf("failed to create OCI store at %s: %w", destDir, err)
	}

	// We need a full source that implements oras.ReadOnlyTarget.
	// The memory store implements this.
	srcTarget, ok := src.(oras.ReadOnlyTarget)
	if !ok {
		return fmt.Errorf("source storage does not implement ReadOnlyTarget")
	}

	// Copy the manifest and all referenced blobs using the tag reference.
	_, err = oras.Copy(ctx, srcTarget, srcTag, store, srcTag, oras.CopyOptions{})
	if err != nil {
		return fmt.Errorf("failed to copy to OCI layout: %w", err)
	}

	return nil
}

// digestSHA256 computes the SHA-256 hex digest of data.
func digestSHA256(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}

// yamlUnmarshal is a package-level wrapper for yaml.Unmarshal.
func yamlUnmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}

// writeOCILayout creates a minimal OCI layout with a single manifest.
// This is used when we already have blobs on disk and just need the metadata files.
func writeOCILayout(destDir string, manifestJSON []byte, manifestDigest string) error {
	blobDir := filepath.Join(destDir, "blobs", "sha256")
	if err := os.MkdirAll(blobDir, 0755); err != nil {
		return err
	}

	// Write oci-layout
	if err := os.WriteFile(filepath.Join(destDir, "oci-layout"), []byte(`{"imageLayoutVersion":"1.0.0"}`), 0644); err != nil {
		return err
	}

	// Write manifest blob
	if err := os.WriteFile(filepath.Join(blobDir, manifestDigest), manifestJSON, 0644); err != nil {
		return err
	}

	// Write index.json
	index := ociutil.OCIIndex{
		Manifests: []ociutil.OCIDescriptor{
			{
				MediaType: "application/vnd.oci.image.manifest.v1+json",
				Digest:    "sha256:" + manifestDigest,
				Size:      int64(len(manifestJSON)),
			},
		},
	}
	indexJSON, err := json.Marshal(index)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(destDir, "index.json"), indexJSON, 0644)
}
