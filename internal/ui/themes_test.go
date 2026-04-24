package ui

import (
	"fmt"
	"testing"
)

func TestThemeForHour(t *testing.T) {
	tests := []struct {
		hour      int
		wantTheme string
	}{
		{0, "Midnight"},
		{2, "Midnight"},
		{4, "Midnight"},
		{5, "Dawn"},
		{8, "Dawn"},
		{9, "Classic Amber"},
		{11, "Classic Amber"},
		{12, "Electric Cyan"},
		{16, "Electric Cyan"},
		{17, "Safety Orange"},
		{19, "Safety Orange"},
		{20, "Evening"},
		{23, "Evening"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("hour_%02d", tt.hour), func(t *testing.T) {
			idx := ThemeForHour(tt.hour)
			if got := Themes[idx].Name; got != tt.wantTheme {
				t.Errorf("ThemeForHour(%d) = index %d (%q), want %q", tt.hour, idx, got, tt.wantTheme)
			}
		})
	}
}
