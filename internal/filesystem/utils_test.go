package filesystem

import "testing"

func TestIsBinary(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{"empty slice", []byte{}, false},
		{"plain text", []byte("hello world\n"), false},
		{"null in middle", []byte("some\x00data"), true},
		{"null at start", []byte{0, 1, 2, 3}, true},
		{"null at end", []byte{1, 2, 3, 0}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBinary(tt.data); got != tt.want {
				t.Errorf("IsBinary(%q) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}

func TestHumanSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := HumanSize(tt.bytes); got != tt.want {
				t.Errorf("HumanSize(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		maxLen int
		want   string
	}{
		{"zero maxLen", "hello", 0, ""},
		{"negative maxLen", "hello", -1, ""},
		{"shorter than max", "hi", 10, "hi"},
		{"exact fit", "hello", 5, "hello"},
		{"truncated", "hello world", 5, "hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Truncate(tt.s, tt.maxLen); got != tt.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.s, tt.maxLen, got, tt.want)
			}
		})
	}
}
