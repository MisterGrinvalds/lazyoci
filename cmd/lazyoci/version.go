package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// These variables are set at build time via ldflags.
// goreleaser sets them automatically; local builds use defaults.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print the version, commit hash, and build date of lazyoci.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return printResult(versionInfo{
			Version: version,
			Commit:  commit,
			Date:    date,
		}, func() {
			fmt.Printf("lazyoci version %s (commit: %s, built: %s)\n", version, commit, date)
		})
	},
}

type versionInfo struct {
	Version string `json:"version" yaml:"version"`
	Commit  string `json:"commit" yaml:"commit"`
	Date    string `json:"date" yaml:"date"`
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
