package git

import (
	"context"
	"os/exec"
	"strings"
)

// GetBranch returns the current git branch name for dir, or "" if not a repo.
func GetBranch(ctx context.Context, dir string) string {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// GetStatus runs "git status --porcelain" in the given directory and
// returns a map of filename -> status code, relative to dir.
func GetStatus(ctx context.Context, dir string) map[string]string {
	result := make(map[string]string)

	// git reports paths relative to the repo root regardless of cmd.Dir,
	// so we need the prefix of dir within the repo to strip it.
	prefixCmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-prefix")
	prefixCmd.Dir = dir
	prefixOut, err := prefixCmd.Output()
	if err != nil {
		return result
	}
	prefix := strings.TrimSpace(string(prefixOut)) // e.g. "internal/filesystem/"

	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return result
	}

	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if len(line) < 4 {
			continue
		}
		// Porcelain format: XY filename
		xy := strings.TrimSpace(line[:2])
		name := strings.TrimSpace(line[3:])

		// Handle renamed files: "R  old -> new"
		if idx := strings.Index(name, " -> "); idx >= 0 {
			name = name[idx+4:]
		}

		// Strip the repo-root prefix to make the path relative to dir
		if prefix != "" {
			if !strings.HasPrefix(name, prefix) {
				continue // file is outside this directory
			}
			name = name[len(prefix):]
		}

		// Strip subdirectory components to match top-level entries in dir
		if parts := strings.SplitN(name, "/", 2); len(parts) > 1 {
			name = parts[0]
		}

		if xy == "??" {
			result[name] = "?"
		} else {
			status := string(xy[0])
			if status == " " && len(xy) > 1 {
				status = string(xy[1])
			}
			result[name] = status
		}
	}

	return result
}
