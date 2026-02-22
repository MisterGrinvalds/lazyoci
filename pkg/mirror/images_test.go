package mirror

import (
	"testing"
)

func TestImageLineRE(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string // expected capture group 1, empty if no match
	}{
		{
			name: "standard indented",
			line: "          image: nginx:alpine",
			want: "nginx:alpine",
		},
		{
			name: "quoted value",
			line: `          image: "ghcr.io/foo/bar:v1"`,
			want: `"ghcr.io/foo/bar:v1"`,
		},
		{
			name: "single quoted",
			line: `          image: 'busybox:1.36'`,
			want: `'busybox:1.36'`,
		},
		{
			name: "no indent",
			line: "image: registry.k8s.io/ingress-nginx/controller:v1.11.2",
			want: "registry.k8s.io/ingress-nginx/controller:v1.11.2",
		},
		{
			name: "tab indented",
			line: "\t\timage: docker.io/library/redis:7.2",
			want: "docker.io/library/redis:7.2",
		},
		{
			name: "no match - different key",
			line: "          name: mycontainer",
			want: "",
		},
		{
			name: "no match - comment",
			line: "  # image: nginx:latest",
			want: "",
		},
		{
			name: "with digest",
			line: "          image: ghcr.io/foo/bar@sha256:abcdef1234567890",
			want: "ghcr.io/foo/bar@sha256:abcdef1234567890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := imageLineRE.FindStringSubmatch(tt.line)
			if tt.want == "" {
				if len(matches) > 0 {
					t.Errorf("expected no match, got %v", matches)
				}
				return
			}
			if len(matches) < 2 {
				t.Fatalf("expected match, got none for line %q", tt.line)
			}
			got := matches[1]
			if got != tt.want {
				t.Errorf("captured %q, want %q", got, tt.want)
			}
		})
	}
}
