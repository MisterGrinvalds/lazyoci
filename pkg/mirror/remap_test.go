package mirror

import "testing"

func TestRemapImage(t *testing.T) {
	target := "registry.digitalocean.com/greenforests"

	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "ghcr with tag",
			src:  "ghcr.io/kyverno/kyverno:v1.13.2",
			want: "registry.digitalocean.com/greenforests/kyverno/kyverno:v1.13.2",
		},
		{
			name: "docker hub library image",
			src:  "docker.io/library/redis:7.2.4-alpine",
			want: "registry.digitalocean.com/greenforests/library/redis:7.2.4-alpine",
		},
		{
			name: "docker hub user image",
			src:  "docker.io/grafana/grafana:11.3.0",
			want: "registry.digitalocean.com/greenforests/grafana/grafana:11.3.0",
		},
		{
			name: "quay image",
			src:  "quay.io/jetstack/cert-manager-controller:v1.16.1",
			want: "registry.digitalocean.com/greenforests/jetstack/cert-manager-controller:v1.16.1",
		},
		{
			name: "registry.k8s.io nested path",
			src:  "registry.k8s.io/ingress-nginx/controller:v1.11.2",
			want: "registry.digitalocean.com/greenforests/ingress-nginx/controller:v1.11.2",
		},
		{
			name: "digest reference",
			src:  "registry.k8s.io/ingress-nginx/controller:v1.11.2@sha256:abcdef1234",
			want: "registry.digitalocean.com/greenforests/ingress-nginx/controller:v1.11.2@sha256:abcdef1234",
		},
		{
			name: "no tag defaults to latest",
			src:  "ghcr.io/foo/bar",
			want: "registry.digitalocean.com/greenforests/foo/bar:latest",
		},
		{
			name: "public ECR",
			src:  "public.ecr.aws/docker/library/redis:7.2.4-alpine",
			want: "registry.digitalocean.com/greenforests/docker/library/redis:7.2.4-alpine",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemapImage(tt.src, target)
			if got != tt.want {
				t.Errorf("RemapImage(%q, %q)\n  got  %q\n  want %q", tt.src, target, got, tt.want)
			}
		})
	}
}

func TestNormalizeImage(t *testing.T) {
	tests := []struct {
		name string
		img  string
		want string
	}{
		{
			name: "bare image",
			img:  "nginx:alpine",
			want: "docker.io/library/nginx:alpine",
		},
		{
			name: "user image",
			img:  "stakater/reloader:v1.0",
			want: "docker.io/stakater/reloader:v1.0",
		},
		{
			name: "already qualified",
			img:  "ghcr.io/foo/bar:v1",
			want: "ghcr.io/foo/bar:v1",
		},
		{
			name: "quoted image",
			img:  `"nginx:alpine"`,
			want: "docker.io/library/nginx:alpine",
		},
		{
			name: "template placeholder",
			img:  "{{ .Values.image }}",
			want: "",
		},
		{
			name: "empty string",
			img:  "",
			want: "",
		},
		{
			name: "localhost registry",
			img:  "localhost:5000/myapp:v1",
			want: "localhost:5000/myapp:v1",
		},
		{
			name: "registry with port",
			img:  "registry.example.com:5000/org/app:v1",
			want: "registry.example.com:5000/org/app:v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeImage(tt.img)
			if got != tt.want {
				t.Errorf("NormalizeImage(%q)\n  got  %q\n  want %q", tt.img, got, tt.want)
			}
		})
	}
}

func TestSourceRegistryHost(t *testing.T) {
	tests := []struct {
		ref  string
		want string
	}{
		{"ghcr.io/kyverno/kyverno:v1.13.2", "ghcr.io"},
		{"docker.io/library/redis:7.2.4-alpine", "docker.io"},
		{"registry.k8s.io/ingress-nginx/controller:v1.11.2", "registry.k8s.io"},
		{"nginx:alpine", ""},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			got := SourceRegistryHost(tt.ref)
			if got != tt.want {
				t.Errorf("SourceRegistryHost(%q) = %q, want %q", tt.ref, got, tt.want)
			}
		})
	}
}
