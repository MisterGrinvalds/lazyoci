package theme

func init() {
	RegisterTheme("tokyonight", NewTokyoNightTheme())
}

// Tokyo Night palette
// https://github.com/enkia/tokyo-night-vscode-theme
const (
	tokyoBg       = "#1a1b26"
	tokyoBgDark   = "#16161e"
	tokyoFg       = "#c0caf5"
	tokyoComment  = "#565f89"
	tokyoDark5    = "#292e42"
	tokyoBlue     = "#7aa2f7"
	tokyoCyan     = "#7dcfff"
	tokyoGreen    = "#9ece6a"
	tokyoMagenta  = "#bb9af7"
	tokyoOrange   = "#ff9e64"
	tokyoRed      = "#f7768e"
	tokyoYellow   = "#e0af68"
	tokyoTeal     = "#73daca"
	tokyoTermBlue = "#2ac3de"

	// Tokyo Night Day palette
	tokyoDayBg      = "#e1e2e7"
	tokyoDayBgDark  = "#d0d5e3"
	tokyoDayFg      = "#3760bf"
	tokyoDayComment = "#848cb5"
	tokyoDaySurface = "#c4c8da"
	tokyoDayBlue    = "#2e7de9"
	tokyoDayCyan    = "#007197"
	tokyoDayGreen   = "#587539"
	tokyoDayMagenta = "#9854f1"
	tokyoDayOrange  = "#b15c00"
	tokyoDayRed     = "#f52a65"
	tokyoDayYellow  = "#8c6c3e"
	tokyoDayTeal    = "#118c74"
)

// NewTokyoNightTheme creates the Tokyo Night theme.
func NewTokyoNightTheme() Theme {
	return &BaseTheme{
		ThemeName: "Tokyo Night",

		PrimaryColor:   ColorPair{Dark: tokyoBlue, Light: tokyoDayBlue},
		SecondaryColor: ColorPair{Dark: tokyoMagenta, Light: tokyoDayMagenta},
		AccentColor:    ColorPair{Dark: tokyoCyan, Light: tokyoDayCyan},

		TextColor:         ColorPair{Dark: tokyoFg, Light: tokyoDayFg},
		TextMutedColor:    ColorPair{Dark: tokyoComment, Light: tokyoDayComment},
		TextEmphasisColor: ColorPair{Dark: tokyoYellow, Light: tokyoDayYellow},

		BackgroundColor:          ColorPair{Dark: tokyoBg, Light: tokyoDayBg},
		BackgroundSecondaryColor: ColorPair{Dark: tokyoBgDark, Light: tokyoDayBgDark},
		BackgroundElementColor:   ColorPair{Dark: tokyoDark5, Light: tokyoDaySurface},

		BorderNormalColor:  ColorPair{Dark: tokyoComment, Light: tokyoDayComment},
		BorderFocusedColor: ColorPair{Dark: tokyoBlue, Light: tokyoDayBlue},

		SuccessColor: ColorPair{Dark: tokyoGreen, Light: tokyoDayGreen},
		WarningColor: ColorPair{Dark: tokyoYellow, Light: tokyoDayYellow},
		ErrorColor:   ColorPair{Dark: tokyoRed, Light: tokyoDayRed},
		InfoColor:    ColorPair{Dark: tokyoTermBlue, Light: tokyoDayCyan},

		SelectionBgColor: ColorPair{Dark: tokyoDark5, Light: tokyoDaySurface},
		SelectionFgColor: ColorPair{Dark: tokyoFg, Light: tokyoDayFg},

		TableHeaderColor: ColorPair{Dark: tokyoMagenta, Light: tokyoDayMagenta},

		TypeImageColor:       ColorPair{Dark: tokyoBlue, Light: tokyoDayBlue},
		TypeHelmColor:        ColorPair{Dark: tokyoGreen, Light: tokyoDayGreen},
		TypeSBOMColor:        ColorPair{Dark: tokyoYellow, Light: tokyoDayYellow},
		TypeSignatureColor:   ColorPair{Dark: tokyoMagenta, Light: tokyoDayMagenta},
		TypeAttestationColor: ColorPair{Dark: tokyoTeal, Light: tokyoDayTeal},
		TypeWASMColor:        ColorPair{Dark: tokyoRed, Light: tokyoDayRed},
		TypeUnknownColor:     ColorPair{Dark: tokyoComment, Light: tokyoDayComment},
	}
}
