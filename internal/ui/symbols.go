package ui

import "os"

// Glyphs is the set of UI marker characters scout renders in width-sensitive
// positions. Two presets exist: a Unicode set (default) and a narrow-safe
// ASCII set for terminals that render ambiguous-width glyphs as 2 cells
// (e.g. RUNEWIDTH_EASTASIAN=1), which otherwise misaligns the fixed columns.
type Glyphs struct {
	Dir      string // directory marker
	Symlink  string // symlink marker
	SubCount string // directory child-count marker
	Bullet   string // file-preview header bullet
	Branch   string // git branch symbol
	Prompt   string // "scout >" prompt caret
	Changed  string // changed-file count marker
	Block    string // search input cursor
	Dot      string // clean-file marker, separators, spinner
	Ellipsis string // truncation indicator
	Rule     string // horizontal divider
	Up       string
	Down     string
	Left     string
	Right    string
}

var unicodeGlyphs = Glyphs{
	Dir: "▸", Symlink: "↳", SubCount: "≡", Bullet: "•", Branch: "⎇",
	Prompt: "›", Changed: "±", Block: "█", Dot: "·", Ellipsis: "…", Rule: "─",
	Up: "↑", Down: "↓", Left: "←", Right: "→",
}

var asciiGlyphs = Glyphs{
	Dir: ">", Symlink: "@", SubCount: "=", Bullet: "*", Branch: "#",
	Prompt: ">", Changed: "~", Block: "_", Dot: ".", Ellipsis: "...", Rule: "-",
	Up: "^", Down: "v", Left: "<", Right: ">",
}

// selectGlyphs returns the ASCII set when SCOUT_UNICODE_SAFE is set (to any
// non-empty value), otherwise the Unicode set.
func selectGlyphs() Glyphs {
	if os.Getenv("SCOUT_UNICODE_SAFE") != "" {
		return asciiGlyphs
	}
	return unicodeGlyphs
}
