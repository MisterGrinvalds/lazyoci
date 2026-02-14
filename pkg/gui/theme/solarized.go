package theme

func init() {
	RegisterTheme("solarized-dark", NewSolarizedDarkTheme())
}

// Solarized palette
// https://ethanschoonover.com/solarized/
const (
	solBase03  = "#002b36"
	solBase02  = "#073642"
	solBase01  = "#586e75"
	solBase00  = "#657b83"
	solBase0   = "#839496"
	solBase1   = "#93a1a1"
	solBase2   = "#eee8d5"
	solBase3   = "#fdf6e3"
	solYellow  = "#b58900"
	solOrange  = "#cb4b16"
	solRed     = "#dc322f"
	solMagenta = "#d33682"
	solViolet  = "#6c71c4"
	solBlue    = "#268bd2"
	solCyan    = "#2aa198"
	solGreen   = "#859900"
)

// NewSolarizedDarkTheme creates the Solarized Dark theme.
func NewSolarizedDarkTheme() Theme {
	return &BaseTheme{
		ThemeName: "Solarized Dark",

		PrimaryColor:   ColorPair{Dark: solBlue, Light: solBlue},
		SecondaryColor: ColorPair{Dark: solCyan, Light: solCyan},
		AccentColor:    ColorPair{Dark: solViolet, Light: solViolet},

		TextColor:         ColorPair{Dark: solBase0, Light: solBase00},
		TextMutedColor:    ColorPair{Dark: solBase01, Light: solBase1},
		TextEmphasisColor: ColorPair{Dark: solYellow, Light: solYellow},

		BackgroundColor:          ColorPair{Dark: solBase03, Light: solBase3},
		BackgroundSecondaryColor: ColorPair{Dark: solBase02, Light: solBase2},
		BackgroundElementColor:   ColorPair{Dark: solBase02, Light: solBase2},

		BorderNormalColor:  ColorPair{Dark: solBase01, Light: solBase1},
		BorderFocusedColor: ColorPair{Dark: solBlue, Light: solBlue},

		SuccessColor: ColorPair{Dark: solGreen, Light: solGreen},
		WarningColor: ColorPair{Dark: solYellow, Light: solYellow},
		ErrorColor:   ColorPair{Dark: solRed, Light: solRed},
		InfoColor:    ColorPair{Dark: solCyan, Light: solCyan},

		SelectionBgColor: ColorPair{Dark: solBase02, Light: solBase2},
		SelectionFgColor: ColorPair{Dark: solBase1, Light: solBase01},

		TableHeaderColor: ColorPair{Dark: solYellow, Light: solYellow},

		TypeImageColor:       ColorPair{Dark: solBlue, Light: solBlue},
		TypeHelmColor:        ColorPair{Dark: solGreen, Light: solGreen},
		TypeSBOMColor:        ColorPair{Dark: solYellow, Light: solYellow},
		TypeSignatureColor:   ColorPair{Dark: solMagenta, Light: solMagenta},
		TypeAttestationColor: ColorPair{Dark: solCyan, Light: solCyan},
		TypeWASMColor:        ColorPair{Dark: solRed, Light: solRed},
		TypeUnknownColor:     ColorPair{Dark: solBase01, Light: solBase1},
	}
}
