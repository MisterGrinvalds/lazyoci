package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mistergrinvalds/lazyoci/pkg/config"
	"github.com/mistergrinvalds/lazyoci/pkg/pull"
	"github.com/mistergrinvalds/lazyoci/pkg/registry"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

var (
	pullDest     string
	pullPlatform string
	pullDocker   bool
	pullQuiet    bool
)

var pullCmd = &cobra.Command{
	Use:   "pull <reference>",
	Short: "Pull an OCI artifact from a registry",
	Long: `Pull an OCI artifact (image, Helm chart, etc.) from a registry to local storage.

By default, artifacts are stored as OCI layouts in the configured artifact directory.
The artifact directory is resolved in priority order:
  1. --artifact-dir flag (global)
  2. LAZYOCI_ARTIFACT_DIR environment variable
  3. artifactDir in config file
  4. Default: ~/.cache/lazyoci/artifacts

Use --docker to load the pulled image into the Docker daemon.

Examples:
  # Pull to local OCI layout
  lazyoci pull localhost:5050/test/hello:v1

  # Pull from Docker Hub
  lazyoci pull nginx:latest
  lazyoci pull docker.io/library/alpine:3.19

  # Pull and load into Docker
  lazyoci pull alpine:latest --docker

  # Pull specific platform
  lazyoci pull nginx:latest --platform linux/arm64

  # Pull to custom directory
  lazyoci pull nginx:latest --artifact-dir ~/my-artifacts

  # Use environment variable
  LAZYOCI_ARTIFACT_DIR=/tmp/oci lazyoci pull nginx:latest

  # Quiet mode with JSON output
  lazyoci pull nginx:latest -q -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reference := args[0]

		// Load config to check for insecure registries
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Parse the reference to get registry
		ref, err := pull.ParseReference(reference)
		if err != nil {
			return fmt.Errorf("invalid reference: %w", err)
		}

		// Check if this registry is marked as insecure
		insecure := false
		for _, r := range cfg.Registries {
			if r.URL == ref.Registry && r.Insecure {
				insecure = true
				break
			}
		}

		// Resolve registry credentials through the credential chain
		regClient := registry.NewClient(cfg)
		credFn := regClient.CredentialFunc(ref.Registry)

		// Parse platform if specified (nil means no platform filter — for generic OCI artifacts)
		var platform *ocispec.Platform
		if pullPlatform != "" {
			parts := strings.Split(pullPlatform, "/")
			if len(parts) < 2 {
				return fmt.Errorf("invalid platform format %q, expected os/arch (e.g., linux/amd64)", pullPlatform)
			}
			platform = &ocispec.Platform{
				OS:           parts[0],
				Architecture: parts[1],
			}
			if len(parts) > 2 {
				platform.Variant = parts[2]
			}
		}
		// Note: if platform is nil, we pull all platforms / generic artifacts

		opts := pull.PullOptions{
			Reference:      reference,
			Destination:    pullDest, // Empty means puller will use type-aware default
			ArtifactBase:   cfg.GetArtifactDir(),
			Platform:       platform,
			ToDocker:       pullDocker,
			Quiet:          pullQuiet || isStructuredOutput(),
			Insecure:       insecure,
			CredentialFunc: credFn,
		}

		// Show what we're doing
		if !opts.Quiet && !isStructuredOutput() {
			fmt.Fprintf(cmd.ErrOrStderr(), "Pulling %s...\n", reference)
			if opts.Platform != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Platform: %s/%s\n", opts.Platform.OS, opts.Platform.Architecture)
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		puller := pull.NewPuller(opts.Quiet)
		result, err := puller.Pull(ctx, opts)
		if err != nil {
			return err
		}

		return printResult(result, func() {
			fmt.Println()
			fmt.Printf("Successfully pulled %s\n", result.Reference)
			fmt.Printf("  Type:        %s\n", result.ArtifactType.String())
			if result.TypeDetail != "" {
				fmt.Printf("  Detail:      %s\n", result.TypeDetail)
			}
			fmt.Printf("  Digest:      %s\n", result.Digest)
			fmt.Printf("  Size:        %s\n", formatBytes(result.Size))
			fmt.Printf("  Layers:      %d\n", result.Layers)
			fmt.Printf("  Destination: %s\n", result.Destination)
			if result.LoadedToDocker {
				fmt.Println("  Loaded into Docker daemon")
			}
		})
	},
}

func init() {
	pullCmd.Flags().StringVarP(&pullDest, "dest", "d", "", "Explicit destination directory (overrides artifact-dir)")
	pullCmd.Flags().StringVar(&pullPlatform, "platform", "", "Target platform (e.g., linux/amd64)")
	pullCmd.Flags().BoolVar(&pullDocker, "docker", false, "Load pulled image into Docker daemon")
	pullCmd.Flags().BoolVarP(&pullQuiet, "quiet", "q", false, "Suppress progress output")

	rootCmd.AddCommand(pullCmd)
}
