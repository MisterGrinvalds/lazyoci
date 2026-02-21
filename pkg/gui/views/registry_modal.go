package views

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mistergrinvalds/lazyoci/pkg/config"
	"github.com/mistergrinvalds/lazyoci/pkg/gui/theme"
	"github.com/rivo/tview"
)

// registryPreset defines a quick-add registry template.
type registryPreset struct {
	Label       string // dropdown display label
	Name        string // auto-filled name field
	URL         string // auto-filled URL (empty = user must type)
	Placeholder string // URL placeholder hint when URL is empty
}

var registryPresets = []registryPreset{
	{Label: "Custom...", Name: "", URL: "", Placeholder: "e.g., harbor.example.com"},
	{Label: "Docker Hub", Name: "Docker Hub", URL: "docker.io", Placeholder: ""},
	{Label: "GitHub Packages", Name: "GitHub Packages", URL: "ghcr.io", Placeholder: ""},
	{Label: "Quay.io", Name: "Quay.io", URL: "quay.io", Placeholder: ""},
	{Label: "GitLab Registry", Name: "GitLab Registry", URL: "registry.gitlab.com", Placeholder: ""},
	{Label: "AWS ECR", Name: "AWS ECR", URL: "", Placeholder: "<account>.dkr.ecr.<region>.amazonaws.com"},
	{Label: "Google Artifact Registry", Name: "Google AR", URL: "", Placeholder: "<region>-docker.pkg.dev"},
	{Label: "Azure ACR", Name: "Azure ACR", URL: "", Placeholder: "<name>.azurecr.io"},
	{Label: "Red Hat Registry", Name: "Red Hat", URL: "registry.access.redhat.com", Placeholder: ""},
}

// RegistryModal is a modal dialog for adding/editing registries.
type RegistryModal struct {
	Form     *tview.Form
	Flex     *tview.Flex
	onSave   func(name, url, username, password string, insecure bool)
	onCancel func()

	presetDropdown *tview.DropDown
	nameField      *tview.InputField
	urlField       *tview.InputField
	userField      *tview.InputField
	passField      *tview.InputField
	insecureCheck  *tview.Checkbox

	isEditing  bool
	editingURL string
}

// NewRegistryModal creates a new registry modal.
func NewRegistryModal(onSave func(name, url, username, password string, insecure bool), onCancel func()) *RegistryModal {
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

	// --- Name and URL fields (created first so preset callback can access them) ---
	rm.nameField = tview.NewInputField().
		SetLabel("  Name: ").
		SetFieldWidth(40).
		SetPlaceholder("e.g., My Harbor (optional)")

	rm.urlField = tview.NewInputField().
		SetLabel("   URL: ").
		SetFieldWidth(40).
		SetPlaceholder("e.g., harbor.example.com")

	// --- Preset dropdown (created after fields it references) ---
	presetLabels := make([]string, len(registryPresets))
	for i, p := range registryPresets {
		presetLabels[i] = p.Label
	}

	rm.presetDropdown = tview.NewDropDown().
		SetLabel("Preset: ").
		SetOptions(presetLabels, rm.onPresetSelected).
		SetCurrentOption(0)

	// --- Auth fields ---
	rm.userField = tview.NewInputField().
		SetLabel("  User: ").
		SetFieldWidth(40).
		SetPlaceholder("(optional)")

	rm.passField = tview.NewInputField().
		SetLabel("  Pass: ").
		SetFieldWidth(40).
		SetPlaceholder("(optional)").
		SetMaskCharacter('*')

	// --- Insecure checkbox ---
	rm.insecureCheck = tview.NewCheckbox().
		SetLabel("Insecure (HTTP): ").
		SetChecked(false)

	// --- Assemble form ---
	rm.Form.AddFormItem(rm.presetDropdown)
	rm.Form.AddFormItem(rm.nameField)
	rm.Form.AddFormItem(rm.urlField)
	rm.Form.AddFormItem(rm.userField)
	rm.Form.AddFormItem(rm.passField)
	rm.Form.AddFormItem(rm.insecureCheck)

	rm.Form.AddButton("Save", func() {
		url := rm.urlField.GetText()
		if url != "" && rm.onSave != nil {
			rm.onSave(
				rm.nameField.GetText(),
				url,
				rm.userField.GetText(),
				rm.passField.GetText(),
				rm.insecureCheck.IsChecked(),
			)
		}
	})

	rm.Form.AddButton("Cancel", func() {
		if rm.onCancel != nil {
			rm.onCancel()
		}
	})

	rm.Form.SetBorder(true).SetTitle(" Add Registry ")
	rm.Form.SetButtonsAlign(tview.AlignCenter)

	// Handle escape key on the form
	rm.Form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			if rm.onCancel != nil {
				rm.onCancel()
			}
			return nil
		}
		return event
	})

	// Center the form: 62 cols wide, 14 rows tall
	rm.Flex = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(rm.Form, 14, 1, true).
			AddItem(nil, 0, 1, false), 62, 1, true).
		AddItem(nil, 0, 1, false)
}

func (rm *RegistryModal) onPresetSelected(_ string, index int) {
	if index < 0 || index >= len(registryPresets) {
		return
	}
	preset := registryPresets[index]

	rm.nameField.SetText(preset.Name)
	rm.urlField.SetText(preset.URL)

	// Update URL placeholder for cloud presets
	if preset.Placeholder != "" {
		rm.urlField.SetPlaceholder(preset.Placeholder)
	} else {
		rm.urlField.SetPlaceholder("e.g., harbor.example.com")
	}
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

	// Preset dropdown
	rm.presetDropdown.SetBackgroundColor(theme.BackgroundColor())
	rm.presetDropdown.SetFieldBackgroundColor(theme.ElementBgColor())
	rm.presetDropdown.SetFieldTextColor(theme.TextColor())
	rm.presetDropdown.SetLabelColor(theme.TextColor())
	rm.presetDropdown.SetListStyles(
		tcell.StyleDefault.Background(theme.ElementBgColor()).Foreground(theme.TextColor()),
		tcell.StyleDefault.Background(theme.SelectionBgColor()).Foreground(theme.SelectionFgColor()),
	)

	// InputFields
	for _, field := range []*tview.InputField{rm.nameField, rm.urlField, rm.userField, rm.passField} {
		field.SetBackgroundColor(theme.BackgroundColor())
		field.SetFieldBackgroundColor(theme.ElementBgColor())
		field.SetFieldTextColor(theme.TextColor())
		field.SetLabelColor(theme.TextColor())
		field.SetPlaceholderTextColor(theme.PlaceholderColor())
	}

	// Checkbox
	rm.insecureCheck.SetBackgroundColor(theme.BackgroundColor())
	rm.insecureCheck.SetFieldBackgroundColor(theme.ElementBgColor())
	rm.insecureCheck.SetFieldTextColor(theme.TextColor())
	rm.insecureCheck.SetLabelColor(theme.TextColor())

	// Flex
	rm.Flex.SetBackgroundColor(theme.BackgroundColor())
}

// Clear resets the form fields and switches back to add mode.
func (rm *RegistryModal) Clear() {
	rm.presetDropdown.SetCurrentOption(0)
	rm.nameField.SetText("")
	rm.urlField.SetText("")
	rm.urlField.SetPlaceholder("e.g., harbor.example.com")
	rm.userField.SetText("")
	rm.passField.SetText("")
	rm.insecureCheck.SetChecked(false)
	rm.isEditing = false
	rm.editingURL = ""
	rm.Form.SetTitle(" Add Registry ")
}

// SetRegistry pre-fills the modal for editing an existing registry.
func (rm *RegistryModal) SetRegistry(reg config.Registry) {
	rm.isEditing = true
	rm.editingURL = reg.URL

	// Set preset dropdown to "Custom..." since we're editing
	rm.presetDropdown.SetCurrentOption(0)

	rm.nameField.SetText(reg.Name)
	rm.urlField.SetText(reg.URL)
	rm.urlField.SetPlaceholder("e.g., harbor.example.com")
	rm.userField.SetText(reg.Username)
	rm.passField.SetText(reg.Password)
	rm.insecureCheck.SetChecked(reg.Insecure)

	rm.Form.SetTitle(" Edit Registry ")
}

// IsEditing returns whether the modal is in edit mode.
func (rm *RegistryModal) IsEditing() bool {
	return rm.isEditing
}

// EditingURL returns the original URL of the registry being edited.
func (rm *RegistryModal) EditingURL() string {
	return rm.editingURL
}

// SetURL sets the URL field.
func (rm *RegistryModal) SetURL(url string) {
	rm.urlField.SetText(url)
}

// GetPrimitive returns the flex container for display.
func (rm *RegistryModal) GetPrimitive() tview.Primitive {
	return rm.Flex
}
