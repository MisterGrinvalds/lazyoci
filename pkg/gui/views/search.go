package views

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/mistergrinvalds/lazyoci/pkg/gui/theme"
	"github.com/mistergrinvalds/lazyoci/pkg/registry"
	"github.com/rivo/tview"
)

// SearchView displays search results
type SearchView struct {
	Table      *tview.Table
	InputField *tview.InputField
	Flex       *tview.Flex
	registry   *registry.Client
	onSelect   func(registryURL, repoName string)
	results    []*registry.SearchResult
	currentReg string
	statusText *tview.TextView
	app        *tview.Application
}

// NewSearchView creates a new search view
func NewSearchView(reg *registry.Client, onSelect func(registryURL, repoName string)) *SearchView {
	sv := &SearchView{
		registry: reg,
		onSelect: onSelect,
	}

	sv.setupUI()
	sv.ApplyTheme()
	return sv
}

// SetApp sets the application reference for async updates
func (sv *SearchView) SetApp(app *tview.Application) {
	sv.app = app

	// Handle Enter and navigation in input
	sv.InputField.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			query := sv.InputField.GetText()
			if query != "" {
				sv.Search(query)
				// Move to results after search
				if len(sv.results) > 0 {
					app.SetFocus(sv.Table)
					sv.Table.Select(1, 0)
				}
			}
		case tcell.KeyDown:
			if len(sv.results) > 0 {
				app.SetFocus(sv.Table)
				sv.Table.Select(1, 0)
			}
		}
	})

	// Arrow down from input goes to results
	sv.InputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyDown && len(sv.results) > 0 {
			app.SetFocus(sv.Table)
			sv.Table.Select(1, 0)
			return nil
		}
		return event
	})

	// Handle navigation in table
	sv.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := sv.Table.GetSelection()
		switch event.Key() {
		case tcell.KeyUp:
			if row <= 1 {
				app.SetFocus(sv.InputField)
				return nil
			}
		case tcell.KeyRune:
			// Typing in table redirects to input
			app.SetFocus(sv.InputField)
			sv.InputField.SetText(string(event.Rune()))
			return nil
		}
		return event
	})
}

func (sv *SearchView) setupUI() {
	// Search input
	sv.InputField = tview.NewInputField().
		SetLabel(" Search: ").
		SetFieldWidth(0).
		SetPlaceholder("Type to search (e.g., nginx, postgres, redis)...")

	// Results table
	sv.Table = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0)

	sv.Table.SetBorder(true).SetTitle(" [2] Search Results ")
	sv.setupHeaders()

	// Status text
	sv.statusText = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	// Combined layout
	sv.Flex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(sv.InputField, 1, 0, true).
		AddItem(sv.Table, 0, 1, false).
		AddItem(sv.statusText, 1, 0, false)

	sv.Table.SetSelectedFunc(func(row, col int) {
		if row > 0 && row-1 < len(sv.results) {
			result := sv.results[row-1]
			if sv.onSelect != nil {
				// Use result's registry or fall back to current search registry
				regURL := result.RegistryURL
				if regURL == "" {
					regURL = sv.currentReg
				}
				sv.onSelect(regURL, result.Name)
			}
		}
	})

	sv.showWelcome()
}

// ApplyTheme applies the current theme to this view's widgets.
func (sv *SearchView) ApplyTheme() {
	// InputField
	sv.InputField.SetBackgroundColor(theme.BackgroundColor())
	sv.InputField.SetFieldBackgroundColor(theme.ElementBgColor())
	sv.InputField.SetFieldTextColor(theme.TextColor())
	sv.InputField.SetLabelColor(theme.TextColor())
	sv.InputField.SetPlaceholderTextColor(theme.PlaceholderColor())

	// Table
	sv.Table.SetBackgroundColor(theme.BackgroundColor())
	sv.Table.SetBorderColor(theme.BorderNormalColor())
	sv.Table.SetTitleColor(theme.TitleColor())
	sv.Table.SetSelectedStyle(tcell.StyleDefault.
		Background(theme.SelectionBgColor()).
		Foreground(theme.SelectionFgColor()))

	// Status text
	sv.statusText.SetBackgroundColor(theme.BackgroundColor())
	sv.statusText.SetTextColor(theme.TextColor())

	// Flex
	sv.Flex.SetBackgroundColor(theme.BackgroundColor())

	// Re-apply header colors
	sv.setupHeaders()
}

func (sv *SearchView) setupHeaders() {
	headers := []string{"", "Repository", "Description", "Stars", "Pulls"}
	widths := []int{3, 30, 40, 8, 10}

	for col, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(theme.HeaderColor()).
			SetSelectable(false).
			SetExpansion(0)
		if col < len(widths) {
			cell.SetMaxWidth(widths[col])
		}
		sv.Table.SetCell(0, col, cell)
	}
}

func (sv *SearchView) showWelcome() {
	sv.Table.Clear()
	sv.setupHeaders()
	sv.statusText.SetText(theme.Tag("muted") + "Enter a search term and press Enter" + theme.ResetTag())
}

// SetRegistry sets the current registry for searches
func (sv *SearchView) SetRegistry(registryURL string) {
	sv.currentReg = registryURL
	sv.InputField.SetLabel(fmt.Sprintf(" Search %s: ", registryURL))
}

// Search performs a search and updates the results
func (sv *SearchView) Search(query string) {
	if query == "" {
		sv.showWelcome()
		return
	}

	// Default to docker.io if no registry selected
	if sv.currentReg == "" {
		sv.currentReg = "docker.io"
		sv.InputField.SetLabel(" Search docker.io: ")
	}

	sv.statusText.SetText(theme.Tag("warning") + "Searching " + sv.currentReg + "..." + theme.ResetTag())
	sv.Table.Clear()
	sv.setupHeaders()

	// Perform search in background
	go func() {
		results, err := sv.registry.Search(sv.currentReg, query)

		// Update UI on main thread
		if sv.app != nil {
			sv.app.QueueUpdateDraw(func() {
				if err != nil {
					sv.statusText.SetText(fmt.Sprintf("%sError: %v%s", theme.Tag("error"), err, theme.ResetTag()))
					return
				}

				sv.results = results
				sv.renderResults()

				if len(results) == 0 {
					sv.statusText.SetText(theme.Tag("muted") + "No results found" + theme.ResetTag())
				} else {
					sv.statusText.SetText(fmt.Sprintf("%sFound %d repositories%s", theme.Tag("success"), len(results), theme.ResetTag()))
				}
			})
		}
	}()
}

func (sv *SearchView) renderResults() {
	for i, result := range sv.results {
		row := i + 1

		// Official badge
		badge := ""
		if result.IsOfficial {
			badge = theme.Tag("success") + "✓" + theme.ResetTag()
		}

		// Truncate description
		desc := result.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}

		sv.Table.SetCell(row, 0, tview.NewTableCell(badge))
		sv.Table.SetCell(row, 1, tview.NewTableCell(result.Name).SetExpansion(1).SetTextColor(theme.TextColor()))
		sv.Table.SetCell(row, 2, tview.NewTableCell(desc).SetExpansion(1).SetTextColor(theme.DescriptionColor()))
		sv.Table.SetCell(row, 3, tview.NewTableCell(fmt.Sprintf("%d", result.StarCount)).SetTextColor(theme.TextColor()))
		sv.Table.SetCell(row, 4, tview.NewTableCell(registry.FormatPullCount(result.PullCount)).SetTextColor(theme.TextColor()))
	}

	if len(sv.results) > 0 {
		sv.Table.Select(1, 0)
	}
}

// GetResults returns current search results
func (sv *SearchView) GetResults() []*registry.SearchResult {
	return sv.results
}

// Clear clears the search view
func (sv *SearchView) Clear() {
	sv.InputField.SetText("")
	sv.results = nil
	sv.showWelcome()
}
