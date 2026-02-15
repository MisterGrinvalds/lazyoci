package build

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"
)

// Helm OCI media types (per the Helm spec).
const (
	helmConfigMediaType = "application/vnd.cncf.helm.config.v1+json"
	helmChartMediaType  = "application/vnd.cncf.helm.chart.content.v1.tar+gzip"
)

// helmChartMeta is the minimal config blob for a Helm OCI artifact.
type helmChartMeta struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
	APIVersion  string `json:"apiVersion,omitempty"`
	AppVersion  string `json:"appVersion,omitempty"`
	Type        string `json:"type,omitempty"`
}

// buildHelm packages a Helm chart directory as an OCI artifact.
// Returns the path to a temporary OCI layout directory.
func (b *Builder) buildHelm(ctx context.Context, artifact *Artifact, chartVersion string) (string, error) {
	chartPath := b.resolvePath(artifact.ChartPath)

	// Verify chart directory exists
	info, err := os.Stat(chartPath)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("chart directory %s does not exist", chartPath)
	}

	b.logf("  Packaging Helm chart from %s...\n", chartPath)

	// Read Chart.yaml for metadata
	chartYamlData, err := os.ReadFile(filepath.Join(chartPath, "Chart.yaml"))
	if err != nil {
		return "", fmt.Errorf("failed to read Chart.yaml: %w", err)
	}

	// Parse minimal chart metadata for the config blob
	meta, err := parseChartMeta(chartYamlData)
	if err != nil {
		return "", fmt.Errorf("failed to parse Chart.yaml: %w", err)
	}

	// Create the chart tarball (.tgz)
	tgzData, err := packageChart(chartPath)
	if err != nil {
		return "", fmt.Errorf("failed to package chart: %w", err)
	}

	// Create config blob (JSON with chart metadata)
	configData, err := json.Marshal(meta)
	if err != nil {
		return "", fmt.Errorf("failed to marshal chart config: %w", err)
	}

	// Build OCI manifest using oras memory store
	store := memory.New()

	// Push config blob
	configDesc, err := pushBlob(ctx, store, helmConfigMediaType, configData)
	if err != nil {
		return "", fmt.Errorf("failed to store config blob: %w", err)
	}

	// Push chart layer
	chartDesc, err := pushBlob(ctx, store, helmChartMediaType, tgzData)
	if err != nil {
		return "", fmt.Errorf("failed to store chart layer: %w", err)
	}

	// Pack manifest
	packOpts := oras.PackManifestOptions{
		Layers:           []ocispec.Descriptor{chartDesc},
		ConfigDescriptor: &configDesc,
		ManifestAnnotations: map[string]string{
			"org.opencontainers.image.title":   meta.Name,
			"org.opencontainers.image.version": meta.Version,
		},
	}

	manifestDesc, err := oras.PackManifest(ctx, store, oras.PackManifestVersion1_1, "", packOpts)
	if err != nil {
		return "", fmt.Errorf("failed to pack manifest: %w", err)
	}

	// Tag the manifest so we can reference it
	if err := store.Tag(ctx, manifestDesc, chartVersion); err != nil {
		return "", fmt.Errorf("failed to tag manifest: %w", err)
	}

	// Export to OCI layout on disk (for push.go to consume)
	tmpDir, err := os.MkdirTemp("", "lazyoci-helm-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	if err := exportToOCILayout(ctx, store, chartVersion, tmpDir); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to export OCI layout: %w", err)
	}

	b.logf("  Chart %s:%s packaged (%d bytes)\n", meta.Name, meta.Version, len(tgzData))

	return tmpDir, nil
}

// parseChartMeta parses Chart.yaml bytes into the minimal metadata struct.
func parseChartMeta(data []byte) (*helmChartMeta, error) {
	// Use a map first since Chart.yaml has more fields than we need
	var raw map[string]interface{}
	if err := decodeYAML(data, &raw); err != nil {
		return nil, err
	}

	meta := &helmChartMeta{}
	if v, ok := raw["name"].(string); ok {
		meta.Name = v
	}
	if v, ok := raw["version"].(string); ok {
		meta.Version = v
	}
	if v, ok := raw["description"].(string); ok {
		meta.Description = v
	}
	if v, ok := raw["apiVersion"].(string); ok {
		meta.APIVersion = v
	}
	if v, ok := raw["appVersion"].(string); ok {
		meta.AppVersion = v
	}
	if v, ok := raw["type"].(string); ok {
		meta.Type = v
	}

	if meta.Name == "" {
		return nil, fmt.Errorf("Chart.yaml missing name")
	}
	if meta.Version == "" {
		return nil, fmt.Errorf("Chart.yaml missing version")
	}

	return meta, nil
}

// decodeYAML is a helper that uses the yaml package already imported in config.go.
// We re-import it here to keep helm.go self-contained.
func decodeYAML(data []byte, v interface{}) error {
	// Use encoding/json-style interface — yaml.v3 supports this via Unmarshal
	return yamlUnmarshal(data, v)
}

// packageChart creates a tar.gz of the chart directory.
// The chart is archived as chartName/ (the base directory name).
func packageChart(chartPath string) ([]byte, error) {
	chartName := filepath.Base(chartPath)

	var buf strings.Builder // use bytes.Buffer via io pipe for large charts
	_ = buf                 // suppress unused

	// Create a pipe to stream tar.gz
	pr, pw := io.Pipe()

	errCh := make(chan error, 1)
	go func() {
		gw := gzip.NewWriter(pw)
		tw := tar.NewWriter(gw)

		err := filepath.Walk(chartPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Get relative path from chartPath's parent
			relPath, err := filepath.Rel(filepath.Dir(chartPath), path)
			if err != nil {
				return err
			}

			// Skip hidden files/directories (except the chart root itself)
			base := filepath.Base(path)
			if base != chartName && strings.HasPrefix(base, ".") {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			header.Name = relPath

			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tw, file)
			return err
		})

		tw.Close()
		gw.Close()
		pw.CloseWithError(err)
		errCh <- err
	}()

	data, readErr := io.ReadAll(pr)
	walkErr := <-errCh

	if walkErr != nil {
		return nil, walkErr
	}
	if readErr != nil {
		return nil, readErr
	}

	return data, nil
}
