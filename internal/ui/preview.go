package ui

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/mirageglobe/scout/internal/filesystem"
)

// ansiWrap renders a sentinel through style once and splits on it, returning the
// leading and trailing ANSI escape sequences. Callers can then wrap many strings
// in the same style by string concatenation, avoiding a lipgloss render per item.
// Safe for colour-only styles (no padding/width); returns empty strings when the
// active colour profile emits no escapes.
func ansiWrap(style lipgloss.Style) (pre, suf string) {
	const sentinel = "\x00"
	rendered := style.Render(sentinel)
	i := strings.Index(rendered, sentinel)
	if i < 0 {
		return "", ""
	}
	return rendered[:i], rendered[i+len(sentinel):]
}

// BuildPreview generates the preview text for the currently selected entry.
func (m Model) BuildPreview() string {
	if len(m.Entries) == 0 {
		return "  (empty directory)"
	}
	if m.Cursor >= len(m.Entries) {
		return ""
	}

	selected := m.Entries[m.Cursor]
	fullPath := filepath.Join(m.Cwd, selected.Name)
	t := Themes[m.ThemeIdx]

	if selected.IsDir {
		return m.previewDir(fullPath, selected, t)
	}
	// large files render as plain text immediately; the highlighted version is
	// filled asynchronously into highlightCache and picked up here on the next
	// build (nav-back, theme change, or the fill's own re-render).
	if fileNeedsAsyncHighlight(selected) {
		if v, ok := highlightCacheGet(highlightKey(fullPath, selected, t)); ok {
			return v
		}
		return m.previewFile(fullPath, selected, t, false)
	}
	return m.previewFile(fullPath, selected, t, true)
}

// asyncHighlightThreshold is the file size above which syntax highlighting is done
// off the event loop instead of synchronously (chroma on a large file blocks the UI
// for tens to >100 ms; see the BuildPreview benchmark).
const asyncHighlightThreshold = 32 * 1024

// fileNeedsAsyncHighlight reports whether e is a regular file large enough to warrant
// async highlighting rather than a blocking synchronous highlight.
func fileNeedsAsyncHighlight(e filesystem.Entry) bool {
	return !e.IsDir && e.Info != nil && e.Info.Size() > asyncHighlightThreshold
}

// highlightCache memoizes highlighted previews for large files so a revisit renders
// instantly. Keyed by path/size/mtime/theme; filled by HighlightPreview off the loop.
// Bounded to avoid unbounded growth over a long session.
var (
	highlightMu    sync.Mutex
	highlightCache = map[string]string{}
)

const highlightCacheMax = 128

// highlightKey identifies a highlighted preview by path, size, mtime, and theme, so
// an edited file (new mtime) or a theme switch misses and re-highlights.
func highlightKey(path string, e filesystem.Entry, t Theme) string {
	var size, mod int64
	if e.Info != nil {
		size = e.Info.Size()
		mod = e.Info.ModTime().UnixNano()
	}
	return fmt.Sprintf("%s|%d|%d|%s", path, size, mod, t.ChromaStyle)
}

func highlightCacheGet(key string) (string, bool) {
	highlightMu.Lock()
	defer highlightMu.Unlock()
	v, ok := highlightCache[key]
	return v, ok
}

func highlightCachePut(key, val string) {
	highlightMu.Lock()
	defer highlightMu.Unlock()
	if len(highlightCache) >= highlightCacheMax {
		highlightCache = map[string]string{} // crude bound: reset when full
	}
	highlightCache[key] = val
}

// buildHighlightedFile builds the selected file's preview with highlighting forced
// on; used by the async fill command for large files.
func (m Model) buildHighlightedFile(path string) string {
	if len(m.Entries) == 0 || m.Cursor >= len(m.Entries) {
		return ""
	}
	e := m.Entries[m.Cursor]
	if e.IsDir {
		return ""
	}
	return m.previewFile(path, e, Themes[m.ThemeIdx], true)
}

func (m Model) previewDir(path string, e filesystem.Entry, t Theme) string {
	accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Accent)).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Dim))

	var sb strings.Builder
	sb.WriteString(accentStyle.Render(m.Sym.Dir+" Directory: "+e.Name+"/") + "\n")
	sb.WriteString(dimStyle.Render(strings.Repeat(m.Sym.Rule, 30)) + "\n")

	if e.Info != nil {
		sb.WriteString(accentStyle.Render("Modified: ") + e.Info.ModTime().Format("2006-01-02 15:04") + "\n")
		sb.WriteString(accentStyle.Render("Mode:     ") + e.Info.Mode().String() + "\n")
	}

	children, err := os.ReadDir(path)
	if err != nil {
		sb.WriteString("\n" + dimStyle.Render("(cannot read directory)"))
		return sb.String()
	}

	sb.WriteString(accentStyle.Render("Children: ") + fmt.Sprintf("%d", len(children)) + "\n")
	sb.WriteString(dimStyle.Render(strings.Repeat(m.Sym.Rule, 30)) + "\n")

	shown := 0
	for _, c := range children {
		if shown >= 20 {
			sb.WriteString(fmt.Sprintf("  %s and %d more\n", m.Sym.Ellipsis, len(children)-shown))
			break
		}
		name := c.Name()
		if c.IsDir() {
			name += "/"
		}
		sb.WriteString("  " + name + "\n")
		shown++
	}

	return sb.String()
}

func (m Model) previewFile(path string, e filesystem.Entry, t Theme, highlight bool) string {
	accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Accent)).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Dim))

	var sb strings.Builder
	sb.WriteString(accentStyle.Render(m.Sym.Bullet+" File: "+e.Name) + "\n")
	sb.WriteString(dimStyle.Render(strings.Repeat(m.Sym.Rule, 30)) + "\n")

	if e.Info != nil {
		sb.WriteString(accentStyle.Render("Size:     ") + filesystem.HumanSize(e.Info.Size()) + "\n")
		sb.WriteString(accentStyle.Render("Modified: ") + e.Info.ModTime().Format("2006-01-02 15:04") + "\n")
		sb.WriteString(accentStyle.Render("Mode:     ") + e.Info.Mode().String() + "\n")
	}

	sb.WriteString(dimStyle.Render(strings.Repeat(m.Sym.Rule, 30)) + "\n")

	data, err := os.ReadFile(path)
	if err != nil {
		sb.WriteString("\n(cannot read file)")
		return sb.String()
	}

	previewData := data
	if len(previewData) > 131072 {
		previewData = previewData[:131072]
	}

	if filesystem.IsBinary(previewData) {
		sb.WriteString("\n(binary file – no preview)")
		return sb.String()
	}

	previewStr := string(previewData)

	// cap to maxLines BEFORE highlighting so chroma only tokenises what is shown,
	// not the whole (up to 128 KB) buffer that would then be discarded.
	maxLines := 2500
	rawLines := strings.Split(previewStr, "\n")
	truncatedLines := len(rawLines) > maxLines
	if truncatedLines {
		previewStr = strings.Join(rawLines[:maxLines], "\n")
	}

	// highlight is skipped for large files on the synchronous path; they render as
	// plain text immediately and gain colour from the async cache fill (see BuildPreview).
	if highlight {
		lang := filepath.Ext(path)
		if len(lang) > 0 {
			lang = lang[1:]
		} else {
			lang = filepath.Base(path)
		}
		var b bytes.Buffer
		if err := quick.Highlight(&b, previewStr, lang, "terminal256", t.ChromaStyle); err == nil && b.Len() > 0 {
			previewStr = b.String()
		}
	}

	// render the constant dim-gutter wrapper once; only the line number varies,
	// avoiding a lipgloss render call per line (up to maxLines of them).
	gPre, gSuf := ansiWrap(dimStyle)
	for i, l := range strings.Split(previewStr, "\n") {
		l = strings.ReplaceAll(l, "\t", "    ")
		sb.WriteString(gPre + fmt.Sprintf("%3d │", i+1) + gSuf + " " + l + "\n")
	}

	if len(data) > 131072 || truncatedLines {
		sb.WriteString("\n  " + m.Sym.Ellipsis + " (truncated)")
	}

	return sb.String()
}

// renderGitPreview formats async git diff/log output for the preview pane.
// diff output is highlighted with the chroma "diff" lexer; log is plain.
func (m Model) renderGitPreview(mode int, content string) string {
	t := Themes[m.ThemeIdx]
	accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Accent)).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Dim))

	header := m.Sym.Branch + " git diff"
	if mode == GitLog {
		header = m.Sym.Branch + " git log"
	}

	var sb strings.Builder
	sb.WriteString(accentStyle.Render(header) + "\n")
	sb.WriteString(dimStyle.Render(strings.Repeat(m.Sym.Rule, 30)) + "\n")

	body := content
	if mode == GitDiff {
		var b bytes.Buffer
		if err := quick.Highlight(&b, content, "diff", "terminal256", t.ChromaStyle); err == nil && b.Len() > 0 {
			body = b.String()
		}
	}

	lines := strings.Split(strings.TrimSuffix(body, "\n"), "\n")
	maxLines := 2500
	truncated := false
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		truncated = true
	}
	for _, l := range lines {
		sb.WriteString(strings.ReplaceAll(l, "\t", "    ") + "\n")
	}
	if truncated {
		sb.WriteString("\n  " + m.Sym.Ellipsis + " (truncated)")
	}
	return sb.String()
}
