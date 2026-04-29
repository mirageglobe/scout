package ui

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mirageglobe/scout/internal/filesystem"
)

// mockFileInfo satisfies os.FileInfo with a configurable ModTime.
type mockFileInfo struct{ modTime time.Time }

func (f mockFileInfo) Name() string      { return "" }
func (f mockFileInfo) Size() int64       { return 0 }
func (f mockFileInfo) Mode() os.FileMode { return 0 }
func (f mockFileInfo) ModTime() time.Time { return f.modTime }
func (f mockFileInfo) IsDir() bool       { return false }
func (f mockFileInfo) Sys() any          { return nil }

func TestComputeSearchMatches(t *testing.T) {
	preview := "hello world\nfoo bar\nHELLO again"
	tests := []struct {
		query string
		want  []int
	}{
		{"hello", []int{0, 2}}, // case-insensitive match across lines
		{"foo", []int{1}},
		{"xyz", nil},
		{"", nil},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got := computeSearchMatches(preview, tt.query)
			if len(got) != len(tt.want) {
				t.Errorf("computeSearchMatches(%q) = %v, want %v", tt.query, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("computeSearchMatches(%q)[%d] = %d, want %d", tt.query, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestChangedFileCount(t *testing.T) {
	tests := []struct {
		name      string
		gitStatus map[string]string
		want      int
	}{
		{"empty", map[string]string{}, 0},
		{"one modified", map[string]string{"foo.go": "M"}, 1},
		{"two changes", map[string]string{"a.go": "M", "b.go": "?"}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := changedFileCount(tt.gitStatus); got != tt.want {
				t.Errorf("changedFileCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDirEntriesChanged(t *testing.T) {
	base := []filesystem.Entry{
		{Name: "foo", IsDir: true},
		{Name: "bar.txt"},
	}
	same := []filesystem.Entry{
		{Name: "foo", IsDir: true},
		{Name: "bar.txt"},
	}
	renamed := []filesystem.Entry{
		{Name: "foo", IsDir: true},
		{Name: "baz.txt"},
	}
	typeChanged := []filesystem.Entry{
		{Name: "foo", IsDir: false},
		{Name: "bar.txt"},
	}

	if dirEntriesChanged(base, same) {
		t.Error("identical slices reported as changed")
	}
	if !dirEntriesChanged(base, renamed) {
		t.Error("renamed entry not detected as changed")
	}
	if !dirEntriesChanged(base, typeChanged) {
		t.Error("IsDir change not detected")
	}
	if !dirEntriesChanged(base, base[:1]) {
		t.Error("different lengths not detected as changed")
	}

	t0 := time.Now().Add(-time.Minute)
	t1 := time.Now()
	withModtime := []filesystem.Entry{
		{Name: "foo", IsDir: true, Info: mockFileInfo{modTime: t0}},
		{Name: "bar.txt", Info: mockFileInfo{modTime: t0}},
	}
	modtimeChanged := []filesystem.Entry{
		{Name: "foo", IsDir: true, Info: mockFileInfo{modTime: t0}},
		{Name: "bar.txt", Info: mockFileInfo{modTime: t1}},
	}
	if dirEntriesChanged(withModtime, withModtime) {
		t.Error("same modtimes reported as changed")
	}
	if !dirEntriesChanged(withModtime, modtimeChanged) {
		t.Error("modtime change not detected")
	}
}

func TestDirWatchMsgRebuildsPreviewOnModtimeChange(t *testing.T) {
	t0 := time.Now().Add(-time.Minute)
	t1 := time.Now()

	m := Model{
		Entries: []filesystem.Entry{
			{Name: "file.txt", Info: mockFileInfo{modTime: t0}},
		},
		Preview: "stale",
	}

	msg := filesystem.DirWatchMsg{
		Entries: []filesystem.Entry{
			{Name: "file.txt", Info: mockFileInfo{modTime: t1}},
		},
	}
	updated, _ := m.Update(msg)
	um := updated.(Model)

	if um.Preview == "stale" {
		t.Error("preview was not rebuilt after modtime change in DirWatchMsg")
	}
}

func TestClampedScrollFor(t *testing.T) {
	// 30 lines of content, Height=20 → contentHeight=15, maxScroll=15
	m := Model{
		Height:  20,
		Preview: strings.Repeat("line\n", 30),
	}

	tests := []struct {
		name    string
		lineIdx int
		want    int
	}{
		{"top of file", 0, 0},
		{"beyond max scroll", 29, 15},
		{"centred mid-file", 15, 8}, // 15 - 15/2 = 8
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clampedScrollFor(m, tt.lineIdx); got != tt.want {
				t.Errorf("clampedScrollFor(line %d) = %d, want %d", tt.lineIdx, got, tt.want)
			}
		})
	}
}
