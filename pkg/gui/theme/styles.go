package theme

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ApplyToTview configures tview's global styles with the current theme.
// This should be called once at startup and again whenever the theme changes.
func ApplyToTview() {
	tview.Styles.PrimitiveBackgroundColor = BackgroundColor()
	tview.Styles.ContrastBackgroundColor = ElementBgColor()
	tview.Styles.MoreContrastBackgroundColor = ElementBgColor()
	tview.Styles.BorderColor = BorderNormalColor()
	tview.Styles.TitleColor = TitleColor()
	tview.Styles.GraphicsColor = BorderNormalColor()
	tview.Styles.PrimaryTextColor = TextColor()
	tview.Styles.SecondaryTextColor = TextMutedColor()
	tview.Styles.TertiaryTextColor = TextMutedColor()
	tview.Styles.InverseTextColor = SelectionFgColor()
	tview.Styles.ContrastSecondaryTextColor = TextColor()
}

// --- tcell.Color accessors (for widget properties) ---

// PrimaryColor returns the primary color for the current theme.
func PrimaryColor() tcell.Color {
	return CurrentTheme().Primary().Resolve(IsDark())
}

// SecondaryColor returns the secondary color.
func SecondaryColor() tcell.Color {
	return CurrentTheme().Secondary().Resolve(IsDark())
}

// AccentColor returns the accent color.
func AccentColor() tcell.Color {
	return CurrentTheme().Accent().Resolve(IsDark())
}

// TextColor returns the main text color.
func TextColor() tcell.Color {
	return CurrentTheme().Text().Resolve(IsDark())
}

// TextMutedColor returns the muted/dim text color.
func TextMutedColor() tcell.Color {
	return CurrentTheme().TextMuted().Resolve(IsDark())
}

// TextEmphasisColor returns the emphasized text color (for titles, headings).
func TextEmphasisColor() tcell.Color {
	return CurrentTheme().TextEmphasis().Resolve(IsDark())
}

// BackgroundColor returns the main background color.
func BackgroundColor() tcell.Color {
	return CurrentTheme().Background().Resolve(IsDark())
}

// BackgroundSecondaryColor returns the secondary background color.
func BackgroundSecondaryColor() tcell.Color {
	return CurrentTheme().BackgroundSecondary().Resolve(IsDark())
}

// ElementBgColor returns the background color for inputs and modals.
func ElementBgColor() tcell.Color {
	return CurrentTheme().BackgroundElement().Resolve(IsDark())
}

// BorderNormalColor returns the normal border color.
func BorderNormalColor() tcell.Color {
	return CurrentTheme().BorderNormal().Resolve(IsDark())
}

// BorderFocusedColor returns the focused border color.
func BorderFocusedColor() tcell.Color {
	return CurrentTheme().BorderFocused().Resolve(IsDark())
}

// SelectionBgColor returns the selection background color.
func SelectionBgColor() tcell.Color {
	return CurrentTheme().SelectionBg().Resolve(IsDark())
}

// SelectionFgColor returns the selection foreground color.
func SelectionFgColor() tcell.Color {
	return CurrentTheme().SelectionFg().Resolve(IsDark())
}

// HeaderColor returns the table header color.
func HeaderColor() tcell.Color {
	return CurrentTheme().TableHeader().Resolve(IsDark())
}

// TitleColor returns the title color (alias for TextEmphasis).
func TitleColor() tcell.Color {
	return CurrentTheme().TextEmphasis().Resolve(IsDark())
}

// DescriptionColor returns the color for description/muted text in tables.
func DescriptionColor() tcell.Color {
	return CurrentTheme().TextMuted().Resolve(IsDark())
}

// SuccessColor returns the success color.
func SuccessColor() tcell.Color {
	return CurrentTheme().Success().Resolve(IsDark())
}

// WarningColor returns the warning color.
func WarningColor() tcell.Color {
	return CurrentTheme().Warning().Resolve(IsDark())
}

// ErrorColor returns the error color.
func ErrorColor() tcell.Color {
	return CurrentTheme().Error().Resolve(IsDark())
}

// InfoColor returns the info color.
func InfoColor() tcell.Color {
	return CurrentTheme().Info().Resolve(IsDark())
}

// ModalBgColor returns the modal background color.
func ModalBgColor() tcell.Color {
	return CurrentTheme().BackgroundElement().Resolve(IsDark())
}

// ButtonBgColor returns the button background color.
func ButtonBgColor() tcell.Color {
	return CurrentTheme().BackgroundElement().Resolve(IsDark())
}

// ButtonTextColor returns the button text color.
func ButtonTextColor() tcell.Color {
	return CurrentTheme().Text().Resolve(IsDark())
}

// PlaceholderColor returns the placeholder text color.
func PlaceholderColor() tcell.Color {
	return CurrentTheme().TextMuted().Resolve(IsDark())
}

// --- Artifact type colors ---

// TypeImageColor returns the color for image artifacts.
func TypeImageColor() tcell.Color {
	return CurrentTheme().TypeImage().Resolve(IsDark())
}

// TypeHelmColor returns the color for helm chart artifacts.
func TypeHelmColor() tcell.Color {
	return CurrentTheme().TypeHelm().Resolve(IsDark())
}

// TypeSBOMColor returns the color for SBOM artifacts.
func TypeSBOMColor() tcell.Color {
	return CurrentTheme().TypeSBOM().Resolve(IsDark())
}

// TypeSignatureColor returns the color for signature artifacts.
func TypeSignatureColor() tcell.Color {
	return CurrentTheme().TypeSignature().Resolve(IsDark())
}

// TypeAttestationColor returns the color for attestation artifacts.
func TypeAttestationColor() tcell.Color {
	return CurrentTheme().TypeAttestation().Resolve(IsDark())
}

// TypeWASMColor returns the color for WASM artifacts.
func TypeWASMColor() tcell.Color {
	return CurrentTheme().TypeWASM().Resolve(IsDark())
}

// TypeUnknownColor returns the color for unknown artifacts.
func TypeUnknownColor() tcell.Color {
	return CurrentTheme().TypeUnknown().Resolve(IsDark())
}

// --- tview inline color tag helpers ---

// Tag returns a tview inline color tag for a semantic color name.
// Example: Tag("success") returns "[#a6e3a1]" for catppuccin-mocha dark.
func Tag(name string) string {
	t := CurrentTheme()
	dark := IsDark()

	switch name {
	// Text
	case "text":
		return t.Text().Tag(dark)
	case "muted":
		return t.TextMuted().Tag(dark)
	case "emphasis":
		return t.TextEmphasis().Tag(dark)

	// Status
	case "success":
		return t.Success().Tag(dark)
	case "warning":
		return t.Warning().Tag(dark)
	case "error":
		return t.Error().Tag(dark)
	case "info":
		return t.Info().Tag(dark)

	// Brand
	case "primary":
		return t.Primary().Tag(dark)
	case "secondary":
		return t.Secondary().Tag(dark)
	case "accent":
		return t.Accent().Tag(dark)

	// Table
	case "header":
		return t.TableHeader().Tag(dark)

	// Dim (alias for muted)
	case "dim":
		return t.TextMuted().Tag(dark)

	default:
		return "[-]"
	}
}

// ResetTag returns the tview color reset tag.
func ResetTag() string {
	return "[-]"
}

// ArtifactTypeTag returns a tview color tag for an artifact type short name.
func ArtifactTypeTag(typeName string) string {
	t := CurrentTheme()
	dark := IsDark()

	var cp ColorPair
	switch typeName {
	case "image":
		cp = t.TypeImage()
	case "helm":
		cp = t.TypeHelm()
	case "sbom":
		cp = t.TypeSBOM()
	case "sig":
		cp = t.TypeSignature()
	case "att":
		cp = t.TypeAttestation()
	case "wasm":
		cp = t.TypeWASM()
	case "...", "-":
		cp = t.TextMuted()
	case "?":
		cp = t.Error()
	default:
		cp = t.TypeUnknown()
	}

	return fmt.Sprintf("%s%s%s", cp.Tag(dark), typeName, ResetTag())
}

// StatusTag returns a colored status string.
func StatusTag(status string) string {
	t := CurrentTheme()
	dark := IsDark()

	switch status {
	case "available":
		return fmt.Sprintf("%savailable%s", t.Success().Tag(dark), ResetTag())
	default:
		return status
	}
}
