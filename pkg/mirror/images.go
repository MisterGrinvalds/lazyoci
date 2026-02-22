package mirror

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/mistergrinvalds/lazyoci/pkg/ociutil"
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

	_, err = oras.Copy(ctx, srcRepo, srcParsed.Tag, dstRepo, dstParsed.Tag, oras.CopyOptions{})
	if err != nil {
		return fmt.Errorf("copy: %w", err)
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
