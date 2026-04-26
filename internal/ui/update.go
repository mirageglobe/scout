package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/mirageglobe/scout/internal/filesystem"
)

// Update is the bubbletea entry point. It delegates to handleMsg and ensures
// the hint idle timer is always reset after a keypress.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	newModel, cmd := m.handleMsg(msg)
	if _, ok := msg.(tea.KeyPressMsg); ok {
		nm := newModel.(Model)
		return nm, tea.Batch(cmd, DoHintIdleTick(nm.HintIdleSeq))
	}
	return newModel, cmd
}

// handleMsg handles all state transitions in response to messages.
func (m Model) handleMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case filesystem.GitRefreshMsg:
		m.GitStatus = msg.GitStatus
		m.GitBranch = msg.GitBranch
		return m, nil

	case filesystem.TickMsg:
		return m, tea.Batch(filesystem.DoTick(), filesystem.GetStats(m.Cwd), m.RefreshGit(), m.WatchDir(m.Cwd))

	case filesystem.DirWatchMsg:
		if msg.Err != nil {
			return m, nil
		}
		entries := msg.Entries
		if !m.ShowHidden {
			filtered := entries[:0]
			for _, e := range entries {
				if !strings.HasPrefix(e.Name, ".") {
					filtered = append(filtered, e)
				}
			}
			entries = filtered
		}
		m.GitStatus = msg.GitStatus
		m.GitBranch = msg.GitBranch
		if dirEntriesChanged(m.Entries, entries) {
			m.Entries = entries
			if m.Cursor >= len(m.Entries) {
				m.Cursor = max(0, len(m.Entries)-1)
			}
			m.Preview = m.BuildPreview()
		}
		return m, nil

	case filesystem.StatsMsg:
		m.Stats.CPU = msg.CPU
		m.Stats.Mem = msg.Mem
		m.Stats.DirSize = msg.DirSize
		return m, nil

	case SpinnerTickMsg:
		if m.Loading {
			m.SpinnerFrame = (m.SpinnerFrame + 1) % 3
			return m, DoSpinnerTick()
		}
		return m, nil

	case HintIdleTickMsg:
		if msg.Seq != m.HintIdleSeq {
			return m, nil // stale tick, a keypress already cancelled it
		}
		m.HintCycling = true
		m.HintTipIdx = 0
		return m, DoHintTipTick()

	case HintTipTickMsg:
		if !m.HintCycling {
			return m, nil
		}
		m.HintTipIdx++
		if m.HintTipIdx >= len(HintTips) {
			// one full cycle complete — return to normal bar
			m.HintCycling = false
			m.HintTipIdx = 0
			return m, nil
		}
		return m, DoHintTipTick()

	case filesystem.DirLoadedMsg:
		m.Loading = false
		if msg.Err != nil {
			m.StatusMsg = fmt.Sprintf("[error] %v", msg.Err)
			return m, nil
		}
		entries := msg.Entries
		if !m.ShowHidden {
			filtered := entries[:0]
			for _, e := range entries {
				if !strings.HasPrefix(e.Name, ".") {
					filtered = append(filtered, e)
				}
			}
			entries = filtered
		}
		m.Entries = entries
		m.GitStatus = msg.GitStatus
		m.GitBranch = msg.GitBranch
		m.Err = nil
		m.StatusMsg = ""
		m.PreviewScroll = 0
		if m.PendingCursor != "" {
			m.Cursor = 0
			for i, e := range m.Entries {
				if e.Name == m.PendingCursor {
					m.Cursor = i
					break
				}
			}
			m.PendingCursor = ""
		} else if m.Cursor >= len(m.Entries) {
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
		// cancel any active tip cycling; wrapper re-arms the idle timer
		m.HintCycling = false
		m.HintTipIdx = 0
		m.HintIdleSeq++

		if m.ShowHelp {
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			m.ShowHelp = false
			return m, nil
		}

		// explorer search input mode: intercept keys until enter or escape
		if m.ExplorerSearchActive {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				m.ExplorerSearchActive = false
			case "esc":
				m = clearExplorerSearch(m)
			case "backspace":
				if len(m.ExplorerSearchInput) > 0 {
					_, size := utf8.DecodeLastRuneInString(m.ExplorerSearchInput)
					m.ExplorerSearchInput = m.ExplorerSearchInput[:len(m.ExplorerSearchInput)-size]
					if filtered := m.explorerFiltered(); len(filtered) > 0 {
						m.Cursor = filtered[0]
						m.Preview = m.BuildPreview()
					}
				}
			default:
				if utf8.RuneCountInString(msg.String()) == 1 {
					m.ExplorerSearchInput += msg.String()
					if filtered := m.explorerFiltered(); len(filtered) > 0 {
						m.Cursor = filtered[0]
						m.Preview = m.BuildPreview()
					}
				}
			}
			return m, nil
		}

		// search input mode: intercept all keys until enter or escape
		if m.SearchActive {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				m.SearchActive = false
			case "esc":
				m = clearSearch(m)
			case "backspace":
				if len(m.SearchInput) > 0 {
					_, size := utf8.DecodeLastRuneInString(m.SearchInput)
					m.SearchInput = m.SearchInput[:len(m.SearchInput)-size]
					m.SearchQuery = m.SearchInput
					m.SearchMatches = computeSearchMatches(m.Preview, m.SearchQuery)
					m.SearchMatchIdx = 0
					if len(m.SearchMatches) > 0 {
						m.PreviewScroll = clampedScrollFor(m, m.SearchMatches[0])
					}
				}
			default:
				if utf8.RuneCountInString(msg.String()) == 1 {
					m.SearchInput += msg.String()
					m.SearchQuery = m.SearchInput
					m.SearchMatches = computeSearchMatches(m.Preview, m.SearchQuery)
					m.SearchMatchIdx = 0
					if len(m.SearchMatches) > 0 {
						m.PreviewScroll = clampedScrollFor(m, m.SearchMatches[0])
					}
				}
			}
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
						m.StatusMsg = fmt.Sprintf("[error] %v", err)
					} else {
						m.StatusMsg = fmt.Sprintf("[ok] opened: %s", selected.Name)
					}
				}
			}
			return m, nil

		case "r":
			if m.FocusRight {
				m.Preview = m.BuildPreview()
				m.StatusMsg = "[ok] preview refreshed"
			}
			return m, nil

		case "t":
			m.ThemeIdx = (m.ThemeIdx + 1) % len(Themes)
			filesystem.SaveConfig(filesystem.Config{ThemeIdx: m.ThemeIdx})
			m.Preview = m.BuildPreview()
			return m, nil

		case "i":
			m.ShowHidden = !m.ShowHidden
			m.Cursor = 0
			m, cmd := startLoading(m)
			return m, tea.Batch(m.LoadDir(m.Cwd), cmd)

		case "f":
			m.RootFocus = !m.RootFocus
			if m.RootFocus {
				m.StatusMsg = "[info] root focus enabled"
			} else {
				m.StatusMsg = "[info] root focus disabled"
			}
			return m, nil

		case "tab":
			m.ExplorerCollapsed = !m.ExplorerCollapsed
			return m, nil

		// search/find: "/" activates in whichever pane is focused
		case "/":
			if len(m.Entries) > 0 {
				if m.FocusRight {
					m.SearchActive = true
					m.SearchInput = ""
				} else {
					m.ExplorerSearchActive = true
					m.ExplorerSearchInput = ""
				}
			}
			return m, nil

		// n / N: navigate matches in whichever pane is active
		case "n":
			if m.FocusRight && len(m.SearchMatches) > 0 {
				m.SearchMatchIdx = (m.SearchMatchIdx + 1) % len(m.SearchMatches)
				m.PreviewScroll = clampedScrollFor(m, m.SearchMatches[m.SearchMatchIdx])
			} else if !m.FocusRight && m.ExplorerSearchInput != "" {
				filtered := m.explorerFiltered()
				for pos, idx := range filtered {
					if idx == m.Cursor && pos < len(filtered)-1 {
						m.Cursor = filtered[pos+1]
						m.PreviewScroll = 0
						m.Preview = m.BuildPreview()
						break
					}
				}
			}
			return m, nil

		case "N", "shift+n":
			if m.FocusRight && len(m.SearchMatches) > 0 {
				m.SearchMatchIdx = (m.SearchMatchIdx - 1 + len(m.SearchMatches)) % len(m.SearchMatches)
				m.PreviewScroll = clampedScrollFor(m, m.SearchMatches[m.SearchMatchIdx])
			} else if !m.FocusRight && m.ExplorerSearchInput != "" {
				filtered := m.explorerFiltered()
				for pos, idx := range filtered {
					if idx == m.Cursor && pos > 0 {
						m.Cursor = filtered[pos-1]
						m.PreviewScroll = 0
						m.Preview = m.BuildPreview()
						break
					}
				}
			}
			return m, nil

		// search: clear highlights
		case "esc":
			m = clearSearch(m)
			m = clearExplorerSearch(m)
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
				m = clearSearch(m)
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
				m = clearSearch(m)
			}
			return m, nil

		// Navigation: go to parent directory or unfocus right pane
		case "h", "left", "backspace":
			if m.FocusRight {
				m.FocusRight = false
				m = clearSearch(m)
				return m, nil
			}
			parent := filepath.Dir(m.Cwd)
			if parent != m.Cwd && !(m.RootFocus && m.Cwd == m.RootPath) {
				m.PendingCursor = filepath.Base(m.Cwd)
				m.Cwd = parent
				m.PreviewScroll = 0
				m.Preview = ""
				m = clearSearch(m)
				m = clearExplorerSearch(m)
				m, cmd := startLoading(m)
				return m, tea.Batch(m.LoadDir(m.Cwd), cmd)
			}
			return m, nil

		// Action: open directory or file
		case "e":
			if len(m.Entries) == 0 {
				return m, nil
			}
			selected := m.Entries[m.Cursor]
			fullPath := filepath.Join(m.Cwd, selected.Name)

			if selected.IsDir {
				return m, nil
			}

			// Security check before launching Editor
			f, _ := os.Open(fullPath)
			if f != nil {
				buf := make([]byte, 1024)
				n, _ := f.Read(buf)
				f.Close()
				if filesystem.IsBinary(buf[:n]) {
					m.StatusMsg = fmt.Sprintf("[error] cannot open binary file: %s", selected.Name)
					return m, nil
				}
			}

			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vim"
			}
			m.StatusMsg = fmt.Sprintf("[info] opening with %s", editor)
			c := exec.Command(editor, fullPath)
			return m, tea.ExecProcess(c, func(err error) tea.Msg {
				return EditorFinishedMsg{Err: err}
			})

		case "enter", "l", "right":
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
				m = clearSearch(m)
				m = clearExplorerSearch(m)
				m, cmd := startLoading(m)
				return m, tea.Batch(m.LoadDir(m.Cwd), cmd)
			}
			if !m.FocusRight {
				m.FocusRight = true
				m = clearExplorerSearch(m)
			}
			return m, nil

		case "g":
			m.Cursor = 0
			m.PreviewScroll = 0
			m.Preview = m.BuildPreview()
			m.StatusMsg = ""
			m = clearSearch(m)
			return m, nil

		case "G", "shift+g":
			if len(m.Entries) > 0 {
				m.Cursor = len(m.Entries) - 1
			}
			m.PreviewScroll = 0
			m.Preview = m.BuildPreview()
			m.StatusMsg = ""
			m = clearSearch(m)
			return m, nil
		}

	case EditorFinishedMsg:
		if msg.Err != nil {
			m.StatusMsg = fmt.Sprintf("[error] %v", msg.Err)
		}
		m, cmd := startLoading(m)
		return m, tea.Batch(m.LoadDir(m.Cwd), cmd)
	}

	return m, nil
}

// stripANSI removes ANSI/CSI escape sequences for plain-text matching.
func stripANSI(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			// skip parameter bytes (0x30-0x3F) and intermediate bytes (0x20-0x2F)
			for i < len(s) && (s[i] < 0x40 || s[i] > 0x7E) {
				i++
			}
			if i < len(s) {
				i++ // skip final byte (e.g. 'm')
			}
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	return result.String()
}

// computeSearchMatches returns the line indices in preview that contain query (case-insensitive).
func computeSearchMatches(preview, query string) []int {
	if query == "" {
		return nil
	}
	lines := strings.Split(strings.TrimSuffix(preview, "\n"), "\n")
	lowerQuery := strings.ToLower(query)
	var matches []int
	for i, line := range lines {
		if strings.Contains(strings.ToLower(stripANSI(line)), lowerQuery) {
			matches = append(matches, i)
		}
	}
	return matches
}

// clampedScrollFor returns a PreviewScroll value that centres lineIdx in the viewport.
func clampedScrollFor(m Model, lineIdx int) int {
	contentHeight := m.Height - 5
	if contentHeight < 1 {
		contentHeight = 1
	}
	previewLines := strings.Split(strings.TrimSuffix(m.Preview, "\n"), "\n")
	maxScroll := len(previewLines) - contentHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	scroll := lineIdx - contentHeight/2
	if scroll < 0 {
		scroll = 0
	}
	if scroll > maxScroll {
		scroll = maxScroll
	}
	return scroll
}

// explorerFiltered returns the m.Entries indices whose names match ExplorerSearchInput.
// Returns nil when ExplorerSearchInput is empty (meaning: show all entries).
func (m Model) explorerFiltered() []int {
	if m.ExplorerSearchInput == "" {
		return nil
	}
	query := strings.ToLower(m.ExplorerSearchInput)
	var result []int
	for i, e := range m.Entries {
		if strings.Contains(strings.ToLower(e.Name), query) {
			result = append(result, i)
		}
	}
	return result
}

// clearExplorerSearch resets all explorer search state on the model.
func clearExplorerSearch(m Model) Model {
	m.ExplorerSearchActive = false
	m.ExplorerSearchInput = ""
	return m
}

// clearSearch resets all search state on the model.
func clearSearch(m Model) Model {
	m.SearchActive = false
	m.SearchQuery = ""
	m.SearchInput = ""
	m.SearchMatches = nil
	m.SearchMatchIdx = 0
	return m
}

// startLoading sets Loading state and returns the spinner tick command.
func startLoading(m Model) (Model, tea.Cmd) {
	m.Loading = true
	m.SpinnerFrame = 0
	m.StatusMsg = ""
	return m, DoSpinnerTick()
}

// dirEntriesChanged returns true if the entry list has changed by name, type, count, or modification time.
func dirEntriesChanged(a, b []filesystem.Entry) bool {
	if len(a) != len(b) {
		return true
	}
	for i := range a {
		if a[i].Name != b[i].Name || a[i].IsDir != b[i].IsDir || a[i].SubCount != b[i].SubCount {
			return true
		}
		if a[i].Info != nil && b[i].Info != nil && !a[i].Info.ModTime().Equal(b[i].Info.ModTime()) {
			return true
		}
	}
	return false
}
