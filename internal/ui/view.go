package ui

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mirageglobe/scout/internal/filesystem"
)

// preview highlight styles; the colours are fixed, only .Width() varies per frame,
// so the bases are built once here instead of on every View call.
var (
	previewMatchBase        = lipgloss.NewStyle().Background(lipgloss.Color("#44475A")).Foreground(lipgloss.Color("#F1FA8C"))
	previewCurrentMatchBase = lipgloss.NewStyle().Background(lipgloss.Color("#F1FA8C")).Foreground(lipgloss.Color("#282A36")).Bold(true)
	previewSelectionBase    = lipgloss.NewStyle().Background(lipgloss.Color("#3D59A1")).Foreground(lipgloss.Color("#C0CAF5"))
)

// withPreviewDisplay recomputes the cached display rows for m.Preview: each source
// line wrapped (PreviewWrap on) or truncated to the pane width. Update calls this
// only when an input changes, so View and scrolling read the cache instead of
// re-expanding every line on every frame.
func (m Model) withPreviewDisplay() Model {
	w := previewWrapWidth(m)
	raw := strings.Split(strings.TrimSuffix(m.Preview, "\n"), "\n")
	dimEllipsis := lipgloss.NewStyle().Foreground(lipgloss.Color(Themes[m.ThemeIdx].Dim)).Render(m.Sym.Ellipsis)
	out := make([]displayLine, 0, len(raw))
	for origIdx, l := range raw {
		if m.PreviewWrap {
			for _, sub := range strings.Split(lipgloss.Wrap(l, w, " "), "\n") {
				out = append(out, displayLine{sub, origIdx})
			}
		} else {
			out = append(out, displayLine{filesystem.TruncateWithTail(l, w, dimEllipsis), origIdx})
		}
	}
	m.previewDisplay = out
	m.previewDisplayFor = m.Preview
	m.previewDisplayW = w
	m.previewDisplayWrap = m.PreviewWrap
	m.previewDisplayTheme = m.ThemeIdx
	return m
}

// hintTips returns the rotating [key, description] pairs shown after 10s idle,
// built from the active glyph set so arrows honour SCOUT_UNICODE_SAFE.
func hintTips(g Glyphs) [][2]string {
	return [][2]string{
		{g.Up + " / " + g.Down, "move up and down through files"},
		{g.Left + " / " + g.Right, "jump between the explorer and preview panes"},
		{"e", "open the selected file in your $EDITOR"},
		{"o", "open the file with your system default app"},
		{"d", "cycle the preview between the file, its git diff, and git log"},
		{"i", "show or hide dotfiles and hidden entries"},
		{"f", "lock navigation to the folder you launched scout from"},
		{"r", "force a refresh of the current directory"},
		{"t", "switch to the next colour theme"},
		{"/", "search for text inside the preview pane"},
		{"?", "open the full help overlay"},
		{"q", "quit scout"},
	}
}

// View renders the entire application UI.
func (m Model) View() tea.View {
	if m.Width == 0 {
		return tea.NewView("Loading" + m.Sym.Ellipsis)
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

	leftWidth := ExplorerLeftWidth(m.ExplorerWidthMode, usableWidth)
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

	scrollOffset := m.ExplorerScroll

	normalItem := lipgloss.NewStyle().Foreground(textColor).Width(leftWidth - 4)
	selectedItem := lipgloss.NewStyle().
		Foreground(selectedFg).
		Background(selectedBg).
		Bold(true).
		Width(leftWidth - 4)
	dirStyle := lipgloss.NewStyle().Foreground(dirColor)
	dirCountStyle := lipgloss.NewStyle().Foreground(dimColor)

	for i := scrollOffset; i < len(m.Entries) && len(listLines) < contentHeight-1; i++ {
		e := m.Entries[i]
		name := e.Name

		var symbol string
		var symStyle lipgloss.Style
		var dirBaseName, dirCountStr, fileSizeStr string

		if e.IsDir {
			symbol = m.Sym.Dir
			symStyle = lipgloss.NewStyle().Foreground(dirColor)
			dirCountStr = fmt.Sprintf("%d %s", e.SubCount, m.Sym.SubCount)
			dirBaseName = e.Name + "/"
			nameWidth := leftWidth - 6 // leftWidth-4 content minus 2 for symbol+space
			if padWidth := nameWidth - lipgloss.Width(dirCountStr); padWidth >= len(dirBaseName) {
				name = fmt.Sprintf("%-*s%s", padWidth, dirBaseName, dirCountStr)
			} else {
				name = dirBaseName
				dirCountStr = "" // no room for count
			}
		} else if e.IsSymlink {
			symbol = m.Sym.Symlink
			symStyle = lipgloss.NewStyle().Foreground(accentColor)
		} else {
			symbol = m.Sym.Dot
			symStyle = lipgloss.NewStyle().Foreground(dimColor)
		}

		if !e.IsDir && e.Info != nil {
			fileSizeStr = filesystem.HumanSize(e.Info.Size())
			nameWidth := leftWidth - 6
			if padWidth := nameWidth - lipgloss.Width(fileSizeStr); padWidth >= len(e.Name) {
				name = fmt.Sprintf("%-*s%s", padWidth, e.Name, fileSizeStr)
			} else {
				fileSizeStr = ""
			}
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
			} else if fileSizeStr != "" {
				nameWidth := leftWidth - 6
				padWidth := nameWidth - lipgloss.Width(fileSizeStr)
				paddedBase := fmt.Sprintf("%-*s", padWidth, e.Name)
				normalStyle := lipgloss.NewStyle().Foreground(textColor)
				lineStyled = symStyled + " " + normalStyle.Render(paddedBase) + dirCountStyle.Render(fileSizeStr)
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
	if m.ExplorerWidthMode == 1 || m.ExplorerWidthMode == 2 { // sliver/narrow: too narrow for stats
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
		changeStat := ""
		if n := changedFileCount(m.GitStatus); n > 0 {
			changeStat = fmt.Sprintf(" %s %d", m.Sym.Changed, n)
		}
		leftStat := fmt.Sprintf("%s%s", itemStat, changeStat)
		rightStat := fmt.Sprintf("%s%s", curFileSize, filesystem.HumanSize(m.Stats.DirSize))
		paneWidth := leftWidth - 4
		gap := paneWidth - len(leftStat) - len(rightStat)
		if gap < 1 {
			gap = 1
		}
		dirStatLine := dirStatStyle.Render(leftStat + strings.Repeat(" ", gap) + rightStat)
		listLines = append(listLines, dirStatLine)
	}

	leftContent := strings.Join(listLines, "\n")

	// ── Right pane: preview ────────────────────────────────────────────
	// build match lookup for highlighting
	matchSet := make(map[int]bool, len(m.SearchMatches))
	for _, idx := range m.SearchMatches {
		matchSet[idx] = true
	}
	currentMatchLine := -1
	if len(m.SearchMatches) > 0 {
		currentMatchLine = m.SearchMatches[m.SearchMatchIdx]
	}
	matchStyle := previewMatchBase.Width(rightWidth - 4)
	currentMatchStyle := previewCurrentMatchBase.Width(rightWidth - 4)
	selectionStyle := previewSelectionBase.Width(rightWidth - 4)

	// display lines are precomputed and cached on the model (see withPreviewDisplay);
	// Update rebuilds the cache only when the preview, width, wrap, or theme changes,
	// so this render path stays O(visible) instead of re-expanding every line.
	displayLines := m.previewDisplay

	startIdx := min(m.PreviewScroll, len(displayLines))
	endIdx := min(startIdx+contentHeight, len(displayLines))

	dragLo, dragHi := -1, -1
	if m.DragActive {
		dragLo = m.DragStartRow
		dragHi = m.DragEndRow
		if dragLo > dragHi {
			dragLo, dragHi = dragHi, dragLo
		}
	}

	visiblePreview := make([]string, endIdx-startIdx)
	for i, dl := range displayLines[startIdx:endIdx] {
		switch {
		case m.DragActive && i >= dragLo && i <= dragHi:
			visiblePreview[i] = selectionStyle.Render(stripANSI(dl.text))
		case dl.origIdx == currentMatchLine:
			visiblePreview[i] = currentMatchStyle.Render(stripANSI(dl.text))
		case matchSet[dl.origIdx]:
			visiblePreview[i] = matchStyle.Render(stripANSI(dl.text))
		default:
			visiblePreview[i] = dl.text
		}
	}

	for len(visiblePreview) < contentHeight {
		visiblePreview = append(visiblePreview, "")
	}
	rightContent := strings.Join(visiblePreview, "\n")
	previewLines := displayLines // alias for scrollbar math below

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
		Height(contentHeight+2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(rightBorderColor).
		Padding(0, 1).
		Render(rightContent)

	totalLines := len(previewLines)
	if totalLines > contentHeight {
		thumbH := contentHeight * contentHeight / totalLines
		if thumbH < 1 {
			thumbH = 1
		}
		maxOffset := contentHeight - thumbH
		maxScroll := totalLines - contentHeight
		thumbStart := 0
		if maxScroll > 0 {
			thumbStart = m.PreviewScroll * maxOffset / maxScroll
		}
		thumbChar := lipgloss.NewStyle().Foreground(accentColor).Render("▐")
		rightPane = injectScrollbar(rightPane, contentHeight, thumbStart, thumbStart+thumbH-1, thumbChar)
	}

	leftPane := lipgloss.NewStyle().
		Width(leftWidth).
		Height(contentHeight+2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(leftBorderColor).
		Padding(0, 1).
		Render(leftContent)

	visibleEntryRows := contentHeight - 2 // minus cwd header and stats line
	if len(m.Entries) > visibleEntryRows && visibleEntryRows > 0 {
		thumbH := visibleEntryRows * visibleEntryRows / len(m.Entries)
		if thumbH < 1 {
			thumbH = 1
		}
		maxOffset := visibleEntryRows - thumbH
		maxScroll := len(m.Entries) - visibleEntryRows
		thumbStart := 0
		if maxScroll > 0 {
			thumbStart = scrollOffset * maxOffset / maxScroll
		}
		thumbChar := lipgloss.NewStyle().Foreground(accentColor).Render("▐")
		leftPane = injectScrollbar(leftPane, contentHeight, thumbStart+1, thumbStart+thumbH, thumbChar)
	}

	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	// ── Status bar ─────────────────────────────────────────────────────
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	dimHint := lipgloss.NewStyle().Foreground(dimColor)
	activeHint := lipgloss.NewStyle().Foreground(accentColor).Bold(true)

	var statusBar string
	if m.GitBranch != "" {
		statusBar = dimHint.Render(" " + m.Sym.Branch + " " + m.GitBranch + "  │")
	}
	sep := "  "

	if m.HintCycling {
		tip := hintTips(m.Sym)[m.HintTipIdx]
		statusBar += " " + activeHint.Render(tip[0]) + dimHint.Render("  "+m.Sym.Dot+"  "+tip[1])
	} else {
		hint := func(key, label string, on bool) string {
			text := key + ":" + label
			if on {
				return activeHint.Render(text)
			}
			return dimHint.Render(text)
		}
		if m.FocusRight {
			// preview pane hints
			statusBar += " " + hint(m.Sym.Up+"/"+m.Sym.Down, "scroll", false) +
				sep + hint(m.Sym.Left, "explorer", false) +
				sep + hint("e", "edit("+editor+")", false) +
				sep + hint("o", "open", false) +
				sep + hint("r", "refresh", false) +
				sep + hint("w", "wrap", m.PreviewWrap) +
				sep + hint("t", "theme", false) +
				sep + hint("/", "search", false) +
				sep + hint("?", "help", false) +
				sep + hint("q", "quit", false)
		} else {
			// explorer pane hints
			statusBar += " " + hint(m.Sym.Up+"/"+m.Sym.Down, "nav", false) +
				sep + hint(m.Sym.Left+"/"+m.Sym.Right, "nav", false) +
				sep + hint("e", "edit("+editor+")", false) +
				sep + hint("o", "open", false) +
				sep + hint("i", "show hidden", m.ShowHidden) +
				sep + hint("l", "root-lock", m.RootLock) +
				sep + hint("d", "git", m.PreviewMode != PreviewFile) +
				sep + hint("t", "theme", false) +
				sep + hint("/", "search", false) +
				sep + hint("?", "help", false) +
				sep + hint("q", "quit", false)
		}
	}
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
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// injectScrollbar replaces the right-border │ on thumb lines with thumbChar.
// paneHeight is the number of content rows (excluding border lines).
// thumbStart and thumbEnd are 0-indexed content row positions.
func injectScrollbar(pane string, paneHeight, thumbStart, thumbEnd int, thumbChar string) string {
	lines := strings.Split(pane, "\n")
	for i, line := range lines {
		contentRow := i - 1 // line 0 = top border, lines 1..paneHeight = content
		if contentRow < thumbStart || contentRow > thumbEnd || contentRow < 0 || contentRow >= paneHeight {
			continue
		}
		idx := strings.LastIndex(line, "│")
		if idx < 0 {
			continue
		}
		lines[i] = line[:idx] + thumbChar + line[idx+len("│"):]
	}
	return strings.Join(lines, "\n")
}

// changedFileCount returns the number of entries with a git status.
func changedFileCount(status map[string]string) int {
	return len(status)
}

// RenderStatusLine generates the persistent "scout: " prompt line between panes and the hint bar.
func (m Model) RenderStatusLine() string {
	accent := lipgloss.Color(Themes[m.ThemeIdx].Accent)
	dim := lipgloss.Color(Themes[m.ThemeIdx].Dim)
	dimStyle := lipgloss.NewStyle().Foreground(dim)
	accentStyle := lipgloss.NewStyle().Foreground(accent)

	// "scout ›" prefix is always rendered bold+accent, followed by state-specific content.
	prefix := " " + lipgloss.NewStyle().Foreground(accent).Bold(true).Render("scout "+m.Sym.Prompt) + " "

	if m.Loading {
		d := m.Sym.Dot
		dots := [3]string{d, d + d, d + d + d}
		return prefix + accentStyle.Render(dots[m.SpinnerFrame])
	}

	if m.ExplorerSearchActive {
		return prefix + accentStyle.Render("/"+m.ExplorerSearchInput+m.Sym.Block) +
			dimStyle.Render("  enter:confirm  esc:clear")
	}

	if m.ExplorerSearchInput != "" {
		count := len(m.explorerFiltered())
		return prefix + accentStyle.Render(fmt.Sprintf("/%s  [%d matches]", m.ExplorerSearchInput, count)) +
			dimStyle.Render("  n/N:next/prev  esc:clear")
	}

	if m.SearchActive {
		return prefix + accentStyle.Render("/"+m.SearchInput+m.Sym.Block) +
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
		return prefix + dimStyle.Render(filesystem.Truncate(m.StatusMsg, m.Width-10))
	}

	// idle: dim prompt awaiting input
	return dimStyle.Render(" scout " + m.Sym.Prompt + " ")
}
