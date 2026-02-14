package pull

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mistergrinvalds/lazyoci/pkg/ociutil"
)

// LoadToDocker loads an OCI layout into the local Docker daemon.
// It creates a tarball from the OCI layout and uses `docker load` to import it.
func LoadToDocker(ociLayoutPath, reference string) error {
	// Check if docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker not found in PATH: %w", err)
	}

	// Create a temporary tar file
	tmpDir, err := os.MkdirTemp("", "lazyoci-docker-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tarPath := filepath.Join(tmpDir, "image.tar")

	// Try skopeo first — it handles OCI layout to Docker daemon natively.
	if skopeoPath, err := exec.LookPath("skopeo"); err == nil {
		return loadWithSkopeo(skopeoPath, ociLayoutPath, reference)
	}

	// Fall back to manual OCI-to-Docker-save conversion.
	return loadManual(ociLayoutPath, tarPath, reference)
}

// loadWithSkopeo uses skopeo to copy from OCI layout to Docker daemon
func loadWithSkopeo(skopeoPath, ociLayoutPath, reference string) error {
	// skopeo copy oci:./path:tag docker-daemon:reference
	cmd := exec.Command(skopeoPath, "copy",
		fmt.Sprintf("oci:%s", ociLayoutPath),
		fmt.Sprintf("docker-daemon:%s", reference),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return fmt.Errorf("skopeo copy failed: %s", errMsg)
	}
	return nil
}

// loadManual converts an OCI layout to a Docker save format tarball and loads it.
//
// docker load expects the Docker save format:
//
//	manifest.json            — [{Config: "<sha>.json", RepoTags: [...], Layers: ["<sha>/layer.tar", ...]}]
//	<config-sha>.json        — image config blob
//	<layer-sha>/layer.tar    — each layer (decompressed)
func loadManual(ociLayoutPath, tarPath, reference string) error {
	// --- 1. Read the OCI index to find the manifest descriptor ---
	index, err := ociutil.ReadOCIIndex(ociLayoutPath)
	if err != nil {
		return fmt.Errorf("failed to read OCI index: %w", err)
	}
	if len(index.Manifests) == 0 {
		return fmt.Errorf("OCI index contains no manifests")
	}

	// Use the first manifest (single-platform images have exactly one).
	manifestDesc := index.Manifests[0]

	// --- 2. Read the image manifest ---
	manifest, err := ociutil.ReadOCIManifest(ociLayoutPath, manifestDesc.Digest)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	// --- 3. Build the Docker save format tarball ---
	tarFile, err := os.Create(tarPath)
	if err != nil {
		return fmt.Errorf("failed to create tarball: %w", err)
	}
	defer tarFile.Close()

	tw := tar.NewWriter(tarFile)
	defer tw.Close()

	// 3a. Add config blob as <sha256>.json
	configDigest := ociutil.StripDigestPrefix(manifest.Config.Digest)
	configData, err := ociutil.ReadBlob(ociLayoutPath, manifest.Config.Digest)
	if err != nil {
		return fmt.Errorf("failed to read config blob: %w", err)
	}
	configName := configDigest + ".json"
	if err := ociutil.AddTarEntry(tw, configName, configData); err != nil {
		return fmt.Errorf("failed to add config to tarball: %w", err)
	}

	// 3b. Add each layer as <sha256>/layer.tar (decompressed if gzipped)
	var layerPaths []string
	for _, layer := range manifest.Layers {
		layerDigest := ociutil.StripDigestPrefix(layer.Digest)
		layerDir := layerDigest + "/layer.tar"
		layerPaths = append(layerPaths, layerDir)

		layerBlobPath := ociutil.BlobPath(ociLayoutPath, layer.Digest)
		layerFile, err := os.Open(layerBlobPath)
		if err != nil {
			return fmt.Errorf("failed to open layer %s: %w", layer.Digest, err)
		}

		// Determine if the layer is gzipped and get actual size.
		// Docker save format expects uncompressed tar layers.
		isGzip := strings.Contains(layer.MediaType, "gzip")

		if isGzip {
			// We need to decompress. Since tar requires the size up front,
			// decompress to a temp file first.
			decompressed, decompSize, err := ociutil.DecompressToTemp(layerFile)
			layerFile.Close()
			if err != nil {
				return fmt.Errorf("failed to decompress layer %s: %w", layer.Digest, err)
			}
			defer os.Remove(decompressed.Name())
			defer decompressed.Close()

			if err := ociutil.AddTarEntryFromReader(tw, layerDir, decompressed, decompSize); err != nil {
				return fmt.Errorf("failed to add layer to tarball: %w", err)
			}
		} else {
			// Uncompressed layer — use file size directly.
			stat, err := layerFile.Stat()
			if err != nil {
				layerFile.Close()
				return fmt.Errorf("failed to stat layer %s: %w", layer.Digest, err)
			}
			if err := ociutil.AddTarEntryFromReader(tw, layerDir, layerFile, stat.Size()); err != nil {
				layerFile.Close()
				return fmt.Errorf("failed to add layer to tarball: %w", err)
			}
			layerFile.Close()
		}
	}

	// 3c. Write manifest.json (Docker save format)
	dockerManifest := []ociutil.DockerSaveManifest{{
		Config:   configName,
		RepoTags: []string{reference},
		Layers:   layerPaths,
	}}
	manifestJSON, err := json.Marshal(dockerManifest)
	if err != nil {
		return fmt.Errorf("failed to marshal Docker manifest: %w", err)
	}
	if err := ociutil.AddTarEntry(tw, "manifest.json", manifestJSON); err != nil {
		return fmt.Errorf("failed to add manifest.json to tarball: %w", err)
	}

	// Close the tar writer before loading.
	if err := tw.Close(); err != nil {
		return fmt.Errorf("failed to finalize tarball: %w", err)
	}
	if err := tarFile.Close(); err != nil {
		return fmt.Errorf("failed to close tarball: %w", err)
	}

	// --- 4. Load into Docker ---
	var stdout, stderr bytes.Buffer
	loadCmd := exec.Command("docker", "load", "-i", tarPath)
	loadCmd.Stdout = &stdout
	loadCmd.Stderr = &stderr

	if err := loadCmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return fmt.Errorf("docker load failed: %s", errMsg)
	}

	return nil
}

// IsDockerAvailable checks if Docker daemon is accessible.
func IsDockerAvailable() bool {
	cmd := exec.Command("docker", "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}
