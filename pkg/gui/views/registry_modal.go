package views

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mistergrinvalds/lazyoci/pkg/gui/theme"
	"github.com/rivo/tview"
)

// RegistryModal is a modal dialog for adding/editing registries
type RegistryModal struct {
	Form     *tview.Form
	Flex     *tview.Flex
	onSave   func(name, url, username, password string)
	onCancel func()

	nameField *tview.InputField
	urlField  *tview.InputField
	userField *tview.InputField
	passField *tview.InputField
}

// NewRegistryModal creates a new registry modal
func NewRegistryModal(onSave func(name, url, username, password string), onCancel func()) *RegistryModal {
	rm := &RegistryModal{
		onSave:   onSave,
		onCancel: onCancel,
	}

	rm.setupUI()
	rm.ApplyTheme()
	return rm
}

func (rm *RegistryModal) setupUI() {
	rm.Form = tview.NewForm()

	rm.nameField = tview.NewInputField().
		SetLabel("Name: ").
		SetFieldWidth(40).
		SetPlaceholder("e.g., My Harbor (optional)")

	rm.urlField = tview.NewInputField().
		SetLabel("URL: ").
		SetFieldWidth(40).
		SetPlaceholder("e.g., harbor.example.com")

	rm.userField = tview.NewInputField().
		SetLabel("Username: ").
		SetFieldWidth(30).
		SetPlaceholder("(optional)")

	rm.passField = tview.NewInputField().
		SetLabel("Password: ").
		SetFieldWidth(30).
		SetPlaceholder("(optional)").
		SetMaskCharacter('*')

	rm.Form.AddFormItem(rm.nameField)
	rm.Form.AddFormItem(rm.urlField)
	rm.Form.AddFormItem(rm.userField)
	rm.Form.AddFormItem(rm.passField)

	rm.Form.AddButton("Save", func() {
		url := rm.urlField.GetText()
		if url != "" && rm.onSave != nil {
			rm.onSave(rm.nameField.GetText(), url, rm.userField.GetText(), rm.passField.GetText())
		}
	})

	rm.Form.AddButton("Cancel", func() {
		if rm.onCancel != nil {
			rm.onCancel()
		}
	})

	rm.Form.SetBorder(true).SetTitle(" Add Registry ")
	rm.Form.SetButtonsAlign(tview.AlignCenter)

	// Handle escape key
	rm.Form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			if rm.onCancel != nil {
				rm.onCancel()
			}
			return nil
		}
		return event
	})

	// Center the form in a flex container
	rm.Flex = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(rm.Form, 15, 1, true).
			AddItem(nil, 0, 1, false), 50, 1, true).
		AddItem(nil, 0, 1, false)
}

// ApplyTheme applies the current theme to this view's widgets.
func (rm *RegistryModal) ApplyTheme() {
	// Form
	rm.Form.SetBackgroundColor(theme.BackgroundColor())
	rm.Form.SetBorderColor(theme.BorderNormalColor())
	rm.Form.SetTitleColor(theme.TitleColor())
	rm.Form.SetFieldBackgroundColor(theme.ElementBgColor())
	rm.Form.SetFieldTextColor(theme.TextColor())
	rm.Form.SetLabelColor(theme.TextColor())
	rm.Form.SetButtonBackgroundColor(theme.ElementBgColor())
	rm.Form.SetButtonTextColor(theme.TextColor())

	// InputFields
	for _, field := range []*tview.InputField{rm.nameField, rm.urlField, rm.userField, rm.passField} {
		field.SetBackgroundColor(theme.BackgroundColor())
		field.SetFieldBackgroundColor(theme.ElementBgColor())
		field.SetFieldTextColor(theme.TextColor())
		field.SetLabelColor(theme.TextColor())
		field.SetPlaceholderTextColor(theme.PlaceholderColor())
	}

	// Flex
	rm.Flex.SetBackgroundColor(theme.BackgroundColor())
}

// Clear resets the form fields
func (rm *RegistryModal) Clear() {
	rm.nameField.SetText("")
	rm.urlField.SetText("")
	rm.userField.SetText("")
	rm.passField.SetText("")
}

// SetURL sets the URL field (for editing)
func (rm *RegistryModal) SetURL(url string) {
	rm.urlField.SetText(url)
}

// GetPrimitive returns the flex container for display
func (rm *RegistryModal) GetPrimitive() tview.Primitive {
	return rm.Flex
}
