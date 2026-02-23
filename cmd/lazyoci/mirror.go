package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mistergrinvalds/lazyoci/pkg/config"
	"github.com/mistergrinvalds/lazyoci/pkg/mirror"
	"github.com/spf13/cobra"
)

var (
	mirrorConfigPath  string
	mirrorChart       string
	mirrorVersions    []string
	mirrorDryRun      bool
	mirrorChartsOnly  bool
	mirrorImagesOnly  bool
	mirrorAll         bool
	mirrorForce       bool
	mirrorConcurrency int
)

var mirrorCmd = &cobra.Command{
	Use:   "mirror",
	Short: "Mirror Helm charts and container images to a target OCI registry",
	Long: `Mirror Helm chart OCI artifacts and their referenced container images from
upstream sources to a target OCI registry.

Configuration is read from a YAML file (--config) that defines upstream chart
sources, version lists, and the target registry.  Artifacts that already exist
in the target are skipped automatically.

Examples:
  # Mirror all charts defined in the config
  lazyoci mirror --config mirror.yaml --all

  # Mirror a specific chart
  lazyoci mirror --config mirror.yaml --chart vault

  # Mirror with version override
  lazyoci mirror --config mirror.yaml --chart vault --version 0.28.0

  # Mirror multiple versions
  lazyoci mirror --config mirror.yaml --chart vault --version 0.28.0 --version 0.29.0

  # Dry run — preview what would be mirrored
  lazyoci mirror --config mirror.yaml --all --dry-run

  # Charts only (skip container images)
  lazyoci mirror --config mirror.yaml --chart vault --charts-only

  # Images only (skip chart push)
  lazyoci mirror --config mirror.yaml --chart vault --images-only

  # JSON output for scripting
  lazyoci mirror --config mirror.yaml --all -o json`,
	RunE: runMirror,
}

func init() {
	mirrorCmd.Flags().StringVarP(&mirrorConfigPath, "config", "c", "mirror.yaml", "Path to mirror config file")
	mirrorCmd.Flags().StringVar(&mirrorChart, "chart", "", "Mirror a specific chart key")
	mirrorCmd.Flags().StringSliceVar(&mirrorVersions, "version", nil, "Override version(s) to mirror (repeatable)")
	mirrorCmd.Flags().BoolVar(&mirrorDryRun, "dry-run", false, "Preview what would be mirrored")
	mirrorCmd.Flags().BoolVar(&mirrorChartsOnly, "charts-only", false, "Mirror chart artifacts only, skip images")
	mirrorCmd.Flags().BoolVar(&mirrorImagesOnly, "images-only", false, "Mirror images only, skip chart push")
	mirrorCmd.Flags().BoolVar(&mirrorAll, "all", false, "Mirror all charts in the config")
	mirrorCmd.Flags().BoolVar(&mirrorForce, "force", false, "Re-copy images even if they already exist in the target")
	mirrorCmd.Flags().IntVar(&mirrorConcurrency, "concurrency", 4, "Parallel image copies per chart")

	rootCmd.AddCommand(mirrorCmd)
}

func runMirror(cmd *cobra.Command, args []string) error {
	if mirrorChart == "" && !mirrorAll {
		return fmt.Errorf("specify --chart <key> or --all")
	}
	if mirrorChart != "" && mirrorAll {
		return fmt.Errorf("--chart and --all are mutually exclusive")
	}

	// Load mirror config.
	mirrorCfg, err := mirror.LoadConfig(mirrorConfigPath)
	if err != nil {
		return fmt.Errorf("loading mirror config: %w", err)
	}

	// Load app config for credential resolution.
	appCfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading app config: %w", err)
	}

	// Determine log output — suppress for structured output.
	var logWriter *os.File
	if !isStructuredOutput() {
		logWriter = os.Stderr
	}

	opts := mirror.Options{
		Config:      mirrorCfg,
		AppConfig:   appCfg,
		DryRun:      mirrorDryRun,
		ChartsOnly:  mirrorChartsOnly,
		ImagesOnly:  mirrorImagesOnly,
		Force:       mirrorForce,
		Concurrency: mirrorConcurrency,
		Log:         logWriter,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	m := mirror.New(opts)

	var result *mirror.MirrorResult
	if mirrorAll {
		result, err = m.MirrorAll(ctx)
	} else {
		result, err = m.MirrorOne(ctx, mirrorChart, mirrorVersions)
	}
	if err != nil {
		return err
	}

	return printResult(result, func() {
		printMirrorSummary(result)
	})
}

func printMirrorSummary(r *mirror.MirrorResult) {
	fmt.Println("════════════════════════════════════════")
	fmt.Println("Mirror Summary")
	fmt.Println("════════════════════════════════════════")

	for _, cr := range r.Charts {
		fmt.Printf("  %s (%s):\n", cr.Key, cr.Chart)
		for _, vr := range cr.Versions {
			fmt.Printf("    %s: chart=%s", vr.Version, vr.ChartStatus)
			if vr.ChartError != "" {
				fmt.Printf(" (%s)", vr.ChartError)
			}
			total := vr.ImagesCopied + vr.ImagesSkipped + vr.ImagesFailed
			if total > 0 {
				fmt.Printf("  images=%d copied, %d skipped, %d failed",
					vr.ImagesCopied, vr.ImagesSkipped, vr.ImagesFailed)
			}
			fmt.Println()
		}
	}

	fmt.Println()
	fmt.Printf("  Charts:  %d pushed, %d skipped, %d failed\n",
		r.ChartsPushed, r.ChartsSkipped, r.ChartsFailed)
	fmt.Printf("  Images:  %d copied, %d skipped, %d failed\n",
		r.ImagesCopied, r.ImagesSkipped, r.ImagesFailed)

	if r.DryRun {
		fmt.Println("  (dry-run — no changes made)")
	}
}
