package views

import (
	"fmt"
	"strings"

	"github.com/mistergrinvalds/lazyoci/pkg/registry"
	"github.com/rivo/tview"
)

// DetailsView displays detailed information about the current context
type DetailsView struct {
	TextView *tview.TextView
}

// NewDetailsView creates a new details view
func NewDetailsView() *DetailsView {
	dv := &DetailsView{}

	dv.TextView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)

	dv.TextView.SetBorder(true).SetTitle(" Details ")
	dv.ShowRegistryHelp()

	return dv
}

// ShowRegistryHelp shows help for the registry view
func (dv *DetailsView) ShowRegistryHelp() {
	dv.TextView.SetTitle(" Details ")
	dv.TextView.SetText(`[yellow]Welcome to lazyoci![white]

Browse OCI registries to find container images,
Helm charts, and other artifacts.

[green]Configured Registries:[white]
  • docker.io  - Docker Hub
  • quay.io    - Red Hat Quay
  • ghcr.io    - GitHub Container Registry

[green]Navigation:[white]
  [yellow]1[white]       Focus registries
  [yellow]2[white] or [yellow]/[white] Focus search
  [yellow]3[white]       Focus artifacts
  [yellow]Tab[white]     Cycle panels
  [yellow]Enter[white]   Select/expand
  [yellow]j/k[white]     Move down/up
  [yellow]?[white]       Help
  [yellow]q[white]       Quit

[green]Quick Start:[white]
  1. Press [yellow]/[white] to search
  2. Type "nginx" and press Enter
  3. Select a result to view tags`)
}

// ShowRepository shows details for a repository
func (dv *DetailsView) ShowRepository(repoPath string) {
	dv.TextView.SetTitle(" Repository ")

	parts := strings.SplitN(repoPath, "/", 2)
	registry := parts[0]
	name := ""
	if len(parts) > 1 {
		name = parts[1]
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "[yellow]%s[white]\n\n", repoPath)
	fmt.Fprintf(&sb, "[green]Registry:[white]  %s\n", registry)
	fmt.Fprintf(&sb, "[green]Name:[white]      %s\n", name)
	sb.WriteString("\n[gray]Loading artifacts...[-]\n")

	sb.WriteString("\n[green]Commands:[white]\n")
	fmt.Fprintf(&sb, "  docker pull %s\n", repoPath)

	dv.TextView.SetText(sb.String())
	dv.TextView.ScrollToBeginning()
}

// ShowArtifact displays details for the given artifact
func (dv *DetailsView) ShowArtifact(artifact *registry.Artifact) {
	if artifact == nil {
		dv.ShowRegistryHelp()
		return
	}

	dv.TextView.SetTitle(" Artifact Details ")

	var sb strings.Builder

	fmt.Fprintf(&sb, "[yellow]%s:%s[white]\n\n", artifact.Repository, artifact.Tag)

	fmt.Fprintf(&sb, "[green]Type:[white]     %s\n", artifact.Type)
	fmt.Fprintf(&sb, "[green]Digest:[white]   %s\n", artifact.Digest)
	fmt.Fprintf(&sb, "[green]Size:[white]     %s\n", formatSize(artifact.Size))

	if !artifact.Created.IsZero() {
		fmt.Fprintf(&sb, "[green]Created:[white]  %s\n", artifact.Created.Format("2006-01-02 15:04:05 MST"))
	}

	if artifact.Platform != "" {
		fmt.Fprintf(&sb, "[green]Platform:[white] %s\n", artifact.Platform)
	}

	if len(artifact.Labels) > 0 {
		sb.WriteString("\n[yellow]Labels:[white]\n")
		for k, v := range artifact.Labels {
			fmt.Fprintf(&sb, "  %s: %s\n", k, v)
		}
	}

	if len(artifact.Layers) > 0 {
		fmt.Fprintf(&sb, "\n[yellow]Layers (%d):[white]\n", len(artifact.Layers))
		for i, layer := range artifact.Layers {
			if i >= 10 {
				fmt.Fprintf(&sb, "  ... and %d more\n", len(artifact.Layers)-10)
				break
			}
			fmt.Fprintf(&sb, "  %d. %s (%s)\n", i+1, truncateDigest(layer.Digest), formatSize(layer.Size))
		}
	}

	sb.WriteString("\n[yellow]Pull Commands:[white]\n")
	fmt.Fprintf(&sb, "  [cyan]docker pull %s:%s[white]\n", artifact.Repository, artifact.Tag)
	fmt.Fprintf(&sb, "  [cyan]docker pull %s@%s[white]\n", artifact.Repository, artifact.Digest)

	dv.TextView.SetText(sb.String())
	dv.TextView.ScrollToBeginning()
}

// ShowSearchHelp shows help for searching
func (dv *DetailsView) ShowSearchHelp(registry string) {
	dv.TextView.SetTitle(" Search ")

	var sb strings.Builder
	fmt.Fprintf(&sb, "[yellow]Search %s[white]\n\n", registry)

	sb.WriteString("[green]Tips:[white]\n")
	sb.WriteString("  • Type a search term and press Enter\n")
	sb.WriteString("  • Use specific names: nginx, postgres, redis\n")
	sb.WriteString("  • Include namespace: bitnami/nginx\n")
	sb.WriteString("\n[green]Examples:[white]\n")
	sb.WriteString("  nginx          Official nginx image\n")
	sb.WriteString("  bitnami/       Bitnami images\n")
	sb.WriteString("  postgres       PostgreSQL database\n")

	dv.TextView.SetText(sb.String())
}

// Clear clears the details view
func (dv *DetailsView) Clear() {
	dv.ShowRegistryHelp()
}
