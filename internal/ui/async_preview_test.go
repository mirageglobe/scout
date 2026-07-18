package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mirageglobe/scout/internal/filesystem"
)

func writeSizedFile(t *testing.T, dir, name, content string) filesystem.Entry {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	return filesystem.Entry{Name: name, Info: info}
}

func TestFileNeedsAsyncHighlight(t *testing.T) {
	dir := t.TempDir()
	small := writeSizedFile(t, dir, "small.go", "package a\n")
	big := writeSizedFile(t, dir, "big.go", strings.Repeat("x", asyncHighlightThreshold+1))
	if fileNeedsAsyncHighlight(small) {
		t.Error("small file should not need async highlight")
	}
	if !fileNeedsAsyncHighlight(big) {
		t.Error("big file should need async highlight")
	}
	if fileNeedsAsyncHighlight(filesystem.Entry{IsDir: true, Info: big.Info}) {
		t.Error("directory should not need async highlight")
	}
}

// A large file renders plain on a cache miss, and the highlighted version once the
// async fill (HighlightPreview) has populated the cache.
func TestBuildPreviewLargeFilePlainThenCached(t *testing.T) {
	dir := t.TempDir()
	e := writeSizedFile(t, dir, "big.go", strings.Repeat("func f() int { return 42 }\n", 2000))
	path := filepath.Join(dir, e.Name)
	if !fileNeedsAsyncHighlight(e) {
		t.Fatalf("expected big file (size %d) to be async", e.Info.Size())
	}
	m := Model{Cwd: dir, ThemeIdx: 0, Sym: selectGlyphs(), Entries: []filesystem.Entry{e}, Cursor: 0}

	key := highlightKey(path, e, Themes[0])
	highlightMu.Lock()
	delete(highlightCache, key)
	highlightMu.Unlock()

	plain := m.BuildPreview()

	// fill the cache via the async command
	msg, ok := m.HighlightPreview(path, key)().(HighlightFilledMsg)
	if !ok || msg.Path != path || msg.Key != key {
		t.Fatalf("HighlightPreview returned unexpected msg: %+v", msg)
	}
	cached := m.BuildPreview()
	if cached != msg.Content {
		t.Errorf("BuildPreview after fill did not return the cached highlighted content")
	}
	if cached == plain {
		t.Errorf("highlighted preview should differ from the plain one")
	}
}

// HighlightFilledMsg swaps in only when the cursor is still on the same file.
func TestHighlightFilledMsgStaleness(t *testing.T) {
	dir := t.TempDir()
	e := writeSizedFile(t, dir, "a.go", "package a\n")
	path := filepath.Join(dir, e.Name)
	base := Model{Cwd: dir, PreviewMode: PreviewFile, ThemeIdx: 0, Sym: selectGlyphs(),
		Entries: []filesystem.Entry{e}, Cursor: 0, Preview: "orig"}

	u, _ := base.Update(HighlightFilledMsg{Path: path, Key: "k", Content: "HL"})
	if got := u.(Model).Preview; got != "HL" {
		t.Errorf("matching highlight should swap in; Preview = %q, want HL", got)
	}

	u2, _ := base.Update(HighlightFilledMsg{Path: filepath.Join(dir, "other.go"), Key: "k", Content: "HL"})
	if got := u2.(Model).Preview; got != "orig" {
		t.Errorf("stale highlight should be dropped; Preview = %q, want orig", got)
	}
}
