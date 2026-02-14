package app

import (
	"github.com/mistergrinvalds/lazyoci/pkg/cache"
	"github.com/mistergrinvalds/lazyoci/pkg/config"
	"github.com/mistergrinvalds/lazyoci/pkg/gui"
	"github.com/mistergrinvalds/lazyoci/pkg/gui/theme"
	"github.com/mistergrinvalds/lazyoci/pkg/registry"
)

// App is the main application controller
type App struct {
	config   *config.Config
	gui      *gui.GUI
	registry *registry.Client
	cache    *cache.Cache
}

// New creates a new application instance
func New(cfg *config.Config) (*App, error) {
	// Initialize theme from config BEFORE creating any GUI widgets
	initTheme(cfg)

	c := cache.New(cfg.CacheDir)

	reg := registry.NewClient(cfg)
	reg.SetCache(c) // Wire cache to registry client

	g, err := gui.New(reg, c, cfg)
	if err != nil {
		return nil, err
	}

	return &App{
		config:   cfg,
		gui:      g,
		registry: reg,
		cache:    c,
	}, nil
}

// initTheme configures the theme system from config.
// Must be called BEFORE creating any tview widgets.
func initTheme(cfg *config.Config) {
	// Set dark/light mode (auto-detects terminal if "auto")
	theme.SetMode(cfg.GetMode())

	// Set the active theme
	themeName := cfg.GetTheme()
	if !theme.SetTheme(themeName) {
		// Fall back to default if configured theme not found
		theme.SetTheme("default")
	}

	// Apply theme colors to tview's global styles.
	// This must happen BEFORE any widgets are created so they
	// inherit the correct default colors.
	theme.ApplyToTview()
}

// Run starts the application
func (a *App) Run() error {
	return a.gui.Run()
}
