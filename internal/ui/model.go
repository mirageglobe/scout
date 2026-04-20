package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/mirageglobe/scout/internal/filesystem"
	"github.com/mirageglobe/scout/internal/git"
)

// EditorFinishedMsg is sent when the external editor (vim) exits.
type EditorFinishedMsg struct{ Err error }

// Model represents the state of the Scout TUI.
type Model struct {
	Cwd           string
	Entries       []filesystem.Entry
	Cursor        int
	Width         int
	Height        int
	Preview       string
	PreviewScroll int
	FocusRight    bool
	ShowHelp      bool
	ThemeIdx      int
	GitStatus     map[string]string
	GitBranch     string
	Stats         filesystem.Stats
	StatusMsg     string
	Err           error
}

// NewModel initializes a fresh UI model.
func NewModel(cwd string) Model {
	return Model{
		Cwd: cwd,
	}
}

// Init initializes the application by loading the starting directory and stats.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.LoadDir(m.Cwd),
		filesystem.DoTick(),
		filesystem.GetStats(m.Cwd),
	)
}

// LoadDir is a command that reads a directory and its git status.
func (m Model) LoadDir(path string) tea.Cmd {
	return func() tea.Msg {
		entries, err := filesystem.ReadDir(path)
		if err != nil {
			return filesystem.DirLoadedMsg{Err: err}
		}
		status := git.GetStatus(path)
		branch := git.GetBranch(path)

		return filesystem.DirLoadedMsg{
			Entries:   entries,
			GitStatus: status,
			GitBranch: branch,
		}
	}
}
