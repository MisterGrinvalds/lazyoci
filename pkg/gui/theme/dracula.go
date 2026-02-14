package theme

func init() {
	RegisterTheme("dracula", NewDraculaTheme())
}

// Dracula palette
// https://draculatheme.com/contribute
const (
	draculaBackground = "#282a36"
	draculaCurrent    = "#44475a"
	draculaForeground = "#f8f8f2"
	draculaComment    = "#6272a4"
	draculaCyan       = "#8be9fd"
	draculaGreen      = "#50fa7b"
	draculaOrange     = "#ffb86c"
	draculaPink       = "#ff79c6"
	draculaPurple     = "#bd93f9"
	draculaRed        = "#ff5555"
	draculaYellow     = "#f1fa8c"

	// Dracula light approximation
	draculaLightBg      = "#f8f8f2"
	draculaLightFg      = "#282a36"
	draculaLightComment = "#7970a9"
	draculaLightCurrent = "#e6e6e6"
	draculaLightCyan    = "#0097a7"
	draculaLightGreen   = "#2e7d32"
	draculaLightOrange  = "#e65100"
	draculaLightPink    = "#c2185b"
	draculaLightPurple  = "#7c4dff"
	draculaLightRed     = "#d32f2f"
	draculaLightYellow  = "#f9a825"
	draculaLightSurface = "#d4d4d8"
)

// NewDraculaTheme creates the Dracula theme.
func NewDraculaTheme() Theme {
	return &BaseTheme{
		ThemeName: "Dracula",

		PrimaryColor:   ColorPair{Dark: draculaPurple, Light: draculaLightPurple},
		SecondaryColor: ColorPair{Dark: draculaCyan, Light: draculaLightCyan},
		AccentColor:    ColorPair{Dark: draculaPink, Light: draculaLightPink},

		TextColor:         ColorPair{Dark: draculaForeground, Light: draculaLightFg},
		TextMutedColor:    ColorPair{Dark: draculaComment, Light: draculaLightComment},
		TextEmphasisColor: ColorPair{Dark: draculaYellow, Light: draculaLightYellow},

		BackgroundColor:          ColorPair{Dark: draculaBackground, Light: draculaLightBg},
		BackgroundSecondaryColor: ColorPair{Dark: "#21222c", Light: "#ededef"},
		BackgroundElementColor:   ColorPair{Dark: draculaCurrent, Light: draculaLightCurrent},

		BorderNormalColor:  ColorPair{Dark: draculaComment, Light: draculaLightComment},
		BorderFocusedColor: ColorPair{Dark: draculaPurple, Light: draculaLightPurple},

		SuccessColor: ColorPair{Dark: draculaGreen, Light: draculaLightGreen},
		WarningColor: ColorPair{Dark: draculaYellow, Light: draculaLightYellow},
		ErrorColor:   ColorPair{Dark: draculaRed, Light: draculaLightRed},
		InfoColor:    ColorPair{Dark: draculaCyan, Light: draculaLightCyan},

		SelectionBgColor: ColorPair{Dark: draculaCurrent, Light: draculaLightSurface},
		SelectionFgColor: ColorPair{Dark: draculaForeground, Light: draculaLightFg},

		TableHeaderColor: ColorPair{Dark: draculaPurple, Light: draculaLightPurple},

		TypeImageColor:       ColorPair{Dark: draculaCyan, Light: draculaLightCyan},
		TypeHelmColor:        ColorPair{Dark: draculaGreen, Light: draculaLightGreen},
		TypeSBOMColor:        ColorPair{Dark: draculaYellow, Light: draculaLightYellow},
		TypeSignatureColor:   ColorPair{Dark: draculaPink, Light: draculaLightPink},
		TypeAttestationColor: ColorPair{Dark: draculaOrange, Light: draculaLightOrange},
		TypeWASMColor:        ColorPair{Dark: draculaRed, Light: draculaLightRed},
		TypeUnknownColor:     ColorPair{Dark: draculaComment, Light: draculaLightComment},
	}
}
