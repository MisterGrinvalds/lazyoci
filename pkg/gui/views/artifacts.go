package views

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/mistergrinvalds/lazyoci/pkg/registry"
	"github.com/rivo/tview"
)

const pageSize = 20

// ArtifactView displays the list of artifacts (tags, manifests) for a repository
type ArtifactView struct {
	Flex        *tview.Flex
	Table       *tview.Table
	FilterInput *tview.InputField
	StatusText  *tview.TextView
	registry    *registry.Client
	onSelect    func(*registry.Artifact)
	app         *tview.Application

	currentRepo string
	artifacts   []*registry.Artifact
	filter      string
	offset      int
	totalCount  int
	loading     bool
	hasMore     bool
}

// NewArtifactView creates a new artifact list view
func NewArtifactView(reg *registry.Client, onSelect func(*registry.Artifact)) *ArtifactView {
	av := &ArtifactView{
		registry: reg,
		onSelect: onSelect,
	}

	av.setupUI()
	return av
}

func (av *ArtifactView) setupUI() {
	// Filter input
	av.FilterInput = tview.NewInputField().
		SetLabel(" Filter: ").
		SetFieldWidth(0).
		SetPlaceholder("Type to filter tags...")
	av.FilterInput.SetFieldBackgroundColor(tcell.ColorDarkSlateGray)

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
			if av.onSelect != nil {
				av.onSelect(av.artifacts[idx])
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
			// Typing in table redirects to filter
			app.SetFocus(av.FilterInput)
			av.FilterInput.SetText(string(event.Rune()))
			return nil
		}
		return event
	})
}

func (av *ArtifactView) setupHeaders() {
	headers := []string{"Tag", "Type", "Status"}
	for col, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false).
			SetExpansion(1)
		av.Table.SetCell(0, col, cell)
	}
}

func (av *ArtifactView) showWelcome() {
	av.Table.Clear()
	av.setupHeaders()
	av.StatusText.SetText("[gray]Select a repository to view artifacts[-]")
}

// LoadArtifacts loads artifacts for the given repository
func (av *ArtifactView) LoadArtifacts(repo string) {
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
	av.Table.SetCell(1, 0, tview.NewTableCell("[yellow]Loading...[-]").SetExpansion(3))
	av.StatusText.SetText("[yellow]Loading artifacts...[-]")

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
						av.StatusText.SetText(fmt.Sprintf("[gray]No tags matching '%s'[-]", av.filter))
					} else {
						av.StatusText.SetText("[gray]No artifacts found[-]")
					}
					av.Table.SetCell(1, 0, tview.NewTableCell("[gray]No artifacts found[-]").SetExpansion(3))
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
	av.Table.SetCell(lastRow, 0, tview.NewTableCell("[yellow]Loading more...[-]").SetExpansion(3))

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
					av.StatusText.SetText(fmt.Sprintf("[red]Error loading more: %v[-]", err))
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
					av.Table.SetCell(row, 0, tview.NewTableCell(artifact.Tag).SetExpansion(1))
					av.Table.SetCell(row, 1, tview.NewTableCell(string(artifact.Type)).SetExpansion(1))
					av.Table.SetCell(row, 2, tview.NewTableCell("[green]available[-]"))
				}

				// Add "Load more" if there are more
				if av.hasMore {
					loadMoreRow := av.Table.GetRowCount()
					av.Table.SetCell(loadMoreRow, 0, tview.NewTableCell("[cyan]▼ Load more...[-]").SetExpansion(3))
				}

				av.updateStatus()
			})
		}
	}()
}

func (av *ArtifactView) renderArtifacts() {
	for i, artifact := range av.artifacts {
		row := i + 1
		av.Table.SetCell(row, 0, tview.NewTableCell(artifact.Tag).SetExpansion(1))
		av.Table.SetCell(row, 1, tview.NewTableCell(string(artifact.Type)).SetExpansion(1))
		av.Table.SetCell(row, 2, tview.NewTableCell("[green]available[-]"))
	}

	// Add "Load more" row if there are more artifacts
	if av.hasMore {
		loadMoreRow := len(av.artifacts) + 1
		av.Table.SetCell(loadMoreRow, 0, tview.NewTableCell("[cyan]▼ Load more...[-]").SetExpansion(3))
	}

	if len(av.artifacts) > 0 {
		av.Table.Select(1, 0)
	}

	av.updateStatus()
}

func (av *ArtifactView) updateStatus() {
	showing := len(av.artifacts)
	status := fmt.Sprintf("[green]Showing %d artifacts[-]", showing)
	if av.hasMore {
		status += " [gray](more available)[-]"
	}
	if av.filter != "" {
		status += fmt.Sprintf(" [yellow]filter: %s[-]", av.filter)
	}
	av.StatusText.SetText(status)
}

func (av *ArtifactView) showError(err error) {
	errMsg := err.Error()
	if len(errMsg) > 60 {
		errMsg = errMsg[:57] + "..."
	}
	av.Table.SetCell(1, 0, tview.NewTableCell("[red]"+errMsg+"[-]").SetExpansion(3))
	av.StatusText.SetText("[red]Error loading artifacts[-]")
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
