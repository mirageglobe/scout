package ui

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

// RenderHelp generates the full-screen help overlay.
func (m Model) RenderHelp() string {
	t := Themes[m.ThemeIdx]
	accentColor := lipgloss.Color(t.Accent)
	dimColor := lipgloss.Color(t.Dim)
	textColor := lipgloss.Color(t.Text)

	titleStyle := lipgloss.NewStyle().Foreground(accentColor).Bold(true).MarginBottom(1)
	keyStyle := lipgloss.NewStyle().Foreground(accentColor).Width(16)
	descStyle := lipgloss.NewStyle().Foreground(textColor)
	sectionStyle := lipgloss.NewStyle().Foreground(dimColor).MarginBottom(1)
	colStyle := lipgloss.NewStyle().PaddingRight(4)

	header := titleStyle.Render(fmt.Sprintf("scout help  -  %s theme  (press any key to dismiss)", t.Name))

	hotkeys := lipgloss.JoinVertical(lipgloss.Left,
		sectionStyle.Render("─ keys ─"),
		keyStyle.Render("↑/↓, j/k")   +descStyle.Render("navigate / scroll preview"),
		keyStyle.Render("←/→, h/l")   +descStyle.Render("nav to parent / nav into directory or preview"),
		keyStyle.Render("backspace")   +descStyle.Render("nav to parent directory"),
		keyStyle.Render("enter")       +descStyle.Render("enter directory / nav to preview"),
		keyStyle.Render("e")           +descStyle.Render("open file in editor"),
		keyStyle.Render("o")           +descStyle.Render("open with system default"),
		keyStyle.Render("i")           +descStyle.Render("toggle hidden files"),
		keyStyle.Render("f")           +descStyle.Render("toggle root-focus mode"),
		keyStyle.Render("tab")         +descStyle.Render("collapse / expand explorer"),
		keyStyle.Render("r")           +descStyle.Render("refresh preview"),
		keyStyle.Render("t")           +descStyle.Render("cycle color themes"),
		keyStyle.Render("/")           +descStyle.Render("search active pane  (n / N: next / prev)"),
		keyStyle.Render("esc")          +descStyle.Render("clear search"),
		keyStyle.Render("?")           +descStyle.Render("show / hide help"),
		keyStyle.Render("q, ctrl+c")   +descStyle.Render("quit"),
	)

	symbols := lipgloss.JoinVertical(lipgloss.Left,
		sectionStyle.Render("─ symbols ─"),
		keyStyle.Render("▸") +descStyle.Render("directory"),
		keyStyle.Render("·") +descStyle.Render("clean file"),
		keyStyle.Render("M") +descStyle.Render("git modified"),
		keyStyle.Render("+") +descStyle.Render("git added / staged"),
		keyStyle.Render("?") +descStyle.Render("git untracked"),
		keyStyle.Render("!") +descStyle.Render("other git change"),
	)

	columns := lipgloss.JoinHorizontal(lipgloss.Top, colStyle.Render(hotkeys), symbols)

	helpBody := lipgloss.JoinVertical(lipgloss.Left, header, columns)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accentColor).
		Padding(1, 4).
		MaxHeight(m.Height - 2).
		Render(helpBody)
}
