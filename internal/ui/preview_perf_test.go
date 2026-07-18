package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/mirageglobe/scout/internal/filesystem"
)

// ansiWrap should reconstruct exactly what a per-item style.Render would produce.
func TestAnsiWrapReconstructs(t *testing.T) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	pre, suf := ansiWrap(style)
	if got, want := pre+"42 │"+suf, style.Render("42 │"); got != want {
		t.Errorf("ansiWrap mismatch:\n got %q\nwant %q", got, want)
	}
}

// withPreviewDisplay caches expanded rows; the line count reads the cache on the
// fast path and recomputes directly when the preview changed without a rebuild.
func TestPreviewDisplayCache(t *testing.T) {
	m := Model{Width: 100, ThemeIdx: 0, Sym: selectGlyphs()}
	m.Preview = "alpha\nbeta\ngamma"
	m = m.withPreviewDisplay()

	if got := len(m.previewDisplay); got != 3 {
		t.Fatalf("previewDisplay len = %d, want 3", got)
	}
	if got := m.previewDisplay[2].origIdx; got != 2 {
		t.Errorf("row 2 origIdx = %d, want 2", got)
	}
	if got := previewDisplayLineCount(m); got != 3 {
		t.Errorf("previewDisplayLineCount (cached) = %d, want 3", got)
	}

	// stale: preview changed but cache not rebuilt -> count computes directly
	m.Preview = "one\ntwo"
	if got := previewDisplayLineCount(m); got != 2 {
		t.Errorf("previewDisplayLineCount (stale) = %d, want 2", got)
	}
}

// previewFile must cap the preview and mark it truncated for very large files,
// regardless of the (now earlier) truncation happening before highlighting.
func TestPreviewFileTruncatesLargeFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "big.txt")
	var sb strings.Builder
	for i := 0; i < 3000; i++ {
		sb.WriteString("line\n")
	}
	if err := os.WriteFile(f, []byte(sb.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(f)
	m := Model{
		Cwd:      dir,
		ThemeIdx: 0,
		Sym:      selectGlyphs(),
		Entries:  []filesystem.Entry{{Name: "big.txt", Info: info}},
		Cursor:   0,
	}
	out := m.BuildPreview()
	if !strings.Contains(out, "(truncated)") {
		t.Errorf("expected a (truncated) marker for a 3000-line file")
	}
}
