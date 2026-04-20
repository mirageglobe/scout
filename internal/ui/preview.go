package ui

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/mirageglobe/scout/internal/filesystem"
)

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
	return m.previewFile(fullPath, selected, t)
}

func (m Model) previewDir(path string, e filesystem.Entry, t Theme) string {
	accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Accent)).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Dim))

	var sb strings.Builder
	sb.WriteString(accentStyle.Render("▸ Directory: "+e.Name+"/") + "\n")
	sb.WriteString(dimStyle.Render(strings.Repeat("─", 30)) + "\n")

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
	sb.WriteString(dimStyle.Render(strings.Repeat("─", 30)) + "\n")

	shown := 0
	for _, c := range children {
		if shown >= 20 {
			sb.WriteString(fmt.Sprintf("  … and %d more\n", len(children)-shown))
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

func (m Model) previewFile(path string, e filesystem.Entry, t Theme) string {
	accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Accent)).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Dim))

	var sb strings.Builder
	sb.WriteString(accentStyle.Render("• File: "+e.Name) + "\n")
	sb.WriteString(dimStyle.Render(strings.Repeat("─", 30)) + "\n")

	if e.Info != nil {
		sb.WriteString(accentStyle.Render("Size:     ") + filesystem.HumanSize(e.Info.Size()) + "\n")
		sb.WriteString(accentStyle.Render("Modified: ") + e.Info.ModTime().Format("2006-01-02 15:04") + "\n")
		sb.WriteString(accentStyle.Render("Mode:     ") + e.Info.Mode().String() + "\n")
	}

	sb.WriteString(dimStyle.Render(strings.Repeat("─", 30)) + "\n")

	data, err := os.ReadFile(path)
	if err != nil {
		sb.WriteString("\n(cannot read file)")
		return sb.String()
	}

	previewData := data
	if len(previewData) > 32768 {
		previewData = previewData[:32768]
	}

	if filesystem.IsBinary(previewData) {
		sb.WriteString("\n(binary file – no preview)")
		return sb.String()
	}

	previewStr := string(previewData)
	var b bytes.Buffer
	lang := filepath.Ext(path)
	if len(lang) > 0 {
		lang = lang[1:]
	} else {
		lang = filepath.Base(path)
	}

	if err := quick.Highlight(&b, previewStr, lang, "terminal256", "dracula"); err == nil && b.Len() > 0 {
		previewStr = b.String()
	}

	lines := strings.Split(previewStr, "\n")
	maxLines := 1000
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	for i, l := range lines {
		l = strings.ReplaceAll(l, "\t", "    ")
		gutter := dimStyle.Render(fmt.Sprintf("%3d │", i+1))
		sb.WriteString(fmt.Sprintf("%s %s\n", gutter, l))
	}

	if len(data) > 32768 || len(lines) >= maxLines {
		sb.WriteString("\n  … (truncated)")
	}

	return sb.String()
}
