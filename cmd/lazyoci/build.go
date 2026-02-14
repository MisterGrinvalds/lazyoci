package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mistergrinvalds/lazyoci/pkg/build"
	"github.com/mistergrinvalds/lazyoci/pkg/config"
	"github.com/mistergrinvalds/lazyoci/pkg/registry"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry/remote/auth"
)

var (
	buildFile     string
	buildTag      string
	buildPush     bool
	buildNoPush   bool
	buildDryRun   bool
	buildArtifact string
	buildPlatform []string
	buildQuiet    bool
	buildInsecure bool
)

var buildCmd = &cobra.Command{
	Use:   "build [path]",
	Short: "Build and push OCI artifacts from a .lazy config",
	Long: `Build and push OCI artifacts defined in a .lazy configuration file.

The .lazy file (YAML) describes one or more artifacts to build and push to
OCI registries. Supported artifact types:

  image     Build a container image from a Dockerfile (uses docker buildx)
  helm      Package a Helm chart directory as an OCI artifact
  artifact  Package generic files as an OCI artifact with custom media types
  docker    Push an existing Docker daemon image to a registry

Tag templates support variables:
  {{ .Tag }}               Value of --tag flag (or LAZYOCI_TAG env var)
  {{ .GitSHA }}            Current git commit SHA (short)
  {{ .GitBranch }}         Current git branch name
  {{ .ChartVersion }}      Chart.yaml version (helm type only)
  {{ .Timestamp }}         Build timestamp (YYYYMMDDHHmmss)
  {{ .Version }}           Semver from git tag (v stripped), e.g. "1.2.3"
  {{ .VersionMajor }}      Major component, e.g. "1"
  {{ .VersionMinor }}      Minor component, e.g. "2"
  {{ .VersionPatch }}      Patch component, e.g. "3"
  {{ .VersionPrerelease }} Prerelease identifier, e.g. "rc.1"
  {{ .VersionMajorMinor }} "MAJOR.MINOR", e.g. "1.2"
  {{ .VersionRaw }}        Raw git tag string, e.g. "v1.2.3-rc.1"

Version auto-detection: {{ .Version }} is resolved automatically from git tags.
Override with LAZYOCI_VERSION env var or by passing a semver to --tag.

Environment variables:
  LAZYOCI_TAG       Fallback for --tag when not set on CLI
  LAZYOCI_VERSION   Override version detection (skips git describe)

Examples:
  # Build all artifacts defined in .lazy
  lazyoci build

  # Build with a specific tag (also populates {{ .Version }} if semver)
  lazyoci build --tag v1.2.3

  # Build from a specific config file
  lazyoci build --file path/to/.lazy

  # Build only, don't push
  lazyoci build --tag v1.0.0 --no-push

  # Dry run — show what would be built/pushed
  lazyoci build --tag v1.0.0 --dry-run

  # Build specific artifact by name
  lazyoci build --tag v1.0.0 --artifact myapp

  # Override platforms for image builds
  lazyoci build --tag v1.0.0 --platform linux/amd64 --platform linux/arm64

  # Build with JSON output
  lazyoci build --tag v1.0.0 -o json

  # CI: use env vars instead of flags
  LAZYOCI_TAG=v1.2.3 lazyoci build`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine config path
		configPath := buildFile
		if len(args) > 0 {
			configPath = args[0]
		}

		// Resolve the config file path
		resolvedPath, err := build.ResolveConfigPath(configPath)
		if err != nil {
			return err
		}

		// Load and validate config
		cfg, err := build.LoadConfig(resolvedPath)
		if err != nil {
			return err
		}
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}

		// Resolve push flag (--push is default true, --no-push overrides)
		push := buildPush
		if buildNoPush {
			push = false
		}

		// Resolve tag: --tag flag > LAZYOCI_TAG env var
		tag := buildTag
		if tag == "" {
			if envTag := os.Getenv("LAZYOCI_TAG"); envTag != "" {
				tag = envTag
			}
		}

		// Load app config for credential resolution
		appCfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load app config: %w", err)
		}

		regClient := registry.NewClient(appCfg)

		// Build options
		opts := build.BuilderOptions{
			Tag:            tag,
			Push:           push,
			DryRun:         buildDryRun,
			Quiet:          buildQuiet || isStructuredOutput(),
			Insecure:       buildInsecure,
			Platforms:      buildPlatform,
			ArtifactFilter: buildArtifact,
			CredentialFunc: func(registryURL string) auth.CredentialFunc {
				return regClient.CredentialFunc(registryURL)
			},
		}

		if !opts.Quiet && !isStructuredOutput() {
			fmt.Fprintf(cmd.ErrOrStderr(), "Building from %s...\n", resolvedPath)
			if opts.Tag != "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "Tag: %s\n", opts.Tag)
			}
			if opts.DryRun {
				fmt.Fprintf(cmd.ErrOrStderr(), "Mode: dry-run\n")
			} else if !opts.Push {
				fmt.Fprintf(cmd.ErrOrStderr(), "Mode: build only (no push)\n")
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		builder := build.NewBuilder(cfg, resolvedPath, opts)
		results, err := builder.Build(ctx)
		if err != nil {
			return err
		}

		return printResult(results, func() {
			fmt.Println()
			for _, r := range results {
				if r.Error != "" {
					fmt.Printf("FAIL  %s (%s): %s\n", r.Name, r.Type, r.Error)
					continue
				}
				fmt.Printf("OK    %s (%s)\n", r.Name, r.Type)
				for _, t := range r.Targets {
					status := "built"
					if t.Pushed {
						status = "pushed"
					}
					if t.Digest != "" {
						fmt.Printf("      %s [%s] %s\n", t.Reference, status, t.Digest)
					} else {
						fmt.Printf("      %s [%s]\n", t.Reference, status)
					}
				}
			}
		})
	},
}

func init() {
	buildCmd.Flags().StringVarP(&buildFile, "file", "f", "", "Path to .lazy config (default: .lazy in current directory)")
	buildCmd.Flags().StringVarP(&buildTag, "tag", "t", "", "Set {{ .Tag }} template variable (fallback: LAZYOCI_TAG env var)")
	buildCmd.Flags().BoolVar(&buildPush, "push", true, "Push to registries after build")
	buildCmd.Flags().BoolVar(&buildNoPush, "no-push", false, "Build only, don't push")
	buildCmd.Flags().BoolVar(&buildDryRun, "dry-run", false, "Show what would be built/pushed")
	buildCmd.Flags().StringVarP(&buildArtifact, "artifact", "a", "", "Build only specific artifact by name, type, or index")
	buildCmd.Flags().StringSliceVar(&buildPlatform, "platform", nil, "Override platforms for image builds (can be specified multiple times)")
	buildCmd.Flags().BoolVarP(&buildQuiet, "quiet", "q", false, "Suppress progress output")
	buildCmd.Flags().BoolVar(&buildInsecure, "insecure", false, "Allow HTTP for push targets")

	rootCmd.AddCommand(buildCmd)
}
