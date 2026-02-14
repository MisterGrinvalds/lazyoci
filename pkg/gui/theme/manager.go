package theme

import (
	"os"
	"sort"
	"strings"
	"sync"
)

// manager is the global theme registry and state holder.
var manager = &Manager{
	themes:      make(map[string]Theme),
	currentName: "default",
	isDark:      true,
	mode:        "auto",
}

// Manager handles theme registration, selection, and dark/light detection.
type Manager struct {
	mu          sync.RWMutex
	themes      map[string]Theme
	currentName string
	isDark      bool
	mode        string // "auto", "dark", "light"
}

// RegisterTheme adds a theme to the global registry.
// Called by init() functions in individual theme files.
func RegisterTheme(name string, t Theme) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	manager.themes[name] = t
}

// SetTheme changes the active theme by name.
// Returns false if the theme name is not found.
func SetTheme(name string) bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if _, ok := manager.themes[name]; !ok {
		return false
	}
	manager.currentName = name
	return true
}

// CurrentTheme returns the currently active theme.
// Falls back to the "default" theme if the current is not found.
func CurrentTheme() Theme {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	if t, ok := manager.themes[manager.currentName]; ok {
		return t
	}
	if t, ok := manager.themes["default"]; ok {
		return t
	}
	// Should never happen if default theme is registered
	return nil
}

// CurrentThemeName returns the name of the active theme.
func CurrentThemeName() string {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return manager.currentName
}

// IsDark returns whether the terminal background is dark.
func IsDark() bool {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return manager.isDark
}

// SetMode sets the dark/light mode ("auto", "dark", "light").
func SetMode(mode string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	manager.mode = mode
	switch mode {
	case "dark":
		manager.isDark = true
	case "light":
		manager.isDark = false
	default:
		// auto: detect from terminal
		manager.isDark = detectDarkBackground()
	}
}

// DetectAndApplyMode runs terminal background detection for "auto" mode.
// Should be called once at startup after mode is configured.
func DetectAndApplyMode() {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if manager.mode == "auto" || manager.mode == "" {
		manager.isDark = detectDarkBackground()
	}
}

// AvailableThemes returns a sorted list of registered theme names,
// with "default" always appearing first.
func AvailableThemes() []string {
	manager.mu.RLock()
	defer manager.mu.RUnlock()

	names := make([]string, 0, len(manager.themes))
	for name := range manager.themes {
		if name != "default" {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	// "default" always first
	if _, ok := manager.themes["default"]; ok {
		names = append([]string{"default"}, names...)
	}

	return names
}

// GetTheme returns a specific theme by name, or nil if not found.
func GetTheme(name string) Theme {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return manager.themes[name]
}

// detectDarkBackground attempts to determine if the terminal has a dark
// background. This is a best-effort heuristic.
func detectDarkBackground() bool {
	// Check common environment hints
	colorScheme := os.Getenv("COLORFGBG")
	if colorScheme != "" {
		// Format is "foreground;background", e.g. "15;0" (white on black = dark)
		parts := strings.Split(colorScheme, ";")
		if len(parts) >= 2 {
			bg := parts[len(parts)-1]
			// Background values 0-6 are typically dark, 7+ are light
			if bg == "0" || bg == "1" || bg == "2" || bg == "3" || bg == "4" || bg == "5" || bg == "6" {
				return true
			}
			if bg == "7" || bg == "8" || bg == "9" || bg == "10" || bg == "11" || bg == "12" || bg == "13" || bg == "14" || bg == "15" {
				return false
			}
		}
	}

	// Check for known dark terminal themes
	termProgram := os.Getenv("TERM_PROGRAM")
	iterm := os.Getenv("ITERM_PROFILE")
	_ = iterm

	// Most terminals default to dark backgrounds
	_ = termProgram

	// Default to dark (most common)
	return true
}
