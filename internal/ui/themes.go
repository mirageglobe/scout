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
		Name:       "Classic Amber",
		Accent:     "#FFAF00",
		Dim:        "#875F00",
		Text:       "#D7D7D7",
		SelectedBg: "#FFAF00",
		SelectedFg: "#000000",
	},
	{
		Name:       "Safety Orange",
		Accent:     "#FF8700",
		Dim:        "#AF5F00",
		Text:       "#D7D7D7",
		SelectedBg: "#FF8700",
		SelectedFg: "#000000",
	},
	{
		Name:       "Mono",
		Accent:     "#FFFFFF",
		Dim:        "#555555",
		Text:       "#BBBBBB",
		SelectedBg: "#FFFFFF",
		SelectedFg: "#000000",
	},
	{
		Name:       "Electric Cyan",
		Accent:     "#00AFFF",
		Dim:        "#005F87",
		Text:       "#D7D7D7",
		SelectedBg: "#00AFFF",
		SelectedFg: "#000000",
	},
}
