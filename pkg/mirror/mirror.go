package mirror

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/mistergrinvalds/lazyoci/pkg/config"
	"github.com/mistergrinvalds/lazyoci/pkg/registry"
	"golang.org/x/sync/errgroup"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// Options configures a mirror operation.
type Options struct {
	// Config is the parsed mirror configuration.
	Config *Config
	// AppConfig is the lazyoci application config (used for credential resolution).
	AppConfig *config.Config
	// DryRun previews what would be mirrored without making changes.
	DryRun bool
	// ChartsOnly mirrors chart OCI artifacts but skips container images.
	ChartsOnly bool
	// ImagesOnly mirrors container images but skips chart artifacts.
	ImagesOnly bool
	// Force re-copies images even if they already exist in the target registry.
	// Useful for re-pushing with updated platform tags.
	Force bool
	// Concurrency is the number of parallel image copies per chart version.
	// Defaults to 4 if zero.
	Concurrency int
	// Log is where human-readable progress is written. If nil, output is discarded.
	Log io.Writer
}

// MirrorResult is the aggregate result of a mirror operation, suitable for
// JSON/YAML serialisation.
type MirrorResult struct {
	Charts []ChartResult `json:"charts" yaml:"charts"`
	// Totals across all charts.
	ChartsPushed  int  `json:"chartsPushed" yaml:"chartsPushed"`
	ChartsSkipped int  `json:"chartsSkipped" yaml:"chartsSkipped"`
	ChartsFailed  int  `json:"chartsFailed" yaml:"chartsFailed"`
	ImagesCopied  int  `json:"imagesCopied" yaml:"imagesCopied"`
	ImagesSkipped int  `json:"imagesSkipped" yaml:"imagesSkipped"`
	ImagesFailed  int  `json:"imagesFailed" yaml:"imagesFailed"`
	DryRun        bool `json:"dryRun,omitempty" yaml:"dryRun,omitempty"`
}

// ChartResult holds the outcome for a single chart key.
type ChartResult struct {
	Key      string          `json:"key" yaml:"key"`
	Chart    string          `json:"chart" yaml:"chart"`
	Versions []VersionResult `json:"versions" yaml:"versions"`
}

// VersionResult holds the outcome for a single chart version.
type VersionResult struct {
	Version       string        `json:"version" yaml:"version"`
	ChartStatus   string        `json:"chartStatus" yaml:"chartStatus"` // "pushed", "skipped", "failed", "dry-run"
	ChartError    string        `json:"chartError,omitempty" yaml:"chartError,omitempty"`
	Images        []ImageResult `json:"images,omitempty" yaml:"images,omitempty"`
	ImagesCopied  int           `json:"imagesCopied" yaml:"imagesCopied"`
	ImagesSkipped int           `json:"imagesSkipped" yaml:"imagesSkipped"`
	ImagesFailed  int           `json:"imagesFailed" yaml:"imagesFailed"`
}

// ImageResult holds the outcome for a single image copy.
type ImageResult struct {
	Source string `json:"source" yaml:"source"`
	Target string `json:"target" yaml:"target"`
	Status string `json:"status" yaml:"status"` // "copied", "skipped", "failed", "dry-run"
	Error  string `json:"error,omitempty" yaml:"error,omitempty"`
}

// Mirrorer orchestrates the mirroring of charts and their container images
// from upstream sources to a target OCI registry.
type Mirrorer struct {
	opts      Options
	regClient *registry.Client
	mu        sync.Mutex
	result    MirrorResult
}

// New creates a Mirrorer with the given options.
func New(opts Options) *Mirrorer {
	if opts.Concurrency <= 0 {
		opts.Concurrency = 4
	}
	return &Mirrorer{
		opts:      opts,
		regClient: registry.NewClient(opts.AppConfig),
	}
}

// MirrorAll mirrors every chart defined in the configuration.
func (m *Mirrorer) MirrorAll(ctx context.Context) (*MirrorResult, error) {
	for key, upstream := range m.opts.Config.Upstreams {
		versions := upstream.Versions
		if len(versions) == 0 {
			m.logf("  %s: no versions configured, skipping\n", key)
			continue
		}
		if err := m.mirrorChart(ctx, key, upstream, versions); err != nil {
			return nil, err
		}
	}

	m.result.DryRun = m.opts.DryRun
	return &m.result, nil
}

// MirrorOne mirrors a single chart by key, optionally overriding the version list.
func (m *Mirrorer) MirrorOne(ctx context.Context, key string, versionOverrides []string) (*MirrorResult, error) {
	upstream, ok := m.opts.Config.Upstreams[key]
	if !ok {
		return nil, fmt.Errorf("chart %q not found in config", key)
	}

	versions := versionOverrides
	if len(versions) == 0 {
		versions = upstream.Versions
	}
	if len(versions) == 0 {
		return nil, fmt.Errorf("chart %q: no versions specified", key)
	}

	if err := m.mirrorChart(ctx, key, upstream, versions); err != nil {
		return nil, err
	}

	m.result.DryRun = m.opts.DryRun
	return &m.result, nil
}

// mirrorChart processes a single chart key across all requested versions.
func (m *Mirrorer) mirrorChart(ctx context.Context, key string, upstream Upstream, versions []string) error {
	target := m.opts.Config.Target
	chartBase := target.ChartOCIBase()
	chartName := upstream.Chart

	m.logf("Mirror: %s (%s) → %s\n", key, chartName, chartBase)
	m.logf("Versions: %s\n\n", strings.Join(versions, ", "))

	cr := ChartResult{Key: key, Chart: chartName}

	for _, ver := range versions {
		vr := m.mirrorVersion(ctx, key, upstream, ver)
		cr.Versions = append(cr.Versions, vr)
	}

	m.mu.Lock()
	m.result.Charts = append(m.result.Charts, cr)
	m.mu.Unlock()

	return nil
}

// mirrorVersion handles a single (chart, version) pair.
func (m *Mirrorer) mirrorVersion(ctx context.Context, key string, upstream Upstream, version string) VersionResult {
	target := m.opts.Config.Target
	chartBase := target.ChartOCIBase()
	chartName := upstream.Chart
	vr := VersionResult{Version: version}

	m.logf("── %s:%s ──\n", chartName, version)

	// --- Chart mirroring ---
	var tgzPath string
	if !m.opts.ImagesOnly {
		chartRef := chartBase + "/" + chartName + ":" + version
		targetCredFn := m.credentialFunc(target.URL)

		if !m.opts.Force && Exists(ctx, chartRef, target.Insecure, targetCredFn) {
			m.logf("  Chart: already exists, skipping\n")
			vr.ChartStatus = "skipped"
			m.addChartSkipped()
		} else if m.opts.DryRun {
			m.logf("  Chart: would push → oci://%s/%s:%s\n", chartBase, chartName, version)
			vr.ChartStatus = "dry-run"
		} else {
			// Pull from upstream.
			tgz, cleanup, err := PullChart(ctx, upstream, version)
			if err != nil {
				m.logf("  Chart: failed to pull — %s\n", err)
				vr.ChartStatus = "failed"
				vr.ChartError = err.Error()
				m.addChartFailed()
			} else {
				defer cleanup()
				tgzPath = tgz

				// Push to target.
				m.logf("  Chart: pushing... ")
				if err := PushChart(ctx, tgzPath, chartBase, chartName, version, target.Insecure, targetCredFn); err != nil {
					m.logf("FAILED (%s)\n", err)
					vr.ChartStatus = "failed"
					vr.ChartError = err.Error()
					m.addChartFailed()
				} else {
					m.logf("OK\n")
					vr.ChartStatus = "pushed"
					m.addChartPushed()
				}
			}
		}
	}

	// --- Image mirroring ---
	if !m.opts.ChartsOnly {
		// If we don't have a tgz yet (chart was skipped or images-only mode), pull it.
		if tgzPath == "" {
			tgz, cleanup, err := PullChart(ctx, upstream, version)
			if err != nil {
				m.logf("  Images: could not pull chart for image extraction — %s\n", err)
				m.logf("\n")
				return vr
			}
			defer cleanup()
			tgzPath = tgz
		}

		images, err := ExtractImages(ctx, tgzPath)
		if err != nil || len(images) == 0 {
			m.logf("  Images: none found\n")
		} else {
			m.logf("  Images: %d found\n", len(images))
			vr.Images = m.mirrorImages(ctx, images)
			for _, ir := range vr.Images {
				switch ir.Status {
				case "copied", "dry-run":
					vr.ImagesCopied++
				case "skipped":
					vr.ImagesSkipped++
				case "failed":
					vr.ImagesFailed++
				}
			}
		}
	}

	m.logf("\n")
	return vr
}

// mirrorImages copies a list of container images to the target registry.
// When not in dry-run mode, images are copied in parallel up to the
// configured concurrency limit.
func (m *Mirrorer) mirrorImages(ctx context.Context, images []string) []ImageResult {
	target := m.opts.Config.Target
	targetURL := target.URL
	targetCredFn := m.credentialFunc(target.URL)

	results := make([]ImageResult, len(images))

	// Dry-run stays sequential — no I/O, fast, keeps output readable.
	if m.opts.DryRun {
		for i, src := range images {
			dst := RemapImage(src, targetURL)
			m.logf("    %s → %s (dry-run)\n", src, dst)
			results[i] = ImageResult{Source: src, Target: dst, Status: "dry-run"}
		}
		return results
	}

	// Use errgroup with a concurrency limit for parallel copies.
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(m.opts.Concurrency)

	for i, src := range images {
		i, src := i, src // capture loop variables
		dst := RemapImage(src, targetURL)
		results[i] = ImageResult{Source: src, Target: dst}

		g.Go(func() error {
			// Check if target image already exists.
			if !m.opts.Force && Exists(gctx, dst, target.Insecure, targetCredFn) {
				m.logf("    %s → exists\n", src)
				results[i].Status = "skipped"
				m.addImageSkipped()
				return nil
			}

			// Resolve source credentials.  Each source registry gets its own
			// credential lookup — no credential leaking between registries.
			srcHost := SourceRegistryHost(src)
			srcCredFn := m.credentialFunc(srcHost)

			m.logf("    %s → copying...\n", src)
			if err := CopyImage(gctx, src, dst, false, target.Insecure, srcCredFn, targetCredFn); err != nil {
				m.logf("    %s → FAILED (%s)\n", src, err)
				results[i].Status = "failed"
				results[i].Error = err.Error()
				m.addImageFailed()
				// Don't return error — we want to continue with other images.
				return nil
			}
			m.logf("    %s → OK\n", src)
			results[i].Status = "copied"
			m.addImageCopied()
			return nil
		})
	}

	// Wait for all copies to complete.  Errors are handled per-image above
	// so this should always return nil.
	_ = g.Wait()

	return results
}

// credentialFunc returns an auth.CredentialFunc for the given registry URL.
// It extracts the hostname from URLs that include a path
// (e.g. "registry.digitalocean.com/greenforests" → "registry.digitalocean.com").
func (m *Mirrorer) credentialFunc(registryURL string) auth.CredentialFunc {
	host := registryURL
	if idx := strings.Index(host, "/"); idx != -1 {
		host = host[:idx]
	}
	return m.regClient.CredentialFunc(host)
}

// --- logging ---

func (m *Mirrorer) logf(format string, args ...any) {
	if m.opts.Log != nil {
		fmt.Fprintf(m.opts.Log, format, args...)
	}
}

// --- thread-safe counter updates ---

func (m *Mirrorer) addChartPushed()  { m.mu.Lock(); m.result.ChartsPushed++; m.mu.Unlock() }
func (m *Mirrorer) addChartSkipped() { m.mu.Lock(); m.result.ChartsSkipped++; m.mu.Unlock() }
func (m *Mirrorer) addChartFailed()  { m.mu.Lock(); m.result.ChartsFailed++; m.mu.Unlock() }
func (m *Mirrorer) addImageCopied()  { m.mu.Lock(); m.result.ImagesCopied++; m.mu.Unlock() }
func (m *Mirrorer) addImageSkipped() { m.mu.Lock(); m.result.ImagesSkipped++; m.mu.Unlock() }
func (m *Mirrorer) addImageFailed()  { m.mu.Lock(); m.result.ImagesFailed++; m.mu.Unlock() }
