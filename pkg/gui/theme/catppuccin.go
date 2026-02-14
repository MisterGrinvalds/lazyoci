package theme

func init() {
	RegisterTheme("catppuccin-mocha", NewCatppuccinMochaTheme())
	RegisterTheme("catppuccin-latte", NewCatppuccinLatteTheme())
}

// Catppuccin Mocha palette
// https://github.com/catppuccin/catppuccin
const (
	mochaRosewater = "#f5e0dc"
	mochaFlamingo  = "#f2cdcd"
	mochaPink      = "#f5c2e7"
	mochaMauve     = "#cba6f7"
	mochaRed       = "#f38ba8"
	mochaMaroon    = "#eba0ac"
	mochaPeach     = "#fab387"
	mochaYellow    = "#f9e2af"
	mochaGreen     = "#a6e3a1"
	mochaTeal      = "#94e2d5"
	mochaSky       = "#89dceb"
	mochaSapphire  = "#74c7ec"
	mochaBlue      = "#89b4fa"
	mochaLavender  = "#b4befe"
	mochaText      = "#cdd6f4"
	mochaSubtext1  = "#bac2de"
	mochaSubtext0  = "#a6adc8"
	mochaOverlay2  = "#9399b2"
	mochaOverlay1  = "#7f849c"
	mochaOverlay0  = "#6c7086"
	mochaSurface2  = "#585b70"
	mochaSurface1  = "#45475a"
	mochaSurface0  = "#313244"
	mochaBase      = "#1e1e2e"
	mochaMantle    = "#181825"
	mochaCrust     = "#11111b"
)

// Catppuccin Latte palette
const (
	latteRosewater = "#dc8a78"
	latteFlamingo  = "#dd7878"
	lattePink      = "#ea76cb"
	latteMauve     = "#8839ef"
	latteRed       = "#d20f39"
	latteMaroon    = "#e64553"
	lattePeach     = "#fe640b"
	latteYellow    = "#df8e1d"
	latteGreen     = "#40a02b"
	latteTeal      = "#179299"
	latteSky       = "#04a5e5"
	latteSapphire  = "#209fb5"
	latteBlue      = "#1e66f5"
	latteLavender  = "#7287fd"
	latteText      = "#4c4f69"
	latteSubtext1  = "#5c5f77"
	latteSubtext0  = "#6c6f85"
	latteOverlay2  = "#7c7f93"
	latteOverlay1  = "#8c8fa1"
	latteOverlay0  = "#9ca0b0"
	latteSurface2  = "#acb0be"
	latteSurface1  = "#bcc0cc"
	latteSurface0  = "#ccd0da"
	latteBase      = "#eff1f5"
	latteMantle    = "#e6e9ef"
	latteCrust     = "#dce0e8"
)

// NewCatppuccinMochaTheme creates the Catppuccin Mocha (dark) theme.
func NewCatppuccinMochaTheme() Theme {
	return &BaseTheme{
		ThemeName: "Catppuccin Mocha",

		PrimaryColor:   ColorPair{Dark: mochaBlue, Light: latteBlue},
		SecondaryColor: ColorPair{Dark: mochaLavender, Light: latteLavender},
		AccentColor:    ColorPair{Dark: mochaMauve, Light: latteMauve},

		TextColor:         ColorPair{Dark: mochaText, Light: latteText},
		TextMutedColor:    ColorPair{Dark: mochaOverlay1, Light: latteOverlay1},
		TextEmphasisColor: ColorPair{Dark: mochaYellow, Light: latteYellow},

		BackgroundColor:          ColorPair{Dark: mochaBase, Light: latteBase},
		BackgroundSecondaryColor: ColorPair{Dark: mochaMantle, Light: latteMantle},
		BackgroundElementColor:   ColorPair{Dark: mochaSurface0, Light: latteSurface0},

		BorderNormalColor:  ColorPair{Dark: mochaSurface1, Light: latteSurface1},
		BorderFocusedColor: ColorPair{Dark: mochaBlue, Light: latteBlue},

		SuccessColor: ColorPair{Dark: mochaGreen, Light: latteGreen},
		WarningColor: ColorPair{Dark: mochaYellow, Light: latteYellow},
		ErrorColor:   ColorPair{Dark: mochaRed, Light: latteRed},
		InfoColor:    ColorPair{Dark: mochaSapphire, Light: latteSapphire},

		SelectionBgColor: ColorPair{Dark: mochaSurface1, Light: latteSurface1},
		SelectionFgColor: ColorPair{Dark: mochaText, Light: latteText},

		TableHeaderColor: ColorPair{Dark: mochaMauve, Light: latteMauve},

		TypeImageColor:       ColorPair{Dark: mochaBlue, Light: latteBlue},
		TypeHelmColor:        ColorPair{Dark: mochaGreen, Light: latteGreen},
		TypeSBOMColor:        ColorPair{Dark: mochaYellow, Light: latteYellow},
		TypeSignatureColor:   ColorPair{Dark: mochaPink, Light: lattePink},
		TypeAttestationColor: ColorPair{Dark: mochaTeal, Light: latteTeal},
		TypeWASMColor:        ColorPair{Dark: mochaRed, Light: latteRed},
		TypeUnknownColor:     ColorPair{Dark: mochaOverlay0, Light: latteOverlay0},
	}
}

// NewCatppuccinLatteTheme creates the Catppuccin Latte (light) theme.
func NewCatppuccinLatteTheme() Theme {
	return &BaseTheme{
		ThemeName: "Catppuccin Latte",

		PrimaryColor:   ColorPair{Dark: mochaBlue, Light: latteBlue},
		SecondaryColor: ColorPair{Dark: mochaLavender, Light: latteLavender},
		AccentColor:    ColorPair{Dark: mochaMauve, Light: latteMauve},

		TextColor:         ColorPair{Dark: mochaText, Light: latteText},
		TextMutedColor:    ColorPair{Dark: mochaOverlay1, Light: latteOverlay1},
		TextEmphasisColor: ColorPair{Dark: mochaYellow, Light: latteYellow},

		BackgroundColor:          ColorPair{Dark: mochaBase, Light: latteBase},
		BackgroundSecondaryColor: ColorPair{Dark: mochaMantle, Light: latteMantle},
		BackgroundElementColor:   ColorPair{Dark: mochaSurface0, Light: latteSurface0},

		BorderNormalColor:  ColorPair{Dark: mochaSurface1, Light: latteSurface1},
		BorderFocusedColor: ColorPair{Dark: mochaBlue, Light: latteBlue},

		SuccessColor: ColorPair{Dark: mochaGreen, Light: latteGreen},
		WarningColor: ColorPair{Dark: mochaYellow, Light: latteYellow},
		ErrorColor:   ColorPair{Dark: mochaRed, Light: latteRed},
		InfoColor:    ColorPair{Dark: mochaSapphire, Light: latteSapphire},

		SelectionBgColor: ColorPair{Dark: mochaSurface1, Light: latteSurface1},
		SelectionFgColor: ColorPair{Dark: mochaText, Light: latteText},

		TableHeaderColor: ColorPair{Dark: mochaMauve, Light: latteMauve},

		TypeImageColor:       ColorPair{Dark: mochaBlue, Light: latteBlue},
		TypeHelmColor:        ColorPair{Dark: mochaGreen, Light: latteGreen},
		TypeSBOMColor:        ColorPair{Dark: mochaYellow, Light: latteYellow},
		TypeSignatureColor:   ColorPair{Dark: mochaPink, Light: lattePink},
		TypeAttestationColor: ColorPair{Dark: mochaTeal, Light: latteTeal},
		TypeWASMColor:        ColorPair{Dark: mochaRed, Light: latteRed},
		TypeUnknownColor:     ColorPair{Dark: mochaOverlay0, Light: latteOverlay0},
	}
}
