package mirror

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mistergrinvalds/lazyoci/pkg/ociutil"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// PullChart downloads a chart from its upstream source into a temporary OCI
// layout directory.  The caller is responsible for cleaning up the returned
// directory.
//
// For repo and oci source types the helm CLI is used (exec).
// For local source types the chart is packaged from the filesystem.
func PullChart(ctx context.Context, upstream Upstream, version string) (tgzPath string, cleanup func(), err error) {
	tmpDir, err := os.MkdirTemp("", "lazyoci-mirror-chart-*")
	if err != nil {
		return "", nil, fmt.Errorf("creating temp dir: %w", err)
	}
	cleanupFn := func() { os.RemoveAll(tmpDir) }

	switch upstream.Type {
	case SourceRepo:
		tgz, err := pullChartRepo(upstream, version, tmpDir)
		if err != nil {
			cleanupFn()
			return "", nil, err
		}
		return tgz, cleanupFn, nil

	case SourceOCI:
		tgz, err := pullChartOCI(upstream, version, tmpDir)
		if err != nil {
			cleanupFn()
			return "", nil, err
		}
		return tgz, cleanupFn, nil

	case SourceLocal:
		tgz, err := pullChartLocal(upstream, version, tmpDir)
		if err != nil {
			cleanupFn()
			return "", nil, err
		}
		return tgz, cleanupFn, nil

	default:
		cleanupFn()
		return "", nil, fmt.Errorf("unknown source type %q", upstream.Type)
	}
}

// PushChart copies a Helm chart .tgz to the target OCI registry using
// oras.Copy through a local OCI layout intermediary.
//
// The targetBase is the OCI base path (e.g. "registry.example.com/ns/charts")
// and chartName is the chart name.  The chart is pushed as
// oci://targetBase/chartName:version.
func PushChart(ctx context.Context, tgzPath, targetBase, chartName, version string, insecure bool, credFn auth.CredentialFunc) error {
	// Use helm push which handles the OCI manifest construction correctly.
	ref := "oci://" + targetBase
	args := []string{"push", tgzPath, ref}
	if insecure {
		args = append(args, "--plain-http")
	}

	cmd := exec.CommandContext(ctx, "helm", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("helm push failed: %s: %w", strings.TrimSpace(stderr.String()), err)
	}
	return nil
}

// CopyOCIChart performs a direct registry-to-registry copy of a chart that
// already exists as an OCI artifact.  Used for oci-type upstreams where we
// can skip the helm pull→push round-trip and copy directly via oras.
func CopyOCIChart(ctx context.Context, srcRef, dstRef string, srcInsecure, dstInsecure bool, srcCredFn, dstCredFn auth.CredentialFunc) error {
	srcParsed, err := ociutil.ParseReference(srcRef)
	if err != nil {
		return fmt.Errorf("parsing source ref: %w", err)
	}
	dstParsed, err := ociutil.ParseReference(dstRef)
	if err != nil {
		return fmt.Errorf("parsing destination ref: %w", err)
	}

	srcRepo, err := ociutil.NewRemoteRepository(srcParsed, srcInsecure, srcCredFn)
	if err != nil {
		return fmt.Errorf("creating source repo: %w", err)
	}
	dstRepo, err := ociutil.NewRemoteRepository(dstParsed, dstInsecure, dstCredFn)
	if err != nil {
		return fmt.Errorf("creating destination repo: %w", err)
	}

	_, err = oras.Copy(ctx, srcRepo, srcParsed.Tag, dstRepo, dstParsed.Tag, oras.CopyOptions{})
	if err != nil {
		return fmt.Errorf("oras copy: %w", err)
	}
	return nil
}

// --- internal helpers ---

// pullChartRepo uses `helm repo add` + `helm pull` to download a chart from
// a traditional Helm repository.
func pullChartRepo(u Upstream, version, destDir string) (string, error) {
	// Derive a stable repo alias from the chart key.
	repoAlias := "lazyoci-" + u.Chart

	// Add the repo (idempotent).
	if err := helmExec("repo", "add", repoAlias, u.Repo); err != nil {
		// Ignore if already exists.
		_ = err
	}
	if err := helmExec("repo", "update", repoAlias); err != nil {
		return "", fmt.Errorf("helm repo update: %w", err)
	}

	if err := helmExec("pull", repoAlias+"/"+u.Chart, "--version", version, "--destination", destDir); err != nil {
		return "", fmt.Errorf("helm pull: %w", err)
	}

	return findTgz(destDir, u.Chart, version)
}

// pullChartOCI uses `helm pull` with an OCI URL.
func pullChartOCI(u Upstream, version, destDir string) (string, error) {
	ociRef := strings.TrimSuffix(u.Registry, "/") + "/" + u.Chart
	if err := helmExec("pull", ociRef, "--version", version, "--destination", destDir); err != nil {
		return "", fmt.Errorf("helm pull (oci): %w", err)
	}

	return findTgz(destDir, u.Chart, version)
}

// pullChartLocal packages a local chart directory using `helm package`.
func pullChartLocal(u Upstream, version, destDir string) (string, error) {
	chartPath := u.Path

	// Build dependencies if Chart.lock or dependencies exist.
	lockPath := filepath.Join(chartPath, "Chart.lock")
	if _, err := os.Stat(lockPath); err == nil {
		_ = helmExec("dependency", "build", chartPath, "--skip-refresh")
	}

	if err := helmExec("package", chartPath, "--destination", destDir); err != nil {
		return "", fmt.Errorf("helm package: %w", err)
	}

	return findTgz(destDir, u.Chart, version)
}

// findTgz locates the .tgz file produced by helm pull or helm package.
// Some charts embed build metadata in the filename so we do a glob match.
func findTgz(dir, chartName, version string) (string, error) {
	// Try exact match first.
	exact := filepath.Join(dir, chartName+"-"+version+".tgz")
	if _, err := os.Stat(exact); err == nil {
		return exact, nil
	}

	// Glob for build-metadata suffixed filenames.
	pattern := filepath.Join(dir, chartName+"-"+version+"*.tgz")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("glob: %w", err)
	}
	if len(matches) > 0 {
		return matches[0], nil
	}

	// Broader glob — sometimes the version in Chart.yaml differs from what
	// was requested (e.g. local charts).
	pattern = filepath.Join(dir, chartName+"-*.tgz")
	matches, err = filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("glob: %w", err)
	}
	if len(matches) > 0 {
		return matches[0], nil
	}

	return "", fmt.Errorf("no .tgz found in %s for %s-%s", dir, chartName, version)
}

// helmExec runs a helm CLI command, discarding stdout and returning any
// error with stderr context.
func helmExec(args ...string) error {
	cmd := exec.Command("helm", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	// Discard stdout to avoid polluting function return values.
	cmd.Stdout = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", strings.Join(args, " "), strings.TrimSpace(stderr.String()))
	}
	return nil
}

// pushOCILayout pushes a local OCI layout directory to a remote repository
// using oras.Copy.  Used as a fallback when helm push is not suitable.
func pushOCILayout(ctx context.Context, layoutDir, targetRef string, insecure bool, credFn auth.CredentialFunc) error {
	store, err := oci.New(layoutDir)
	if err != nil {
		return fmt.Errorf("opening OCI layout: %w", err)
	}

	parsed, err := ociutil.ParseReference(targetRef)
	if err != nil {
		return fmt.Errorf("parsing target ref: %w", err)
	}

	repo, err := ociutil.NewRemoteRepository(parsed, insecure, credFn)
	if err != nil {
		return fmt.Errorf("creating remote repo: %w", err)
	}

	index, err := ociutil.ReadOCIIndex(layoutDir)
	if err != nil {
		return fmt.Errorf("reading OCI index: %w", err)
	}
	if len(index.Manifests) == 0 {
		return fmt.Errorf("OCI layout has no manifests")
	}

	srcDigest := index.Manifests[0].Digest
	_, err = oras.Copy(ctx, store, srcDigest, repo, parsed.Tag, oras.CopyOptions{})
	return err
}
