package ui

import (
	"strings"
	"testing"

	"github.com/mirageglobe/scout/internal/filesystem"
)

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
