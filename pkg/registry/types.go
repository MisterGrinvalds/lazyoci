package registry

import "time"

// ArtifactType represents the type of OCI artifact
type ArtifactType string

const (
	ArtifactTypeImage       ArtifactType = "image"
	ArtifactTypeHelmChart   ArtifactType = "helm"
	ArtifactTypeSBOM        ArtifactType = "sbom"
	ArtifactTypeSignature   ArtifactType = "signature"
	ArtifactTypeAttestation ArtifactType = "attestation"
	ArtifactTypeWasm        ArtifactType = "wasm"
	ArtifactTypeUnknown     ArtifactType = "unknown"
)

// ArtifactInfo contains detailed information about an artifact's type.
// This is resolved lazily when a tag is selected in the TUI.
type ArtifactInfo struct {
	// Type is the detected artifact type
	Type ArtifactType `json:"type" yaml:"type"`

	// MediaType is the manifest media type
	MediaType string `json:"mediaType" yaml:"mediaType"`

	// ConfigMediaType is the config descriptor media type (provides additional type hints)
	ConfigMediaType string `json:"configMediaType,omitempty" yaml:"configMediaType,omitempty"`

	// Digest is the manifest digest
	Digest string `json:"digest" yaml:"digest"`

	// Size is the total artifact size in bytes
	Size int64 `json:"size" yaml:"size"`

	// Platform is the target platform (for images)
	Platform string `json:"platform,omitempty" yaml:"platform,omitempty"`

	// Layers is the number of layers
	Layers int `json:"layers" yaml:"layers"`

	// TypeDetail provides additional context (e.g., "spdx" vs "cyclonedx" for sbom)
	TypeDetail string `json:"typeDetail,omitempty" yaml:"typeDetail,omitempty"`

	// Annotations from the manifest
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

// String returns a human-readable type description
func (t ArtifactType) String() string {
	switch t {
	case ArtifactTypeImage:
		return "Container Image"
	case ArtifactTypeHelmChart:
		return "Helm Chart"
	case ArtifactTypeSBOM:
		return "SBOM"
	case ArtifactTypeSignature:
		return "Signature"
	case ArtifactTypeAttestation:
		return "Attestation"
	case ArtifactTypeWasm:
		return "WebAssembly"
	case ArtifactTypeUnknown:
		return "Unknown"
	default:
		return string(t)
	}
}

// Short returns a short type label for display in tables
func (t ArtifactType) Short() string {
	switch t {
	case ArtifactTypeImage:
		return "image"
	case ArtifactTypeHelmChart:
		return "helm"
	case ArtifactTypeSBOM:
		return "sbom"
	case ArtifactTypeSignature:
		return "sig"
	case ArtifactTypeAttestation:
		return "att"
	case ArtifactTypeWasm:
		return "wasm"
	case ArtifactTypeUnknown:
		return "?"
	default:
		return string(t)
	}
}

// Artifact represents an OCI artifact (image, helm chart, etc.)
type Artifact struct {
	// Repository is the full repository path (e.g., docker.io/library/nginx)
	Repository string

	// Tag is the artifact tag (e.g., latest, v1.0.0)
	Tag string

	// Digest is the content-addressable digest
	Digest string

	// Size is the total size in bytes
	Size int64

	// Type is the artifact type
	Type ArtifactType

	// MediaType is the OCI media type
	MediaType string

	// Platform is the target platform (e.g., linux/amd64)
	Platform string

	// Created is the creation timestamp
	Created time.Time

	// Labels are the artifact labels/annotations
	Labels map[string]string

	// Layers contains information about each layer
	Layers []Layer
}

// Layer represents a layer in an OCI artifact
type Layer struct {
	Digest    string
	Size      int64
	MediaType string
}

// Manifest represents an OCI manifest
type Manifest struct {
	SchemaVersion int
	MediaType     string
	Config        Descriptor
	Layers        []Descriptor
	Annotations   map[string]string
}

// Descriptor represents an OCI content descriptor
type Descriptor struct {
	MediaType   string
	Digest      string
	Size        int64
	Annotations map[string]string
	Platform    *Platform
}

// Platform represents a platform specification
type Platform struct {
	Architecture string
	OS           string
	Variant      string
}

// String returns a string representation of the platform
func (p *Platform) String() string {
	if p == nil {
		return ""
	}
	s := p.OS + "/" + p.Architecture
	if p.Variant != "" {
		s += "/" + p.Variant
	}
	return s
}
