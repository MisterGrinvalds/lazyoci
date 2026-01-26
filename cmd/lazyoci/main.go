package main

import (
	"fmt"
	"os"

	"github.com/mistergrinvalds/lazyoci/pkg/app"
	"github.com/mistergrinvalds/lazyoci/pkg/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	application, err := app.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize app: %v\n", err)
		os.Exit(1)
	}

	if err := application.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "application error: %v\n", err)
		os.Exit(1)
	}
}
