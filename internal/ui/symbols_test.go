package ui

import (
	"strings"
	"testing"
)

func TestSelectGlyphs(t *testing.T) {
	t.Setenv("SCOUT_UNICODE_SAFE", "1")
	if selectGlyphs() != asciiGlyphs {
		t.Error("SCOUT_UNICODE_SAFE=1 should select the ascii glyph set")
	}
	t.Setenv("SCOUT_UNICODE_SAFE", "")
	if selectGlyphs() != unicodeGlyphs {
		t.Error("empty SCOUT_UNICODE_SAFE should select the unicode glyph set")
	}
}

func TestAsciiGlyphsAreNarrow(t *testing.T) {
	g := asciiGlyphs
	for _, f := range []string{g.Dir, g.Symlink, g.SubCount, g.Bullet, g.Branch, g.Prompt, g.Changed, g.Block, g.Dot, g.Ellipsis, g.Rule, g.Up, g.Down, g.Left, g.Right} {
		for _, r := range f {
			if r > 127 {
				t.Errorf("ascii glyph %q contains a non-ascii rune %q", f, r)
			}
		}
	}
}

func TestUnicodeSafeStatusLineNoWideMarkers(t *testing.T) {
	m := Model{Sym: asciiGlyphs, ThemeIdx: 0, Width: 80}
	out := m.RenderStatusLine() // idle -> " scout > "
	for _, w := range []string{"›", "·", "█", "…", "⎇"} {
		if strings.Contains(out, w) {
			t.Errorf("ascii status line leaked wide glyph %q: %q", w, out)
		}
	}
}
