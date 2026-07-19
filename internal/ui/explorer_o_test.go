package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mirageglobe/scout/internal/filesystem"
)

// o on a directory copies its absolute path to the clipboard and reports it in the
// status footer (rather than the file open-with-default action).
func TestOpenKeyOnDirectoryCopiesPath(t *testing.T) {
	orig := copyToClipboard
	defer func() { copyToClipboard = orig }()
	var copied string
	copyToClipboard = func(text string) error { copied = text; return nil }

	m := Model{
		Cwd:      "/home/u",
		Entries:  []filesystem.Entry{{Name: "subdir", IsDir: true}},
		Cursor:   0,
		ThemeIdx: 0,
		Sym:      selectGlyphs(),
	}
	u, _ := m.Update(tea.KeyPressMsg{Code: 'o', Text: "o"})
	m = u.(Model)

	if copied != "/home/u/subdir" {
		t.Errorf("copied path = %q, want /home/u/subdir", copied)
	}
	if !strings.Contains(m.StatusMsg, "copied path") {
		t.Errorf("StatusMsg = %q, want it to mention 'copied path'", m.StatusMsg)
	}
}
