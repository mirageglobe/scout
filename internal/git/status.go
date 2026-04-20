package git

import (
	"os/exec"
	"strings"
)

// GetBranch returns the current git branch name for dir, or "" if not a repo.
func GetBranch(dir string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// GetStatus runs "git status --porcelain" in the given directory and
// returns a map of filename -> status code.
func GetStatus(dir string) map[string]string {
	result := make(map[string]string)

	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		// Not a git repo or git not available
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

		// Strip leading path components to match entries in this directory
		if parts := strings.SplitN(name, "/", 2); len(parts) > 1 {
			name = parts[0]
		}

		if xy == "??" {
			result[name] = "?"
		} else {
			// Use the first non-space character as status
			status := string(xy[0])
			if status == " " && len(xy) > 1 {
				status = string(xy[1])
			}
			result[name] = status
		}
	}

	return result
}
