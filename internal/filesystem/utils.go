package filesystem

import (
	"fmt"
	"github.com/charmbracelet/x/ansi"
	"charm.land/lipgloss/v2"
)

// IsBinary attempts to detect binary content by checking for null bytes.
func IsBinary(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}

// HumanSize formats a byte count into a human-readable string.
func HumanSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Truncate cuts a string to maxLen characters physically, appending "…" if truncated.
func Truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	return ansi.Truncate(s, maxLen, "")
}

// VisibleLen returns the approximate visible length of a string,
// stripping ANSI escape sequences.
func VisibleLen(s string) int {
	return lipgloss.Width(s)
}
