package main

import (
	"fmt"
	"os"

	"github.com/mistergrinvalds/lazyoci/pkg/app"
	"github.com/mistergrinvalds/lazyoci/pkg/config"
	"github.com/spf13/cobra"
)

var artifactDir string
var themeName string

var rootCmd = &cobra.Command{
	Use:   "lazyoci",
	Short: "A TUI for browsing OCI registries",
	Long: `lazyoci is a terminal UI for browsing OCI container registries.

Browse docker.io, quay.io, ghcr.io, and custom registries to find
container images, Helm charts, and other OCI artifacts.

Environment Variables:
  LAZYOCI_ARTIFACT_DIR  Override artifact storage directory`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set CLI override for artifact directory if flag was provided
		if artifactDir != "" {
			config.SetArtifactDirOverride(artifactDir)
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTUI()
	},
}

func runTUI() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// CLI --theme flag overrides config
	if themeName != "" {
		cfg.Theme = themeName
	}

	application, err := app.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize app: %w", err)
	}

	return application.Run()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text", "Output format: text, json, yaml")
	rootCmd.PersistentFlags().StringVar(&artifactDir, "artifact-dir", "", "Override artifact storage directory")
	rootCmd.PersistentFlags().StringVar(&themeName, "theme", "", "Color theme (default, catppuccin-mocha, catppuccin-latte, dracula, tokyonight, gruvbox, solarized-dark)")

	// Set version for --version flag (version var is in version.go, set via ldflags)
	rootCmd.Version = version
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
