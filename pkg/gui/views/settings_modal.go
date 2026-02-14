package views

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/mistergrinvalds/lazyoci/pkg/config"
	"github.com/mistergrinvalds/lazyoci/pkg/gui/theme"
	"github.com/rivo/tview"
)

// LocationOption represents a storage location option
type LocationOption int

const (
	LocationDefault LocationOption = iota
	LocationCurrentDir
	LocationHome
	LocationEnv
	LocationCustom
)

// SettingsModal is a modal dialog for configuring settings
type SettingsModal struct {
	Form   *tview.Form
	Flex   *tview.Flex
	config *config.Config

	// Callbacks
	onSave    func(path string, createDir bool)
	onCancel  func()
	onConfirm func(title, message string, onYes, onNo, onCancel func())

	// Form fields
	radioButtons *tview.DropDown
	customInput  *tview.InputField
	currentLabel *tview.TextView

	// State
	selectedOption LocationOption
	customPath     string
}

// NewSettingsModal creates a new settings modal
func NewSettingsModal(
	cfg *config.Config,
	onSave func(path string, createDir bool),
	onCancel func(),
	onConfirm func(title, message string, onYes, onNo, onCancel func()),
) *SettingsModal {
	sm := &SettingsModal{
		config:    cfg,
		onSave:    onSave,
		onCancel:  onCancel,
		onConfirm: onConfirm,
	}

	sm.setupUI()
	sm.ApplyTheme()
	return sm
}

func (sm *SettingsModal) setupUI() {
	sm.Form = tview.NewForm()

	// Build location options
	options := sm.buildLocationOptions()

	// Custom path input - create FIRST so onLocationSelected can access it
	sm.customInput = tview.NewInputField().
		SetLabel("Custom Path: ").
		SetFieldWidth(40).
		SetPlaceholder("(select Custom to enter path)")

	// Dropdown for location selection - create AFTER customInput
	sm.radioButtons = tview.NewDropDown().
		SetLabel("Storage Location: ").
		SetOptions(options, sm.onLocationSelected).
		SetCurrentOption(0)

	// Current location display
	sm.currentLabel = tview.NewTextView().
		SetDynamicColors(true).
		SetText(sm.getCurrentLocationText())

	sm.Form.AddFormItem(sm.radioButtons)
	sm.Form.AddFormItem(sm.customInput)

	// Add current location as a text display
	sm.Form.AddFormItem(tview.NewTextView().
		SetDynamicColors(true).
		SetText(sm.getCurrentLocationText()))

	sm.Form.AddButton("Save", sm.handleSave)
	sm.Form.AddButton("Cancel", func() {
		if sm.onCancel != nil {
			sm.onCancel()
		}
	})

	sm.Form.SetBorder(true).SetTitle(" Settings ")
	sm.Form.SetButtonsAlign(tview.AlignCenter)

	// Handle escape key
	sm.Form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			if sm.onCancel != nil {
				sm.onCancel()
			}
			return nil
		}
		return event
	})

	// Center the form in a flex container
	sm.Flex = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(sm.Form, 18, 1, true).
			AddItem(nil, 0, 1, false), 60, 1, true).
		AddItem(nil, 0, 1, false)

	// Set initial selection based on current config
	sm.detectCurrentSelection()
}

// ApplyTheme applies the current theme to this view's widgets.
func (sm *SettingsModal) ApplyTheme() {
	// Form
	sm.Form.SetBackgroundColor(theme.BackgroundColor())
	sm.Form.SetBorderColor(theme.BorderNormalColor())
	sm.Form.SetTitleColor(theme.TitleColor())
	sm.Form.SetFieldBackgroundColor(theme.ElementBgColor())
	sm.Form.SetFieldTextColor(theme.TextColor())
	sm.Form.SetLabelColor(theme.TextColor())
	sm.Form.SetButtonBackgroundColor(theme.ElementBgColor())
	sm.Form.SetButtonTextColor(theme.TextColor())

	// DropDown
	sm.radioButtons.SetBackgroundColor(theme.BackgroundColor())
	sm.radioButtons.SetFieldBackgroundColor(theme.ElementBgColor())
	sm.radioButtons.SetFieldTextColor(theme.TextColor())
	sm.radioButtons.SetLabelColor(theme.TextColor())
	sm.radioButtons.SetListStyles(
		tcell.StyleDefault.Background(theme.ElementBgColor()).Foreground(theme.TextColor()),
		tcell.StyleDefault.Background(theme.SelectionBgColor()).Foreground(theme.SelectionFgColor()),
	)

	// Custom input
	sm.customInput.SetBackgroundColor(theme.BackgroundColor())
	sm.customInput.SetFieldBackgroundColor(theme.ElementBgColor())
	sm.customInput.SetFieldTextColor(theme.TextColor())
	sm.customInput.SetLabelColor(theme.TextColor())
	sm.customInput.SetPlaceholderTextColor(theme.PlaceholderColor())

	// Current label
	sm.currentLabel.SetBackgroundColor(theme.BackgroundColor())
	sm.currentLabel.SetTextColor(theme.TextColor())

	// Flex
	sm.Flex.SetBackgroundColor(theme.BackgroundColor())
}

func (sm *SettingsModal) buildLocationOptions() []string {
	options := []string{
		fmt.Sprintf("Default (%s)", shortenPath(config.DefaultArtifactDir())),
		fmt.Sprintf("Current Directory (%s)", shortenPath(filepath.Join(getCurrentDir(), "artifacts"))),
		fmt.Sprintf("Home (~/.lazyoci)"),
	}

	// Only show env option if set
	if envDir := os.Getenv("LAZYOCI_ARTIFACT_DIR"); envDir != "" {
		options = append(options, fmt.Sprintf("Environment (%s)", shortenPath(envDir)))
	}

	options = append(options, "Custom...")

	return options
}

func (sm *SettingsModal) onLocationSelected(text string, index int) {
	// Determine which option was selected
	hasEnv := os.Getenv("LAZYOCI_ARTIFACT_DIR") != ""

	switch index {
	case 0:
		sm.selectedOption = LocationDefault
	case 1:
		sm.selectedOption = LocationCurrentDir
	case 2:
		sm.selectedOption = LocationHome
	case 3:
		if hasEnv {
			sm.selectedOption = LocationEnv
		} else {
			sm.selectedOption = LocationCustom
		}
	case 4:
		sm.selectedOption = LocationCustom
	}

	// Update custom input visibility (show placeholder hint)
	if sm.selectedOption == LocationCustom {
		sm.customInput.SetPlaceholder("Enter path...")
	} else {
		sm.customInput.SetPlaceholder("(select Custom to enter path)")
		sm.customInput.SetText("")
	}
}

func (sm *SettingsModal) detectCurrentSelection() {
	currentDir := sm.config.GetArtifactDir()
	defaultDir := config.DefaultArtifactDir()
	homeDir, _ := os.UserHomeDir()
	cwdDir := filepath.Join(getCurrentDir(), "artifacts")
	homeLazyoci := filepath.Join(homeDir, ".lazyoci")
	envDir := os.Getenv("LAZYOCI_ARTIFACT_DIR")

	hasEnv := envDir != ""

	switch currentDir {
	case defaultDir:
		sm.radioButtons.SetCurrentOption(0)
		sm.selectedOption = LocationDefault
	case cwdDir:
		sm.radioButtons.SetCurrentOption(1)
		sm.selectedOption = LocationCurrentDir
	case homeLazyoci:
		sm.radioButtons.SetCurrentOption(2)
		sm.selectedOption = LocationHome
	case envDir:
		if hasEnv {
			sm.radioButtons.SetCurrentOption(3)
			sm.selectedOption = LocationEnv
		}
	default:
		// Custom path
		if hasEnv {
			sm.radioButtons.SetCurrentOption(4)
		} else {
			sm.radioButtons.SetCurrentOption(3)
		}
		sm.selectedOption = LocationCustom
		sm.customInput.SetText(currentDir)
	}
}

func (sm *SettingsModal) getCurrentLocationText() string {
	current := sm.config.GetArtifactDir()
	source := "default"

	if config.GetArtifactDirOverride() != "" {
		source = "CLI flag"
	} else if os.Getenv("LAZYOCI_ARTIFACT_DIR") != "" {
		source = "environment"
	} else if sm.config.ArtifactDir != "" {
		source = "config"
	}

	return fmt.Sprintf("%sCurrent:%s %s %s(%s)%s",
		theme.Tag("emphasis"), theme.Tag("text"),
		shortenPath(current),
		theme.Tag("muted"), source, theme.ResetTag())
}

func (sm *SettingsModal) handleSave() {
	var path string

	switch sm.selectedOption {
	case LocationDefault:
		path = "" // Empty means default
	case LocationCurrentDir:
		path = filepath.Join(getCurrentDir(), "artifacts")
	case LocationHome:
		path = "~/.lazyoci"
	case LocationEnv:
		// Keep using environment variable (don't save to config)
		if sm.onCancel != nil {
			sm.onCancel()
		}
		return
	case LocationCustom:
		path = sm.customInput.GetText()
		if path == "" {
			// No path entered, treat as cancel
			return
		}
	}

	// Expand and validate path
	expanded := config.ExpandPath(path)

	// Check if path exists
	if path != "" && !config.PathExists(expanded) {
		// Show confirmation to create
		if sm.onConfirm != nil {
			sm.onConfirm(
				"Create Directory",
				fmt.Sprintf("Directory does not exist:\n%s\n\nCreate it?", expanded),
				func() {
					// Yes - create and save
					if sm.onSave != nil {
						sm.onSave(path, true)
					}
				},
				func() {
					// No - save anyway (will be created on first pull)
					if sm.onSave != nil {
						sm.onSave(path, false)
					}
				},
				func() {
					// Cancel - return to settings
				},
			)
			return
		}
	}

	// Path exists or is default, save directly
	if sm.onSave != nil {
		sm.onSave(path, false)
	}
}

// Refresh updates the modal with current config values
func (sm *SettingsModal) Refresh() {
	sm.currentLabel.SetText(sm.getCurrentLocationText())
	sm.detectCurrentSelection()
}

// GetPrimitive returns the flex container for display
func (sm *SettingsModal) GetPrimitive() tview.Primitive {
	return sm.Flex
}

// Helper functions

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}

func shortenPath(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if len(path) > len(homeDir) && path[:len(homeDir)] == homeDir {
		return "~" + path[len(homeDir):]
	}

	// Truncate long paths
	if len(path) > 35 {
		return "..." + path[len(path)-32:]
	}

	return path
}
