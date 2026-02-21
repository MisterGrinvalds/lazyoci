package keybindings

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/mistergrinvalds/lazyoci/pkg/gui/theme"
	"github.com/rivo/tview"
)

// Controller defines the interface for GUI control from keybindings
type Controller interface {
	GetApp() *tview.Application
	FocusRegistry()
	FocusSearch()
	FocusArtifacts()
	FocusDetails()
	CycleFocus()
	IsInputFocused() bool
	IsModalOpen() bool
	ShowSettings()
	ShowHelp()
	ShowThemePicker()
}

// Setup configures all keybindings for the application
func Setup(app *tview.Application, ctrl Controller) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// When a modal is open, let it handle all keys — skip global bindings
		if ctrl.IsModalOpen() {
			return event
		}

		inInput := ctrl.IsInputFocused()

		// Handle Escape - exit input fields
		if event.Key() == tcell.KeyEscape && inInput {
			ctrl.FocusRegistry()
			return nil
		}

		// Handle '/' - enter search/filter mode (don't print the char)
		if event.Rune() == '/' && !inInput {
			ctrl.FocusSearch()
			return nil
		}

		// Global keybindings only when NOT in input field
		if !inInput {
			switch event.Key() {
			case tcell.KeyTab:
				ctrl.CycleFocus()
				return nil

			case tcell.KeyBacktab:
				ctrl.FocusRegistry()
				return nil

			case tcell.KeyCtrlC:
				app.Stop()
				return nil
			}

			switch event.Rune() {
			case 'q':
				app.Stop()
				return nil

			case '?':
				ctrl.ShowHelp()
				return nil

			case '1':
				ctrl.FocusRegistry()
				return nil

			case '2':
				ctrl.FocusSearch()
				return nil

			case '3':
				ctrl.FocusArtifacts()
				return nil

			case '4':
				ctrl.FocusDetails()
				return nil

			case 'S': // Shift+S for settings
				ctrl.ShowSettings()
				return nil

			case 'T': // Shift+T for theme picker
				ctrl.ShowThemePicker()
				return nil
			}
		} else {
			// In input mode, only handle Ctrl+C and Tab
			switch event.Key() {
			case tcell.KeyCtrlC:
				app.Stop()
				return nil
			case tcell.KeyTab:
				ctrl.CycleFocus()
				return nil
			}
		}

		return event
	})
}

// GetHelpText returns the help text for display
func GetHelpText() string {
	emphasis := theme.Tag("emphasis")
	text := theme.Tag("text")
	success := theme.Tag("success")
	muted := theme.Tag("muted")

	return fmt.Sprintf(`%sKeybindings%s

%sNavigation%s
  Tab         Cycle focus between panels
  Shift+Tab   Go back to registry list
  1           Focus registry list
  2 or /      Focus search
  3           Focus artifacts list
  4           Focus details panel
  j/k         Move down/up (vim-style)
  g/G         Scroll to top/bottom (in details)
  Enter       Select item

%sSearch%s
  /           Start search
  Enter       Execute search
  Esc         Cancel search / clear filter

%sArtifact Actions%s
  p           Pull artifact (shows options)
  d           Pull & load to Docker directly

%sSettings%s
  S           Open settings modal
  T           Open theme picker

%sGeneral%s
  ?           Show this help
  q           Quit
  Ctrl+C      Force quit

%sPress Esc or Enter to close this help.%s`,
		emphasis, text,
		success, text,
		success, text,
		success, text,
		success, text,
		success, text,
		muted, theme.ResetTag(),
	)
}
