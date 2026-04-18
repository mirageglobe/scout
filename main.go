package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"runtime"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/x/ansi"
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

// statsMsg carries system statistics.
type statsMsg struct {
	cpu     float64
	mem     uint64
	dirSize int64
}

// tickMsg is sent periodically to refresh stats.
type tickMsg time.Time

// stats represents the current resource usage.
type stats struct {
	cpu     float64
	mem     uint64
	dirSize int64
}

// ---------------------------------------------------------------------------
// Entry represents a single file or directory in the listing.
// ---------------------------------------------------------------------------

type entry struct {
	name  string
	isDir bool
	info  os.FileInfo
}

type theme struct {
	name       string
	accent     string
	dim        string
	text       string
	selectedBg string
	selectedFg string
}

var themes = []theme{
	{
		name:       "Safety Orange",
		accent:     "#FF8700",
		dim:        "#AF5F00",
		text:       "#D7D7D7",
		selectedBg: "#FF8700",
		selectedFg: "#000000",
	},
	{
		name:       "Mono",
		accent:     "#FFFFFF",
		dim:        "#555555",
		text:       "#BBBBBB",
		selectedBg: "#FFFFFF",
		selectedFg: "#000000",
	},
	{
		name:       "Classic Amber",
		accent:     "#FFAF00",
		dim:        "#875F00",
		text:       "#D7D7D7",
		selectedBg: "#FFAF00",
		selectedFg: "#000000",
	},
	{
		name:       "Electric Cyan",
		accent:     "#00AFFF",
		dim:        "#005F87",
		text:       "#D7D7D7",
		selectedBg: "#00AFFF",
		selectedFg: "#000000",
	},
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
	preview       string
	previewScroll int

	// Focus state: true if right pane (preview) is focused.
	focusRight bool

	// Help state: true if help screen is displayed.
	showHelp bool

	// Theme state: index into the global themes slice.
	themeIdx int

	// Git status map: filename -> status indicator like [M] or [?].
	gitStatus map[string]string

	// Stats: resource usage and directory metadata.
	stats stats

	// Error to display, if any.
	err error
}

// ---------------------------------------------------------------------------
// Init / Update / View
// ---------------------------------------------------------------------------

func (m model) Init() tea.Cmd {
	return tea.Batch(
		loadDir(m.cwd),
		doTick(),
		getStats(m.cwd),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tickMsg:
		return m, tea.Batch(doTick(), getStats(m.cwd))

	case statsMsg:
		m.stats.cpu = msg.cpu
		m.stats.mem = msg.mem
		m.stats.dirSize = msg.dirSize
		return m, nil

	case dirLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.entries = msg.entries
		m.gitStatus = msg.gitStatus
		m.err = nil
		m.previewScroll = 0
		if m.cursor >= len(m.entries) {
			m.cursor = max(0, len(m.entries)-1)
		}
		m.preview = buildPreview(m)
		return m, getStats(m.cwd)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.preview = buildPreview(m)
		return m, nil

	case tea.KeyPressMsg:
		if m.showHelp {
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			m.showHelp = false
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "?":
			m.showHelp = true
			return m, nil

		case "t":
			m.themeIdx = (m.themeIdx + 1) % len(themes)
			m.preview = buildPreview(m)
			return m, nil

		// Navigation: move cursor down
		case "j", "down":
			if m.focusRight {
				previewLines := strings.Split(strings.TrimSuffix(m.preview, "\n"), "\n")
				contentHeight := m.height - 4
				maxScroll := len(previewLines) - contentHeight
				if maxScroll < 0 {
					maxScroll = 0
				}
				if m.previewScroll < maxScroll {
					m.previewScroll++
				}
			} else {
				if m.cursor < len(m.entries)-1 {
					m.cursor++
				}
				m.previewScroll = 0
				m.preview = buildPreview(m)
			}
			return m, nil

		// Navigation: move cursor up
		case "k", "up":
			if m.focusRight {
				if m.previewScroll > 0 {
					m.previewScroll--
				}
			} else {
				if m.cursor > 0 {
					m.cursor--
				}
				m.previewScroll = 0
				m.preview = buildPreview(m)
			}
			return m, nil

		// Navigation: go to parent directory or unfocus right pane
		case "h", "left":
			if m.focusRight {
				m.focusRight = false
				return m, nil
			}
			parent := filepath.Dir(m.cwd)
			if parent != m.cwd {
				m.cwd = parent
				m.cursor = 0
				m.preview = ""
				return m, loadDir(m.cwd)
			}
			return m, nil

		// V / Enter / l / right: open directory, focus file preview, or open in vim
		case "v", "enter", "l", "right":
			if len(m.entries) == 0 {
				return m, nil
			}
			selected := m.entries[m.cursor]
			fullPath := filepath.Join(m.cwd, selected.name)

			if selected.isDir {
				m.cwd = fullPath
				m.cursor = 0
				m.preview = ""
				m.focusRight = false
				return m, loadDir(m.cwd)
			}

			// It's a file
			isAction := msg.String() == "enter" || msg.String() == "v"
			if !isAction {
				if !m.focusRight {
					m.focusRight = true
				}
				return m, nil
			}

			// Open file in vim via ExecProcess
			// Safety check: read first 1KB to check if it's binary
			f, _ := os.Open(fullPath)
			if f != nil {
				buf := make([]byte, 1024)
				n, _ := f.Read(buf)
				f.Close()
				if isBinary(buf[:n]) {
					m.err = fmt.Errorf("cannot open binary file: %s", selected.name)
					return m, nil
				}
			}

			c := exec.Command("vim", fullPath)
			return m, tea.ExecProcess(c, func(err error) tea.Msg {
				return editorFinishedMsg{err: err}
			})

		// Jump to top
		case "g":
			m.cursor = 0
			m.previewScroll = 0
			m.preview = buildPreview(m)
			return m, nil

		// Jump to bottom
		case "G", "shift+g":
			if len(m.entries) > 0 {
				m.cursor = len(m.entries) - 1
			}
			m.previewScroll = 0
			m.preview = buildPreview(m)
			return m, nil
		}

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
	t := themes[m.themeIdx]
	accentColor := lipgloss.Color(t.accent)
	dimColor := lipgloss.Color(t.dim)
	textColor := lipgloss.Color(t.text)
	selectedBg := lipgloss.Color(t.selectedBg)
	selectedFg := lipgloss.Color(t.selectedFg)
	dirColor := accentColor // Folders match the primary accent color

	// Reserve space for borders (2 chars each side) and a 1-char gap
	usableWidth := m.width - 5
	if usableWidth < 20 {
		usableWidth = 20
	}

	// We fix the file list at 40 wide, but never more than 40% of total width
	leftWidth := 40
	if leftWidth > usableWidth*2/5 {
		leftWidth = usableWidth * 2 / 5
	}
	rightWidth := usableWidth - leftWidth

	// Usable height inside borders (top + bottom border = 2 lines)
	// We reserve 1 line for the header, 1 line for the status bar and 1 line for safety.
	contentHeight := m.height - 5
	if contentHeight < 1 {
		contentHeight = 1
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

	normalItem := lipgloss.NewStyle().Foreground(textColor).Width(leftWidth - 2)
	selectedItem := lipgloss.NewStyle().
		Foreground(selectedFg).
		Background(selectedBg).
		Bold(true).
		Width(leftWidth - 2)
	dirStyle := lipgloss.NewStyle().Foreground(dirColor).Bold(true).Width(leftWidth - 2)

	for i := scrollOffset; i < len(m.entries) && len(listLines) < contentHeight; i++ {
		e := m.entries[i]
		name := e.name

		// Determine the single indicator and its style
		var symbol string
		var symStyle lipgloss.Style

		if e.isDir {
			symbol = "▸"
			symStyle = lipgloss.NewStyle().Foreground(dirColor)
			name = name + "/"
		} else {
			symbol = "•"
			symStyle = lipgloss.NewStyle().Foreground(textColor)
		}

		// Git override
		if status, ok := m.gitStatus[name]; ok {
			switch status {
			case "M", "A", "R", "C", "U", "MM", "AM":
				symbol = "●"
			case "?":
				symbol = "○"
			default:
				symbol = "◆"
			}
		}

		line := symStyle.Render(symbol) + " " + name
		line = truncate(line, leftWidth-4)

		if i == m.cursor {
			if e.isDir {
				listLines = append(listLines, selectedItem.Render(line))
			} else {
				listLines = append(listLines, selectedItem.Render(line))
			}
		} else {
			if e.isDir {
				listLines = append(listLines, dirStyle.Render(line))
			} else {
				listLines = append(listLines, normalItem.Render(line))
			}
		}
	}

	// Pad remaining lines so the box keeps its shape
	for len(listLines) < contentHeight {
		listLines = append(listLines, strings.Repeat(" ", leftWidth-4))
	}

	leftContent := strings.Join(listLines, "\n")

	// ── Right pane: preview ────────────────────────────────────────────
	previewLines := strings.Split(strings.TrimSuffix(m.preview, "\n"), "\n")

	maxScroll := len(previewLines) - contentHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.previewScroll > maxScroll {
		m.previewScroll = maxScroll
	}

	startIdx := m.previewScroll
	endIdx := startIdx + contentHeight
	if endIdx > len(previewLines) {
		endIdx = len(previewLines)
	}

	// Create a copy to prevent mutating the original preview string slice
	visiblePreview := make([]string, endIdx-startIdx)
	copy(visiblePreview, previewLines[startIdx:endIdx])
	for i, l := range visiblePreview {
		visiblePreview[i] = truncate(l, rightWidth-4)
	}
	for len(visiblePreview) < contentHeight {
		visiblePreview = append(visiblePreview, "")
	}
	rightContent := strings.Join(visiblePreview, "\n")

	// ── Pane styles ────────────────────────────────────────────────────
	leftBorderColor := dimColor
	rightBorderColor := dimColor
	if m.focusRight {
		rightBorderColor = accentColor
	} else {
		leftBorderColor = accentColor
	}

	leftPane := lipgloss.NewStyle().
		Width(leftWidth).
		Height(contentHeight+2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(leftBorderColor).
		Padding(0, 1).
		Render(leftContent)

	rightPane := lipgloss.NewStyle().
		Width(rightWidth).
		Height(contentHeight+2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(rightBorderColor).
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
	help := " q:quit  ←/→:focus  ↑/↓:nav/scroll  v/enter:vim  t:theme ?:help"

	statusBar := statusStyle.Render(
		truncate(count+pos+"  │"+help, m.width),
	)

	var layout string
	if m.showHelp {
		// Center the help screen
		helpScreen := renderHelp(m)
		layout = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, helpScreen)
	} else {
		header := renderHeader(m)
		layout = lipgloss.JoinVertical(lipgloss.Left, header, panes, statusBar)
	}

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
	t := themes[m.themeIdx]

	if selected.isDir {
		return previewDir(fullPath, selected, t)
	}
	return previewFile(fullPath, selected, t)
}

// previewDir shows directory metadata and a listing of its contents.
func previewDir(path string, e entry, t theme) string {
	accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.accent)).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.dim))

	var sb strings.Builder
	sb.WriteString(accentStyle.Render("▸ Directory: "+e.name+"/") + "\n")
	sb.WriteString(dimStyle.Render(strings.Repeat("─", 30)) + "\n")

	if e.info != nil {
		sb.WriteString(accentStyle.Render("Modified: ") + e.info.ModTime().Format("2006-01-02 15:04") + "\n")
		sb.WriteString(accentStyle.Render("Mode:     ") + e.info.Mode().String() + "\n")
	}

	children, err := os.ReadDir(path)
	if err != nil {
		sb.WriteString("\n" + dimStyle.Render("(cannot read directory)"))
		return sb.String()
	}

	sb.WriteString(accentStyle.Render("Children: ") + fmt.Sprintf("%d", len(children)) + "\n")
	sb.WriteString(dimStyle.Render(strings.Repeat("─", 30)) + "\n")

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
func previewFile(path string, e entry, t theme) string {
	accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.accent)).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.dim))

	var sb strings.Builder
	sb.WriteString(accentStyle.Render("• File: "+e.name) + "\n")
	sb.WriteString(dimStyle.Render(strings.Repeat("─", 30)) + "\n")

	if e.info != nil {
		sb.WriteString(accentStyle.Render("Size:     ") + humanSize(e.info.Size()) + "\n")
		sb.WriteString(accentStyle.Render("Modified: ") + e.info.ModTime().Format("2006-01-02 15:04") + "\n")
		sb.WriteString(accentStyle.Render("Mode:     ") + e.info.Mode().String() + "\n")
	}

	sb.WriteString(dimStyle.Render(strings.Repeat("─", 30)) + "\n")

	// Try to read as text (first 4KB)
	data, err := os.ReadFile(path)
	if err != nil {
		sb.WriteString("\n(cannot read file)")
		return sb.String()
	}

	// Limit to first 32KB for preview
	preview := data
	if len(preview) > 32768 {
		preview = preview[:32768]
	}

	// Check if the content looks like binary
	if isBinary(preview) {
		sb.WriteString("\n(binary file – no preview)")
		return sb.String()
	}

	previewStr := string(preview)

	// Apply Chroma Syntax Highlighting
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
		// Replace tabs with spaces for cleaner rendering
		l = strings.ReplaceAll(l, "\t", "    ")
		sb.WriteString(fmt.Sprintf("%3d │ %s\n", i+1, l))
	}

	if len(data) > 32768 || len(lines) >= maxLines {
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

// truncate cuts a string to maxLen characters physically, appending "…" if truncated.
func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	return ansi.Truncate(s, maxLen, "…")
}

// visibleLen returns the approximate visible length of a string,
// stripping ANSI escape sequences.
func visibleLen(s string) int {
	return lipgloss.Width(s)
}

func renderHelp(m model) string {
	t := themes[m.themeIdx]
	accentColor := lipgloss.Color(t.accent)
	dimColor := lipgloss.Color(t.dim)
	textColor := lipgloss.Color(t.text)

	titleStyle := lipgloss.NewStyle().Foreground(accentColor).Bold(true).MarginBottom(1)
	keyStyle := lipgloss.NewStyle().Foreground(accentColor).Width(15)
	descStyle := lipgloss.NewStyle().Foreground(textColor)
	sectionStyle := lipgloss.NewStyle().MarginTop(1)

	header := titleStyle.Render(fmt.Sprintf("Scout Help - %s Theme (press any key to dismiss)", t.name))

	hotkeys := []string{
		keyStyle.Render("j, down") + descStyle.Render("Move cursor down / Scroll preview"),
		keyStyle.Render("k, up") + descStyle.Render("Move cursor up / Scroll preview"),
		keyStyle.Render("h, left") + descStyle.Render("Back to parent / Unfocus preview"),
		keyStyle.Render("l, right") + descStyle.Render("Enter directory / Focus preview"),
		keyStyle.Render("v, enter") + descStyle.Render("Open file in Vim"),
		keyStyle.Render("g") + descStyle.Render("Go to top"),
		keyStyle.Render("G") + descStyle.Render("Go to bottom"),
		keyStyle.Render("t") + descStyle.Render("Cycle color themes"),
		keyStyle.Render("?") + descStyle.Render("Show/hide this help"),
		keyStyle.Render("q, ctrl+c") + descStyle.Render("Quit scout"),
	}

	symbols := []string{
		keyStyle.Render("●") + descStyle.Render("Modified file"),
		keyStyle.Render("○") + descStyle.Render("Untracked/New file"),
		keyStyle.Render("◆") + descStyle.Render("Other git change"),
		keyStyle.Render("▸") + descStyle.Render("Directory"),
		keyStyle.Render("•") + descStyle.Render("Regular file"),
	}

	helpBody := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		sectionStyle.Render(lipgloss.NewStyle().Foreground(dimColor).Render("─ KEYBOARD SHORTCUTS ─")),
		lipgloss.JoinVertical(lipgloss.Left, hotkeys...),
		sectionStyle.Render(lipgloss.NewStyle().Foreground(dimColor).Render("─ SYMBOLS ─")),
		lipgloss.JoinVertical(lipgloss.Left, symbols...),
	)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accentColor).
		Padding(1, 4).
		Render(helpBody)
}

// getStats returns a command that fetches memory and directory size stats.
func getStats(path string) tea.Cmd {
	return func() tea.Msg {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		dirSize := int64(0)
		entries, _ := os.ReadDir(path)
		for _, e := range entries {
			if !e.IsDir() {
				info, _ := e.Info()
				if info != nil {
					dirSize += info.Size()
				}
			}
		}

		return statsMsg{
			cpu:     0.1, // Simplified placeholder
			mem:     m.Alloc,
			dirSize: dirSize,
		}
	}
}

func doTick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func renderHeader(m model) string {
	t := themes[m.themeIdx]
	accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.accent)).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.dim))

	left := accentStyle.Render(" scout ") + dimStyle.Render("│ github.com/mirageglobe/scout")

	memMB := float64(m.stats.mem) / 1024 / 1024
	dirSizeStr := humanSize(m.stats.dirSize)

	// Build stats string
	statsStr := fmt.Sprintf("MEM: %.1fMB  DIR: %s", memMB, dirSizeStr)
	right := dimStyle.Render(statsStr + " ")

	// Calculate space between
	width := m.width
	paddingCount := width - lipgloss.Width(left) - lipgloss.Width(right)
	if paddingCount < 0 {
		paddingCount = 0
	}

	return left + strings.Repeat(" ", paddingCount) + right
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
