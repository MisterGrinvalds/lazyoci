package main

import (
	"fmt"

	"github.com/mistergrinvalds/lazyoci/pkg/config"
	"github.com/mistergrinvalds/lazyoci/pkg/registry"
	"github.com/spf13/cobra"
)

var (
	registryName     string
	registryUser     string
	registryPassword string
	registryInsecure bool
)

// Response types for structured output.

type registryListItem struct {
	Name     string `json:"name" yaml:"name"`
	URL      string `json:"url" yaml:"url"`
	Auth     string `json:"auth" yaml:"auth"`
	Insecure bool   `json:"insecure" yaml:"insecure"`
}

type registryAddResult struct {
	Name     string `json:"name" yaml:"name"`
	URL      string `json:"url" yaml:"url"`
	Insecure bool   `json:"insecure" yaml:"insecure"`
	Status   string `json:"status" yaml:"status"`
	Message  string `json:"message,omitempty" yaml:"message,omitempty"`
}

type registryTestResult struct {
	URL    string `json:"url" yaml:"url"`
	Status string `json:"status" yaml:"status"`
	Error  string `json:"error,omitempty" yaml:"error,omitempty"`
}

type registryRemoveResult struct {
	URL    string `json:"url" yaml:"url"`
	Status string `json:"status" yaml:"status"`
}

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage OCI registries",
	Long:  `Add, remove, list, and test OCI registries.`,
}

var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured registries",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		var items []registryListItem
		for _, reg := range cfg.Registries {
			auth := ""
			if reg.Username != "" {
				auth = reg.Username
			}
			items = append(items, registryListItem{
				Name:     reg.Name,
				URL:      reg.URL,
				Auth:     auth,
				Insecure: reg.Insecure,
			})
		}

		return printResult(items, func() {
			w := newTabWriter()
			fmt.Fprintln(w, "NAME\tURL\tAUTH\tINSECURE")
			fmt.Fprintln(w, "----\t---\t----\t--------")
			for _, item := range items {
				auth := "-"
				if item.Auth != "" {
					auth = item.Auth
				}
				insecure := "-"
				if item.Insecure {
					insecure = "yes"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", item.Name, item.URL, auth, insecure)
			}
			w.Flush()
		})
	},
}

var registryAddCmd = &cobra.Command{
	Use:   "add <url>",
	Short: "Add a registry",
	Long: `Add a new OCI registry to the configuration.

Examples:
  lazyoci registry add harbor.example.com
  lazyoci registry add harbor.example.com --name "My Harbor"
  lazyoci registry add private.io --user=admin --pass=secret
  lazyoci registry add localhost:5050 --insecure`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Check if already exists
		if cfg.GetRegistry(url) != nil {
			return fmt.Errorf("registry %s already exists", url)
		}

		// Set insecure flag before adding so connectivity test uses HTTP
		name := registryName

		// Add the registry first (so insecure flag is available for test)
		if registryUser != "" {
			err = cfg.AddRegistryWithAuth(name, url, registryUser, registryPassword)
		} else {
			err = cfg.AddRegistry(name, url)
		}
		if err != nil {
			return fmt.Errorf("failed to add registry: %w", err)
		}

		// Set insecure flag if requested
		if registryInsecure {
			reg := cfg.GetRegistry(url)
			if reg != nil {
				reg.Insecure = true
				cfg.Save()
			}
		}

		// Test connectivity
		client := registry.NewClient(cfg)
		var testMsg string
		if !isStructuredOutput() {
			fmt.Fprintf(cmd.ErrOrStderr(), "Testing connection to %s...\n", url)
		}
		if err := client.TestRegistry(url); err != nil {
			testMsg = fmt.Sprintf("warning: %v", err)
			if !isStructuredOutput() {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %v\n", err)
				fmt.Fprintln(cmd.ErrOrStderr(), "Registry added. You may need to configure authentication.")
			}
		} else {
			if !isStructuredOutput() {
				fmt.Fprintln(cmd.ErrOrStderr(), "Connection successful!")
			}
		}

		// Resolve actual name stored
		storedReg := cfg.GetRegistry(url)
		actualName := url
		if storedReg != nil {
			actualName = storedReg.Name
		}

		result := registryAddResult{
			Name:     actualName,
			URL:      url,
			Insecure: registryInsecure,
			Status:   "added",
			Message:  testMsg,
		}

		return printResult(result, func() {
			fmt.Printf("Registry %s (%s) added successfully.\n", actualName, url)
		})
	},
}

var registryRemoveCmd = &cobra.Command{
	Use:     "remove <url>",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a registry",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if cfg.GetRegistry(url) == nil {
			return fmt.Errorf("registry %s not found", url)
		}

		if err := cfg.RemoveRegistry(url); err != nil {
			return fmt.Errorf("failed to remove registry: %w", err)
		}

		result := registryRemoveResult{URL: url, Status: "removed"}
		return printResult(result, func() {
			fmt.Printf("Registry %s removed.\n", url)
		})
	},
}

var registryTestCmd = &cobra.Command{
	Use:   "test <url>",
	Short: "Test connectivity to a registry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		client := registry.NewClient(cfg)

		if !isStructuredOutput() {
			fmt.Fprintf(cmd.ErrOrStderr(), "Testing connection to %s...\n", url)
		}

		result := registryTestResult{URL: url}

		if err := client.TestRegistry(url); err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			if isStructuredOutput() {
				return printResult(result, nil)
			}
			return fmt.Errorf("connection failed: %w", err)
		}

		result.Status = "ok"
		return printResult(result, func() {
			fmt.Println("Connection successful!")
		})
	},
}

func init() {
	// Add flags to add command
	registryAddCmd.Flags().StringVarP(&registryName, "name", "n", "", "Display name for the registry")
	registryAddCmd.Flags().StringVarP(&registryUser, "user", "u", "", "Username for authentication")
	registryAddCmd.Flags().StringVarP(&registryPassword, "pass", "p", "", "Password for authentication")
	registryAddCmd.Flags().BoolVar(&registryInsecure, "insecure", false, "Allow insecure connections (HTTP)")

	// Build command hierarchy
	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registryAddCmd)
	registryCmd.AddCommand(registryRemoveCmd)
	registryCmd.AddCommand(registryTestCmd)

	rootCmd.AddCommand(registryCmd)
}
