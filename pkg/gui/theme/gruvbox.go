package theme

func init() {
	RegisterTheme("gruvbox", NewGruvboxTheme())
}

// Gruvbox Dark palette
// https://github.com/morhetz/gruvbox
const (
	gruvboxDarkBg0    = "#282828"
	gruvboxDarkBg1    = "#3c3836"
	gruvboxDarkBg2    = "#504945"
	gruvboxDarkFg0    = "#fbf1c7"
	gruvboxDarkFg1    = "#ebdbb2"
	gruvboxDarkFg4    = "#a89984"
	gruvboxDarkGray   = "#928374"
	gruvboxDarkRed    = "#fb4934"
	gruvboxDarkGreen  = "#b8bb26"
	gruvboxDarkYellow = "#fabd2f"
	gruvboxDarkBlue   = "#83a598"
	gruvboxDarkPurple = "#d3869b"
	gruvboxDarkAqua   = "#8ec07c"
	gruvboxDarkOrange = "#fe8019"

	// Gruvbox Light palette
	gruvboxLightBg0    = "#fbf1c7"
	gruvboxLightBg1    = "#ebdbb2"
	gruvboxLightBg2    = "#d5c4a1"
	gruvboxLightFg0    = "#282828"
	gruvboxLightFg1    = "#3c3836"
	gruvboxLightFg4    = "#665c54"
	gruvboxLightGray   = "#928374"
	gruvboxLightRed    = "#9d0006"
	gruvboxLightGreen  = "#79740e"
	gruvboxLightYellow = "#b57614"
	gruvboxLightBlue   = "#076678"
	gruvboxLightPurple = "#8f3f71"
	gruvboxLightAqua   = "#427b58"
	gruvboxLightOrange = "#af3a03"
)

// NewGruvboxTheme creates the Gruvbox theme.
func NewGruvboxTheme() Theme {
	return &BaseTheme{
		ThemeName: "Gruvbox",

		PrimaryColor:   ColorPair{Dark: gruvboxDarkAqua, Light: gruvboxLightAqua},
		SecondaryColor: ColorPair{Dark: gruvboxDarkBlue, Light: gruvboxLightBlue},
		AccentColor:    ColorPair{Dark: gruvboxDarkOrange, Light: gruvboxLightOrange},

		TextColor:         ColorPair{Dark: gruvboxDarkFg1, Light: gruvboxLightFg1},
		TextMutedColor:    ColorPair{Dark: gruvboxDarkGray, Light: gruvboxLightGray},
		TextEmphasisColor: ColorPair{Dark: gruvboxDarkYellow, Light: gruvboxLightYellow},

		BackgroundColor:          ColorPair{Dark: gruvboxDarkBg0, Light: gruvboxLightBg0},
		BackgroundSecondaryColor: ColorPair{Dark: "#1d2021", Light: "#f9f5d7"},
		BackgroundElementColor:   ColorPair{Dark: gruvboxDarkBg1, Light: gruvboxLightBg1},

		BorderNormalColor:  ColorPair{Dark: gruvboxDarkFg4, Light: gruvboxLightFg4},
		BorderFocusedColor: ColorPair{Dark: gruvboxDarkAqua, Light: gruvboxLightAqua},

		SuccessColor: ColorPair{Dark: gruvboxDarkGreen, Light: gruvboxLightGreen},
		WarningColor: ColorPair{Dark: gruvboxDarkYellow, Light: gruvboxLightYellow},
		ErrorColor:   ColorPair{Dark: gruvboxDarkRed, Light: gruvboxLightRed},
		InfoColor:    ColorPair{Dark: gruvboxDarkBlue, Light: gruvboxLightBlue},

		SelectionBgColor: ColorPair{Dark: gruvboxDarkBg2, Light: gruvboxLightBg2},
		SelectionFgColor: ColorPair{Dark: gruvboxDarkFg1, Light: gruvboxLightFg1},

		TableHeaderColor: ColorPair{Dark: gruvboxDarkOrange, Light: gruvboxLightOrange},

		TypeImageColor:       ColorPair{Dark: gruvboxDarkBlue, Light: gruvboxLightBlue},
		TypeHelmColor:        ColorPair{Dark: gruvboxDarkGreen, Light: gruvboxLightGreen},
		TypeSBOMColor:        ColorPair{Dark: gruvboxDarkYellow, Light: gruvboxLightYellow},
		TypeSignatureColor:   ColorPair{Dark: gruvboxDarkPurple, Light: gruvboxLightPurple},
		TypeAttestationColor: ColorPair{Dark: gruvboxDarkAqua, Light: gruvboxLightAqua},
		TypeWASMColor:        ColorPair{Dark: gruvboxDarkRed, Light: gruvboxLightRed},
		TypeUnknownColor:     ColorPair{Dark: gruvboxDarkGray, Light: gruvboxLightGray},
	}
}
