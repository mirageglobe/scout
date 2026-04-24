package filesystem

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

// OpenWithSystem opens the given path using the default system application.
func OpenWithSystem(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		// macOS
		cmd = exec.Command("open", path)
	case "linux":
		// Linux
		cmd = exec.Command("xdg-open", path)
	case "windows":
		// Windows
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// ReadDirContext runs ReadDir in a goroutine and returns early if ctx is cancelled.
// the underlying goroutine may outlive the context but its result is discarded on timeout.
func ReadDirContext(ctx context.Context, path string) ([]Entry, error) {
	type result struct {
		entries []Entry
		err     error
	}
	ch := make(chan result, 1)
	go func() {
		e, err := ReadDir(path)
		ch <- result{e, err}
	}()
	select {
	case r := <-ch:
		return r.entries, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// ReadDir reads the directory at path and returns a slice of sorted entries.
func ReadDir(path string) ([]Entry, error) {
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var entries []Entry
	for _, de := range dirEntries {
		info, _ := de.Info()
		subCount := 0
		if de.IsDir() {
			if subs, err := os.ReadDir(filepath.Join(path, de.Name())); err == nil {
				subCount = len(subs)
			}
		}
		entries = append(entries, Entry{
			Name:      de.Name(),
			IsDir:     de.IsDir(),
			IsSymlink: de.Type()&os.ModeSymlink != 0,
			Info:      info,
			SubCount:  subCount,
		})
	}

	// Sort: directories first, then alphabetical
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})

	return entries, nil
}

// GetStats returns a command that fetches memory and directory size stats.
// operations are bounded by a 5-second timeout to avoid blocking on slow mounts.
func GetStats(path string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)

		dirSize := int64(0)
		if entries, err := ReadDirContext(ctx, path); err == nil {
			for _, e := range entries {
				if !e.IsDir && e.Info != nil {
					dirSize += e.Info.Size()
				}
			}
		}

		cpu := 0.0
		if out, err := exec.CommandContext(ctx, "ps", "-p", strconv.Itoa(os.Getpid()), "-o", "%cpu=").Output(); err == nil {
			cpu, _ = strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
		}

		return StatsMsg{
			CPU:     cpu,
			Mem:     mem.Alloc,
			DirSize: dirSize,
		}
	}
}

// DoTick returns a command that sends a TickMsg every 2 seconds.
func DoTick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}
