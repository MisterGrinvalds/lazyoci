package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mistergrinvalds/lazyoci/pkg/config"
	"github.com/spf13/cobra"
)

var configCreateDir bool

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage lazyoci configuration",
	Long: `View and modify lazyoci configuration.

Supported configuration keys:
  artifact-dir       Directory for storing pulled artifacts
  cache-dir          Directory for metadata cache
  default-registry   Default registry shown in TUI

Examples:
  # Get a configuration value
  lazyoci config get artifact-dir

  # Set artifact directory (creates if needed)
  lazyoci config set artifact-dir ~/my-artifacts --create

  # List all configuration
  lazyoci config list

  # Show config file path
  lazyoci config path`,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		value, err := getConfigValue(cfg, key)
		if err != nil {
			return err
		}

		return printResult(value, func() {
			fmt.Println(value)
		})
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if err := setConfigValue(cfg, key, value, configCreateDir); err != nil {
			return err
		}

		result := map[string]string{
			"key":   key,
			"value": value,
		}

		return printResult(result, func() {
			fmt.Printf("Set %s = %s\n", key, value)
		})
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		configMap := map[string]interface{}{
			"artifact-dir":     cfg.GetArtifactDir(),
			"cache-dir":        cfg.CacheDir,
			"default-registry": cfg.DefaultRegistry,
			"registries":       len(cfg.Registries),
		}

		// Add source information for artifact-dir
		artifactDirSource := getArtifactDirSource(cfg)

		return printResult(configMap, func() {
			fmt.Printf("artifact-dir:     %s", cfg.GetArtifactDir())
			if artifactDirSource != "" {
				fmt.Printf(" (%s)", artifactDirSource)
			}
			fmt.Println()
			fmt.Printf("cache-dir:        %s\n", cfg.CacheDir)
			fmt.Printf("default-registry: %s\n", cfg.DefaultRegistry)
			fmt.Printf("registries:       %d configured\n", len(cfg.Registries))
		})
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file path",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.GetConfigPath()

		return printResult(path, func() {
			fmt.Println(path)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				fmt.Fprintln(os.Stderr, "(file does not exist, using defaults)")
			}
		})
	},
}

func init() {
	configSetCmd.Flags().BoolVar(&configCreateDir, "create", false, "Create directory if it doesn't exist")

	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configPathCmd)

	rootCmd.AddCommand(configCmd)
}

// getConfigValue returns the value for a configuration key.
func getConfigValue(cfg *config.Config, key string) (string, error) {
	switch strings.ToLower(key) {
	case "artifact-dir", "artifactdir":
		return cfg.GetArtifactDir(), nil
	case "cache-dir", "cachedir":
		return cfg.CacheDir, nil
	case "default-registry", "defaultregistry":
		return cfg.DefaultRegistry, nil
	default:
		return "", fmt.Errorf("unknown configuration key: %s", key)
	}
}

// setConfigValue sets a configuration value and saves the config.
func setConfigValue(cfg *config.Config, key, value string, createDir bool) error {
	switch strings.ToLower(key) {
	case "artifact-dir", "artifactdir":
		// Allow unsetting by setting to empty string
		if value == "" || value == "default" {
			cfg.ArtifactDir = ""
			return cfg.Save()
		}
		return cfg.SetArtifactDir(value, createDir)
	case "cache-dir", "cachedir":
		expanded := config.ExpandPath(value)
		if createDir {
			if err := os.MkdirAll(expanded, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		}
		cfg.CacheDir = value
		return cfg.Save()
	case "default-registry", "defaultregistry":
		cfg.DefaultRegistry = value
		return cfg.Save()
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}
}

// getArtifactDirSource returns a description of where the artifact-dir value comes from.
func getArtifactDirSource(cfg *config.Config) string {
	if config.GetArtifactDirOverride() != "" {
		return "from --artifact-dir flag"
	}
	if os.Getenv("LAZYOCI_ARTIFACT_DIR") != "" {
		return "from $LAZYOCI_ARTIFACT_DIR"
	}
	if cfg.ArtifactDir != "" {
		return "from config file"
	}
	return "default"
}
