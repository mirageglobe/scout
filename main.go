package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ---------------------------------------------------------------------------
// Messages
// ---------------------------------------------------------------------------

// editorFinishedMsg is sent when the external editor (vim) exits.
type editorFinishedMsg struct{ err error }

// dirLoadedMsg carries the result of loading a directory.
type dirLoadedMsg struct {
	entries   []entry
	gitStatus map[string]string
	err       error
}

// ---------------------------------------------------------------------------
// Entry represents a single file or directory in the listing.
// ---------------------------------------------------------------------------

type entry struct {
	name  string
	isDir bool
	info  os.FileInfo
}

// ---------------------------------------------------------------------------
// Model
// ---------------------------------------------------------------------------

type model struct {
	// Current working directory (absolute path).
	cwd string

	// Directory entries displayed in the left pane.
	entries []entry

	// Cursor position in the entry list.
	cursor int

	// Terminal dimensions.
	width  int
	height int

	// Preview content shown in the right pane.
	preview string

	// Git status map: filename -> status indicator like [M] or [?].
	gitStatus map[string]string

	// Error to display, if any.
	err error
}

// ---------------------------------------------------------------------------
// Init / Update / View
// ---------------------------------------------------------------------------

func (m model) Init() tea.Cmd {
	return loadDir(m.cwd)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.preview = buildPreview(m)
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		// Navigation: move cursor down
		case "j", "down":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}
			m.preview = buildPreview(m)
			return m, nil

		// Navigation: move cursor up
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
			m.preview = buildPreview(m)
			return m, nil

		// Navigation: go to parent directory
		case "h", "left":
			parent := filepath.Dir(m.cwd)
			if parent != m.cwd {
				m.cwd = parent
				m.cursor = 0
				m.preview = ""
				return m, loadDir(m.cwd)
			}
			return m, nil

		// Enter: open directory or file
		case "enter":
			if len(m.entries) == 0 {
				return m, nil
			}
			selected := m.entries[m.cursor]
			fullPath := filepath.Join(m.cwd, selected.name)

			if selected.isDir {
				m.cwd = fullPath
				m.cursor = 0
				m.preview = ""
				return m, loadDir(m.cwd)
			}

			// Open file in vim via ExecProcess
			c := exec.Command("vim", fullPath)
			return m, tea.ExecProcess(c, func(err error) tea.Msg {
				return editorFinishedMsg{err: err}
			})

		// Jump to top
		case "g":
			m.cursor = 0
			m.preview = buildPreview(m)
			return m, nil

		// Jump to bottom
		case "G", "shift+g":
			if len(m.entries) > 0 {
				m.cursor = len(m.entries) - 1
			}
			m.preview = buildPreview(m)
			return m, nil
		}

	case dirLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.entries = msg.entries
		m.gitStatus = msg.gitStatus
		m.err = nil
		if m.cursor >= len(m.entries) {
			m.cursor = max(0, len(m.entries)-1)
		}
		m.preview = buildPreview(m)
		return m, nil

	case editorFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		// Reload the directory after returning from vim
		return m, loadDir(m.cwd)
	}

	return m, nil
}

func (m model) View() tea.View {
	if m.width == 0 {
		return tea.NewView("Loading…")
	}

	// ── Colours & metrics ──────────────────────────────────────────────
	borderColor := lipgloss.Color("#7D56F4")
	accentColor := lipgloss.Color("#FF79C6")
	dimColor := lipgloss.Color("#6272A4")
	textColor := lipgloss.Color("#F8F8F2")
	selectedBg := lipgloss.Color("#44475A")
	dirColor := lipgloss.Color("#8BE9FD")
	gitModColor := lipgloss.Color("#FFB86C")
	gitNewColor := lipgloss.Color("#50FA7B")

	// Reserve space for borders (2 chars each side) and a 1-char gap
	usableWidth := m.width - 5
	if usableWidth < 20 {
		usableWidth = 20
	}
	leftWidth := usableWidth * 2 / 5
	rightWidth := usableWidth - leftWidth

	// Usable height inside borders (top + bottom border = 2 lines, title = 1)
	contentHeight := m.height - 4
	if contentHeight < 3 {
		contentHeight = 3
	}

	// ── Left pane: file list ───────────────────────────────────────────
	var listLines []string

	// Show current path as a subtle header
	cwdDisplay := m.cwd
	home, _ := os.UserHomeDir()
	if strings.HasPrefix(cwdDisplay, home) {
		cwdDisplay = "~" + cwdDisplay[len(home):]
	}

	headerStyle := lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true)

	listLines = append(listLines, headerStyle.Render(truncate(cwdDisplay, leftWidth-2)))

	if m.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
		listLines = append(listLines, errStyle.Render("Error: "+m.err.Error()))
	}

	// Visible window of entries
	visibleRows := contentHeight - len(listLines)
	if visibleRows < 1 {
		visibleRows = 1
	}

	// Scroll offset so the cursor stays visible
	scrollOffset := 0
	if m.cursor >= visibleRows {
		scrollOffset = m.cursor - visibleRows + 1
	}

	normalItem := lipgloss.NewStyle().Foreground(textColor)
	selectedItem := lipgloss.NewStyle().
		Foreground(textColor).
		Background(selectedBg).
		Bold(true)
	dirStyle := lipgloss.NewStyle().Foreground(dirColor).Bold(true)
	gitModStyle := lipgloss.NewStyle().Foreground(gitModColor)
	gitNewStyle := lipgloss.NewStyle().Foreground(gitNewColor)

	for i := scrollOffset; i < len(m.entries) && len(listLines) < contentHeight; i++ {
		e := m.entries[i]
		name := e.name

		// Git status badge
		badge := "   "
		if status, ok := m.gitStatus[name]; ok {
			switch status {
			case "M", "A", "R", "C", "U", "MM", "AM":
				badge = gitModStyle.Render("[M]")
			case "?":
				badge = gitNewStyle.Render("[?]")
			default:
				badge = gitModStyle.Render("[" + status[:1] + "]")
			}
		}

		// Directory indicator
		if e.isDir {
			name = name + "/"
		}

		line := badge + " " + name
		line = truncate(line, leftWidth-2)

		// Pad to full width for consistent highlighting
		padded := line + strings.Repeat(" ", max(0, leftWidth-2-visibleLen(line)))

		if i == m.cursor {
			if e.isDir {
				listLines = append(listLines, selectedItem.Render(padded))
			} else {
				listLines = append(listLines, selectedItem.Render(padded))
			}
		} else {
			if e.isDir {
				listLines = append(listLines, dirStyle.Render(padded))
			} else {
				listLines = append(listLines, normalItem.Render(padded))
			}
		}
	}

	// Pad remaining lines so the box keeps its shape
	for len(listLines) < contentHeight {
		listLines = append(listLines, strings.Repeat(" ", leftWidth-2))
	}

	leftContent := strings.Join(listLines, "\n")

	// ── Right pane: preview ────────────────────────────────────────────
	previewLines := strings.Split(m.preview, "\n")
	if len(previewLines) > contentHeight {
		previewLines = previewLines[:contentHeight]
	}
	for i, l := range previewLines {
		previewLines[i] = truncate(l, rightWidth-2)
	}
	for len(previewLines) < contentHeight {
		previewLines = append(previewLines, "")
	}
	rightContent := strings.Join(previewLines, "\n")

	// ── Pane styles ────────────────────────────────────────────────────
	leftPane := lipgloss.NewStyle().
		Width(leftWidth).
		Height(contentHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Render(leftContent)

	rightPane := lipgloss.NewStyle().
		Width(rightWidth).
		Height(contentHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Render(rightContent)

	// ── Compose layout ─────────────────────────────────────────────────
	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	// ── Status bar ─────────────────────────────────────────────────────
	statusStyle := lipgloss.NewStyle().
		Foreground(dimColor)

	count := fmt.Sprintf(" %d items", len(m.entries))
	pos := ""
	if len(m.entries) > 0 {
		pos = fmt.Sprintf(" %d/%d", m.cursor+1, len(m.entries))
	}
	help := " q:quit  j/k:navigate  h:up  enter:open  g/G:top/bottom"

	statusBar := statusStyle.Render(
		truncate(count+pos+"  │"+help, m.width),
	)

	layout := lipgloss.JoinVertical(lipgloss.Left, panes, statusBar)
	v := tea.NewView(layout)
	v.AltScreen = true
	return v
}

// ---------------------------------------------------------------------------
// Commands
// ---------------------------------------------------------------------------

// loadDir reads the directory at path and returns a dirLoadedMsg.
func loadDir(path string) tea.Cmd {
	return func() tea.Msg {
		dirEntries, err := os.ReadDir(path)
		if err != nil {
			return dirLoadedMsg{err: err}
		}

		var entries []entry
		for _, de := range dirEntries {
			info, _ := de.Info()
			entries = append(entries, entry{
				name:  de.Name(),
				isDir: de.IsDir(),
				info:  info,
			})
		}

		// Sort: directories first, then alphabetical
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].isDir != entries[j].isDir {
				return entries[i].isDir
			}
			return strings.ToLower(entries[i].name) < strings.ToLower(entries[j].name)
		})

		gitStatus := getGitStatus(path)

		return dirLoadedMsg{
			entries:   entries,
			gitStatus: gitStatus,
		}
	}
}

// ---------------------------------------------------------------------------
// Git integration
// ---------------------------------------------------------------------------

// getGitStatus runs "git status --porcelain" in the given directory and
// returns a map of filename -> status code.
func getGitStatus(dir string) map[string]string {
	result := make(map[string]string)

	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		// Not a git repo or git not available – that's fine
		return result
	}

	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if len(line) < 4 {
			continue
		}
		// Porcelain format: XY filename
		xy := strings.TrimSpace(line[:2])
		name := strings.TrimSpace(line[3:])

		// Handle renamed files: "R  old -> new"
		if idx := strings.Index(name, " -> "); idx >= 0 {
			name = name[idx+4:]
		}

		// Strip leading path components to match entries in this directory
		// For nested changes, we tag the top-level directory
		if parts := strings.SplitN(name, "/", 2); len(parts) > 1 {
			name = parts[0]
		}

		if xy == "??" {
			result[name] = "?"
		} else {
			// Use the first non-space character as status
			status := string(xy[0])
			if status == " " && len(xy) > 1 {
				status = string(xy[1])
			}
			result[name] = status
		}
	}

	return result
}

// ---------------------------------------------------------------------------
// Preview builder
// ---------------------------------------------------------------------------

// buildPreview generates the preview text for the currently selected entry.
func buildPreview(m model) string {
	if len(m.entries) == 0 {
		return "  (empty directory)"
	}
	if m.cursor >= len(m.entries) {
		return ""
	}

	selected := m.entries[m.cursor]
	fullPath := filepath.Join(m.cwd, selected.name)

	if selected.isDir {
		return previewDir(fullPath, selected)
	}
	return previewFile(fullPath, selected)
}

// previewDir shows directory metadata and a listing of its contents.
func previewDir(path string, e entry) string {
	var sb strings.Builder
	sb.WriteString("📁 Directory: " + e.name + "/\n")
	sb.WriteString(strings.Repeat("─", 30) + "\n")

	if e.info != nil {
		sb.WriteString(fmt.Sprintf("Modified: %s\n", e.info.ModTime().Format("2006-01-02 15:04")))
		sb.WriteString(fmt.Sprintf("Mode:     %s\n", e.info.Mode()))
	}

	children, err := os.ReadDir(path)
	if err != nil {
		sb.WriteString("\n(cannot read directory)")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("Children: %d\n", len(children)))
	sb.WriteString(strings.Repeat("─", 30) + "\n")

	// Show up to 20 children
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

// previewFile shows file metadata and the first lines of text content.
func previewFile(path string, e entry) string {
	var sb strings.Builder
	sb.WriteString("📄 File: " + e.name + "\n")
	sb.WriteString(strings.Repeat("─", 30) + "\n")

	if e.info != nil {
		sb.WriteString(fmt.Sprintf("Size:     %s\n", humanSize(e.info.Size())))
		sb.WriteString(fmt.Sprintf("Modified: %s\n", e.info.ModTime().Format("2006-01-02 15:04")))
		sb.WriteString(fmt.Sprintf("Mode:     %s\n", e.info.Mode()))
	}

	sb.WriteString(strings.Repeat("─", 30) + "\n")

	// Try to read as text (first 4KB)
	data, err := os.ReadFile(path)
	if err != nil {
		sb.WriteString("\n(cannot read file)")
		return sb.String()
	}

	// Limit to first 4KB for preview
	preview := data
	if len(preview) > 4096 {
		preview = preview[:4096]
	}

	// Check if the content looks like binary
	if isBinary(preview) {
		sb.WriteString("\n(binary file – no preview)")
		return sb.String()
	}

	lines := strings.Split(string(preview), "\n")
	maxLines := 40
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	for i, l := range lines {
		// Replace tabs with spaces for cleaner rendering
		l = strings.ReplaceAll(l, "\t", "    ")
		sb.WriteString(fmt.Sprintf("%3d │ %s\n", i+1, l))
	}

	if len(data) > 4096 {
		sb.WriteString("\n  … (truncated)")
	}

	return sb.String()
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// isBinary attempts to detect binary content by checking for null bytes.
func isBinary(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}

// humanSize formats a byte count into a human-readable string.
func humanSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// truncate cuts a string to maxLen characters, appending "…" if truncated.
func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return "…"
	}
	return string(runes[:maxLen-1]) + "…"
}

// visibleLen returns the approximate visible length of a string,
// stripping ANSI escape sequences.
func visibleLen(s string) int {
	n := 0
	inEsc := false
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		n++
	}
	return n
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "scout: %v\n", err)
		os.Exit(1)
	}

	m := model{
		cwd: cwd,
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "scout: %v\n", err)
		os.Exit(1)
	}
}
