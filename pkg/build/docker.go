package build

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mistergrinvalds/lazyoci/pkg/ociutil"
)

// buildDocker exports an existing Docker daemon image to OCI layout for pushing.
// Returns the path to a temporary OCI layout directory.
func (b *Builder) buildDocker(ctx context.Context, artifact *Artifact) (string, error) {
	image := artifact.Image

	b.logf("  Exporting Docker image %s...\n", image)

	// Verify docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		return "", fmt.Errorf("docker not found in PATH: %w", err)
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "lazyoci-docker-export-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Run docker save to get Docker save format tarball
	saveTar := filepath.Join(tmpDir, "docker-save.tar")
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "docker", "save", "-o", saveTar, image)
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("docker save failed: %s", errMsg)
	}

	// Convert Docker save format to OCI layout
	ociLayoutDir := filepath.Join(tmpDir, "oci-layout")
	if err := dockerSaveToOCILayout(saveTar, ociLayoutDir); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to convert Docker save to OCI layout: %w", err)
	}

	// Clean up the Docker save tarball
	os.Remove(saveTar)

	b.logf("  Docker image %s exported to OCI layout\n", image)

	return ociLayoutDir, nil
}

// dockerSaveToOCILayout converts a Docker save format tarball to an OCI layout directory.
//
// Docker save format:
//
//	manifest.json       — [{Config: "<sha>.json", RepoTags: [...], Layers: ["<sha>/layer.tar", ...]}]
//	<config-sha>.json   — image config
//	<layer-sha>/layer.tar — uncompressed layers
//
// OCI layout:
//
//	oci-layout          — {"imageLayoutVersion": "1.0.0"}
//	index.json          — {manifests: [{digest, mediaType, size}]}
//	blobs/sha256/<hash> — config, layers (gzipped), manifest
func dockerSaveToOCILayout(saveTarPath, destDir string) error {
	// Extract Docker save tarball to a temp directory
	extractDir, err := os.MkdirTemp("", "lazyoci-docker-extract-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(extractDir)

	if err := extractTar(saveTarPath, extractDir); err != nil {
		return fmt.Errorf("failed to extract Docker save tar: %w", err)
	}

	// Read Docker save manifest.json
	manifestData, err := os.ReadFile(filepath.Join(extractDir, "manifest.json"))
	if err != nil {
		return fmt.Errorf("failed to read manifest.json: %w", err)
	}

	var dockerManifests []ociutil.DockerSaveManifest
	if err := json.Unmarshal(manifestData, &dockerManifests); err != nil {
		return fmt.Errorf("failed to parse manifest.json: %w", err)
	}
	if len(dockerManifests) == 0 {
		return fmt.Errorf("Docker save manifest.json is empty")
	}

	dm := dockerManifests[0]

	// Create OCI layout directory structure
	blobDir := filepath.Join(destDir, "blobs", "sha256")
	if err := os.MkdirAll(blobDir, 0755); err != nil {
		return err
	}

	// Write oci-layout file
	ociLayoutContent := []byte(`{"imageLayoutVersion":"1.0.0"}`)
	if err := os.WriteFile(filepath.Join(destDir, "oci-layout"), ociLayoutContent, 0644); err != nil {
		return err
	}

	// Copy config blob
	configData, err := os.ReadFile(filepath.Join(extractDir, dm.Config))
	if err != nil {
		return fmt.Errorf("failed to read config %s: %w", dm.Config, err)
	}
	configDigest := digestSHA256(configData)
	if err := os.WriteFile(filepath.Join(blobDir, configDigest), configData, 0644); err != nil {
		return err
	}

	// Process layers: compress and write to blobs
	var layerDescs []ociutil.OCIDescriptor
	for _, layerPath := range dm.Layers {
		layerFile, err := os.Open(filepath.Join(extractDir, layerPath))
		if err != nil {
			return fmt.Errorf("failed to open layer %s: %w", layerPath, err)
		}

		// Gzip compress the layer
		var compressed bytes.Buffer
		gw := gzip.NewWriter(&compressed)
		if _, err := io.Copy(gw, layerFile); err != nil {
			layerFile.Close()
			return fmt.Errorf("failed to compress layer %s: %w", layerPath, err)
		}
		layerFile.Close()
		gw.Close()

		compressedData := compressed.Bytes()
		layerDigest := digestSHA256(compressedData)

		if err := os.WriteFile(filepath.Join(blobDir, layerDigest), compressedData, 0644); err != nil {
			return err
		}

		layerDescs = append(layerDescs, ociutil.OCIDescriptor{
			MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
			Digest:    "sha256:" + layerDigest,
			Size:      int64(len(compressedData)),
		})
	}

	// Build OCI manifest
	manifest := struct {
		SchemaVersion int                     `json:"schemaVersion"`
		MediaType     string                  `json:"mediaType"`
		Config        ociutil.OCIDescriptor   `json:"config"`
		Layers        []ociutil.OCIDescriptor `json:"layers"`
	}{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		Config: ociutil.OCIDescriptor{
			MediaType: "application/vnd.oci.image.config.v1+json",
			Digest:    "sha256:" + configDigest,
			Size:      int64(len(configData)),
		},
		Layers: layerDescs,
	}

	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	manifestDigest := digestSHA256(manifestJSON)
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
	if err := os.WriteFile(filepath.Join(destDir, "index.json"), indexJSON, 0644); err != nil {
		return err
	}

	return nil
}
