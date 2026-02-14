package theme

func init() {
	RegisterTheme("default", NewDefaultTheme())
}

// NewDefaultTheme creates the default theme that matches lazyoci's original colors.
// The dark palette is a 1:1 mapping of the previously hardcoded colors.
// The light palette is a hand-crafted complement.
func NewDefaultTheme() Theme {
	return &BaseTheme{
		ThemeName: "Default",

		// Base
		PrimaryColor:   ColorPair{Dark: "#00ced1", Light: "#008b8b"}, // DarkCyan-ish
		SecondaryColor: ColorPair{Dark: "#87ceeb", Light: "#4682b4"},
		AccentColor:    ColorPair{Dark: "#ffd700", Light: "#b8860b"}, // Gold

		// Text
		TextColor:         ColorPair{Dark: "#ffffff", Light: "#1a1a2e"},
		TextMutedColor:    ColorPair{Dark: "#808080", Light: "#808080"}, // Gray
		TextEmphasisColor: ColorPair{Dark: "#ffff00", Light: "#b8860b"}, // Yellow / dark gold

		// Background
		BackgroundColor:          ColorPair{Dark: "", Light: ""}, // Terminal default
		BackgroundSecondaryColor: ColorPair{Dark: "#1a1a2e", Light: "#f0f0f0"},
		BackgroundElementColor:   ColorPair{Dark: "#2f4f4f", Light: "#d3d3d3"}, // DarkSlateGray / LightGray

		// Border
		BorderNormalColor:  ColorPair{Dark: "#808080", Light: "#a0a0a0"},
		BorderFocusedColor: ColorPair{Dark: "#00ced1", Light: "#008b8b"},

		// Status
		SuccessColor: ColorPair{Dark: "#00ff00", Light: "#228b22"},
		WarningColor: ColorPair{Dark: "#ffff00", Light: "#b8860b"},
		ErrorColor:   ColorPair{Dark: "#ff0000", Light: "#cc0000"},
		InfoColor:    ColorPair{Dark: "#00ffff", Light: "#008b8b"},

		// Selection
		SelectionBgColor: ColorPair{Dark: "#008b8b", Light: "#b0e0e6"}, // DarkCyan
		SelectionFgColor: ColorPair{Dark: "#ffffff", Light: "#1a1a2e"},

		// Table
		TableHeaderColor: ColorPair{Dark: "#ffff00", Light: "#b8860b"},

		// Artifact types
		TypeImageColor:       ColorPair{Dark: "#6495ed", Light: "#4169e1"}, // Blue
		TypeHelmColor:        ColorPair{Dark: "#00ff00", Light: "#228b22"}, // Green
		TypeSBOMColor:        ColorPair{Dark: "#ffff00", Light: "#b8860b"}, // Yellow
		TypeSignatureColor:   ColorPair{Dark: "#ff00ff", Light: "#8b008b"}, // Magenta
		TypeAttestationColor: ColorPair{Dark: "#00ffff", Light: "#008b8b"}, // Cyan
		TypeWASMColor:        ColorPair{Dark: "#ff0000", Light: "#cc0000"}, // Red
		TypeUnknownColor:     ColorPair{Dark: "#808080", Light: "#808080"}, // Gray
	}
}
