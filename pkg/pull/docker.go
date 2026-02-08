package pull

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	indexData, err := os.ReadFile(filepath.Join(ociLayoutPath, "index.json"))
	if err != nil {
		return fmt.Errorf("failed to read OCI index: %w", err)
	}
	var index ociIndex
	if err := json.Unmarshal(indexData, &index); err != nil {
		return fmt.Errorf("failed to parse OCI index: %w", err)
	}
	if len(index.Manifests) == 0 {
		return fmt.Errorf("OCI index contains no manifests")
	}

	// Use the first manifest (single-platform images have exactly one).
	manifestDesc := index.Manifests[0]

	// --- 2. Read the image manifest ---
	manifestData, err := readBlob(ociLayoutPath, manifestDesc.Digest)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}
	var manifest ociManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
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
	configDigest := stripPrefix(manifest.Config.Digest)
	configData, err := readBlob(ociLayoutPath, manifest.Config.Digest)
	if err != nil {
		return fmt.Errorf("failed to read config blob: %w", err)
	}
	configName := configDigest + ".json"
	if err := addTarEntry(tw, configName, configData); err != nil {
		return fmt.Errorf("failed to add config to tarball: %w", err)
	}

	// 3b. Add each layer as <sha256>/layer.tar (decompressed if gzipped)
	var layerPaths []string
	for _, layer := range manifest.Layers {
		layerDigest := stripPrefix(layer.Digest)
		layerDir := layerDigest + "/layer.tar"
		layerPaths = append(layerPaths, layerDir)

		layerBlobPath := blobPath(ociLayoutPath, layer.Digest)
		layerFile, err := os.Open(layerBlobPath)
		if err != nil {
			return fmt.Errorf("failed to open layer %s: %w", layer.Digest, err)
		}

		// Determine if the layer is gzipped and get actual size.
		// Docker save format expects uncompressed tar layers.
		var layerReader io.Reader = layerFile
		isGzip := strings.Contains(layer.MediaType, "gzip")

		if isGzip {
			// We need to decompress. Since tar requires the size up front,
			// decompress to a temp file first.
			decompressed, decompSize, err := decompressToTemp(layerFile)
			layerFile.Close()
			if err != nil {
				return fmt.Errorf("failed to decompress layer %s: %w", layer.Digest, err)
			}
			defer os.Remove(decompressed.Name())
			defer decompressed.Close()

			if err := addTarEntryFromReader(tw, layerDir, decompressed, decompSize); err != nil {
				return fmt.Errorf("failed to add layer to tarball: %w", err)
			}
		} else {
			// Uncompressed layer — use file size directly.
			stat, err := layerFile.Stat()
			if err != nil {
				layerFile.Close()
				return fmt.Errorf("failed to stat layer %s: %w", layer.Digest, err)
			}
			layerReader = layerFile
			if err := addTarEntryFromReader(tw, layerDir, layerReader, stat.Size()); err != nil {
				layerFile.Close()
				return fmt.Errorf("failed to add layer to tarball: %w", err)
			}
			layerFile.Close()
		}
	}

	// 3c. Write manifest.json (Docker save format)
	dockerManifest := []dockerSaveManifest{{
		Config:   configName,
		RepoTags: []string{reference},
		Layers:   layerPaths,
	}}
	manifestJSON, err := json.Marshal(dockerManifest)
	if err != nil {
		return fmt.Errorf("failed to marshal Docker manifest: %w", err)
	}
	if err := addTarEntry(tw, "manifest.json", manifestJSON); err != nil {
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

// ---------------------------------------------------------------------------
// OCI / Docker save format types
// ---------------------------------------------------------------------------

type ociIndex struct {
	Manifests []ociDescriptor `json:"manifests"`
}

type ociDescriptor struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
}

type ociManifest struct {
	Config ociDescriptor   `json:"config"`
	Layers []ociDescriptor `json:"layers"`
}

type dockerSaveManifest struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// stripPrefix removes the "sha256:" (or any algo:) prefix from a digest.
func stripPrefix(digest string) string {
	if idx := strings.Index(digest, ":"); idx != -1 {
		return digest[idx+1:]
	}
	return digest
}

// blobPath returns the filesystem path for a blob given its digest.
func blobPath(ociLayoutPath, digest string) string {
	parts := strings.SplitN(digest, ":", 2)
	if len(parts) != 2 {
		return filepath.Join(ociLayoutPath, "blobs", digest)
	}
	return filepath.Join(ociLayoutPath, "blobs", parts[0], parts[1])
}

// readBlob reads a blob from the OCI layout by digest.
func readBlob(ociLayoutPath, digest string) ([]byte, error) {
	return os.ReadFile(blobPath(ociLayoutPath, digest))
}

// addTarEntry adds a file entry with the given name and data to a tar writer.
func addTarEntry(tw *tar.Writer, name string, data []byte) error {
	hdr := &tar.Header{
		Name: name,
		Mode: 0644,
		Size: int64(len(data)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err := tw.Write(data)
	return err
}

// addTarEntryFromReader adds a file entry from a reader with a known size.
func addTarEntryFromReader(tw *tar.Writer, name string, r io.Reader, size int64) error {
	hdr := &tar.Header{
		Name: name,
		Mode: 0644,
		Size: size,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err := io.Copy(tw, r)
	return err
}

// decompressToTemp decompresses a gzip stream to a temporary file,
// returning the file (seeked to start) and its uncompressed size.
func decompressToTemp(r io.Reader) (*os.File, int64, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, 0, err
	}
	defer gz.Close()

	tmp, err := os.CreateTemp("", "lazyoci-layer-*")
	if err != nil {
		return nil, 0, err
	}

	n, err := io.Copy(tmp, gz)
	if err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return nil, 0, err
	}

	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return nil, 0, err
	}

	return tmp, n, nil
}

// IsDockerAvailable checks if Docker daemon is accessible.
func IsDockerAvailable() bool {
	cmd := exec.Command("docker", "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}
