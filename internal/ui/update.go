package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/mirageglobe/scout/internal/filesystem"
)

// Update handles all state transitions in response to messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case filesystem.TickMsg:
		return m, tea.Batch(filesystem.DoTick(), filesystem.GetStats(m.Cwd))

	case filesystem.StatsMsg:
		m.Stats.CPU = msg.CPU
		m.Stats.Mem = msg.Mem
		m.Stats.DirSize = msg.DirSize
		return m, nil

	case filesystem.DirLoadedMsg:
		if msg.Err != nil {
			m.StatusMsg = fmt.Sprintf("Error: %v", msg.Err)
			return m, nil
		}
		m.Entries = msg.Entries
		m.GitStatus = msg.GitStatus
		m.GitBranch = msg.GitBranch
		m.Err = nil
		m.StatusMsg = ""
		m.PreviewScroll = 0
		if m.Cursor >= len(m.Entries) {
			m.Cursor = max(0, len(m.Entries)-1)
		}
		m.Preview = m.BuildPreview()
		return m, filesystem.GetStats(m.Cwd)

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Preview = m.BuildPreview()
		return m, nil

	case tea.KeyPressMsg:
		if m.ShowHelp {
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			m.ShowHelp = false
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "?":
			m.ShowHelp = true
			return m, nil

		case "o":
			if len(m.Entries) > 0 {
				selected := m.Entries[m.Cursor]
				fullPath := filepath.Join(m.Cwd, selected.Name)
				if !selected.IsDir {
					if err := filesystem.OpenWithSystem(fullPath); err != nil {
						m.StatusMsg = fmt.Sprintf("Error: %v", err)
					} else {
						m.StatusMsg = fmt.Sprintf("Opened: %s", selected.Name)
					}
				}
			}
			return m, nil

		case "t":
			m.ThemeIdx = (m.ThemeIdx + 1) % len(Themes)
			m.Preview = m.BuildPreview()
			return m, nil

		// Navigation: move cursor down
		case "j", "down":
			if m.FocusRight {
				previewLines := strings.Split(strings.TrimSuffix(m.Preview, "\n"), "\n")
				contentHeight := m.Height - 5
				maxScroll := len(previewLines) - contentHeight
				if maxScroll < 0 {
					maxScroll = 0
				}
				if m.PreviewScroll < maxScroll {
					m.PreviewScroll++
				}
			} else {
				if m.Cursor < len(m.Entries)-1 {
					m.Cursor++
				}
				m.PreviewScroll = 0
				m.Preview = m.BuildPreview()
				m.StatusMsg = ""
			}
			return m, nil

		// Navigation: move cursor up
		case "k", "up":
			if m.FocusRight {
				if m.PreviewScroll > 0 {
					m.PreviewScroll--
				}
			} else {
				if m.Cursor > 0 {
					m.Cursor--
				}
				m.PreviewScroll = 0
				m.Preview = m.BuildPreview()
				m.StatusMsg = ""
			}
			return m, nil

		// Navigation: go to parent directory or unfocus right pane
		case "h", "left":
			if m.FocusRight {
				m.FocusRight = false
				return m, nil
			}
			parent := filepath.Dir(m.Cwd)
			if parent != m.Cwd {
				m.Cwd = parent
				m.PreviewScroll = 0
				m.Preview = ""
				m.StatusMsg = ""
				return m, m.LoadDir(m.Cwd)
			}
			return m, nil

		// Action: open directory or file
		case "v", "enter", "l", "right":
			if len(m.Entries) == 0 {
				return m, nil
			}
			selected := m.Entries[m.Cursor]
			fullPath := filepath.Join(m.Cwd, selected.Name)

			if selected.IsDir {
				m.Cwd = fullPath
				m.Cursor = 0
				m.Preview = ""
				m.FocusRight = false
				return m, m.LoadDir(m.Cwd)
			}

			// Handle file selection
			isAction := msg.String() == "enter" || msg.String() == "v"
			if !isAction {
				if !m.FocusRight {
					m.FocusRight = true
				}
				return m, nil
			}

			// Security check before launching Vim
			f, _ := os.Open(fullPath)
			if f != nil {
				buf := make([]byte, 1024)
				n, _ := f.Read(buf)
				f.Close()
				if filesystem.IsBinary(buf[:n]) {
					m.StatusMsg = fmt.Sprintf("Error: cannot open binary file: %s", selected.Name)
					return m, nil
				}
			}

			c := exec.Command("vim", fullPath)
			return m, tea.ExecProcess(c, func(err error) tea.Msg {
				return EditorFinishedMsg{Err: err}
			})

		case "g":
			m.Cursor = 0
			m.PreviewScroll = 0
			m.Preview = m.BuildPreview()
			m.StatusMsg = ""
			return m, nil

		case "G", "shift+g":
			if len(m.Entries) > 0 {
				m.Cursor = len(m.Entries) - 1
			}
			m.PreviewScroll = 0
			m.Preview = m.BuildPreview()
			m.StatusMsg = ""
			return m, nil
		}

	case EditorFinishedMsg:
		if msg.Err != nil {
			m.StatusMsg = fmt.Sprintf("Error: %v", msg.Err)
		}
		return m, m.LoadDir(m.Cwd)
	}

	return m, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
