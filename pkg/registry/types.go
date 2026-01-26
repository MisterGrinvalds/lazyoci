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
	ArtifactTypeUnknown     ArtifactType = "unknown"
)

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
