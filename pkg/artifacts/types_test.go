package artifacts

import (
	"testing"

	"github.com/mistergrinvalds/lazyoci/pkg/registry"
)

func TestGetHandler(t *testing.T) {
	tests := []struct {
		name     string
		artifact *registry.Artifact
		wantType string
	}{
		{
			name:     "image artifact gets ImageHandler",
			artifact: &registry.Artifact{Type: registry.ArtifactTypeImage},
			wantType: "*artifacts.ImageHandler",
		},
		{
			name:     "helm artifact gets HelmHandler",
			artifact: &registry.Artifact{Type: registry.ArtifactTypeHelmChart},
			wantType: "*artifacts.HelmHandler",
		},
		{
			name:     "unknown artifact falls back to ImageHandler",
			artifact: &registry.Artifact{Type: registry.ArtifactTypeUnknown},
			wantType: "*artifacts.ImageHandler",
		},
		{
			name:     "sbom artifact falls back to ImageHandler",
			artifact: &registry.Artifact{Type: registry.ArtifactTypeSBOM},
			wantType: "*artifacts.ImageHandler",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := GetHandler(tt.artifact)
			if handler == nil {
				t.Fatal("GetHandler() returned nil")
			}

			// Check handler type via CanHandle behavior
			switch tt.wantType {
			case "*artifacts.HelmHandler":
				if _, ok := handler.(*HelmHandler); !ok {
					t.Errorf("expected HelmHandler, got %T", handler)
				}
			case "*artifacts.ImageHandler":
				if _, ok := handler.(*ImageHandler); !ok {
					t.Errorf("expected ImageHandler, got %T", handler)
				}
			}
		})
	}
}

func TestImageHandlerCanHandle(t *testing.T) {
	h := &ImageHandler{}

	if !h.CanHandle(&registry.Artifact{Type: registry.ArtifactTypeImage}) {
		t.Error("ImageHandler.CanHandle() = false for image artifact")
	}
	if h.CanHandle(&registry.Artifact{Type: registry.ArtifactTypeHelmChart}) {
		t.Error("ImageHandler.CanHandle() = true for helm artifact")
	}
}

func TestHelmHandlerCanHandle(t *testing.T) {
	h := &HelmHandler{}

	if !h.CanHandle(&registry.Artifact{Type: registry.ArtifactTypeHelmChart}) {
		t.Error("HelmHandler.CanHandle() = false for helm artifact")
	}
	if h.CanHandle(&registry.Artifact{Type: registry.ArtifactTypeImage}) {
		t.Error("HelmHandler.CanHandle() = true for image artifact")
	}
}

func TestImageHandlerGetActions(t *testing.T) {
	h := &ImageHandler{}
	artifact := &registry.Artifact{
		Repository: "docker.io/library/nginx",
		Tag:        "1.25",
	}

	actions := h.GetActions(artifact)
	if len(actions) != 3 {
		t.Fatalf("len(actions) = %d, want 3", len(actions))
	}

	// Verify Pull action
	if actions[0].Name != "Pull" {
		t.Errorf("actions[0].Name = %q, want %q", actions[0].Name, "Pull")
	}
	if actions[0].Command != "docker pull docker.io/library/nginx:1.25" {
		t.Errorf("actions[0].Command = %q", actions[0].Command)
	}

	// Verify Inspect action
	if actions[1].Name != "Inspect" {
		t.Errorf("actions[1].Name = %q, want %q", actions[1].Name, "Inspect")
	}

	// Verify Copy Digest action
	if actions[2].Name != "Copy Digest" {
		t.Errorf("actions[2].Name = %q, want %q", actions[2].Name, "Copy Digest")
	}
}

func TestHelmHandlerGetActions(t *testing.T) {
	h := &HelmHandler{}
	artifact := &registry.Artifact{
		Repository: "ghcr.io/helm/charts/mychart",
		Tag:        "0.1.0",
	}

	actions := h.GetActions(artifact)
	if len(actions) != 3 {
		t.Fatalf("len(actions) = %d, want 3", len(actions))
	}

	// Pull
	if actions[0].Name != "Pull" {
		t.Errorf("actions[0].Name = %q, want %q", actions[0].Name, "Pull")
	}
	if actions[0].Command != "helm pull oci://ghcr.io/helm/charts/mychart --version 0.1.0" {
		t.Errorf("actions[0].Command = %q", actions[0].Command)
	}

	// Show Values
	if actions[1].Name != "Show Values" {
		t.Errorf("actions[1].Name = %q, want %q", actions[1].Name, "Show Values")
	}

	// Template
	if actions[2].Name != "Template" {
		t.Errorf("actions[2].Name = %q, want %q", actions[2].Name, "Template")
	}
}

func TestImageHandlerGetDetails(t *testing.T) {
	h := &ImageHandler{}
	artifact := &registry.Artifact{
		Tag:      "latest",
		Digest:   "sha256:abc123",
		Platform: "linux/amd64",
		Layers: []registry.Layer{
			{Digest: "sha256:layer111111", Size: 1024},
			{Digest: "sha256:layer222222", Size: 2048},
		},
	}

	details, err := h.GetDetails(artifact)
	if err != nil {
		t.Fatalf("GetDetails() error = %v", err)
	}

	if details.Summary != "Container Image" {
		t.Errorf("Summary = %q, want %q", details.Summary, "Container Image")
	}

	if details.Properties["Tag"] != "latest" {
		t.Errorf("Properties[Tag] = %q", details.Properties["Tag"])
	}

	if len(details.Components) != 2 {
		t.Fatalf("len(Components) = %d, want 2", len(details.Components))
	}

	// Components should use first 12 chars of digest
	if details.Components[0].Name != "sha256:layer" {
		t.Errorf("Components[0].Name = %q", details.Components[0].Name)
	}
	if details.Components[0].Type != "layer" {
		t.Errorf("Components[0].Type = %q, want %q", details.Components[0].Type, "layer")
	}
	if details.Components[0].Size != 1024 {
		t.Errorf("Components[0].Size = %d, want 1024", details.Components[0].Size)
	}
}

func TestHelmHandlerGetDetails(t *testing.T) {
	h := &HelmHandler{}
	artifact := &registry.Artifact{
		Tag:    "0.1.0",
		Digest: "sha256:helmdigest",
	}

	details, err := h.GetDetails(artifact)
	if err != nil {
		t.Fatalf("GetDetails() error = %v", err)
	}

	if details.Summary != "Helm Chart" {
		t.Errorf("Summary = %q, want %q", details.Summary, "Helm Chart")
	}

	if details.Properties["Version"] != "0.1.0" {
		t.Errorf("Properties[Version] = %q", details.Properties["Version"])
	}
	if details.Properties["Digest"] != "sha256:helmdigest" {
		t.Errorf("Properties[Digest] = %q", details.Properties["Digest"])
	}
}
