package app

import (
	"github.com/mistergrinvalds/lazyoci/pkg/cache"
	"github.com/mistergrinvalds/lazyoci/pkg/config"
	"github.com/mistergrinvalds/lazyoci/pkg/gui"
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
	c := cache.New(cfg.CacheDir)

	reg := registry.NewClient(cfg)
	reg.SetCache(c) // Wire cache to registry client

	g, err := gui.New(reg, c)
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

// Run starts the application
func (a *App) Run() error {
	return a.gui.Run()
}
