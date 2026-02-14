package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// outputFormat is the global output format flag ("text", "json", "yaml").
var outputFormat string

// isStructuredOutput returns true if the output format is json or yaml.
func isStructuredOutput() bool {
	return outputFormat == "json" || outputFormat == "yaml"
}

// printResult outputs v as structured data (json/yaml) or calls textFn for
// human-readable text output. This is the primary output helper for all CLI commands.
func printResult(v any, textFn func()) error {
	if isStructuredOutput() {
		return formatOutput(v)
	}
	textFn()
	return nil
}

// formatOutput marshals v to stdout as JSON or YAML based on the global outputFormat.
func formatOutput(v any) error {
	switch outputFormat {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	case "yaml":
		enc := yaml.NewEncoder(os.Stdout)
		enc.SetIndent(2)
		defer enc.Close()
		return enc.Encode(v)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

// newTabWriter creates a standard tabwriter for text output.
func newTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
}
