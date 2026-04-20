package ui

// Theme represents a color palette for the application.
type Theme struct {
	Name       string
	Accent     string
	Dim        string
	Text       string
	SelectedBg string
	SelectedFg string
}

// Themes is the global list of available color palettes.
var Themes = []Theme{
	{
		// index 0 — morning (09:00–12:00)
		Name:       "Classic Amber",
		Accent:     "#FFAF00",
		Dim:        "#875F00",
		Text:       "#D7D7D7",
		SelectedBg: "#FFAF00",
		SelectedFg: "#000000",
	},
	{
		// index 1 — dusk (17:00–20:00)
		Name:       "Safety Orange",
		Accent:     "#FF8700",
		Dim:        "#AF5F00",
		Text:       "#D7D7D7",
		SelectedBg: "#FF8700",
		SelectedFg: "#000000",
	},
	{
		// index 2 — manual only
		Name:       "Mono",
		Accent:     "#FFFFFF",
		Dim:        "#555555",
		Text:       "#BBBBBB",
		SelectedBg: "#FFFFFF",
		SelectedFg: "#000000",
	},
	{
		// index 3 — afternoon (12:00–17:00)
		Name:       "Electric Cyan",
		Accent:     "#00AFFF",
		Dim:        "#005F87",
		Text:       "#D7D7D7",
		SelectedBg: "#00AFFF",
		SelectedFg: "#000000",
	},
	{
		// index 4 — dawn (05:00–09:00)
		Name:       "Dawn",
		Accent:     "#FF8787",
		Dim:        "#875F5F",
		Text:       "#D7D7D7",
		SelectedBg: "#FF8787",
		SelectedFg: "#000000",
	},
	{
		// index 5 — night (00:00–05:00)
		Name:       "Midnight",
		Accent:     "#875FFF",
		Dim:        "#5F5F87",
		Text:       "#AFAFD7",
		SelectedBg: "#875FFF",
		SelectedFg: "#000000",
	},
	{
		// index 6 — evening (20:00–24:00)
		Name:       "Evening",
		Accent:     "#FF5FAF",
		Dim:        "#87005F",
		Text:       "#D7D7D7",
		SelectedBg: "#FF5FAF",
		SelectedFg: "#000000",
	},
	{
		// index 7 — Solarized Dark (Ethan Schoonover palette)
		Name:       "Solarized Dark",
		Accent:     "#268BD2",
		Dim:        "#586E75",
		Text:       "#839496",
		SelectedBg: "#268BD2",
		SelectedFg: "#002B36",
	},
	{
		// index 8 — Solarized Light (Ethan Schoonover palette)
		Name:       "Solarized Light",
		Accent:     "#268BD2",
		Dim:        "#93A1A1",
		Text:       "#657B83",
		SelectedBg: "#268BD2",
		SelectedFg: "#FDF6E3",
	},
}

// ThemeForHour returns the theme index suited to the given hour (0–23).
func ThemeForHour(hour int) int {
	switch {
	case hour >= 0 && hour < 5:
		return 5 // Midnight
	case hour >= 5 && hour < 9:
		return 4 // Dawn
	case hour >= 9 && hour < 12:
		return 0 // Classic Amber
	case hour >= 12 && hour < 17:
		return 3 // Electric Cyan
	case hour >= 17 && hour < 20:
		return 1 // Safety Orange
	default:
		return 6 // Evening
	}
}
