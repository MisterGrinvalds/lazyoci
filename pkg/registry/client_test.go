package registry

import (
	"testing"
)

func TestDetectArtifactType(t *testing.T) {
	tests := []struct {
		name       string
		manifest   string
		config     string
		layers     []string
		wantType   ArtifactType
		wantDetail string
	}{
		// --- Helm Charts ---
		{
			name:     "helm chart via config media type",
			manifest: "application/vnd.oci.image.manifest.v1+json",
			config:   "application/vnd.cncf.helm.config.v1+json",
			wantType: ArtifactTypeHelmChart,
		},
		{
			name:     "helm chart via layer media type",
			manifest: "application/vnd.oci.image.manifest.v1+json",
			layers:   []string{"application/vnd.cncf.helm.chart.content.v1.tar+gzip"},
			wantType: ArtifactTypeHelmChart,
		},

		// --- SBOM ---
		{
			name:       "sbom spdx via layer",
			manifest:   "application/vnd.oci.image.manifest.v1+json",
			layers:     []string{"application/spdx+json"},
			wantType:   ArtifactTypeSBOM,
			wantDetail: "spdx",
		},
		{
			name:       "sbom cyclonedx via layer",
			manifest:   "application/vnd.oci.image.manifest.v1+json",
			layers:     []string{"application/vnd.cyclonedx+json"},
			wantType:   ArtifactTypeSBOM,
			wantDetail: "cyclonedx",
		},
		{
			name:     "sbom generic via config",
			manifest: "application/vnd.oci.image.manifest.v1+json",
			config:   "application/vnd.example.sbom.v1+json",
			wantType: ArtifactTypeSBOM,
		},

		// --- Signatures ---
		{
			name:       "cosign signature",
			manifest:   "application/vnd.oci.image.manifest.v1+json",
			layers:     []string{"application/vnd.dev.cosign.simplesigning.v1+json"},
			wantType:   ArtifactTypeSignature,
			wantDetail: "cosign",
		},
		{
			name:       "notary signature",
			manifest:   "application/vnd.oci.image.manifest.v1+json",
			config:     "application/vnd.cncf.notary.signature",
			wantType:   ArtifactTypeSignature,
			wantDetail: "notary",
		},
		{
			name:     "generic signature",
			manifest: "application/vnd.oci.image.manifest.v1+json",
			layers:   []string{"application/vnd.example.signature.v1+json"},
			wantType: ArtifactTypeSignature,
		},

		// --- Attestations ---
		{
			name:       "in-toto attestation",
			manifest:   "application/vnd.oci.image.manifest.v1+json",
			layers:     []string{"application/vnd.in-toto+json"},
			wantType:   ArtifactTypeAttestation,
			wantDetail: "in-toto",
		},
		{
			name:       "dsse attestation",
			manifest:   "application/vnd.oci.image.manifest.v1+json",
			layers:     []string{"application/vnd.dsse.envelope.v1+json"},
			wantType:   ArtifactTypeAttestation,
			wantDetail: "dsse",
		},
		{
			name:     "generic attestation",
			manifest: "application/vnd.oci.image.manifest.v1+json",
			config:   "application/vnd.example.attestation.v1+json",
			wantType: ArtifactTypeAttestation,
		},

		// --- WebAssembly ---
		{
			name:     "wasm via layer",
			manifest: "application/vnd.oci.image.manifest.v1+json",
			layers:   []string{"application/vnd.wasm.content.layer.v1+wasm"},
			wantType: ArtifactTypeWasm,
		},
		{
			name:     "wasm via config",
			manifest: "application/vnd.oci.image.manifest.v1+json",
			config:   "application/wasm",
			wantType: ArtifactTypeWasm,
		},

		// --- Container Images ---
		{
			name:     "oci image manifest",
			manifest: "application/vnd.oci.image.manifest.v1+json",
			config:   "application/vnd.oci.image.config.v1+json",
			layers:   []string{"application/vnd.oci.image.layer.v1.tar+gzip"},
			wantType: ArtifactTypeImage,
		},
		{
			name:     "docker manifest",
			manifest: "application/vnd.docker.distribution.manifest.v2+json",
			config:   "application/vnd.docker.container.image.v1+json",
			wantType: ArtifactTypeImage,
		},
		{
			name:     "image via config only",
			manifest: "application/vnd.oci.image.manifest.v1+json",
			config:   "application/vnd.docker.container.image.v1+json",
			wantType: ArtifactTypeImage,
		},

		// --- Unknown ---
		{
			name:     "completely unknown types",
			manifest: "application/octet-stream",
			config:   "application/octet-stream",
			wantType: ArtifactTypeUnknown,
		},

		// --- Case insensitivity ---
		{
			name:     "case insensitive helm detection",
			manifest: "application/vnd.oci.image.manifest.v1+json",
			config:   "Application/Vnd.CNCF.HELM.Config.V1+json",
			wantType: ArtifactTypeHelmChart,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotDetail := detectArtifactType(tt.manifest, tt.config, tt.layers)
			if gotType != tt.wantType {
				t.Errorf("detectArtifactType() type = %q, want %q", gotType, tt.wantType)
			}
			if gotDetail != tt.wantDetail {
				t.Errorf("detectArtifactType() detail = %q, want %q", gotDetail, tt.wantDetail)
			}
		})
	}
}

func TestSortTags(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want []string
	}{
		{
			name: "latest comes first",
			tags: []string{"v1.0.0", "latest", "v2.0.0"},
			want: []string{"latest", "v2.0.0", "v1.0.0"},
		},
		{
			name: "semver descending",
			tags: []string{"v1.0.0", "v1.2.0", "v1.1.0", "v2.0.0"},
			want: []string{"v2.0.0", "v1.2.0", "v1.1.0", "v1.0.0"},
		},
		{
			name: "semver before non-semver",
			tags: []string{"alpine", "v1.0.0", "slim"},
			want: []string{"v1.0.0", "slim", "alpine"},
		},
		{
			name: "non-semver alphabetical descending",
			tags: []string{"alpine", "bullseye", "slim"},
			want: []string{"slim", "bullseye", "alpine"},
		},
		{
			name: "patch version ordering",
			tags: []string{"1.0.1", "1.0.0", "1.0.10", "1.0.2"},
			want: []string{"1.0.10", "1.0.2", "1.0.1", "1.0.0"},
		},
		{
			name: "stable before prerelease",
			tags: []string{"v1.0.0-rc1", "v1.0.0"},
			want: []string{"v1.0.0", "v1.0.0-rc1"},
		},
		{
			name: "mixed with latest",
			tags: []string{"v3.0.0", "latest", "alpine", "v1.0.0"},
			want: []string{"latest", "v3.0.0", "v1.0.0", "alpine"},
		},
		{
			name: "empty list",
			tags: []string{},
			want: []string{},
		},
		{
			name: "single element",
			tags: []string{"v1.0.0"},
			want: []string{"v1.0.0"},
		},
		{
			name: "major-only versions",
			tags: []string{"2", "1", "3"},
			want: []string{"3", "2", "1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying test input
			tags := make([]string, len(tt.tags))
			copy(tags, tt.tags)

			sortTags(tags)

			if len(tags) != len(tt.want) {
				t.Fatalf("sortTags() length = %d, want %d", len(tags), len(tt.want))
			}
			for i := range tags {
				if tags[i] != tt.want[i] {
					t.Errorf("sortTags()[%d] = %q, want %q (full result: %v)", i, tags[i], tt.want[i], tags)
					break
				}
			}
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		tag        string
		wantOK     bool
		wantMajor  int
		wantMinor  int
		wantPatch  int
		wantPrerel string
	}{
		{"v1.2.3", true, 1, 2, 3, ""},
		{"1.2.3", true, 1, 2, 3, ""},
		{"v1.0.0-rc1", true, 1, 0, 0, "rc1"},
		{"v1.0.0-alpha.1", true, 1, 0, 0, "alpha.1"},
		{"1.2", true, 1, 2, 0, ""},
		{"1", true, 1, 0, 0, ""},
		{"v2", true, 2, 0, 0, ""},
		{"v10.20.30", true, 10, 20, 30, ""},
		{"latest", false, 0, 0, 0, ""},
		{"alpine", false, 0, 0, 0, ""},
		{"bullseye-slim", false, 0, 0, 0, ""},
		{"v1.2.3+build42", true, 1, 2, 3, "build42"},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			v, ok := parseVersion(tt.tag)
			if ok != tt.wantOK {
				t.Fatalf("parseVersion(%q) ok = %v, want %v", tt.tag, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if v.major != tt.wantMajor {
				t.Errorf("major = %d, want %d", v.major, tt.wantMajor)
			}
			if v.minor != tt.wantMinor {
				t.Errorf("minor = %d, want %d", v.minor, tt.wantMinor)
			}
			if v.patch != tt.wantPatch {
				t.Errorf("patch = %d, want %d", v.patch, tt.wantPatch)
			}
			if v.prerelease != tt.wantPrerel {
				t.Errorf("prerelease = %q, want %q", v.prerelease, tt.wantPrerel)
			}
		})
	}
}

func TestFormatPullCount(t *testing.T) {
	tests := []struct {
		count int64
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{999, "999"},
		{1000, "1.0K"},
		{1500, "1.5K"},
		{10000, "10.0K"},
		{999999, "1000.0K"},
		{1000000, "1.0M"},
		{1500000, "1.5M"},
		{1000000000, "1.0B"},
		{2500000000, "2.5B"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatPullCount(tt.count)
			if got != tt.want {
				t.Errorf("FormatPullCount(%d) = %q, want %q", tt.count, got, tt.want)
			}
		})
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		s, substr string
		want      bool
	}{
		{"Hello World", "hello", true},
		{"Hello World", "WORLD", true},
		{"Hello World", "xyz", false},
		{"nginx", "nginx", true},
		{"NGINX", "nginx", true},
		{"", "", true},
		{"abc", "", true},
		{"", "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			got := containsIgnoreCase(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("containsIgnoreCase(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

func TestNormalizeRegistry(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"docker.io", "https://index.docker.io/v1/"},
		{"registry-1.docker.io", "https://index.docker.io/v1/"},
		{"https://docker.io", "https://index.docker.io/v1/"},
		{"http://docker.io", "https://index.docker.io/v1/"},
		{"ghcr.io", "ghcr.io"},
		{"quay.io", "quay.io"},
		{"https://ghcr.io", "ghcr.io"},
		{"http://localhost:5000", "localhost:5000"},
		{"myregistry.example.com", "myregistry.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeRegistry(tt.input)
			if got != tt.want {
				t.Errorf("normalizeRegistry(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDockerConfigGetCredentials(t *testing.T) {
	tests := []struct {
		name     string
		auths    map[string]DockerAuth
		registry string
		wantUser string
		wantPass string
		wantOK   bool
	}{
		{
			name: "direct match with username/password",
			auths: map[string]DockerAuth{
				"ghcr.io": {Username: "user", Password: "pass"},
			},
			registry: "ghcr.io",
			wantUser: "user",
			wantPass: "pass",
			wantOK:   true,
		},
		{
			name: "base64 auth field",
			auths: map[string]DockerAuth{
				"ghcr.io": {Auth: "dXNlcjpwYXNz"}, // base64("user:pass")
			},
			registry: "ghcr.io",
			wantUser: "user",
			wantPass: "pass",
			wantOK:   true,
		},
		{
			name: "docker.io normalizes to index.docker.io",
			auths: map[string]DockerAuth{
				"https://index.docker.io/v1/": {Username: "dockeruser", Password: "dockerpass"},
			},
			registry: "docker.io",
			wantUser: "dockeruser",
			wantPass: "dockerpass",
			wantOK:   true,
		},
		{
			name: "alternative https prefix",
			auths: map[string]DockerAuth{
				"https://myregistry.io": {Username: "u", Password: "p"},
			},
			registry: "myregistry.io",
			wantUser: "u",
			wantPass: "p",
			wantOK:   true,
		},
		{
			name:     "no match",
			auths:    map[string]DockerAuth{},
			registry: "ghcr.io",
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := &DockerConfig{Auths: tt.auths}
			user, pass, ok := dc.GetCredentials(tt.registry)
			if ok != tt.wantOK {
				t.Fatalf("GetCredentials() ok = %v, want %v", ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if user != tt.wantUser {
				t.Errorf("username = %q, want %q", user, tt.wantUser)
			}
			if pass != tt.wantPass {
				t.Errorf("password = %q, want %q", pass, tt.wantPass)
			}
		})
	}
}
