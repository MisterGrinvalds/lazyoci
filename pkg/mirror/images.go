package mirror

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/mistergrinvalds/lazyoci/pkg/ociutil"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// imageLineRE matches lines of the form `  image: <ref>` in rendered Helm
// template output.
var imageLineRE = regexp.MustCompile(`(?m)^\s*image:\s*(.+)$`)

// ExtractImages renders a Helm chart with `helm template` and extracts all
// unique, fully-qualified container image references from the output.
//
// chartPath can be a .tgz file or an unpacked chart directory.
func ExtractImages(ctx context.Context, chartPath string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "helm", "template", "extract-images", chartPath,
		"--no-hooks",
		"--include-crds=false",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Best-effort — some charts have template errors with default values.
	_ = cmd.Run()

	if stdout.Len() == 0 {
		return nil, nil
	}

	seen := make(map[string]struct{})
	var images []string

	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		line := scanner.Text()
		matches := imageLineRE.FindStringSubmatch(line)
		if len(matches) < 2 {
			continue
		}
		raw := strings.TrimSpace(matches[1])
		raw = strings.Trim(raw, `"'`)
		if raw == "" || strings.Contains(raw, "{{") || raw == "null" {
			continue
		}

		img := NormalizeImage(raw)
		if img == "" {
			continue
		}
		if _, ok := seen[img]; ok {
			continue
		}
		seen[img] = struct{}{}
		images = append(images, img)
	}

	sort.Strings(images)
	return images, nil
}

// CopyImage performs a registry-to-registry copy of a single container image
// using oras.Copy.  Both source and destination get independent auth clients
// so credentials never leak between registries.
//
// For multi-arch images (OCI index / manifest list), each platform manifest
// is tagged with "<tag>-<os>-<arch>" in the destination registry so that
// registries like DOCR don't surface untagged child manifests.
func CopyImage(ctx context.Context, srcRef, dstRef string, srcInsecure, dstInsecure bool, srcCredFn, dstCredFn auth.CredentialFunc) error {
	srcParsed, err := ociutil.ParseReference(srcRef)
	if err != nil {
		return fmt.Errorf("parsing source: %w", err)
	}
	dstParsed, err := ociutil.ParseReference(dstRef)
	if err != nil {
		return fmt.Errorf("parsing destination: %w", err)
	}

	srcRepo, err := ociutil.NewRemoteRepository(srcParsed, srcInsecure, srcCredFn)
	if err != nil {
		return fmt.Errorf("source repo: %w", err)
	}
	dstRepo, err := ociutil.NewRemoteRepository(dstParsed, dstInsecure, dstCredFn)
	if err != nil {
		return fmt.Errorf("destination repo: %w", err)
	}

	// Copy the image (including all child manifests for multi-arch).
	rootDesc, err := oras.Copy(ctx, srcRepo, srcParsed.Ref(), dstRepo, dstParsed.Ref(), oras.CopyOptions{})
	if err != nil && srcCredFn != nil && isForbidden(err) {
		// Credentials were provided but rejected (e.g. expired PAT for a
		// public image).  Retry with anonymous auth — many registries like
		// ghcr.io allow unauthenticated pulls for public packages.
		anonRepo, anonErr := ociutil.NewRemoteRepository(srcParsed, srcInsecure, nil)
		if anonErr == nil {
			rootDesc, err = oras.Copy(ctx, anonRepo, srcParsed.Ref(), dstRepo, dstParsed.Ref(), oras.CopyOptions{})
			if err == nil {
				srcRepo = anonRepo // use anon repo for platform tagging below
			}
		}
	}
	if err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	// If the root is a manifest list / OCI index, tag each platform manifest
	// so that the destination registry doesn't have untagged child entries.
	if isIndex(rootDesc.MediaType) {
		if err := tagPlatformManifests(ctx, srcRepo, dstRepo, rootDesc, dstParsed.Tag); err != nil {
			// Non-fatal — the copy itself succeeded. Log-worthy but not an error.
			_ = err
		}
	}

	return nil
}

// isForbidden returns true if the error message indicates a 403 Forbidden
// response from a registry.  This is a heuristic check on the error string
// because oras-go doesn't expose structured HTTP status codes.
func isForbidden(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "403") || strings.Contains(msg, "forbidden") || strings.Contains(msg, "denied")
}

// isIndex returns true if the media type is a manifest list or OCI index.
func isIndex(mediaType string) bool {
	return mediaType == ocispec.MediaTypeImageIndex ||
		mediaType == "application/vnd.docker.distribution.manifest.list.v2+json"
}

// tagPlatformManifests reads the manifest list from the source, and for each
// platform entry tags the corresponding manifest in the destination with
// "<baseTag>-<os>-<arch>".
func tagPlatformManifests(ctx context.Context, srcRepo, dstRepo oras.ReadOnlyGraphTarget, rootDesc ocispec.Descriptor, baseTag string) error {
	if baseTag == "" {
		return nil
	}

	// Fetch the index manifest to get platform entries.
	rc, err := srcRepo.Fetch(ctx, rootDesc)
	if err != nil {
		return fmt.Errorf("fetching index: %w", err)
	}
	defer rc.Close()

	var index ocispec.Index
	if err := json.NewDecoder(rc).Decode(&index); err != nil {
		return fmt.Errorf("decoding index: %w", err)
	}

	// Tag each platform manifest in the destination.
	tagger, ok := dstRepo.(interface {
		Tag(ctx context.Context, desc ocispec.Descriptor, reference string) error
	})
	if !ok {
		return nil // destination doesn't support tagging
	}

	for _, m := range index.Manifests {
		if m.Platform == nil {
			continue
		}
		platformTag := baseTag + "-" + m.Platform.OS + "-" + m.Platform.Architecture
		if err := tagger.Tag(ctx, m, platformTag); err != nil {
			// Best-effort — continue tagging others even if one fails.
			continue
		}
	}

	return nil
}

// SourceRegistryHost extracts the registry host from a fully-qualified image
// reference.  Returns empty string if no host can be determined.
func SourceRegistryHost(ref string) string {
	// Strip any tag/digest first.
	r, _ := splitRefIdentifier(ref)
	if idx := strings.Index(r, "/"); idx != -1 {
		return r[:idx]
	}
	return ""
}
