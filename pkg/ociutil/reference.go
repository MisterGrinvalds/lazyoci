package ociutil

import (
	"fmt"
	"strings"
)

// Reference represents a parsed OCI reference.
type Reference struct {
	Registry   string
	Repository string
	Tag        string
	Digest     string // e.g. "sha256:abc123..." (without leading @)
}

// String returns the full reference string (registry/repository:tag or
// registry/repository@digest).
func (r *Reference) String() string {
	base := r.Registry + "/" + r.Repository
	if r.Tag != "" {
		base += ":" + r.Tag
	}
	if r.Digest != "" {
		base += "@" + r.Digest
	}
	return base
}

// Ref returns the tag if present, otherwise the digest prefixed with @.
// This is the value to pass to oras.Copy as the reference string.
func (r *Reference) Ref() string {
	if r.Tag != "" {
		return r.Tag
	}
	if r.Digest != "" {
		return "@" + r.Digest
	}
	return "latest"
}

// ParseReference parses an image reference like "docker.io/library/nginx:latest"
// or "quay.io/cilium/cilium:v1.18.7@sha256:99b02...".
func ParseReference(ref string) (*Reference, error) {
	// Handle docker.io shorthand
	if !strings.Contains(ref, "/") {
		// Single name like "nginx" → docker.io/library/nginx
		ref = "docker.io/library/" + ref
	} else if !strings.Contains(strings.Split(ref, "/")[0], ".") &&
		!strings.Contains(strings.Split(ref, "/")[0], ":") &&
		!strings.Contains(strings.Split(ref, "/")[0], "localhost") {
		// No registry specified, e.g., "library/nginx" → docker.io/library/nginx
		ref = "docker.io/" + ref
	}

	// Extract digest if present (e.g., @sha256:abc123...).
	var digest string
	if idx := strings.Index(ref, "@"); idx != -1 {
		digest = ref[idx+1:] // "sha256:abc123..."
		ref = ref[:idx]      // strip digest from ref before parsing tag
	}

	// Split off tag
	tag := ""
	if idx := strings.LastIndex(ref, ":"); idx != -1 {
		// Make sure this is a tag, not a port
		afterColon := ref[idx+1:]
		if !strings.Contains(afterColon, "/") {
			tag = afterColon
			ref = ref[:idx]
		}
	}

	// Default to "latest" only when neither tag nor digest is specified.
	if tag == "" && digest == "" {
		tag = "latest"
	}

	// Split registry and repository
	parts := strings.SplitN(ref, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid reference format: %s", ref)
	}

	return &Reference{
		Registry:   parts[0],
		Repository: parts[1],
		Tag:        tag,
		Digest:     digest,
	}, nil
}
