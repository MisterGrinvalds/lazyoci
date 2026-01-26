package gui

import (
	"github.com/mistergrinvalds/lazyoci/pkg/cache"
	"github.com/mistergrinvalds/lazyoci/pkg/gui/keybindings"
	"github.com/mistergrinvalds/lazyoci/pkg/gui/views"
	"github.com/mistergrinvalds/lazyoci/pkg/registry"
	"github.com/rivo/tview"
)

// GUI is the main terminal user interface controller
type GUI struct {
	app      *tview.Application
	pages    *tview.Pages
	registry *registry.Client
	cache    *cache.Cache

	// Views
	registryView *views.RegistryView
	searchView   *views.SearchView
	artifactView *views.ArtifactView
	detailsView  *views.DetailsView
	statusBar    *tview.TextView

	// Layout
	mainFlex *tview.Flex
}

// New creates a new GUI instance
func New(reg *registry.Client, c *cache.Cache) (*GUI, error) {
	g := &GUI{
		app:      tview.NewApplication(),
		pages:    tview.NewPages(),
		registry: reg,
		cache:    c,
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
	g.registryView = views.NewRegistryView(g.registry, g.onRepositorySelected)
	g.searchView = views.NewSearchView(g.registry, g.onSearchResultSelected)
	g.artifactView = views.NewArtifactView(g.registry, g.onArtifactSelected)
	g.detailsView = views.NewDetailsView()

	g.statusBar = tview.NewTextView().
		SetDynamicColors(true).
		SetText("[yellow]lazyoci[white] | [green]1[white] Registry [green]2[white] Search [green]3[white] Artifacts [green]4[white] Details | [green]Tab[white] cycle | [green]?[white] help [green]q[white] quit")
}

func (g *GUI) setupLayout() {
	// Left panel: Registry (1) -> Search/Repos (2) -> Artifacts (3)
	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(g.registryView.TreeView, 0, 1, true).
		AddItem(g.searchView.Flex, 0, 1, false).
		AddItem(g.artifactView.Flex, 0, 1, false)

	// Right panel: Details (context-aware)
	rightPanel := g.detailsView.TextView

	// Main content area
	content := tview.NewFlex().
		AddItem(leftPanel, 0, 1, true).
		AddItem(rightPanel, 0, 1, false)

	// Main layout with status bar
	g.mainFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(content, 0, 1, true).
		AddItem(g.statusBar, 1, 0, false)

	// Set up pages
	g.pages.AddPage("main", g.mainFlex, true, true)

	// Wire up views with app for async updates
	g.searchView.SetApp(g.app)
	g.artifactView.SetApp(g.app)

	// Set default registry for search
	regs := g.registry.GetRegistries()
	if len(regs) > 0 {
		g.searchView.SetRegistry(regs[0].URL)
	}
}

func (g *GUI) onRepositorySelected(repo string) {
	g.statusBar.SetText("[yellow]Loading " + repo + "...[white]")
	g.detailsView.ShowRepository(repo)
	g.artifactView.LoadArtifacts(repo)
	g.app.SetFocus(g.artifactView.Table)
}

func (g *GUI) onSearchResultSelected(registryURL, repoName string) {
	fullPath := registryURL + "/" + repoName
	g.onRepositorySelected(fullPath)
}

func (g *GUI) onArtifactSelected(artifact *registry.Artifact) {
	g.detailsView.ShowArtifact(artifact)
}

// GetApp returns the tview application
func (g *GUI) GetApp() *tview.Application {
	return g.app
}

// FocusRegistry focuses the registry tree view
func (g *GUI) FocusRegistry() {
	g.app.SetFocus(g.registryView.TreeView)
	g.detailsView.ShowRegistryHelp()
	g.updateStatus()
}

// FocusSearch focuses the search input and sets the registry based on current selection
func (g *GUI) FocusSearch() {
	selectedReg := g.registryView.GetSelectedRegistry()
	if selectedReg != "" {
		g.searchView.SetRegistry(selectedReg)
	}
	g.app.SetFocus(g.searchView.InputField)
	g.updateStatus()
}

// FocusArtifacts focuses the artifact filter (typing goes to filter)
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
	g.statusBar.SetText("[yellow]lazyoci[white] | [green]1[white] Registry [green]2[white] Search [green]3[white] Artifacts [green]4[white] Details | [green]/[white] search [green]Tab[white] cycle [green]Esc[white] back | [green]?[white] help [green]q[white] quit")
}

// CycleFocus cycles through the panels
func (g *GUI) CycleFocus() {
	current := g.app.GetFocus()

	switch current {
	case g.registryView.TreeView:
		g.app.SetFocus(g.searchView.InputField)
	case g.searchView.InputField, g.searchView.Table:
		g.app.SetFocus(g.artifactView.FilterInput)
	case g.artifactView.GetTable(), g.artifactView.FilterInput:
		g.app.SetFocus(g.detailsView.TextView)
	default:
		g.app.SetFocus(g.registryView.TreeView)
	}
	g.updateStatus()
}
