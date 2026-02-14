package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/mistergrinvalds/lazyoci/pkg/config"
	"github.com/mistergrinvalds/lazyoci/pkg/registry"
	"github.com/spf13/cobra"
)

// browse response types for structured output.

type repoItem struct {
	Namespace  string `json:"namespace" yaml:"namespace"`
	Repository string `json:"repository" yaml:"repository"`
	FullPath   string `json:"fullPath" yaml:"fullPath"`
}

type tagItem struct {
	Tag    string `json:"tag" yaml:"tag"`
	Type   string `json:"type,omitempty" yaml:"type,omitempty"`
	Digest string `json:"digest,omitempty" yaml:"digest,omitempty"`
	Size   int64  `json:"size,omitempty" yaml:"size,omitempty"`
}

type manifestResult struct {
	Repository string            `json:"repository" yaml:"repository"`
	Tag        string            `json:"tag" yaml:"tag"`
	Digest     string            `json:"digest" yaml:"digest"`
	Size       int64             `json:"size" yaml:"size"`
	Type       string            `json:"type" yaml:"type"`
	MediaType  string            `json:"mediaType,omitempty" yaml:"mediaType,omitempty"`
	Platform   string            `json:"platform,omitempty" yaml:"platform,omitempty"`
	Created    *time.Time        `json:"created,omitempty" yaml:"created,omitempty"`
	Labels     map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

type searchItem struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Stars       int    `json:"stars,omitempty" yaml:"stars,omitempty"`
	Pulls       int64  `json:"pulls,omitempty" yaml:"pulls,omitempty"`
	Official    bool   `json:"official,omitempty" yaml:"official,omitempty"`
	Registry    string `json:"registry" yaml:"registry"`
}

var (
	browseLimit  int
	browseOffset int
	browseFilter string
)

var browseCmd = &cobra.Command{
	Use:   "browse",
	Short: "Browse OCI registry content",
	Long: `Browse repositories, tags, manifests, and search across OCI registries.

All commands support --output json|yaml|text for structured output.`,
}

var browseReposCmd = &cobra.Command{
	Use:   "repos <registry-url>",
	Short: "List repositories in a registry",
	Long: `List all namespaces and repositories in an OCI registry.

Examples:
  lazyoci browse repos localhost:5050
  lazyoci browse repos docker.io -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		registryURL := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		client := registry.NewClient(cfg)

		// Some registries (Docker Hub, GHCR) don't support the _catalog API.
		// Detect this and provide a helpful error instead of a cryptic 400.
		namespaces, err := client.ListNamespaces(registryURL)
		if err != nil {
			switch registryURL {
			case "docker.io":
				return fmt.Errorf("Docker Hub does not support listing all repositories.\n  Use 'lazyoci browse search docker.io <query>' instead")
			case "ghcr.io":
				return fmt.Errorf("GitHub Packages does not support listing all repositories.\n  Use 'lazyoci browse tags ghcr.io/<owner>/<repo>' for a known repository")
			default:
				return fmt.Errorf("failed to list namespaces: %w", err)
			}
		}

		var items []repoItem

		for _, ns := range namespaces {
			repos, err := client.ListRepositories(registryURL, ns)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to list repos in %s: %v\n", ns, err)
				continue
			}
			if len(repos) == 0 {
				// Namespace exists but no sub-repos — treat namespace as a repo
				items = append(items, repoItem{
					Namespace:  ns,
					Repository: "",
					FullPath:   registryURL + "/" + ns,
				})
				continue
			}
			for _, repo := range repos {
				items = append(items, repoItem{
					Namespace:  ns,
					Repository: repo,
					FullPath:   registryURL + "/" + ns + "/" + repo,
				})
			}
		}

		return printResult(items, func() {
			w := newTabWriter()
			fmt.Fprintln(w, "NAMESPACE\tREPOSITORY\tFULL PATH")
			fmt.Fprintln(w, "---------\t----------\t---------")
			for _, item := range items {
				repo := item.Repository
				if repo == "" {
					repo = "-"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", item.Namespace, repo, item.FullPath)
			}
			w.Flush()
		})
	},
}

var browseTagsCmd = &cobra.Command{
	Use:   "tags <registry/repo>",
	Short: "List tags in a repository",
	Long: `List artifact tags in an OCI repository with optional pagination and filtering.

Examples:
  lazyoci browse tags localhost:5050/test/hello
  lazyoci browse tags docker.io/library/nginx --limit 10
  lazyoci browse tags docker.io/library/nginx --filter alpine -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		client := registry.NewClient(cfg)

		opts := registry.ListArtifactsOptions{
			Limit:  browseLimit,
			Offset: browseOffset,
			Filter: browseFilter,
		}

		artifacts, err := client.ListArtifactsWithOptions(repoPath, opts)
		if err != nil {
			return fmt.Errorf("failed to list tags: %w", err)
		}

		var items []tagItem
		for _, a := range artifacts {
			items = append(items, tagItem{
				Tag:    a.Tag,
				Type:   string(a.Type),
				Digest: a.Digest,
				Size:   a.Size,
			})
		}

		return printResult(items, func() {
			w := newTabWriter()
			fmt.Fprintln(w, "TAG\tTYPE\tDIGEST\tSIZE")
			fmt.Fprintln(w, "---\t----\t------\t----")
			for _, item := range items {
				digest := item.Digest
				if len(digest) > 19 {
					digest = digest[:19] + "..."
				}
				size := "-"
				if item.Size > 0 {
					size = formatBytes(item.Size)
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", item.Tag, item.Type, digest, size)
			}
			w.Flush()
		})
	},
}

var browseManifestCmd = &cobra.Command{
	Use:   "manifest <registry/repo:tag>",
	Short: "Show manifest details for an artifact",
	Long: `Resolve a tag and display its full manifest details.

Examples:
  lazyoci browse manifest localhost:5050/test/hello:v1
  lazyoci browse manifest docker.io/library/nginx:latest -o yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]

		// Parse registry/repo:tag
		colonIdx := strings.LastIndex(ref, ":")
		if colonIdx == -1 {
			return fmt.Errorf("invalid reference %q: expected <registry/repo:tag>", ref)
		}

		repoPath := ref[:colonIdx]
		tag := ref[colonIdx+1:]

		// Avoid splitting on port numbers (e.g. localhost:5050/test/hello:v1)
		// If there's no "/" after the colon, it's a port, not a tag
		if !strings.Contains(repoPath, "/") {
			return fmt.Errorf("invalid reference %q: expected <registry/repo:tag>", ref)
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		client := registry.NewClient(cfg)

		artifact, err := client.GetArtifactDetails(repoPath, tag)
		if err != nil {
			return fmt.Errorf("failed to resolve manifest: %w", err)
		}

		result := manifestResult{
			Repository: artifact.Repository,
			Tag:        artifact.Tag,
			Digest:     artifact.Digest,
			Size:       artifact.Size,
			Type:       string(artifact.Type),
			MediaType:  artifact.MediaType,
			Platform:   artifact.Platform,
			Labels:     artifact.Labels,
		}
		if !artifact.Created.IsZero() {
			result.Created = &artifact.Created
		}

		return printResult(result, func() {
			fmt.Printf("Repository:  %s\n", result.Repository)
			fmt.Printf("Tag:         %s\n", result.Tag)
			fmt.Printf("Digest:      %s\n", result.Digest)
			fmt.Printf("Size:        %s\n", formatBytes(result.Size))
			fmt.Printf("Type:        %s\n", result.Type)
			if result.MediaType != "" {
				fmt.Printf("Media Type:  %s\n", result.MediaType)
			}
			if result.Platform != "" {
				fmt.Printf("Platform:    %s\n", result.Platform)
			}
			if result.Created != nil {
				fmt.Printf("Created:     %s\n", result.Created.Format(time.RFC3339))
			}
			if len(result.Labels) > 0 {
				fmt.Println("Labels:")
				for k, v := range result.Labels {
					fmt.Printf("  %s: %s\n", k, v)
				}
			}
		})
	},
}

var browseSearchCmd = &cobra.Command{
	Use:   "search <registry> <query>",
	Short: "Search for repositories in a registry",
	Long: `Search for repositories across an OCI registry.

Examples:
  lazyoci browse search docker.io nginx
  lazyoci browse search quay.io prometheus -o json`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		registryURL := args[0]
		query := args[1]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		client := registry.NewClient(cfg)

		results, err := client.Search(registryURL, query)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		var items []searchItem
		for _, r := range results {
			items = append(items, searchItem{
				Name:        r.Name,
				Description: r.Description,
				Stars:       r.StarCount,
				Pulls:       r.PullCount,
				Official:    r.IsOfficial,
				Registry:    r.RegistryURL,
			})
		}

		return printResult(items, func() {
			w := newTabWriter()
			fmt.Fprintln(w, "NAME\tDESCRIPTION\tSTARS\tPULLS\tOFFICIAL")
			fmt.Fprintln(w, "----\t-----------\t-----\t-----\t--------")
			for _, item := range items {
				desc := item.Description
				if len(desc) > 50 {
					desc = desc[:47] + "..."
				}
				official := "-"
				if item.Official {
					official = "yes"
				}
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
					item.Name, desc, item.Stars,
					registry.FormatPullCount(item.Pulls), official)
			}
			w.Flush()
		})
	},
}

func formatBytes(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(b)/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func init() {
	browseTagsCmd.Flags().IntVar(&browseLimit, "limit", 20, "Maximum number of tags to return")
	browseTagsCmd.Flags().IntVar(&browseOffset, "offset", 0, "Number of tags to skip")
	browseTagsCmd.Flags().StringVar(&browseFilter, "filter", "", "Filter tags containing this string")

	browseCmd.AddCommand(browseReposCmd)
	browseCmd.AddCommand(browseTagsCmd)
	browseCmd.AddCommand(browseManifestCmd)
	browseCmd.AddCommand(browseSearchCmd)

	rootCmd.AddCommand(browseCmd)
}
