// Package theme provides a theming system for the lazyoci TUI.
//
// Themes define semantic color palettes that adapt to light and dark terminal
// backgrounds. The architecture follows the interface + base struct pattern:
// concrete themes embed BaseTheme and set color fields, which satisfy the
// Theme interface via getter methods.
package theme

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// ColorPair holds colors for both dark and light terminal backgrounds.
type ColorPair struct {
	Dark  string // hex color for dark backgrounds, e.g. "#cdd6f4"
	Light string // hex color for light backgrounds, e.g. "#4c4f69"
}

// Resolve returns the appropriate tcell.Color based on the isDark flag.
func (cp ColorPair) Resolve(isDark bool) tcell.Color {
	hex := cp.Light
	if isDark {
		hex = cp.Dark
	}
	if hex == "" {
		return tcell.ColorDefault
	}
	return tcell.GetColor(hex)
}

// Hex returns the appropriate hex string based on the isDark flag.
func (cp ColorPair) Hex(isDark bool) string {
	if isDark {
		return cp.Dark
	}
	return cp.Light
}

// Tag returns a tview inline color tag like "[#cdd6f4]" based on isDark.
func (cp ColorPair) Tag(isDark bool) string {
	hex := cp.Light
	if isDark {
		hex = cp.Dark
	}
	if hex == "" {
		return "[-]"
	}
	return fmt.Sprintf("[%s]", hex)
}

// Theme defines the semantic color interface for lazyoci.
// All color methods accept an isDark flag to support adaptive rendering.
type Theme interface {
	// Name returns the human-readable theme name.
	Name() string

	// Base colors
	Primary() ColorPair
	Secondary() ColorPair
	Accent() ColorPair

	// Text colors
	Text() ColorPair
	TextMuted() ColorPair
	TextEmphasis() ColorPair

	// Background colors
	Background() ColorPair
	BackgroundSecondary() ColorPair
	BackgroundElement() ColorPair // inputs, modals

	// Border colors
	BorderNormal() ColorPair
	BorderFocused() ColorPair

	// Status colors
	Success() ColorPair
	Warning() ColorPair
	Error() ColorPair
	Info() ColorPair

	// Selection colors
	SelectionBg() ColorPair
	SelectionFg() ColorPair

	// Table colors
	TableHeader() ColorPair

	// Artifact type colors
	TypeImage() ColorPair
	TypeHelm() ColorPair
	TypeSBOM() ColorPair
	TypeSignature() ColorPair
	TypeAttestation() ColorPair
	TypeWASM() ColorPair
	TypeUnknown() ColorPair
}

// BaseTheme provides a default implementation of Theme using exported ColorPair fields.
// Concrete themes embed this struct and set the fields in their constructor.
type BaseTheme struct {
	ThemeName string

	PrimaryColor   ColorPair
	SecondaryColor ColorPair
	AccentColor    ColorPair

	TextColor         ColorPair
	TextMutedColor    ColorPair
	TextEmphasisColor ColorPair

	BackgroundColor          ColorPair
	BackgroundSecondaryColor ColorPair
	BackgroundElementColor   ColorPair

	BorderNormalColor  ColorPair
	BorderFocusedColor ColorPair

	SuccessColor ColorPair
	WarningColor ColorPair
	ErrorColor   ColorPair
	InfoColor    ColorPair

	SelectionBgColor ColorPair
	SelectionFgColor ColorPair

	TableHeaderColor ColorPair

	TypeImageColor       ColorPair
	TypeHelmColor        ColorPair
	TypeSBOMColor        ColorPair
	TypeSignatureColor   ColorPair
	TypeAttestationColor ColorPair
	TypeWASMColor        ColorPair
	TypeUnknownColor     ColorPair
}

// Interface compliance
func (b *BaseTheme) Name() string                   { return b.ThemeName }
func (b *BaseTheme) Primary() ColorPair             { return b.PrimaryColor }
func (b *BaseTheme) Secondary() ColorPair           { return b.SecondaryColor }
func (b *BaseTheme) Accent() ColorPair              { return b.AccentColor }
func (b *BaseTheme) Text() ColorPair                { return b.TextColor }
func (b *BaseTheme) TextMuted() ColorPair           { return b.TextMutedColor }
func (b *BaseTheme) TextEmphasis() ColorPair        { return b.TextEmphasisColor }
func (b *BaseTheme) Background() ColorPair          { return b.BackgroundColor }
func (b *BaseTheme) BackgroundSecondary() ColorPair { return b.BackgroundSecondaryColor }
func (b *BaseTheme) BackgroundElement() ColorPair   { return b.BackgroundElementColor }
func (b *BaseTheme) BorderNormal() ColorPair        { return b.BorderNormalColor }
func (b *BaseTheme) BorderFocused() ColorPair       { return b.BorderFocusedColor }
func (b *BaseTheme) Success() ColorPair             { return b.SuccessColor }
func (b *BaseTheme) Warning() ColorPair             { return b.WarningColor }
func (b *BaseTheme) Error() ColorPair               { return b.ErrorColor }
func (b *BaseTheme) Info() ColorPair                { return b.InfoColor }
func (b *BaseTheme) SelectionBg() ColorPair         { return b.SelectionBgColor }
func (b *BaseTheme) SelectionFg() ColorPair         { return b.SelectionFgColor }
func (b *BaseTheme) TableHeader() ColorPair         { return b.TableHeaderColor }
func (b *BaseTheme) TypeImage() ColorPair           { return b.TypeImageColor }
func (b *BaseTheme) TypeHelm() ColorPair            { return b.TypeHelmColor }
func (b *BaseTheme) TypeSBOM() ColorPair            { return b.TypeSBOMColor }
func (b *BaseTheme) TypeSignature() ColorPair       { return b.TypeSignatureColor }
func (b *BaseTheme) TypeAttestation() ColorPair     { return b.TypeAttestationColor }
func (b *BaseTheme) TypeWASM() ColorPair            { return b.TypeWASMColor }
func (b *BaseTheme) TypeUnknown() ColorPair         { return b.TypeUnknownColor }
