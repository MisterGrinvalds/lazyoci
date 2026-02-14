package ociutil

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ---------------------------------------------------------------------------
// OCI / Docker save format types
// ---------------------------------------------------------------------------

// OCIIndex represents an OCI image index (index.json).
type OCIIndex struct {
	Manifests []OCIDescriptor `json:"manifests"`
}

// OCIDescriptor is a content-addressable descriptor in an OCI layout.
type OCIDescriptor struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
}

// OCIManifest represents an OCI image manifest.
type OCIManifest struct {
	Config OCIDescriptor   `json:"config"`
	Layers []OCIDescriptor `json:"layers"`
}

// DockerSaveManifest represents an entry in Docker save format's manifest.json.
type DockerSaveManifest struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

// ---------------------------------------------------------------------------
// OCI layout helpers
// ---------------------------------------------------------------------------

// StripDigestPrefix removes the "sha256:" (or any algo:) prefix from a digest.
func StripDigestPrefix(digest string) string {
	if idx := strings.Index(digest, ":"); idx != -1 {
		return digest[idx+1:]
	}
	return digest
}

// BlobPath returns the filesystem path for a blob given its digest.
func BlobPath(ociLayoutPath, digest string) string {
	parts := strings.SplitN(digest, ":", 2)
	if len(parts) != 2 {
		return filepath.Join(ociLayoutPath, "blobs", digest)
	}
	return filepath.Join(ociLayoutPath, "blobs", parts[0], parts[1])
}

// ReadBlob reads a blob from the OCI layout by digest.
func ReadBlob(ociLayoutPath, digest string) ([]byte, error) {
	return os.ReadFile(BlobPath(ociLayoutPath, digest))
}

// ReadOCIIndex reads and parses the index.json from an OCI layout.
func ReadOCIIndex(ociLayoutPath string) (*OCIIndex, error) {
	data, err := os.ReadFile(filepath.Join(ociLayoutPath, "index.json"))
	if err != nil {
		return nil, err
	}
	var index OCIIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, err
	}
	return &index, nil
}

// ReadOCIManifest reads and parses a manifest blob from an OCI layout.
func ReadOCIManifest(ociLayoutPath, digest string) (*OCIManifest, error) {
	data, err := ReadBlob(ociLayoutPath, digest)
	if err != nil {
		return nil, err
	}
	var manifest OCIManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

// ---------------------------------------------------------------------------
// Tar helpers
// ---------------------------------------------------------------------------

// AddTarEntry adds a file entry with the given name and data to a tar writer.
func AddTarEntry(tw *tar.Writer, name string, data []byte) error {
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

// AddTarEntryFromReader adds a file entry from a reader with a known size.
func AddTarEntryFromReader(tw *tar.Writer, name string, r io.Reader, size int64) error {
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

// DecompressToTemp decompresses a gzip stream to a temporary file,
// returning the file (seeked to start) and its uncompressed size.
func DecompressToTemp(r io.Reader) (*os.File, int64, error) {
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
