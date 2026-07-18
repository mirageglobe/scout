package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mirageglobe/scout/internal/filesystem"
)

// After enter commits an explorer search with multiple matches, n/N must still step
// the cursor between those matches (regression: enter used to clear the query that
// n/N keys off, leaving stepping dead).
func TestExplorerSearchNextPrevAfterCommit(t *testing.T) {
	entries := []filesystem.Entry{
		{Name: "alpha.go"},  // 0 matches "go"
		{Name: "beta.go"},   // 1 matches "go"
		{Name: "gamma.txt"}, // 2 no "go"
		{Name: "server.go"}, // 3 matches "go"
	}
	m := Model{Cwd: "/tmp", Entries: entries, Cursor: 0, ThemeIdx: 0, Sym: selectGlyphs(), ExplorerSearchActive: true}

	// type the query "go" (live filtering jumps the cursor to the first match)
	for _, r := range "go" {
		u, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = u.(Model)
	}
	if m.Cursor != 0 {
		t.Fatalf("pre-enter cursor = %d, want 0 (first match)", m.Cursor)
	}

	// enter commits: query retained, input/active cleared, cursor unchanged
	u, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = u.(Model)
	if m.ExplorerSearchActive || m.ExplorerSearchInput != "" {
		t.Fatalf("after enter: active=%v input=%q, want inactive + empty", m.ExplorerSearchActive, m.ExplorerSearchInput)
	}
	if m.ExplorerSearchQuery != "go" {
		t.Fatalf("after enter: query=%q, want go", m.ExplorerSearchQuery)
	}

	step := func(key rune, want int) {
		t.Helper()
		u, _ := m.Update(tea.KeyPressMsg{Code: key, Text: string(key)})
		m = u.(Model)
		if m.Cursor != want {
			t.Errorf("after %q: cursor = %d (%s), want %d", string(key), m.Cursor, entries[m.Cursor].Name, want)
		}
	}
	step('n', 1) // beta.go
	step('n', 3) // server.go
	step('N', 1) // back to beta.go
}
