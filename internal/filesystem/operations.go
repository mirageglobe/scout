package filesystem

import (
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
			Name:     de.Name(),
			IsDir:    de.IsDir(),
			Info:     info,
			SubCount: subCount,
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
func GetStats(path string) tea.Cmd {
	return func() tea.Msg {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		dirSize := int64(0)
		entries, _ := os.ReadDir(path)
		for _, e := range entries {
			if !e.IsDir() {
				info, _ := e.Info()
				if info != nil {
					dirSize += info.Size()
				}
			}
		}

		cpu := 0.0
		if out, err := exec.Command("ps", "-p", strconv.Itoa(os.Getpid()), "-o", "%cpu=").Output(); err == nil {
			cpu, _ = strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
		}

		return StatsMsg{
			CPU:     cpu,
			Mem:     m.Alloc,
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
