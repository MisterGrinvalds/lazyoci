package mirror

import (
	"strings"
)

// RemapImage rewrites a fully-qualified source image reference so it lives
// under targetRegistry.  The source registry host is stripped and the
// remaining repository path + tag/digest are preserved.
//
// Examples:
//
//	ghcr.io/kyverno/kyverno:v1.13.2
//	  → registry.digitalocean.com/greenforests/kyverno/kyverno:v1.13.2
//
//	docker.io/library/redis:7.2.4-alpine
//	  → registry.digitalocean.com/greenforests/library/redis:7.2.4-alpine
//
//	registry.k8s.io/ingress-nginx/controller:v1.11.2@sha256:abc...
//	  → registry.digitalocean.com/greenforests/ingress-nginx/controller:v1.11.2@sha256:abc...
func RemapImage(src, targetRegistry string) string {
	// Separate the reference portion (registry+path) from the identifier
	// (tag or digest).  A digest reference uses '@', a tag uses ':'.
	ref, identifier := splitRefIdentifier(src)

	// Strip the source registry host — everything before the first '/'.
	path := ref
	if idx := strings.Index(ref, "/"); idx != -1 {
		path = ref[idx+1:]
	}

	return targetRegistry + "/" + path + identifier
}

// splitRefIdentifier splits an image reference into the registry+repo part
// and the identifier part (":tag", "@sha256:...", or ":latest" default).
func splitRefIdentifier(src string) (ref, identifier string) {
	// Digest reference takes precedence.
	if idx := strings.Index(src, "@sha256:"); idx != -1 {
		return src[:idx], src[idx:]
	}

	// Tag reference — find the last ':' that is after the last '/'.
	// This avoids confusing a port (registry.io:5000/repo) with a tag.
	lastSlash := strings.LastIndex(src, "/")
	lastColon := strings.LastIndex(src, ":")
	if lastColon > lastSlash && lastColon != -1 {
		return src[:lastColon], src[lastColon:]
	}

	// No tag or digest — default to :latest.
	return src, ":latest"
}

// NormalizeImage ensures a bare image reference is fully qualified.
//
//	"nginx:alpine"           → "docker.io/library/nginx:alpine"
//	"stakater/reloader:v1.0" → "docker.io/stakater/reloader:v1.0"
//	"ghcr.io/foo/bar:v1"     → "ghcr.io/foo/bar:v1" (unchanged)
func NormalizeImage(img string) string {
	img = strings.TrimSpace(img)

	// Strip surrounding quotes.
	img = strings.Trim(img, `"'`)

	if img == "" || strings.Contains(img, "{{") {
		return ""
	}

	// No slash → bare Docker Hub library image.
	if !strings.Contains(img, "/") {
		return "docker.io/library/" + img
	}

	// First path component has no dot and no colon → Docker Hub user image.
	first := img[:strings.Index(img, "/")]
	if !strings.Contains(first, ".") && !strings.Contains(first, ":") && first != "localhost" {
		return "docker.io/" + img
	}

	return img
}
