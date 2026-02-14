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
}

// String returns the full reference string (registry/repository:tag).
func (r *Reference) String() string {
	return r.Registry + "/" + r.Repository + ":" + r.Tag
}

// ParseReference parses an image reference like "docker.io/library/nginx:latest".
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

	// Split off tag
	tag := "latest"
	if idx := strings.LastIndex(ref, ":"); idx != -1 {
		// Make sure this is a tag, not a port
		afterColon := ref[idx+1:]
		if !strings.Contains(afterColon, "/") {
			tag = afterColon
			ref = ref[:idx]
		}
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
	}, nil
}
