package filesystem

import "os"

// Entry represents a single file or directory in the listing.
type Entry struct {
	Name     string
	IsDir    bool
	Info     os.FileInfo
	SubCount int // number of items inside (dirs only)
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

// TickMsg is sent periodically to refresh stats.
type TickMsg struct{}
