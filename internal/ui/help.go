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
	keyStyle := lipgloss.NewStyle().Foreground(accentColor).Width(15)
	descStyle := lipgloss.NewStyle().Foreground(textColor)
	sectionStyle := lipgloss.NewStyle().MarginTop(1)

	header := titleStyle.Render(fmt.Sprintf("Scout Help - %s Theme (press any key to dismiss)", t.Name))

	hotkeys := []string{
		keyStyle.Render("j, down") + descStyle.Render("Move cursor down / Scroll preview"),
		keyStyle.Render("k, up") + descStyle.Render("Move cursor up / Scroll preview"),
		keyStyle.Render("h, left") + descStyle.Render("Back to parent / Unfocus preview"),
		keyStyle.Render("l, right") + descStyle.Render("Enter directory / Focus preview"),
		keyStyle.Render("v, enter") + descStyle.Render("Open file in Vim"),
		keyStyle.Render("g") + descStyle.Render("Go to top"),
		keyStyle.Render("G") + descStyle.Render("Go to bottom"),
		keyStyle.Render("t") + descStyle.Render("Cycle color themes"),
		keyStyle.Render("o") + descStyle.Render("Open with system default"),
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
