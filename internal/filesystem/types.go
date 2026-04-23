package filesystem

import "os"

// Entry represents a single file or directory in the listing.
type Entry struct {
	Name      string
	IsDir     bool
	IsSymlink bool
	Info      os.FileInfo
	SubCount  int // number of items inside (dirs only)
}

// Stats represents the current resource usage and directory metadata.
type Stats struct {
	CPU     float64
	Mem     uint64
	DirSize int64
}

// DirLoadedMsg carries the result of loading a directory.
type DirLoadedMsg struct {
	Entries    []Entry
	GitStatus  map[string]string
	GitBranch  string
	Err        error
}

// StatsMsg carries system statistics.
type StatsMsg struct {
	CPU     float64
	Mem     uint64
	DirSize int64
}

// GitRefreshMsg carries updated git status and branch for the current directory.
type GitRefreshMsg struct {
	GitStatus map[string]string
	GitBranch string
}

// DirWatchMsg carries the result of a background directory poll.
// Unlike DirLoadedMsg, it is handled without resetting cursor or scroll state.
type DirWatchMsg struct {
	Entries   []Entry
	GitStatus map[string]string
	GitBranch string
	Err       error
}

// TickMsg is sent periodically to refresh stats.
type TickMsg struct{}
