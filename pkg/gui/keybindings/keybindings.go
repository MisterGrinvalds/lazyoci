package keybindings

import (
	"github.com/gdamore/tcell/v2"
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
}

// Setup configures all keybindings for the application
func Setup(app *tview.Application, ctrl Controller) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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
				showHelp(app)
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

func showHelp(app *tview.Application) {
	helpText := `[yellow]Keybindings[white]

[green]Navigation[white]
  Tab         Cycle focus between panels
  Shift+Tab   Go back to registry tree
  1           Focus registry tree
  2 or /      Focus search
  3           Focus artifacts list
  4           Focus details
  j/k         Move down/up (vim-style)
  Enter       Select item

[green]Search[white]
  /           Start search
  Enter       Execute search
  Esc         Cancel search

[green]Actions[white]
  y           Copy digest to clipboard
  p           Pull artifact (shows command)

[green]General[white]
  ?           Show this help
  q           Quit
  Ctrl+C      Force quit

Press Esc or Enter to close this help.`

	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetText(helpText).
		SetScrollable(true)
	textView.SetBorder(true).SetTitle(" Help ")

	// Capture the current root to restore later
	// Create a modal-like overlay
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(textView, 24, 1, true).
			AddItem(nil, 0, 1, false), 70, 1, true).
		AddItem(nil, 0, 1, false)

	// Store current root
	pages := tview.NewPages().
		AddPage("help", flex, true, true)

	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyEnter {
			app.Stop()
			return nil
		}
		return event
	})

	// This is a simplified help - in production, we'd overlay on the existing UI
	app.SetRoot(pages, true)
}
