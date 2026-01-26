package artifacts

import (
	"github.com/mistergrinvalds/lazyoci/pkg/registry"
)

// Handler defines the interface for artifact type handlers
type Handler interface {
	// CanHandle returns true if this handler can process the artifact
	CanHandle(artifact *registry.Artifact) bool

	// GetDetails returns detailed information about the artifact
	GetDetails(artifact *registry.Artifact) (*Details, error)

	// GetActions returns available actions for the artifact
	GetActions(artifact *registry.Artifact) []Action
}

// Details contains detailed artifact information
type Details struct {
	// Summary is a brief description
	Summary string

	// Properties are key-value pairs of metadata
	Properties map[string]string

	// Components lists sub-components (e.g., layers, files)
	Components []Component

	// RelatedArtifacts lists related artifacts (e.g., signatures, SBOMs)
	RelatedArtifacts []string
}

// Component represents a sub-component of an artifact
type Component struct {
	Name        string
	Type        string
	Size        int64
	Description string
}

// Action represents an available action for an artifact
type Action struct {
	Name        string
	Description string
	Command     string
	Dangerous   bool
}

// ImageHandler handles container image artifacts
type ImageHandler struct{}

// CanHandle returns true for image artifacts
func (h *ImageHandler) CanHandle(artifact *registry.Artifact) bool {
	return artifact.Type == registry.ArtifactTypeImage
}

// GetDetails returns details for an image artifact
func (h *ImageHandler) GetDetails(artifact *registry.Artifact) (*Details, error) {
	details := &Details{
		Summary: "Container Image",
		Properties: map[string]string{
			"Tag":      artifact.Tag,
			"Digest":   artifact.Digest,
			"Platform": artifact.Platform,
		},
	}

	for _, layer := range artifact.Layers {
		details.Components = append(details.Components, Component{
			Name: layer.Digest[:12],
			Type: "layer",
			Size: layer.Size,
		})
	}

	return details, nil
}

// GetActions returns available actions for an image
func (h *ImageHandler) GetActions(artifact *registry.Artifact) []Action {
	return []Action{
		{
			Name:        "Pull",
			Description: "Pull the image to local Docker",
			Command:     "docker pull " + artifact.Repository + ":" + artifact.Tag,
		},
		{
			Name:        "Inspect",
			Description: "Inspect the image manifest",
			Command:     "docker manifest inspect " + artifact.Repository + ":" + artifact.Tag,
		},
		{
			Name:        "Copy Digest",
			Description: "Copy the image digest to clipboard",
		},
	}
}

// HelmHandler handles Helm chart artifacts
type HelmHandler struct{}

// CanHandle returns true for Helm chart artifacts
func (h *HelmHandler) CanHandle(artifact *registry.Artifact) bool {
	return artifact.Type == registry.ArtifactTypeHelmChart
}

// GetDetails returns details for a Helm chart artifact
func (h *HelmHandler) GetDetails(artifact *registry.Artifact) (*Details, error) {
	details := &Details{
		Summary: "Helm Chart",
		Properties: map[string]string{
			"Version": artifact.Tag,
			"Digest":  artifact.Digest,
		},
	}

	return details, nil
}

// GetActions returns available actions for a Helm chart
func (h *HelmHandler) GetActions(artifact *registry.Artifact) []Action {
	return []Action{
		{
			Name:        "Pull",
			Description: "Pull the Helm chart",
			Command:     "helm pull oci://" + artifact.Repository + " --version " + artifact.Tag,
		},
		{
			Name:        "Show Values",
			Description: "Show default values",
			Command:     "helm show values oci://" + artifact.Repository + " --version " + artifact.Tag,
		},
		{
			Name:        "Template",
			Description: "Render chart templates",
			Command:     "helm template oci://" + artifact.Repository + " --version " + artifact.Tag,
		},
	}
}

// GetHandler returns the appropriate handler for an artifact
func GetHandler(artifact *registry.Artifact) Handler {
	handlers := []Handler{
		&HelmHandler{},
		&ImageHandler{}, // Default fallback
	}

	for _, h := range handlers {
		if h.CanHandle(artifact) {
			return h
		}
	}

	return &ImageHandler{}
}
