package fileutil

import (
	"testing"
)

func TestShouldSkipDir(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{".git", true},
		{".svn", true},
		{".hg", true},
		{"node_modules", true},
		{"vendor", true},
		{".vscode", true},
		{".idea", true},
		{"__pycache__", true},
		{".venv", true},
		{".hidden", true},
		{".DS_Store", true},
		// Should NOT skip
		{"src", false},
		{"cmd", false},
		{"internal", false},
		{"pkg", false},
		{"main", false},
		{"testdata", false},
	}

	for _, tt := range tests {
		got := ShouldSkipDir(tt.name)
		if got != tt.want {
			t.Errorf("ShouldSkipDir(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}
