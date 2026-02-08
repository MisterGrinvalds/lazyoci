package gui

import (
	"context"
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/mistergrinvalds/lazyoci/pkg/cache"
	"github.com/mistergrinvalds/lazyoci/pkg/config"
	"github.com/mistergrinvalds/lazyoci/pkg/gui/keybindings"
	"github.com/mistergrinvalds/lazyoci/pkg/gui/theme"
	"github.com/mistergrinvalds/lazyoci/pkg/gui/views"
	"github.com/mistergrinvalds/lazyoci/pkg/pull"
	"github.com/mistergrinvalds/lazyoci/pkg/registry"
	"github.com/rivo/tview"
)

// GUI is the main terminal user interface controller
type GUI struct {
	app      *tview.Application
	pages    *tview.Pages
	registry *registry.Client
	cache    *cache.Cache
	config   *config.Config

	// Views
	registryView  *views.RegistryView
	registryModal *views.RegistryModal
	settingsModal *views.SettingsModal
	searchView    *views.SearchView
	artifactView  *views.ArtifactView
	detailsView   *views.DetailsView
	statusBar     *tview.TextView

	// Layout
	mainFlex  *tview.Flex
	leftPanel *tview.Flex
	content   *tview.Flex
}

// New creates a new GUI instance
func New(reg *registry.Client, c *cache.Cache, cfg *config.Config) (*GUI, error) {
	g := &GUI{
		app:      tview.NewApplication(),
		pages:    tview.NewPages(),
		registry: reg,
		cache:    c,
		config:   cfg,
	}

	g.setupViews()
	g.setupLayout()
	keybindings.Setup(g.app, g)

	return g, nil
}

// Run starts the GUI event loop
func (g *GUI) Run() error {
	return g.app.SetRoot(g.pages, true).EnableMouse(true).Run()
}

func (g *GUI) setupViews() {
	// Registry selection → moves to search with that registry
	g.registryView = views.NewRegistryView(g.registry, g.onRegistrySelected)

	// Wire up add/delete callbacks
	g.registryView.SetOnAdd(g.showAddRegistryModal)
	g.registryView.SetOnDelete(g.showDeleteConfirmation)

	// Registry modal for adding new registries
	g.registryModal = views.NewRegistryModal(g.onRegistrySave, g.hideModal)

	// Settings modal for configuring artifact storage
	g.settingsModal = views.NewSettingsModal(
		g.config,
		g.onSettingsSave,
		g.hideSettingsModal,
		g.showSettingsConfirmation,
	)

	// Search result selection → loads artifacts
	g.searchView = views.NewSearchView(g.registry, g.onSearchResultSelected)

	// Artifact selection → shows details
	g.artifactView = views.NewArtifactView(g.registry, g.onArtifactSelected)

	// Wire up pull callbacks for artifacts view
	g.artifactView.SetOnPull(g.showPullModal)
	g.artifactView.SetOnPullDirect(g.executePullDirect)

	// Wire up selection with info callback for type-aware details
	g.artifactView.SetOnSelectWithInfo(g.onArtifactSelectedWithInfo)

	g.detailsView = views.NewDetailsView()

	// Wire up pull callbacks for details view
	g.detailsView.SetOnPull(g.showPullModal)
	g.detailsView.SetOnPullDirect(g.executePullDirect)

	g.statusBar = tview.NewTextView().
		SetDynamicColors(true)
	g.applyStatusBarTheme()
	g.updateStatus()
}

func (g *GUI) setupLayout() {
	// Left panel: Registry (1) -> Search/Repos (2) -> Artifacts (3)
	g.leftPanel = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(g.registryView.List, 0, 1, true).
		AddItem(g.searchView.Flex, 0, 2, false).
		AddItem(g.artifactView.Flex, 0, 2, false)
	g.leftPanel.SetBackgroundColor(theme.BackgroundColor())

	// Right panel: Details (context-aware)
	rightPanel := g.detailsView.TextView

	// Main content area
	g.content = tview.NewFlex().
		AddItem(g.leftPanel, 0, 1, true).
		AddItem(rightPanel, 0, 1, false)
	g.content.SetBackgroundColor(theme.BackgroundColor())

	// Main layout with status bar
	g.mainFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(g.content, 0, 1, true).
		AddItem(g.statusBar, 1, 0, false)
	g.mainFlex.SetBackgroundColor(theme.BackgroundColor())

	// Pages container
	g.pages.SetBackgroundColor(theme.BackgroundColor())

	// Set up pages
	g.pages.AddPage("main", g.mainFlex, true, true)
	g.pages.AddPage("add-registry", g.registryModal.GetPrimitive(), true, false)
	g.pages.AddPage("settings", g.settingsModal.GetPrimitive(), true, false)

	// Wire up views with app for async updates
	g.searchView.SetApp(g.app)
	g.artifactView.SetApp(g.app)

	// Set default registry for search
	regs := g.registry.GetRegistries()
	if len(regs) > 0 {
		g.searchView.SetRegistry(regs[0].URL)
	}
}

// applyThemeToAllViews re-applies theme colors to all views.
// Called when the theme is changed at runtime.
func (g *GUI) applyThemeToAllViews() {
	// Update tview global styles
	theme.ApplyToTview()

	// Update all views
	g.registryView.ApplyTheme()
	g.searchView.ApplyTheme()
	g.artifactView.ApplyTheme()
	g.detailsView.ApplyTheme()
	g.registryModal.ApplyTheme()
	g.settingsModal.ApplyTheme()

	// Update layout containers
	g.leftPanel.SetBackgroundColor(theme.BackgroundColor())
	g.content.SetBackgroundColor(theme.BackgroundColor())
	g.mainFlex.SetBackgroundColor(theme.BackgroundColor())
	g.pages.SetBackgroundColor(theme.BackgroundColor())

	// Update status bar
	g.applyStatusBarTheme()
	g.updateStatus()
}

// applyStatusBarTheme applies theme colors to the status bar
func (g *GUI) applyStatusBarTheme() {
	g.statusBar.SetBackgroundColor(theme.BackgroundColor())
	g.statusBar.SetTextColor(theme.TextColor())
}

// showAddRegistryModal shows the add registry modal
func (g *GUI) showAddRegistryModal() {
	g.registryModal.Clear()
	g.registryModal.ApplyTheme() // Ensure current theme is applied
	g.pages.ShowPage("add-registry")
	g.app.SetFocus(g.registryModal.Form)
}

// hideModal hides the current modal and returns to main
func (g *GUI) hideModal() {
	g.pages.HidePage("add-registry")
	g.pages.HidePage("confirm-delete")
	g.pages.SwitchToPage("main")
	g.app.SetFocus(g.registryView.List)
}

// ShowSettings shows the settings modal
func (g *GUI) ShowSettings() {
	g.settingsModal.ApplyTheme() // Ensure current theme is applied
	g.settingsModal.Refresh()
	g.pages.ShowPage("settings")
	g.app.SetFocus(g.settingsModal.Form)
}

// hideSettingsModal hides the settings modal
func (g *GUI) hideSettingsModal() {
	g.pages.HidePage("settings")
	g.pages.HidePage("settings-confirm")
	g.pages.SwitchToPage("main")
	g.updateStatus()
}

// onSettingsSave handles saving settings
func (g *GUI) onSettingsSave(path string, createDir bool) {
	var err error
	if path == "" || path == "default" {
		// Reset to default
		g.config.ArtifactDir = ""
		err = g.config.Save()
	} else {
		err = g.config.SetArtifactDir(path, createDir)
	}

	if err != nil {
		g.statusBar.SetText(fmt.Sprintf("%sError: %v%s", theme.Tag("error"), err, theme.ResetTag()))
	} else {
		if path == "" || path == "default" {
			g.statusBar.SetText(theme.Tag("success") + "Artifact directory reset to default" + theme.ResetTag())
		} else {
			g.statusBar.SetText(fmt.Sprintf("%sArtifact directory set to %s%s", theme.Tag("success"), path, theme.ResetTag()))
		}
	}

	g.hideSettingsModal()
}

// showSettingsConfirmation shows a confirmation dialog during settings
func (g *GUI) showSettingsConfirmation(title, message string, onYes, onNo, onCancel func()) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"Yes", "No", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			g.pages.RemovePage("settings-confirm")
			switch buttonLabel {
			case "Yes":
				if onYes != nil {
					onYes()
				}
			case "No":
				if onNo != nil {
					onNo()
				}
			case "Cancel":
				if onCancel != nil {
					onCancel()
				}
				// Return to settings modal
				g.app.SetFocus(g.settingsModal.Form)
			}
		})

	// Apply theme to modal
	modal.SetBackgroundColor(theme.ModalBgColor())
	modal.SetBorderColor(theme.BorderNormalColor())
	modal.SetTitleColor(theme.TitleColor())
	modal.SetTextColor(theme.TextColor())
	modal.SetButtonBackgroundColor(theme.ButtonBgColor())
	modal.SetButtonTextColor(theme.ButtonTextColor())
	if title != "" {
		modal.SetTitle(" " + title + " ")
	}

	g.pages.AddPage("settings-confirm", modal, true, true)
}

// ShowHelp shows the help modal
func (g *GUI) ShowHelp() {
	helpText := keybindings.GetHelpText()

	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetText(helpText).
		SetScrollable(true)
	textView.SetBorder(true).SetTitle(" Help ")

	// Apply theme to help view
	textView.SetBackgroundColor(theme.BackgroundColor())
	textView.SetTextColor(theme.TextColor())
	textView.SetBorderColor(theme.BorderNormalColor())
	textView.SetTitleColor(theme.TitleColor())

	// Create a centered modal
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(textView, 26, 1, true).
			AddItem(nil, 0, 1, false), 55, 1, true).
		AddItem(nil, 0, 1, false)
	flex.SetBackgroundColor(theme.BackgroundColor())

	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyEnter || event.Rune() == 'q' {
			g.pages.RemovePage("help")
			g.app.SetFocus(g.registryView.List)
			return nil
		}
		return event
	})

	g.pages.AddPage("help", flex, true, true)
	g.app.SetFocus(textView)
}

// ShowThemePicker shows the theme selection modal
func (g *GUI) ShowThemePicker() {
	themes := theme.AvailableThemes()
	currentTheme := theme.CurrentThemeName()

	list := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true)

	list.SetBorder(true).SetTitle(" Select Theme ")

	// Apply theme to list
	list.SetBackgroundColor(theme.BackgroundColor())
	list.SetBorderColor(theme.BorderNormalColor())
	list.SetTitleColor(theme.TitleColor())
	list.SetMainTextColor(theme.TextColor())
	list.SetSelectedBackgroundColor(theme.SelectionBgColor())
	list.SetSelectedTextColor(theme.SelectionFgColor())

	for _, name := range themes {
		display := name
		if name == currentTheme {
			display = name + " (current)"
		}
		list.AddItem(display, "", 0, nil)
	}

	// Pre-select current theme
	for i, name := range themes {
		if name == currentTheme {
			list.SetCurrentItem(i)
			break
		}
	}

	list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(themes) {
			selectedTheme := themes[index]
			theme.SetTheme(selectedTheme)
			// Persist to config
			_ = g.config.SetTheme(selectedTheme)

			// Apply theme to all views
			g.applyThemeToAllViews()

			g.pages.RemovePage("theme-picker")
			g.app.SetFocus(g.registryView.List)
		}
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Rune() == 'q' {
			g.pages.RemovePage("theme-picker")
			g.app.SetFocus(g.registryView.List)
			return nil
		}
		return event
	})

	// Center the list
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(list, len(themes)+2, 1, true).
			AddItem(nil, 0, 1, false), 40, 1, true).
		AddItem(nil, 0, 1, false)
	flex.SetBackgroundColor(theme.BackgroundColor())

	g.pages.AddPage("theme-picker", flex, true, true)
	g.app.SetFocus(list)
}

// onRegistrySave handles saving a new registry
func (g *GUI) onRegistrySave(name, url, username, password string) {
	var err error
	if username != "" {
		err = g.registry.AddRegistryWithAuth(name, url, username, password)
	} else {
		err = g.registry.AddRegistry(name, url)
	}

	if err != nil {
		g.statusBar.SetText(fmt.Sprintf("%sError: %v%s", theme.Tag("error"), err, theme.ResetTag()))
	} else {
		label := url
		if name != "" {
			label = name
		}
		g.statusBar.SetText(fmt.Sprintf("%sRegistry %s added%s", theme.Tag("success"), label, theme.ResetTag()))
		g.registryView.Refresh()
	}

	g.hideModal()
}

// showDeleteConfirmation shows a confirmation dialog for deleting a registry
func (g *GUI) showDeleteConfirmation(url string) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Delete registry %s?", url)).
		AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Delete" {
				if err := g.registry.RemoveRegistry(url); err != nil {
					g.statusBar.SetText(fmt.Sprintf("%sError: %v%s", theme.Tag("error"), err, theme.ResetTag()))
				} else {
					g.statusBar.SetText(fmt.Sprintf("%sRegistry %s removed%s", theme.Tag("warning"), url, theme.ResetTag()))
					g.registryView.Refresh()
				}
			}
			g.pages.RemovePage("confirm-delete")
			g.app.SetFocus(g.registryView.List)
		})

	// Apply theme to modal
	modal.SetBackgroundColor(theme.ModalBgColor())
	modal.SetBorderColor(theme.BorderNormalColor())
	modal.SetTitleColor(theme.TitleColor())
	modal.SetTextColor(theme.TextColor())
	modal.SetButtonBackgroundColor(theme.ButtonBgColor())
	modal.SetButtonTextColor(theme.ButtonTextColor())

	g.pages.AddPage("confirm-delete", modal, true, true)
}

// onRegistrySelected - Enter in (1) → set registry and move to (2)
func (g *GUI) onRegistrySelected(registryURL string) {
	g.searchView.SetRegistry(registryURL)
	g.detailsView.ShowRegistryInfo(registryURL)
	g.app.SetFocus(g.searchView.InputField)
}

// onSearchResultSelected - Enter in (2) → load artifacts and move to (3)
func (g *GUI) onSearchResultSelected(registryURL, repoName string) {
	fullPath := registryURL + "/" + repoName
	g.detailsView.ShowRepository(fullPath)
	g.artifactView.LoadArtifacts(fullPath)
	g.app.SetFocus(g.artifactView.FilterInput)
}

// onArtifactSelected - Enter in (3) → show details in (4)
func (g *GUI) onArtifactSelected(artifact *registry.Artifact) {
	g.detailsView.ShowArtifact(artifact)
}

// onArtifactSelectedWithInfo - called when artifact info is resolved
func (g *GUI) onArtifactSelectedWithInfo(artifact *registry.Artifact, info *registry.ArtifactInfo) {
	g.detailsView.ShowArtifactWithInfo(artifact, info)
}

// showPullModal shows a modal to confirm pulling an artifact
func (g *GUI) showPullModal(artifact *registry.Artifact) {
	if artifact == nil {
		return
	}

	ref := artifact.Repository + ":" + artifact.Tag

	// Determine available options based on artifact type
	// Only images can be loaded to Docker
	artifactType := g.detailsView.GetCurrentArtifactType()
	isImage := artifactType == registry.ArtifactTypeImage || artifactType == registry.ArtifactTypeUnknown

	var buttons []string
	if isImage {
		buttons = []string{"To Disk", "To Docker", "Cancel"}
	} else {
		buttons = []string{"Pull", "Cancel"}
	}

	// Build modal text with type info
	text := fmt.Sprintf("Pull %s?", ref)
	if artifactType != registry.ArtifactTypeImage && artifactType != registry.ArtifactTypeUnknown {
		text = fmt.Sprintf("Pull %s (%s)?", ref, artifactType.String())
	}

	modal := tview.NewModal().
		SetText(text).
		AddButtons(buttons).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			g.pages.RemovePage("pull-modal")
			g.app.SetFocus(g.artifactView.GetTable())

			switch buttonLabel {
			case "To Disk", "Pull":
				g.executePull(artifact, false)
			case "To Docker":
				g.executePull(artifact, true)
			}
		})

	// Apply theme to modal
	modal.SetBackgroundColor(theme.ModalBgColor())
	modal.SetBorderColor(theme.BorderNormalColor())
	modal.SetTitleColor(theme.TitleColor())
	modal.SetTextColor(theme.TextColor())
	modal.SetButtonBackgroundColor(theme.ButtonBgColor())
	modal.SetButtonTextColor(theme.ButtonTextColor())

	g.pages.AddPage("pull-modal", modal, true, true)
}

// executePull pulls an artifact in the background
func (g *GUI) executePull(artifact *registry.Artifact, toDocker bool) {
	ref := artifact.Repository + ":" + artifact.Tag

	g.statusBar.SetText(fmt.Sprintf("%sPulling %s...%s", theme.Tag("warning"), ref, theme.ResetTag()))

	go func() {
		ctx := context.Background()

		// Check if registry is insecure
		insecure := false
		registryURL := ""
		if parts := splitRepoPath(artifact.Repository); len(parts) >= 1 {
			registryURL = parts[0]
			for _, r := range g.registry.GetRegistries() {
				if r.URL == registryURL && r.Insecure {
					insecure = true
					break
				}
			}
		}

		opts := pull.PullOptions{
			Reference:    ref,
			ArtifactBase: g.config.GetArtifactDir(),
			ToDocker:     toDocker,
			Quiet:        true, // No progress bars in TUI
			Insecure:     insecure,
		}

		// Resolve credentials for the registry
		if registryURL != "" {
			opts.CredentialFunc = g.registry.CredentialFunc(registryURL)
		}

		puller := pull.NewPuller(true)
		result, err := puller.Pull(ctx, opts)

		g.app.QueueUpdateDraw(func() {
			if err != nil {
				g.statusBar.SetText(fmt.Sprintf("%sPull failed: %v%s", theme.Tag("error"), err, theme.ResetTag()))
			} else {
				msg := fmt.Sprintf("%sPulled %s to %s%s", theme.Tag("success"), ref, result.Destination, theme.ResetTag())
				if result.LoadedToDocker {
					msg = fmt.Sprintf("%sPulled %s and loaded into Docker%s", theme.Tag("success"), ref, theme.ResetTag())
				}
				g.statusBar.SetText(msg)
			}
		})
	}()
}

// executePullDirect pulls an artifact directly without showing a modal
func (g *GUI) executePullDirect(artifact *registry.Artifact, toDocker bool) {
	if artifact == nil {
		return
	}

	// Check if Docker pull is appropriate for this artifact type
	if toDocker {
		artifactType := g.detailsView.GetCurrentArtifactType()
		if artifactType != registry.ArtifactTypeImage && artifactType != registry.ArtifactTypeUnknown {
			g.statusBar.SetText(fmt.Sprintf("%sCannot load %s to Docker (only images supported)%s", theme.Tag("error"), artifactType.String(), theme.ResetTag()))
			return
		}
	}

	g.executePull(artifact, toDocker)
}

// splitRepoPath splits a repository path into registry and repo parts
func splitRepoPath(repoPath string) []string {
	parts := make([]string, 0, 2)
	idx := 0
	for i, c := range repoPath {
		if c == '/' {
			parts = append(parts, repoPath[idx:i])
			idx = i + 1
			if len(parts) == 1 {
				parts = append(parts, repoPath[idx:])
				return parts
			}
		}
	}
	if idx < len(repoPath) {
		parts = append(parts, repoPath[idx:])
	}
	return parts
}

// GetApp returns the tview application
func (g *GUI) GetApp() *tview.Application {
	return g.app
}

// FocusRegistry focuses the registry list
func (g *GUI) FocusRegistry() {
	g.app.SetFocus(g.registryView.List)
	g.detailsView.ShowRegistryHelp()
	g.updateStatus()
}

// FocusSearch focuses the search input
func (g *GUI) FocusSearch() {
	selectedReg := g.registryView.GetSelectedRegistry()
	if selectedReg != "" {
		g.searchView.SetRegistry(selectedReg)
	}
	g.app.SetFocus(g.searchView.InputField)
	g.updateStatus()
}

// FocusArtifacts focuses the artifact filter
func (g *GUI) FocusArtifacts() {
	g.app.SetFocus(g.artifactView.FilterInput)
	g.updateStatus()
}

// FocusDetails focuses the details view
func (g *GUI) FocusDetails() {
	g.app.SetFocus(g.detailsView.TextView)
	g.updateStatus()
}

// IsInputFocused returns true if an input field is currently focused
func (g *GUI) IsInputFocused() bool {
	current := g.app.GetFocus()
	return current == g.searchView.InputField || current == g.artifactView.FilterInput
}

func (g *GUI) updateStatus() {
	emphasis := theme.Tag("emphasis")
	text := theme.Tag("text")
	success := theme.Tag("success")
	info := theme.Tag("info")

	status := fmt.Sprintf("%slazyoci%s | %s1%s Registry %s2%s Search %s3%s Artifacts %s4%s Details",
		emphasis, text,
		success, text,
		success, text,
		success, text,
		success, text,
	)

	// Show artifact directory indicator if non-default
	if !g.config.IsArtifactDirDefault() {
		artifactDir := g.config.GetArtifactDir()
		// Shorten the path for display
		shortPath := shortenPathForStatus(artifactDir)
		status += fmt.Sprintf(" | %s[%s]%s", info, shortPath, theme.ResetTag())
	}

	status += fmt.Sprintf(" | %s/%s search %sp%s pull %sd%s docker %sS%s settings %sT%s theme %s?%s help | %sq%s quit",
		success, text,
		success, text,
		success, text,
		success, text,
		success, text,
		success, text,
		success, text,
	)
	g.statusBar.SetText(status)
}

// shortenPathForStatus shortens a path for status bar display
func shortenPathForStatus(path string) string {
	// Try to use ~ for home directory
	if home, err := os.UserHomeDir(); err == nil {
		if len(path) > len(home) && path[:len(home)] == home {
			path = "~" + path[len(home):]
		}
	}

	// Truncate if still too long
	if len(path) > 25 {
		return "..." + path[len(path)-22:]
	}
	return path
}

// CycleFocus cycles through the panels
func (g *GUI) CycleFocus() {
	current := g.app.GetFocus()

	switch current {
	case g.registryView.List:
		g.app.SetFocus(g.searchView.InputField)
	case g.searchView.InputField, g.searchView.Table:
		g.app.SetFocus(g.artifactView.FilterInput)
	case g.artifactView.GetTable(), g.artifactView.FilterInput:
		g.app.SetFocus(g.detailsView.TextView)
	default:
		g.app.SetFocus(g.registryView.List)
	}
	g.updateStatus()
}
