package ui

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mirageglobe/scout/internal/filesystem"
)

// View renders the entire application UI.
func (m Model) View() tea.View {
	if m.Width == 0 {
		return tea.NewView("Loading…")
	}

	// ── Colours & metrics ──────────────────────────────────────────────
	t := Themes[m.ThemeIdx]
	accentColor := lipgloss.Color(t.Accent)
	dimColor := lipgloss.Color(t.Dim)
	textColor := lipgloss.Color(t.Text)
	selectedBg := lipgloss.Color(t.SelectedBg)
	selectedFg := lipgloss.Color(t.SelectedFg)
	dirColor := accentColor

	// Use the full width of the terminal
	usableWidth := m.Width
	if usableWidth < 20 {
		usableWidth = 20
	}

	leftWidth := 40
	if m.ExplorerCollapsed {
		leftWidth = 10
	} else if leftWidth > usableWidth*2/5 {
		leftWidth = usableWidth * 2 / 5
	}
	rightWidth := usableWidth - leftWidth

	contentHeight := m.Height - 5
	if contentHeight < 1 {
		contentHeight = 1
	}

	// ── Left pane: file list ───────────────────────────────────────────
	var listLines []string
	cwdDisplay := m.Cwd
	home, _ := os.UserHomeDir()
	if strings.HasPrefix(cwdDisplay, home) {
		cwdDisplay = "~" + cwdDisplay[len(home):]
	}

	headerStyle := lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	listLines = append(listLines, headerStyle.Render(filesystem.Truncate(cwdDisplay, leftWidth-4)))

	visibleRows := contentHeight - len(listLines) - 1 // -1 for the stats line
	if visibleRows < 1 {
		visibleRows = 1
	}

	scrollOffset := 0
	if m.Cursor >= visibleRows {
		scrollOffset = m.Cursor - visibleRows + 1
	}

	normalItem := lipgloss.NewStyle().Foreground(textColor).Width(leftWidth - 4)
	selectedItem := lipgloss.NewStyle().
		Foreground(selectedFg).
		Background(selectedBg).
		Bold(true).
		Width(leftWidth - 4)
	dirStyle     := lipgloss.NewStyle().Foreground(dirColor)
	dirCountStyle := lipgloss.NewStyle().Foreground(dimColor)

	for i := scrollOffset; i < len(m.Entries) && len(listLines) < contentHeight-1; i++ {
		e := m.Entries[i]
		name := e.Name

		var symbol string
		var symStyle lipgloss.Style
		var dirBaseName, dirCountStr string

		if e.IsDir {
			symbol = "▸"
			symStyle = lipgloss.NewStyle().Foreground(dirColor)
			dirCountStr = fmt.Sprintf("%d ≡", e.SubCount)
			dirBaseName = e.Name + "/"
			nameWidth := leftWidth - 6 // leftWidth-4 content minus 2 for symbol+space
			if padWidth := nameWidth - lipgloss.Width(dirCountStr); padWidth >= len(dirBaseName) {
				name = fmt.Sprintf("%-*s%s", padWidth, dirBaseName, dirCountStr)
			} else {
				name = dirBaseName
				dirCountStr = "" // no room for count
			}
		} else if e.IsSymlink {
			symbol = "↳"
			symStyle = lipgloss.NewStyle().Foreground(accentColor)
		} else {
			symbol = "·"
			symStyle = lipgloss.NewStyle().Foreground(dimColor)
		}

		if status, ok := m.GitStatus[e.Name]; ok {
			switch status {
			case "M":
				symbol = "M"
				symStyle = lipgloss.NewStyle().Foreground(accentColor)
			case "A", "AM":
				symbol = "+"
				symStyle = lipgloss.NewStyle().Foreground(accentColor)
			case "?":
				symbol = "?"
				symStyle = lipgloss.NewStyle().Foreground(accentColor)
			default:
				symbol = "!"
				symStyle = lipgloss.NewStyle().Foreground(accentColor)
			}
		}

		// Raw text for selection (no ANSI)
		rawLine := symbol + " " + name
		truncated := filesystem.Truncate(rawLine, leftWidth-4)

		if i == m.Cursor {
			// SELECTED: Render plain text on a solid background
			listLines = append(listLines, selectedItem.Render(truncated))
		} else {
			// NORMAL: Render with themed symbol and name colors
			symStyled := symStyle.Render(symbol)
			var lineStyled string
			if e.IsDir && dirCountStr != "" {
				nameWidth := leftWidth - 6
				padWidth := nameWidth - lipgloss.Width(dirCountStr)
				paddedBase := fmt.Sprintf("%-*s", padWidth, dirBaseName)
				lineStyled = symStyled + " " + dirStyle.Render(paddedBase) + dirCountStyle.Render(dirCountStr)
			} else if e.IsDir {
				lineStyled = symStyled + " " + dirStyle.Render(name)
			} else {
				lineStyled = symStyled + " " + normalItem.Render(name)
			}
			listLines = append(listLines, filesystem.Truncate(lineStyled, leftWidth-4))
		}
	}

	for len(listLines) < contentHeight-1 {
		listLines = append(listLines, strings.Repeat(" ", leftWidth-4))
	}

	// Last line: directory stats (hidden when explorer is collapsed)
	if m.ExplorerCollapsed {
		listLines = append(listLines, strings.Repeat(" ", leftWidth-4))
	} else {
		dirStatStyle := lipgloss.NewStyle().Foreground(dimColor)
		itemStat := fmt.Sprintf("%d items", len(m.Entries))
		if len(m.Entries) > 0 {
			itemStat = fmt.Sprintf("%d/%d items", m.Cursor+1, len(m.Entries))
		}
		curFileSize := ""
		if len(m.Entries) > 0 && m.Entries[m.Cursor].Info != nil && !m.Entries[m.Cursor].IsDir {
			curFileSize = filesystem.HumanSize(m.Entries[m.Cursor].Info.Size()) + " / "
		}
		dirStatLine := dirStatStyle.Render(fmt.Sprintf(" %s  %s%s", itemStat, curFileSize, filesystem.HumanSize(m.Stats.DirSize)))
		listLines = append(listLines, dirStatLine)
	}

	leftContent := strings.Join(listLines, "\n")

	// ── Right pane: preview ────────────────────────────────────────────
	previewLines := strings.Split(strings.TrimSuffix(m.Preview, "\n"), "\n")
	startIdx := min(m.PreviewScroll, len(previewLines))
	endIdx := min(startIdx+contentHeight, len(previewLines))

	visiblePreview := make([]string, endIdx-startIdx)
	copy(visiblePreview, previewLines[startIdx:endIdx])

	// build match lookup for highlighting
	matchSet := make(map[int]bool, len(m.SearchMatches))
	for _, idx := range m.SearchMatches {
		matchSet[idx] = true
	}
	currentMatchLine := -1
	if len(m.SearchMatches) > 0 {
		currentMatchLine = m.SearchMatches[m.SearchMatchIdx]
	}
	matchStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#44475A")).
		Foreground(lipgloss.Color("#F1FA8C")).
		Width(rightWidth - 4)
	currentMatchStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#F1FA8C")).
		Foreground(lipgloss.Color("#282A36")).
		Bold(true).
		Width(rightWidth - 4)

	dimEllipsis := lipgloss.NewStyle().Foreground(dimColor).Render("…")
	for i, l := range visiblePreview {
		absIdx := startIdx + i
		truncated := filesystem.TruncateWithTail(l, rightWidth-4, dimEllipsis)
		if absIdx == currentMatchLine {
			visiblePreview[i] = currentMatchStyle.Render(stripANSI(truncated))
		} else if matchSet[absIdx] {
			visiblePreview[i] = matchStyle.Render(stripANSI(truncated))
		} else {
			visiblePreview[i] = truncated
		}
	}

	for len(visiblePreview) < contentHeight {
		visiblePreview = append(visiblePreview, "")
	}
	rightContent := strings.Join(visiblePreview, "\n")

	// ── Pane styles ────────────────────────────────────────────────────
	leftBorderColor := dimColor
	rightBorderColor := dimColor
	if m.FocusRight {
		rightBorderColor = accentColor
	} else {
		leftBorderColor = accentColor
	}

	rightPane := lipgloss.NewStyle().
		Width(rightWidth).
		Height(contentHeight + 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(rightBorderColor).
		Padding(0, 1).
		Render(rightContent)

	leftPane := lipgloss.NewStyle().
		Width(leftWidth).
		Height(contentHeight + 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(leftBorderColor).
		Padding(0, 1).
		Render(leftContent)

	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	// ── Status bar ─────────────────────────────────────────────────────
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	dimHint := lipgloss.NewStyle().Foreground(dimColor)
	activeHint := lipgloss.NewStyle().Foreground(accentColor).Bold(true)

	hint := func(key, label string, on bool) string {
		text := key + ":" + label
		if on {
			return activeHint.Render(text)
		}
		return dimHint.Render(text)
	}

	var statusBar string
	if m.GitBranch != "" {
		statusBar = dimHint.Render(" ⎇ " + m.GitBranch + "  │")
	}
	sep := "  "
	statusBar += " " + hint("↑/↓", "nav", false) +
		sep + hint("←/→", "nav", false) +
		sep + hint("e", "edit("+editor+")", false) +
		sep + hint("o", "open", false) +
		sep + hint("i", "hidden", !m.ShowHidden) +
		sep + hint("f", "root-focus", m.RootFocus) +
		sep + hint("tab", "explorer", m.ExplorerCollapsed) +
		sep + hint("r", "refresh", false) +
		sep + hint("t", "theme", false) +
		sep + hint("/", "search", false) +
		sep + hint("?", "help", false) +
		sep + hint("q", "quit", false)
	statusBar = filesystem.Truncate(statusBar, m.Width)

	var layout string
	if m.ShowHelp {
		helpScreen := m.RenderHelp()
		layout = lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, helpScreen)
	} else {
		header := m.RenderHeader()
		statusLine := m.RenderStatusLine()
		layout = lipgloss.JoinVertical(lipgloss.Left, header, panes, statusLine, statusBar)
	}

	v := tea.NewView(layout)
	v.AltScreen = true
	return v
}

// RenderStatusLine generates the persistent "scout: " prompt line between panes and the hint bar.
func (m Model) RenderStatusLine() string {
	accent := lipgloss.Color(Themes[m.ThemeIdx].Accent)
	dim := lipgloss.Color(Themes[m.ThemeIdx].Dim)
	dimStyle := lipgloss.NewStyle().Foreground(dim)
	accentStyle := lipgloss.NewStyle().Foreground(accent)

	// "scout ›" prefix is always rendered bold+accent, followed by state-specific content.
	prefix := " " + lipgloss.NewStyle().Foreground(accent).Bold(true).Render("scout ›") + " "

	if m.Loading {
		dots := [3]string{"·", "··", "···"}
		return prefix + accentStyle.Render(dots[m.SpinnerFrame])
	}

	if m.ExplorerSearchActive {
		return prefix + accentStyle.Render("/"+m.ExplorerSearchInput+"█") +
			dimStyle.Render("  enter:confirm  esc:clear")
	}

	if m.ExplorerSearchInput != "" {
		count := len(m.explorerFiltered())
		return prefix + accentStyle.Render(fmt.Sprintf("/%s  [%d matches]", m.ExplorerSearchInput, count)) +
			dimStyle.Render("  n/N:next/prev  esc:clear")
	}

	if m.SearchActive {
		return prefix + accentStyle.Render("/"+m.SearchInput+"█") +
			dimStyle.Render("  enter:confirm  esc:exit")
	}

	if m.SearchQuery != "" {
		if len(m.SearchMatches) == 0 {
			return prefix + lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).
				Render("/"+m.SearchQuery+"  [no matches]") + dimStyle.Render("  esc:clear")
		}
		return prefix + accentStyle.Render(fmt.Sprintf("/%s  [%d/%d]", m.SearchQuery, m.SearchMatchIdx+1, len(m.SearchMatches))) +
			dimStyle.Render("  n/N:next/prev  esc:clear")
	}

	if m.StatusMsg != "" {
		msgStyle := accentStyle.Italic(true)
		if strings.HasPrefix(m.StatusMsg, "error:") {
			msgStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
		}
		return prefix + msgStyle.Render(filesystem.Truncate(m.StatusMsg, m.Width-10))
	}

	// idle: dim prompt awaiting input
	return dimStyle.Render(" scout › ")
}
