package ui

import (
	"time"

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
	ShowHidden      bool
	ExplorerCollapsed bool
	Stats         filesystem.Stats
	StatusMsg     string
	Err           error
	SearchActive   bool   // "/" mode active, user is typing a query
	SearchQuery    string // committed search term
	SearchInput    string // in-progress buffer while SearchActive
	SearchMatches  []int  // preview line indices that contain the query
	SearchMatchIdx int    // index into SearchMatches for current match
	ExplorerSearchActive bool   // "\" mode active in file explorer
	ExplorerSearchInput  string // current explorer search input
	RootFocus            bool   // restrict navigation to RootPath
	RootPath             string // the starting directory path
}

// NewModel initializes a fresh UI model with a time-based theme (or saved config).
func NewModel(cwd string) Model {
	themeIdx := ThemeForHour(time.Now().Hour())
	if cfg, err := filesystem.LoadConfig(); err == nil {
		themeIdx = cfg.ThemeIdx
	}
	return Model{
		Cwd:        cwd,
		RootPath:   cwd,
		ThemeIdx:   themeIdx,
		ShowHidden: true,
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

// RefreshGit is a command that re-fetches git status and branch for the current directory.
func (m Model) RefreshGit() tea.Cmd {
	return func() tea.Msg {
		return filesystem.GitRefreshMsg{
			GitStatus: git.GetStatus(m.Cwd),
			GitBranch: git.GetBranch(m.Cwd),
		}
	}
}

// WatchDir polls a directory for changes in the background, returning DirWatchMsg
// so the handler can update entries without resetting cursor or preview scroll.
func (m Model) WatchDir(path string) tea.Cmd {
	return func() tea.Msg {
		entries, err := filesystem.ReadDir(path)
		if err != nil {
			return filesystem.DirWatchMsg{Err: err}
		}
		status := git.GetStatus(path)
		branch := git.GetBranch(path)
		return filesystem.DirWatchMsg{
			Entries:   entries,
			GitStatus: status,
			GitBranch: branch,
		}
	}
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
