package registry

import "testing"

func TestArtifactTypeString(t *testing.T) {
	tests := []struct {
		input ArtifactType
		want  string
	}{
		{ArtifactTypeImage, "Container Image"},
		{ArtifactTypeHelmChart, "Helm Chart"},
		{ArtifactTypeSBOM, "SBOM"},
		{ArtifactTypeSignature, "Signature"},
		{ArtifactTypeAttestation, "Attestation"},
		{ArtifactTypeWasm, "WebAssembly"},
		{ArtifactTypeUnknown, "Unknown"},
		{ArtifactType("custom"), "custom"},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			got := tt.input.String()
			if got != tt.want {
				t.Errorf("ArtifactType(%q).String() = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestArtifactTypeShort(t *testing.T) {
	tests := []struct {
		input ArtifactType
		want  string
	}{
		{ArtifactTypeImage, "image"},
		{ArtifactTypeHelmChart, "helm"},
		{ArtifactTypeSBOM, "sbom"},
		{ArtifactTypeSignature, "sig"},
		{ArtifactTypeAttestation, "att"},
		{ArtifactTypeWasm, "wasm"},
		{ArtifactTypeUnknown, "?"},
		{ArtifactType("custom"), "custom"},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			got := tt.input.Short()
			if got != tt.want {
				t.Errorf("ArtifactType(%q).Short() = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPlatformString(t *testing.T) {
	tests := []struct {
		name  string
		input *Platform
		want  string
	}{
		{
			name:  "nil platform",
			input: nil,
			want:  "",
		},
		{
			name:  "linux/amd64",
			input: &Platform{OS: "linux", Architecture: "amd64"},
			want:  "linux/amd64",
		},
		{
			name:  "linux/arm/v7",
			input: &Platform{OS: "linux", Architecture: "arm", Variant: "v7"},
			want:  "linux/arm/v7",
		},
		{
			name:  "windows/amd64",
			input: &Platform{OS: "windows", Architecture: "amd64"},
			want:  "windows/amd64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.String()
			if got != tt.want {
				t.Errorf("Platform.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
