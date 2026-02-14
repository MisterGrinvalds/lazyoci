package views

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mistergrinvalds/lazyoci/pkg/gui/theme"
	"github.com/rivo/tview"
)

// ConfirmOption represents a button option in the confirm modal
type ConfirmOption struct {
	Label    string
	Callback func()
}

// ConfirmModal is a generic confirmation dialog
type ConfirmModal struct {
	Modal *tview.Modal
	Flex  *tview.Flex
}

// NewConfirmModal creates a new confirmation modal
func NewConfirmModal(title, message string, options []ConfirmOption) *ConfirmModal {
	cm := &ConfirmModal{}

	// Extract button labels
	labels := make([]string, len(options))
	for i, opt := range options {
		labels[i] = opt.Label
	}

	cm.Modal = tview.NewModal().
		SetText(message).
		AddButtons(labels).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex >= 0 && buttonIndex < len(options) {
				if options[buttonIndex].Callback != nil {
					options[buttonIndex].Callback()
				}
			}
		})

	cm.Modal.SetBorder(true)
	if title != "" {
		cm.Modal.SetTitle(" " + title + " ")
	}

	// Apply theme styling
	cm.ApplyTheme()

	// Center the modal
	cm.Flex = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(cm.Modal, 10, 1, true).
			AddItem(nil, 0, 1, false), 60, 1, true).
		AddItem(nil, 0, 1, false)

	cm.Flex.SetBackgroundColor(theme.BackgroundColor())

	return cm
}

// ApplyTheme applies the current theme to this modal.
func (cm *ConfirmModal) ApplyTheme() {
	cm.Modal.SetBackgroundColor(theme.ModalBgColor())
	cm.Modal.SetBorderColor(theme.BorderNormalColor())
	cm.Modal.SetTitleColor(theme.TitleColor())
	cm.Modal.SetTextColor(theme.TextColor())
	cm.Modal.SetButtonBackgroundColor(theme.ButtonBgColor())
	cm.Modal.SetButtonTextColor(theme.ButtonTextColor())
}

// GetPrimitive returns the flex container for display
func (cm *ConfirmModal) GetPrimitive() tview.Primitive {
	return cm.Flex
}

// SetInputCapture sets a custom input handler
func (cm *ConfirmModal) SetInputCapture(handler func(event *tcell.EventKey) *tcell.EventKey) {
	cm.Modal.SetInputCapture(handler)
}

// ConfirmAction shows a simple Yes/No confirmation
func ConfirmAction(title, message string, onYes, onNo func()) *ConfirmModal {
	return NewConfirmModal(title, message, []ConfirmOption{
		{Label: "Yes", Callback: onYes},
		{Label: "No", Callback: onNo},
	})
}

// ConfirmWithCancel shows a Yes/No/Cancel confirmation
func ConfirmWithCancel(title, message string, onYes, onNo, onCancel func()) *ConfirmModal {
	return NewConfirmModal(title, message, []ConfirmOption{
		{Label: "Yes", Callback: onYes},
		{Label: "No", Callback: onNo},
		{Label: "Cancel", Callback: onCancel},
	})
}
