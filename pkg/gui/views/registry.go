package views

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mistergrinvalds/lazyoci/pkg/config"
	"github.com/mistergrinvalds/lazyoci/pkg/gui/theme"
	"github.com/mistergrinvalds/lazyoci/pkg/registry"
	"github.com/rivo/tview"
)

// RegistryView displays the list of configured registries
type RegistryView struct {
	List     *tview.List
	registry *registry.Client
	onSelect func(registryURL string)
	onAdd    func()
	onEdit   func(registryURL string)
	onDelete func(registryURL string)

	registries []config.Registry
	selected   int
}

// NewRegistryView creates a new registry list view
func NewRegistryView(reg *registry.Client, onSelect func(registryURL string)) *RegistryView {
	rv := &RegistryView{
		registry: reg,
		onSelect: onSelect,
	}

	rv.List = tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true)

	rv.List.SetBorder(true).SetTitle(" [1] Registries (a:add e:edit d:del) ")

	// Apply theme styling
	rv.ApplyTheme()

	rv.loadRegistries()

	// Handle selection
	rv.List.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index < len(rv.registries) {
			rv.selected = index
			if rv.onSelect != nil {
				rv.onSelect(rv.registries[index].URL)
			}
		}
	})

	// Handle keybindings
	rv.List.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'a', 'A':
			if rv.onAdd != nil {
				rv.onAdd()
			}
			return nil
		case 'e', 'E':
			if rv.onEdit != nil {
				url := rv.GetSelectedRegistry()
				if url != "" {
					rv.onEdit(url)
				}
			}
			return nil
		case 'd', 'D':
			if rv.onDelete != nil {
				url := rv.GetSelectedRegistry()
				if url != "" {
					rv.onDelete(url)
				}
			}
			return nil
		}
		return event
	})

	return rv
}

// ApplyTheme applies the current theme to this view's widgets.
func (rv *RegistryView) ApplyTheme() {
	rv.List.SetBackgroundColor(theme.BackgroundColor())
	rv.List.SetBorderColor(theme.BorderNormalColor())
	rv.List.SetTitleColor(theme.TitleColor())
	rv.List.SetMainTextColor(theme.TextColor())
	rv.List.SetSelectedBackgroundColor(theme.SelectionBgColor())
	rv.List.SetSelectedTextColor(theme.SelectionFgColor())
	rv.List.SetSecondaryTextColor(theme.TextMutedColor())
}

// SetOnAdd sets the callback for adding a registry
func (rv *RegistryView) SetOnAdd(fn func()) {
	rv.onAdd = fn
}

// SetOnEdit sets the callback for editing a registry
func (rv *RegistryView) SetOnEdit(fn func(registryURL string)) {
	rv.onEdit = fn
}

// SetOnDelete sets the callback for deleting a registry
func (rv *RegistryView) SetOnDelete(fn func(registryURL string)) {
	rv.onDelete = fn
}

func (rv *RegistryView) loadRegistries() {
	rv.List.Clear()
	rv.registries = rv.registry.GetRegistries()

	for _, reg := range rv.registries {
		name := reg.Name
		if name == "" || name == reg.URL {
			name = reg.URL
		} else {
			name = name + " (" + reg.URL + ")"
		}
		rv.List.AddItem(name, "", 0, nil)
	}

	if len(rv.registries) > 0 {
		rv.List.SetCurrentItem(0)
	}
}

// GetSelectedRegistry returns the currently selected registry URL
func (rv *RegistryView) GetSelectedRegistry() string {
	index := rv.List.GetCurrentItem()
	if index >= 0 && index < len(rv.registries) {
		return rv.registries[index].URL
	}
	if len(rv.registries) > 0 {
		return rv.registries[0].URL
	}
	return "docker.io"
}

// Refresh reloads the registry list
func (rv *RegistryView) Refresh() {
	rv.loadRegistries()
}
