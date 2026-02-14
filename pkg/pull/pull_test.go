package pull

import (
	"testing"

	"github.com/mistergrinvalds/lazyoci/pkg/registry"
)

func TestParseReference(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantRegistry string
		wantRepo     string
		wantTag      string
		wantErr      bool
	}{
		{
			name:         "simple image name",
			input:        "nginx",
			wantRegistry: "docker.io",
			wantRepo:     "library/nginx",
			wantTag:      "latest",
		},
		{
			name:         "image with tag",
			input:        "nginx:1.25",
			wantRegistry: "docker.io",
			wantRepo:     "library/nginx",
			wantTag:      "1.25",
		},
		{
			name:         "docker hub user/repo",
			input:        "myuser/myapp:v1.0.0",
			wantRegistry: "docker.io",
			wantRepo:     "myuser/myapp",
			wantTag:      "v1.0.0",
		},
		{
			name:         "docker hub user/repo no tag",
			input:        "myuser/myapp",
			wantRegistry: "docker.io",
			wantRepo:     "myuser/myapp",
			wantTag:      "latest",
		},
		{
			name:         "full reference with registry",
			input:        "ghcr.io/owner/repo:sha-abc123",
			wantRegistry: "ghcr.io",
			wantRepo:     "owner/repo",
			wantTag:      "sha-abc123",
		},
		{
			name:         "quay.io reference",
			input:        "quay.io/prometheus/node-exporter:v1.7.0",
			wantRegistry: "quay.io",
			wantRepo:     "prometheus/node-exporter",
			wantTag:      "v1.7.0",
		},
		{
			name:         "localhost with port",
			input:        "localhost:5000/myapp:v1",
			wantRegistry: "localhost:5000",
			wantRepo:     "myapp",
			wantTag:      "v1",
		},
		{
			name:         "localhost with port no tag",
			input:        "localhost:5050/test/hello",
			wantRegistry: "localhost:5050",
			wantRepo:     "test/hello",
			wantTag:      "latest",
		},
		{
			name:         "registry with port and tag",
			input:        "myregistry.io:5000/org/app:latest",
			wantRegistry: "myregistry.io:5000",
			wantRepo:     "org/app",
			wantTag:      "latest",
		},
		{
			name:         "docker.io explicit",
			input:        "docker.io/library/alpine:3.19",
			wantRegistry: "docker.io",
			wantRepo:     "library/alpine",
			wantTag:      "3.19",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := ParseReference(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseReference(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if ref.Registry != tt.wantRegistry {
				t.Errorf("Registry = %q, want %q", ref.Registry, tt.wantRegistry)
			}
			if ref.Repository != tt.wantRepo {
				t.Errorf("Repository = %q, want %q", ref.Repository, tt.wantRepo)
			}
			if ref.Tag != tt.wantTag {
				t.Errorf("Tag = %q, want %q", ref.Tag, tt.wantTag)
			}
		})
	}
}

func TestDetectArtifactTypeFromMediaTypes(t *testing.T) {
	tests := []struct {
		name       string
		manifest   string
		config     string
		layers     []string
		wantType   registry.ArtifactType
		wantDetail string
	}{
		// Helm
		{
			name:     "helm chart config",
			manifest: "application/vnd.oci.image.manifest.v1+json",
			config:   "application/vnd.cncf.helm.config.v1+json",
			wantType: registry.ArtifactTypeHelmChart,
		},
		{
			name:     "helm chart layer",
			manifest: "application/vnd.oci.image.manifest.v1+json",
			layers:   []string{"application/vnd.cncf.helm.chart.content.v1.tar+gzip"},
			wantType: registry.ArtifactTypeHelmChart,
		},

		// SBOM
		{
			name:       "spdx sbom",
			manifest:   "application/vnd.oci.image.manifest.v1+json",
			layers:     []string{"application/spdx+json"},
			wantType:   registry.ArtifactTypeSBOM,
			wantDetail: "spdx",
		},
		{
			name:       "cyclonedx sbom",
			manifest:   "application/vnd.oci.image.manifest.v1+json",
			layers:     []string{"application/vnd.cyclonedx+json"},
			wantType:   registry.ArtifactTypeSBOM,
			wantDetail: "cyclonedx",
		},

		// Signature
		{
			name:       "cosign signature",
			manifest:   "application/vnd.oci.image.manifest.v1+json",
			layers:     []string{"application/vnd.dev.cosign.simplesigning.v1+json"},
			wantType:   registry.ArtifactTypeSignature,
			wantDetail: "cosign",
		},
		{
			name:       "notary signature",
			manifest:   "application/vnd.oci.image.manifest.v1+json",
			config:     "application/vnd.cncf.notary.signature",
			wantType:   registry.ArtifactTypeSignature,
			wantDetail: "notary",
		},

		// Attestation
		{
			name:       "in-toto attestation",
			manifest:   "application/vnd.oci.image.manifest.v1+json",
			layers:     []string{"application/vnd.in-toto+json"},
			wantType:   registry.ArtifactTypeAttestation,
			wantDetail: "in-toto",
		},
		{
			name:       "dsse attestation",
			manifest:   "application/vnd.oci.image.manifest.v1+json",
			layers:     []string{"application/vnd.dsse.envelope.v1+json"},
			wantType:   registry.ArtifactTypeAttestation,
			wantDetail: "dsse",
		},

		// WASM
		{
			name:     "wasm layer",
			manifest: "application/vnd.oci.image.manifest.v1+json",
			layers:   []string{"application/vnd.wasm.content.layer.v1+wasm"},
			wantType: registry.ArtifactTypeWasm,
		},

		// Image
		{
			name:     "oci image",
			manifest: "application/vnd.oci.image.manifest.v1+json",
			config:   "application/vnd.oci.image.config.v1+json",
			layers:   []string{"application/vnd.oci.image.layer.v1.tar+gzip"},
			wantType: registry.ArtifactTypeImage,
		},
		{
			name:     "docker image",
			manifest: "application/vnd.docker.distribution.manifest.v2+json",
			config:   "application/vnd.docker.container.image.v1+json",
			wantType: registry.ArtifactTypeImage,
		},

		// Unknown
		{
			name:     "unknown types",
			manifest: "application/octet-stream",
			config:   "application/octet-stream",
			wantType: registry.ArtifactTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotDetail := detectArtifactTypeFromMediaTypes(tt.manifest, tt.config, tt.layers)
			if gotType != tt.wantType {
				t.Errorf("type = %q, want %q", gotType, tt.wantType)
			}
			if gotDetail != tt.wantDetail {
				t.Errorf("detail = %q, want %q", gotDetail, tt.wantDetail)
			}
		})
	}
}

func TestDetectTypeFromMediaType(t *testing.T) {
	tests := []struct {
		mediaType string
		want      registry.ArtifactType
	}{
		{"application/vnd.cncf.helm.config.v1+json", registry.ArtifactTypeHelmChart},
		{"application/spdx+json", registry.ArtifactTypeSBOM},
		{"application/vnd.cyclonedx+json", registry.ArtifactTypeSBOM},
		{"application/vnd.dev.cosign.simplesigning.v1+json", registry.ArtifactTypeSignature},
		{"application/vnd.cncf.notary.signature", registry.ArtifactTypeSignature},
		{"application/vnd.in-toto+json", registry.ArtifactTypeAttestation},
		{"application/vnd.dsse.envelope.v1+json", registry.ArtifactTypeAttestation},
		{"application/wasm", registry.ArtifactTypeWasm},
		{"application/vnd.oci.image.manifest.v1+json", registry.ArtifactTypeImage},
		{"application/vnd.docker.distribution.manifest.v2+json", registry.ArtifactTypeImage},
		{"application/octet-stream", registry.ArtifactTypeImage}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.mediaType, func(t *testing.T) {
			got := detectTypeFromMediaType(tt.mediaType)
			if got != tt.want {
				t.Errorf("detectTypeFromMediaType(%q) = %q, want %q", tt.mediaType, got, tt.want)
			}
		})
	}
}

func TestGetTypeDirectory(t *testing.T) {
	tests := []struct {
		input registry.ArtifactType
		want  string
	}{
		{registry.ArtifactTypeImage, "oci"},
		{registry.ArtifactTypeHelmChart, "helm"},
		{registry.ArtifactTypeSBOM, "sbom"},
		{registry.ArtifactTypeSignature, "sig"},
		{registry.ArtifactTypeAttestation, "att"},
		{registry.ArtifactTypeWasm, "wasm"},
		{registry.ArtifactTypeUnknown, "oci"},
		{registry.ArtifactType("custom"), "oci"},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			got := getTypeDirectory(tt.input)
			if got != tt.want {
				t.Errorf("getTypeDirectory(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsLayer(t *testing.T) {
	tests := []struct {
		mediaType string
		want      bool
	}{
		{"application/vnd.oci.image.layer.v1.tar+gzip", true},
		{"application/vnd.docker.image.rootfs.diff.tar.gzip", true},
		{"application/vnd.oci.image.layer.nondistributable.v1.tar+gzip", true},
		{"application/vnd.cncf.helm.chart.content.v1.tar+gzip", false},
		{"application/vnd.oci.image.manifest.v1+json", false},
		{"application/vnd.oci.image.config.v1+json", false},
		{"application/octet-stream", false},
		// blob media types
		{"application/vnd.oci.image.layer.v1.tar+zstd", true},
	}

	for _, tt := range tests {
		t.Run(tt.mediaType, func(t *testing.T) {
			got := isLayer(tt.mediaType)
			if got != tt.want {
				t.Errorf("isLayer(%q) = %v, want %v", tt.mediaType, got, tt.want)
			}
		})
	}
}
