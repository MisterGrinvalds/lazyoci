package views

import (
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/mistergrinvalds/lazyoci/pkg/gui/theme"
	"github.com/mistergrinvalds/lazyoci/pkg/registry"
	"github.com/rivo/tview"
)

const pageSize = 20

// ArtifactView displays the list of artifacts (tags, manifests) for a repository
type ArtifactView struct {
	Flex             *tview.Flex
	Table            *tview.Table
	FilterInput      *tview.InputField
	StatusText       *tview.TextView
	registry         *registry.Client
	onSelect         func(*registry.Artifact)
	onSelectWithInfo func(*registry.Artifact, *registry.ArtifactInfo)
	onPull           func(*registry.Artifact)       // Shows pull modal
	onPullDirect     func(*registry.Artifact, bool) // Direct pull: bool = toDocker
	app              *tview.Application

	currentRepo string
	artifacts   []*registry.Artifact
	filter      string
	offset      int
	totalCount  int
	loading     bool
	hasMore     bool

	// Type info cache: tag -> ArtifactInfo
	infoCache   map[string]*registry.ArtifactInfo
	infoCacheMu sync.RWMutex

	// Currently resolving tags (to avoid duplicate requests)
	resolving   map[string]bool
	resolvingMu sync.Mutex
}

// NewArtifactView creates a new artifact list view
func NewArtifactView(reg *registry.Client, onSelect func(*registry.Artifact)) *ArtifactView {
	av := &ArtifactView{
		registry:  reg,
		onSelect:  onSelect,
		infoCache: make(map[string]*registry.ArtifactInfo),
		resolving: make(map[string]bool),
	}

	av.setupUI()
	av.ApplyTheme()
	return av
}

func (av *ArtifactView) setupUI() {
	// Filter input
	av.FilterInput = tview.NewInputField().
		SetLabel(" Filter: ").
		SetFieldWidth(0).
		SetPlaceholder("Type to filter tags...")

	// Results table
	av.Table = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0)

	av.Table.SetSelectedFunc(func(row, col int) {
		if row == 0 {
			return
		}
		// Check if "Load more" row
		if av.hasMore && row == av.Table.GetRowCount()-1 {
			av.loadMore()
			return
		}
		idx := row - 1
		if idx >= 0 && idx < len(av.artifacts) {
			artifact := av.artifacts[idx]

			// Resolve artifact info if not cached
			av.resolveArtifactInfo(artifact, row)

			if av.onSelect != nil {
				av.onSelect(artifact)
			}
			// Also call onSelectWithInfo if set (for details view)
			if av.onSelectWithInfo != nil {
				info := av.getCachedInfo(artifact.Tag)
				av.onSelectWithInfo(artifact, info)
			}
		}
	})

	// Status bar
	av.StatusText = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	// Combined layout
	av.Flex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(av.FilterInput, 1, 0, false).
		AddItem(av.Table, 0, 1, true).
		AddItem(av.StatusText, 1, 0, false)

	av.Flex.SetBorder(true).SetTitle(" [3] Artifacts ")

	av.setupHeaders()
	av.showWelcome()
}

// ApplyTheme applies the current theme to this view's widgets.
func (av *ArtifactView) ApplyTheme() {
	// FilterInput
	av.FilterInput.SetBackgroundColor(theme.BackgroundColor())
	av.FilterInput.SetFieldBackgroundColor(theme.ElementBgColor())
	av.FilterInput.SetFieldTextColor(theme.TextColor())
	av.FilterInput.SetLabelColor(theme.TextColor())
	av.FilterInput.SetPlaceholderTextColor(theme.PlaceholderColor())

	// Table
	av.Table.SetBackgroundColor(theme.BackgroundColor())
	av.Table.SetSelectedStyle(tcell.StyleDefault.
		Background(theme.SelectionBgColor()).
		Foreground(theme.SelectionFgColor()))

	// StatusText
	av.StatusText.SetBackgroundColor(theme.BackgroundColor())
	av.StatusText.SetTextColor(theme.TextColor())

	// Flex (container with border)
	av.Flex.SetBackgroundColor(theme.BackgroundColor())
	av.Flex.SetBorderColor(theme.BorderNormalColor())
	av.Flex.SetTitleColor(theme.TitleColor())

	// Re-apply header colors
	av.setupHeaders()
}

// SetApp sets the application reference for async updates
func (av *ArtifactView) SetApp(app *tview.Application) {
	av.app = app

	// Setup filter input handler
	av.FilterInput.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			// Move focus to table for selection
			if len(av.artifacts) > 0 {
				app.SetFocus(av.Table)
				av.Table.Select(1, 0)
			}
		case tcell.KeyEscape:
			av.FilterInput.SetText("")
			av.filter = ""
			av.offset = 0
			av.loadArtifacts()
		case tcell.KeyDown:
			// Move to table
			if len(av.artifacts) > 0 {
				app.SetFocus(av.Table)
				av.Table.Select(1, 0)
			}
		}
	})

	// Handle arrow keys in filter input
	av.FilterInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyDown:
			if len(av.artifacts) > 0 {
				app.SetFocus(av.Table)
				av.Table.Select(1, 0)
			}
			return nil
		}
		return event
	})

	// Live filtering as user types
	av.FilterInput.SetChangedFunc(func(text string) {
		if av.currentRepo != "" {
			av.filter = text
			av.offset = 0
			av.loadArtifacts()
		}
	})

	// Handle going back to filter from table
	av.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := av.Table.GetSelection()
		switch event.Key() {
		case tcell.KeyUp:
			if row <= 1 {
				app.SetFocus(av.FilterInput)
				return nil
			}
		case tcell.KeyRune:
			switch event.Rune() {
			case 'p', 'P':
				// Pull the selected artifact (show modal)
				if av.onPull != nil && row > 0 && row-1 < len(av.artifacts) {
					av.onPull(av.artifacts[row-1])
				}
				return nil
			case 'd', 'D':
				// Direct pull to Docker (no modal)
				if av.onPullDirect != nil && row > 0 && row-1 < len(av.artifacts) {
					av.onPullDirect(av.artifacts[row-1], true)
				}
				return nil
			case 'j':
				// vim-style down
				if row < av.Table.GetRowCount()-1 {
					av.Table.Select(row+1, 0)
				}
				return nil
			case 'k':
				// vim-style up
				if row > 1 {
					av.Table.Select(row-1, 0)
				}
				return nil
			default:
				// Typing in table redirects to filter
				app.SetFocus(av.FilterInput)
				av.FilterInput.SetText(string(event.Rune()))
				return nil
			}
		}
		return event
	})
}

func (av *ArtifactView) setupHeaders() {
	headers := []string{"TAG", "TYPE", "STATUS"}
	for col, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(theme.HeaderColor()).
			SetSelectable(false).
			SetExpansion(1)
		av.Table.SetCell(0, col, cell)
	}
}

func (av *ArtifactView) showWelcome() {
	av.Table.Clear()
	av.setupHeaders()
	av.StatusText.SetText(theme.Tag("muted") + "Select a repository to view artifacts" + theme.ResetTag())
}

// LoadArtifacts loads artifacts for the given repository
func (av *ArtifactView) LoadArtifacts(repo string) {
	// Clear cache when switching repos
	if av.currentRepo != repo {
		av.infoCacheMu.Lock()
		av.infoCache = make(map[string]*registry.ArtifactInfo)
		av.infoCacheMu.Unlock()
	}

	av.currentRepo = repo
	av.offset = 0
	av.filter = ""
	av.FilterInput.SetText("")
	av.loadArtifacts()
}

func (av *ArtifactView) loadArtifacts() {
	if av.loading || av.currentRepo == "" {
		return
	}

	av.loading = true
	av.Table.Clear()
	av.setupHeaders()

	title := fmt.Sprintf(" [3] Artifacts: %s ", av.currentRepo)
	if av.filter != "" {
		title = fmt.Sprintf(" [3] Artifacts: %s (filter: %s) ", av.currentRepo, av.filter)
	}
	av.Flex.SetTitle(title)

	// Show loading indicator
	av.Table.SetCell(1, 0, tview.NewTableCell(theme.Tag("warning")+"Loading..."+theme.ResetTag()).SetExpansion(3))
	av.StatusText.SetText(theme.Tag("warning") + "Loading artifacts..." + theme.ResetTag())

	// Load in background
	go func() {
		opts := registry.ListArtifactsOptions{
			Limit:  pageSize,
			Offset: av.offset,
			Filter: av.filter,
		}
		artifacts, err := av.registry.ListArtifactsWithOptions(av.currentRepo, opts)

		if av.app != nil {
			av.app.QueueUpdateDraw(func() {
				av.loading = false
				av.Table.Clear()
				av.setupHeaders()

				if err != nil {
					av.showError(err)
					return
				}

				if len(artifacts) == 0 && av.offset == 0 {
					if av.filter != "" {
						av.StatusText.SetText(fmt.Sprintf("%sNo tags matching '%s'%s", theme.Tag("muted"), av.filter, theme.ResetTag()))
					} else {
						av.StatusText.SetText(theme.Tag("muted") + "No artifacts found" + theme.ResetTag())
					}
					av.Table.SetCell(1, 0, tview.NewTableCell(theme.Tag("muted")+"No artifacts found"+theme.ResetTag()).SetExpansion(3))
					return
				}

				av.artifacts = artifacts
				av.hasMore = len(artifacts) == pageSize
				av.renderArtifacts()
			})
		}
	}()
}

func (av *ArtifactView) loadMore() {
	if av.loading {
		return
	}

	av.offset += pageSize
	av.loading = true

	// Update the "Load more" row to show loading
	lastRow := av.Table.GetRowCount() - 1
	av.Table.SetCell(lastRow, 0, tview.NewTableCell(theme.Tag("warning")+"Loading more..."+theme.ResetTag()).SetExpansion(3))

	go func() {
		opts := registry.ListArtifactsOptions{
			Limit:  pageSize,
			Offset: av.offset,
			Filter: av.filter,
		}
		moreArtifacts, err := av.registry.ListArtifactsWithOptions(av.currentRepo, opts)

		if av.app != nil {
			av.app.QueueUpdateDraw(func() {
				av.loading = false

				if err != nil {
					av.StatusText.SetText(fmt.Sprintf("%sError loading more: %v%s", theme.Tag("error"), err, theme.ResetTag()))
					return
				}

				// Remove the "Load more" row
				av.Table.RemoveRow(av.Table.GetRowCount() - 1)

				// Append new artifacts
				av.artifacts = append(av.artifacts, moreArtifacts...)
				av.hasMore = len(moreArtifacts) == pageSize

				// Add new rows
				startRow := av.Table.GetRowCount()
				for i, artifact := range moreArtifacts {
					row := startRow + i
					av.Table.SetCell(row, 0, tview.NewTableCell(artifact.Tag).SetExpansion(1).SetTextColor(theme.TextColor()))

					// Check if type is cached, otherwise show "-"
					typeText := theme.ArtifactTypeTag("-")
					if info := av.getCachedInfo(artifact.Tag); info != nil {
						typeText = theme.ArtifactTypeTag(info.Type.Short())
					}
					av.Table.SetCell(row, 1, tview.NewTableCell(typeText).SetExpansion(1))
					av.Table.SetCell(row, 2, tview.NewTableCell(theme.StatusTag("available")))
				}

				// Add "Load more" if there are more
				if av.hasMore {
					loadMoreRow := av.Table.GetRowCount()
					av.Table.SetCell(loadMoreRow, 0, tview.NewTableCell(theme.Tag("info")+"▼ Load more..."+theme.ResetTag()).SetExpansion(3))
				}

				av.updateStatus()
			})
		}
	}()
}

func (av *ArtifactView) renderArtifacts() {
	for i, artifact := range av.artifacts {
		row := i + 1
		av.Table.SetCell(row, 0, tview.NewTableCell(artifact.Tag).SetExpansion(1).SetTextColor(theme.TextColor()))

		// Check if type is cached, otherwise show "-"
		typeText := theme.ArtifactTypeTag("-")
		if info := av.getCachedInfo(artifact.Tag); info != nil {
			typeText = theme.ArtifactTypeTag(info.Type.Short())
			// Update artifact with cached info
			artifact.Type = info.Type
			artifact.Digest = info.Digest
			artifact.Size = info.Size
		}
		av.Table.SetCell(row, 1, tview.NewTableCell(typeText).SetExpansion(1))
		av.Table.SetCell(row, 2, tview.NewTableCell(theme.StatusTag("available")))
	}

	// Add "Load more" row if there are more artifacts
	if av.hasMore {
		loadMoreRow := len(av.artifacts) + 1
		av.Table.SetCell(loadMoreRow, 0, tview.NewTableCell(theme.Tag("info")+"▼ Load more..."+theme.ResetTag()).SetExpansion(3))
	}

	if len(av.artifacts) > 0 {
		av.Table.Select(1, 0)
	}

	av.updateStatus()
}

func (av *ArtifactView) updateStatus() {
	showing := len(av.artifacts)
	status := fmt.Sprintf("%sShowing %d artifacts%s", theme.Tag("success"), showing, theme.ResetTag())
	if av.hasMore {
		status += " " + theme.Tag("muted") + "(more available)" + theme.ResetTag()
	}
	if av.filter != "" {
		status += fmt.Sprintf(" %sfilter: %s%s", theme.Tag("warning"), av.filter, theme.ResetTag())
	}
	av.StatusText.SetText(status)
}

func (av *ArtifactView) showError(err error) {
	errMsg := err.Error()
	if len(errMsg) > 60 {
		errMsg = errMsg[:57] + "..."
	}
	av.Table.SetCell(1, 0, tview.NewTableCell(theme.Tag("error")+errMsg+theme.ResetTag()).SetExpansion(3))
	av.StatusText.SetText(theme.Tag("error") + "Error loading artifacts" + theme.ResetTag())
}

// FocusFilter focuses the filter input
func (av *ArtifactView) FocusFilter() {
	if av.app != nil {
		av.app.SetFocus(av.FilterInput)
	}
}

// GetTable returns the table for focus management
func (av *ArtifactView) GetTable() *tview.Table {
	return av.Table
}

// SetOnPull sets the callback for pulling an artifact (shows modal)
func (av *ArtifactView) SetOnPull(fn func(*registry.Artifact)) {
	av.onPull = fn
}

// SetOnPullDirect sets the callback for direct pull (no modal)
func (av *ArtifactView) SetOnPullDirect(fn func(*registry.Artifact, bool)) {
	av.onPullDirect = fn
}

// SetOnSelectWithInfo sets the callback for selection with artifact info
func (av *ArtifactView) SetOnSelectWithInfo(fn func(*registry.Artifact, *registry.ArtifactInfo)) {
	av.onSelectWithInfo = fn
}

// getCachedInfo returns cached artifact info for a tag, or nil if not cached
func (av *ArtifactView) getCachedInfo(tag string) *registry.ArtifactInfo {
	av.infoCacheMu.RLock()
	defer av.infoCacheMu.RUnlock()
	return av.infoCache[tag]
}

// setCachedInfo stores artifact info in the cache
func (av *ArtifactView) setCachedInfo(tag string, info *registry.ArtifactInfo) {
	av.infoCacheMu.Lock()
	defer av.infoCacheMu.Unlock()
	av.infoCache[tag] = info
}

// isResolving checks if a tag is currently being resolved
func (av *ArtifactView) isResolving(tag string) bool {
	av.resolvingMu.Lock()
	defer av.resolvingMu.Unlock()
	return av.resolving[tag]
}

// setResolving marks a tag as being resolved
func (av *ArtifactView) setResolving(tag string, resolving bool) {
	av.resolvingMu.Lock()
	defer av.resolvingMu.Unlock()
	if resolving {
		av.resolving[tag] = true
	} else {
		delete(av.resolving, tag)
	}
}

// resolveArtifactInfo fetches detailed artifact info and updates the table
func (av *ArtifactView) resolveArtifactInfo(artifact *registry.Artifact, row int) {
	tag := artifact.Tag

	// Check if already cached
	if info := av.getCachedInfo(tag); info != nil {
		av.updateTypeCell(row, info.Type)
		return
	}

	// Check if already resolving
	if av.isResolving(tag) {
		return
	}

	// Mark as resolving and show loading indicator
	av.setResolving(tag, true)
	av.updateTypeCell(row, "...")

	// Resolve in background
	go func() {
		info, err := av.registry.GetArtifactInfo(av.currentRepo, tag)

		if av.app != nil {
			av.app.QueueUpdateDraw(func() {
				av.setResolving(tag, false)

				if err != nil {
					av.updateTypeCell(row, "?")
					return
				}

				// Cache the result
				av.setCachedInfo(tag, info)

				// Update the table cell
				av.updateTypeCell(row, info.Type.Short())

				// Update artifact with resolved info
				artifact.Type = info.Type
				artifact.Digest = info.Digest
				artifact.Size = info.Size

				// If this artifact is still selected, update details
				if av.onSelectWithInfo != nil {
					selectedRow, _ := av.Table.GetSelection()
					if selectedRow == row {
						av.onSelectWithInfo(artifact, info)
					}
				}
			})
		}
	}()
}

// updateTypeCell updates the type column for a specific row
func (av *ArtifactView) updateTypeCell(row int, typeStr interface{}) {
	var text string
	switch t := typeStr.(type) {
	case string:
		text = t
	case registry.ArtifactType:
		text = t.Short()
	default:
		text = fmt.Sprintf("%v", t)
	}

	colored := theme.ArtifactTypeTag(text)

	if row > 0 && row < av.Table.GetRowCount() {
		av.Table.GetCell(row, 1).SetText(colored)
	}
}

// GetSelectedArtifact returns the currently selected artifact, if any
func (av *ArtifactView) GetSelectedArtifact() *registry.Artifact {
	row, _ := av.Table.GetSelection()
	if row > 0 && row-1 < len(av.artifacts) {
		return av.artifacts[row-1]
	}
	return nil
}

func truncateDigest(digest string) string {
	if len(digest) > 19 {
		return digest[:19] + "..."
	}
	return digest
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
