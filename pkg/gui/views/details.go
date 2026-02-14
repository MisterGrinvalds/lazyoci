package views

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mistergrinvalds/lazyoci/pkg/gui/theme"
	"github.com/mistergrinvalds/lazyoci/pkg/registry"
	"github.com/rivo/tview"
)

// DetailsView displays detailed information about the current context
type DetailsView struct {
	TextView        *tview.TextView
	currentInfo     *registry.ArtifactInfo
	currentArtifact *registry.Artifact

	// Callbacks for actions
	onPull       func(*registry.Artifact)       // Shows pull modal
	onPullDirect func(*registry.Artifact, bool) // Direct pull: bool = toDocker
}

// NewDetailsView creates a new details view
func NewDetailsView() *DetailsView {
	dv := &DetailsView{}

	dv.TextView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)

	dv.TextView.SetBorder(true).SetTitle(" [4] Details ")

	// Apply theme styling
	dv.ApplyTheme()

	// Setup input capture for keybindings when focused
	dv.TextView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'p', 'P':
				// Pull with modal
				if dv.onPull != nil && dv.currentArtifact != nil {
					dv.onPull(dv.currentArtifact)
				}
				return nil
			case 'd', 'D':
				// Direct pull to Docker
				if dv.onPullDirect != nil && dv.currentArtifact != nil {
					dv.onPullDirect(dv.currentArtifact, true)
				}
				return nil
			case 'j':
				// Scroll down
				row, col := dv.TextView.GetScrollOffset()
				dv.TextView.ScrollTo(row+1, col)
				return nil
			case 'k':
				// Scroll up
				row, col := dv.TextView.GetScrollOffset()
				if row > 0 {
					dv.TextView.ScrollTo(row-1, col)
				}
				return nil
			case 'G':
				// Scroll to end
				dv.TextView.ScrollToEnd()
				return nil
			case 'g':
				// Scroll to beginning (single 'g' for simplicity)
				dv.TextView.ScrollToBeginning()
				return nil
			}
		}
		return event
	})

	dv.ShowRegistryHelp()

	return dv
}

// ApplyTheme applies the current theme to this view's widgets.
func (dv *DetailsView) ApplyTheme() {
	dv.TextView.SetBackgroundColor(theme.BackgroundColor())
	dv.TextView.SetTextColor(theme.TextColor())
	dv.TextView.SetBorderColor(theme.BorderNormalColor())
	dv.TextView.SetTitleColor(theme.TitleColor())
}

// SetOnPull sets the callback for pull with modal
func (dv *DetailsView) SetOnPull(fn func(*registry.Artifact)) {
	dv.onPull = fn
}

// SetOnPullDirect sets the callback for direct pull
func (dv *DetailsView) SetOnPullDirect(fn func(*registry.Artifact, bool)) {
	dv.onPullDirect = fn
}

// GetCurrentArtifact returns the currently displayed artifact
func (dv *DetailsView) GetCurrentArtifact() *registry.Artifact {
	return dv.currentArtifact
}

// tag helpers for readability
func t(name string) string { return theme.Tag(name) }
func r() string            { return theme.ResetTag() }

// ShowRegistryHelp shows help for the registry view
func (dv *DetailsView) ShowRegistryHelp() {
	dv.currentArtifact = nil
	dv.currentInfo = nil
	dv.TextView.SetTitle(" [4] Details ")

	emphasis := t("emphasis")
	text := t("text")
	success := t("success")

	dv.TextView.SetText(fmt.Sprintf(`%sWelcome to lazyoci!%s

Browse OCI registries to find container images,
Helm charts, and other artifacts.

%sConfigured Registries:%s
  • docker.io  - Docker Hub
  • quay.io    - Red Hat Quay
  • ghcr.io    - GitHub Container Registry

%sNavigation:%s
  %s1%s       Focus registries
  %s2%s or %s/%s Focus search
  %s3%s       Focus artifacts
  %sTab%s     Cycle panels
  %sEnter%s   Select/expand
  %sj/k%s     Move down/up
  %s?%s       Help
  %sq%s       Quit

%sQuick Start:%s
  1. Press %s/%s to search
  2. Type "nginx" and press Enter
  3. Select a result to view tags`,
		emphasis, text,
		success, text,
		success, text,
		emphasis, text, emphasis, text, emphasis, text,
		emphasis, text,
		emphasis, text,
		emphasis, text,
		emphasis, text,
		emphasis, text,
		emphasis, text,
		success, text,
		emphasis, text,
	))
}

// ShowRegistryInfo shows info for a selected registry
func (dv *DetailsView) ShowRegistryInfo(registryURL string) {
	dv.currentArtifact = nil
	dv.currentInfo = nil
	dv.TextView.SetTitle(" [4] Registry ")

	emphasis := t("emphasis")
	text := t("text")
	success := t("success")

	var sb strings.Builder
	fmt.Fprintf(&sb, "%s%s%s\n\n", emphasis, registryURL, text)

	fmt.Fprintf(&sb, "%sStatus:%s Connected\n\n", success, text)

	fmt.Fprintf(&sb, "%sSearch:%s\n", success, text)
	sb.WriteString("  Type in the search box and press Enter\n")
	sb.WriteString("  to find repositories.\n\n")

	fmt.Fprintf(&sb, "%sExamples:%s\n", success, text)
	switch registryURL {
	case "docker.io":
		sb.WriteString("  nginx, postgres, redis\n")
		sb.WriteString("  bitnami/nginx, library/alpine\n")
	case "quay.io":
		sb.WriteString("  prometheus, grafana\n")
		sb.WriteString("  coreos/etcd\n")
	case "ghcr.io":
		sb.WriteString("  Enter owner/repo directly\n")
		sb.WriteString("  (search not supported)\n")
	default:
		sb.WriteString("  Enter repository name to search\n")
	}

	dv.TextView.SetText(sb.String())
	dv.TextView.ScrollToBeginning()
}

// ShowRepository shows details for a repository
func (dv *DetailsView) ShowRepository(repoPath string) {
	dv.currentArtifact = nil
	dv.currentInfo = nil
	dv.TextView.SetTitle(" [4] Repository ")

	parts := strings.SplitN(repoPath, "/", 2)
	reg := parts[0]
	name := ""
	if len(parts) > 1 {
		name = parts[1]
	}

	emphasis := t("emphasis")
	text := t("text")
	success := t("success")
	muted := t("muted")

	var sb strings.Builder
	fmt.Fprintf(&sb, "%s%s%s\n\n", emphasis, repoPath, text)
	fmt.Fprintf(&sb, "%sRegistry:%s  %s\n", success, text, reg)
	fmt.Fprintf(&sb, "%sName:%s      %s\n", success, text, name)
	fmt.Fprintf(&sb, "\n%sLoading artifacts...%s\n", muted, r())

	fmt.Fprintf(&sb, "\n%sCommands:%s\n", success, text)
	fmt.Fprintf(&sb, "  docker pull %s\n", repoPath)

	dv.TextView.SetText(sb.String())
	dv.TextView.ScrollToBeginning()
}

// ShowArtifact displays details for the given artifact
func (dv *DetailsView) ShowArtifact(artifact *registry.Artifact) {
	dv.ShowArtifactWithInfo(artifact, nil)
}

// ShowArtifactWithInfo displays details for the artifact with resolved info
func (dv *DetailsView) ShowArtifactWithInfo(artifact *registry.Artifact, info *registry.ArtifactInfo) {
	if artifact == nil {
		dv.ShowRegistryHelp()
		return
	}

	dv.currentArtifact = artifact
	dv.currentInfo = info

	dv.TextView.SetTitle(" [4] Artifact Details ")

	emphasis := t("emphasis")
	text := t("text")
	success := t("success")
	muted := t("muted")

	var sb strings.Builder

	fmt.Fprintf(&sb, "%s%s:%s%s\n\n", emphasis, artifact.Repository, artifact.Tag, text)

	// Show type with full name if info is available
	if info != nil {
		fmt.Fprintf(&sb, "%sType:%s     %s\n", success, text, info.Type.String())
		if info.TypeDetail != "" {
			fmt.Fprintf(&sb, "%sFormat:%s   %s\n", success, text, info.TypeDetail)
		}
		fmt.Fprintf(&sb, "%sMedia:%s    %s\n", success, text, truncateMediaType(info.MediaType))
		if info.ConfigMediaType != "" {
			fmt.Fprintf(&sb, "%sConfig:%s   %s\n", success, text, truncateMediaType(info.ConfigMediaType))
		}
		fmt.Fprintf(&sb, "%sDigest:%s   %s\n", success, text, truncateDigest(info.Digest))
		fmt.Fprintf(&sb, "%sSize:%s     %s\n", success, text, formatSize(info.Size))
		fmt.Fprintf(&sb, "%sLayers:%s   %d\n", success, text, info.Layers)
	} else {
		// Basic info from artifact (type not yet resolved)
		typeStr := string(artifact.Type)
		if typeStr == "" {
			typeStr = muted + "resolving..." + r()
		}
		fmt.Fprintf(&sb, "%sType:%s     %s\n", success, text, typeStr)
		if artifact.Digest != "" {
			fmt.Fprintf(&sb, "%sDigest:%s   %s\n", success, text, truncateDigest(artifact.Digest))
		}
		if artifact.Size > 0 {
			fmt.Fprintf(&sb, "%sSize:%s     %s\n", success, text, formatSize(artifact.Size))
		}
	}

	if !artifact.Created.IsZero() {
		fmt.Fprintf(&sb, "%sCreated:%s  %s\n", success, text, artifact.Created.Format("2006-01-02 15:04:05 MST"))
	}

	if artifact.Platform != "" {
		fmt.Fprintf(&sb, "%sPlatform:%s %s\n", success, text, artifact.Platform)
	}

	// Show annotations if available
	if info != nil && len(info.Annotations) > 0 {
		fmt.Fprintf(&sb, "\n%sAnnotations:%s\n", emphasis, text)
		count := 0
		for k, v := range info.Annotations {
			if count >= 5 {
				fmt.Fprintf(&sb, "  ... and %d more\n", len(info.Annotations)-5)
				break
			}
			// Truncate long values
			if len(v) > 40 {
				v = v[:37] + "..."
			}
			fmt.Fprintf(&sb, "  %s: %s\n", truncateKey(k), v)
			count++
		}
	}

	// Type-specific actions section
	sb.WriteString("\n")
	dv.writeActionsSection(&sb, artifact, info)

	dv.TextView.SetText(sb.String())
	dv.TextView.ScrollToBeginning()
}

// writeActionsSection writes type-specific actions to the string builder
func (dv *DetailsView) writeActionsSection(sb *strings.Builder, artifact *registry.Artifact, info *registry.ArtifactInfo) {
	emphasis := t("emphasis")
	text := t("text")
	success := t("success")
	muted := t("muted")
	dim := t("dim")

	fmt.Fprintf(sb, "%s━━━ Actions ━━━━━━━━━━━━━━━━━━━━━%s\n", emphasis, text)

	artifactType := registry.ArtifactTypeImage
	if info != nil {
		artifactType = info.Type
	} else if artifact.Type != "" {
		artifactType = artifact.Type
	}

	switch artifactType {
	case registry.ArtifactTypeImage:
		fmt.Fprintf(sb, "%sp%s Pull to disk\n", success, text)
		fmt.Fprintf(sb, "%sd%s Pull & load to Docker\n", success, text)
		fmt.Fprintf(sb, "\n%sPull commands:%s\n", muted, text)
		fmt.Fprintf(sb, "  docker pull %s:%s\n", artifact.Repository, artifact.Tag)

	case registry.ArtifactTypeHelmChart:
		fmt.Fprintf(sb, "%sp%s Pull chart.tgz\n", success, text)
		fmt.Fprintf(sb, "%st%s Template (helm template) %s(coming soon)%s\n", muted, text, dim, r())
		fmt.Fprintf(sb, "%si%s Install (helm install) %s(coming soon)%s\n", muted, text, dim, r())
		fmt.Fprintf(sb, "\n%sPull commands:%s\n", muted, text)
		fmt.Fprintf(sb, "  helm pull oci://%s --version %s\n", artifact.Repository, artifact.Tag)

	case registry.ArtifactTypeSBOM:
		fmt.Fprintf(sb, "%sp%s Pull JSON\n", success, text)
		fmt.Fprintf(sb, "%sv%s View summary %s(coming soon)%s\n", muted, text, dim, r())
		detail := ""
		if info != nil && info.TypeDetail != "" {
			detail = fmt.Sprintf(" (%s)", info.TypeDetail)
		}
		fmt.Fprintf(sb, "\n%sSBOM format:%s%s\n", muted, detail, text)

	case registry.ArtifactTypeSignature:
		fmt.Fprintf(sb, "%sp%s Pull signature\n", success, text)
		fmt.Fprintf(sb, "%sv%s Verify %s(requires cosign)%s\n", muted, text, dim, r())
		if info != nil && info.TypeDetail != "" {
			fmt.Fprintf(sb, "\n%sSignature type: %s%s\n", muted, info.TypeDetail, text)
		}

	case registry.ArtifactTypeAttestation:
		fmt.Fprintf(sb, "%sp%s Pull attestation\n", success, text)
		fmt.Fprintf(sb, "%sv%s View %s(coming soon)%s\n", muted, text, dim, r())
		if info != nil && info.TypeDetail != "" {
			fmt.Fprintf(sb, "\n%sAttestation type: %s%s\n", muted, info.TypeDetail, text)
		}

	case registry.ArtifactTypeWasm:
		fmt.Fprintf(sb, "%sp%s Pull .wasm\n", success, text)
		fmt.Fprintf(sb, "%sr%s Run %s(requires wasmtime)%s\n", muted, text, dim, r())

	default:
		fmt.Fprintf(sb, "%sp%s Pull (OCI layout)\n", success, text)
		fmt.Fprintf(sb, "\n%sUnknown artifact type%s\n", muted, text)
	}
}

// GetCurrentArtifactType returns the current artifact type (for keybinding context)
func (dv *DetailsView) GetCurrentArtifactType() registry.ArtifactType {
	if dv.currentInfo != nil {
		return dv.currentInfo.Type
	}
	if dv.currentArtifact != nil {
		return dv.currentArtifact.Type
	}
	return registry.ArtifactTypeUnknown
}

// truncateMediaType truncates a media type for display
func truncateMediaType(mediaType string) string {
	if len(mediaType) > 50 {
		return "..." + mediaType[len(mediaType)-47:]
	}
	return mediaType
}

// truncateKey truncates an annotation key for display
func truncateKey(key string) string {
	// Remove common prefixes for cleaner display
	key = strings.TrimPrefix(key, "org.opencontainers.image.")
	key = strings.TrimPrefix(key, "io.containerd.")
	key = strings.TrimPrefix(key, "com.docker.")

	if len(key) > 25 {
		return key[:22] + "..."
	}
	return key
}

// ShowSearchHelp shows help for searching
func (dv *DetailsView) ShowSearchHelp(reg string) {
	dv.currentArtifact = nil
	dv.currentInfo = nil
	dv.TextView.SetTitle(" [4] Search ")

	emphasis := t("emphasis")
	text := t("text")
	success := t("success")

	var sb strings.Builder
	fmt.Fprintf(&sb, "%sSearch %s%s\n\n", emphasis, reg, text)

	fmt.Fprintf(&sb, "%sTips:%s\n", success, text)
	sb.WriteString("  • Type a search term and press Enter\n")
	sb.WriteString("  • Use specific names: nginx, postgres, redis\n")
	sb.WriteString("  • Include namespace: bitnami/nginx\n")
	fmt.Fprintf(&sb, "\n%sExamples:%s\n", success, text)
	sb.WriteString("  nginx          Official nginx image\n")
	sb.WriteString("  bitnami/       Bitnami images\n")
	sb.WriteString("  postgres       PostgreSQL database\n")

	dv.TextView.SetText(sb.String())
}

// Clear clears the details view
func (dv *DetailsView) Clear() {
	dv.ShowRegistryHelp()
}
