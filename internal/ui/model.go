package ui

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/mirageglobe/scout/internal/filesystem"
	"github.com/mirageglobe/scout/internal/git"
)

// SpinnerTickMsg drives the loading animation frame.
type SpinnerTickMsg struct{}

// HintIdleTickMsg fires after the idle timeout; Seq guards against stale ticks.
type HintIdleTickMsg struct{ Seq int }

// HintTipTickMsg advances the rotating hint bar tip during a cycling run.
type HintTipTickMsg struct{}

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
	Loading              bool   // true while a LoadDir command is in-flight
	SpinnerFrame         int    // current animation frame (0-2) for the loading indicator
	PendingCursor        string // entry name to restore cursor to after next DirLoadedMsg
	HintCycling          bool   // true while hint bar is cycling through tips
	HintTipIdx           int    // current tip index during a cycling run
	HintIdleSeq          int    // incremented on every keypress to cancel stale idle ticks
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
		RootFocus:  true,
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
		DoHintIdleTick(0),
	)
}

// DoSpinnerTick returns a command that fires SpinnerTickMsg after 200ms.
func DoSpinnerTick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(time.Time) tea.Msg {
		return SpinnerTickMsg{}
	})
}

// DoHintIdleTick fires HintIdleTickMsg after 10s of inactivity.
// Seq is compared on arrival; a mismatched seq means a keypress cancelled this timer.
func DoHintIdleTick(seq int) tea.Cmd {
	return tea.Tick(60*time.Second, func(time.Time) tea.Msg {
		return HintIdleTickMsg{Seq: seq}
	})
}

// DoHintTipTick fires HintTipTickMsg after 5s, advancing the cycling tip.
func DoHintTipTick() tea.Cmd {
	return tea.Tick(5*time.Second, func(time.Time) tea.Msg {
		return HintTipTickMsg{}
	})
}

// RefreshGit is a command that re-fetches git status and branch for the current directory.
// bounded by a 5-second timeout to avoid blocking on slow or hung git repos.
func (m Model) RefreshGit() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return filesystem.GitRefreshMsg{
			GitStatus: git.GetStatus(ctx, m.Cwd),
			GitBranch: git.GetBranch(ctx, m.Cwd),
		}
	}
}

// WatchDir polls a directory for changes in the background, returning DirWatchMsg
// so the handler can update entries without resetting cursor or preview scroll.
// bounded by a 5-second timeout to avoid goroutine pile-up on slow mounts.
func (m Model) WatchDir(path string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		entries, err := filesystem.ReadDirContext(ctx, path)
		if err != nil {
			return filesystem.DirWatchMsg{Err: err}
		}
		return filesystem.DirWatchMsg{
			Entries:   entries,
			GitStatus: git.GetStatus(ctx, path),
			GitBranch: git.GetBranch(ctx, path),
		}
	}
}

// LoadDir is a command that reads a directory and its git status.
// bounded by a 10-second timeout to surface errors on unresponsive paths.
func (m Model) LoadDir(path string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		entries, err := filesystem.ReadDirContext(ctx, path)
		if err != nil {
			return filesystem.DirLoadedMsg{Err: err}
		}
		return filesystem.DirLoadedMsg{
			Entries:   entries,
			GitStatus: git.GetStatus(ctx, path),
			GitBranch: git.GetBranch(ctx, path),
		}
	}
}
