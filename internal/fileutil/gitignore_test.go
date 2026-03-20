package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGitIgnoreMatcher_Basic(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte("*.exe\nbuild/\n# comment\n!important.exe\n"), 0644)

	matcher := NewGitIgnoreMatcher(tmpDir)
	if matcher == nil {
		t.Fatal("NewGitIgnoreMatcher should not return nil when .gitignore exists")
	}

	tests := []struct {
		path   string
		isDir  bool
		ignore bool
	}{
		{"foo.exe", false, true},
		{"important.exe", false, false}, // negated
		{"build", true, true},           // directory match
		{"build", false, false},         // dir-only pattern, not a dir
		{"main.go", false, false},
	}

	for _, tt := range tests {
		got := matcher.ShouldIgnore(tt.path, tt.isDir)
		if got != tt.ignore {
			t.Errorf("ShouldIgnore(%q, isDir=%v) = %v, want %v", tt.path, tt.isDir, got, tt.ignore)
		}
	}
}

func TestGitIgnoreMatcher_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	matcher := NewGitIgnoreMatcher(tmpDir)
	if matcher != nil {
		t.Error("NewGitIgnoreMatcher should return nil when no .gitignore exists")
	}
}

func TestGitIgnoreMatcher_NilSafe(t *testing.T) {
	var matcher *GitIgnoreMatcher
	got := matcher.ShouldIgnore("anything", false)
	if got {
		t.Error("nil matcher should return false")
	}
}

func TestGitIgnoreMatcher_Comments(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte("# This is a comment\n\n# Another comment\n*.log\n"), 0644)

	matcher := NewGitIgnoreMatcher(tmpDir)
	if matcher == nil {
		t.Fatal("matcher should not be nil")
	}

	if !matcher.ShouldIgnore("debug.log", false) {
		t.Error("should ignore *.log files")
	}
	if matcher.ShouldIgnore("main.go", false) {
		t.Error("should not ignore main.go")
	}
}

func TestGitIgnoreMatcher_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .gitignore
	os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte("*.tmp\nbuild/\n"), 0644)

	// Create files
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "data.tmp"), []byte("test"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "build"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "build", "output.go"), []byte("test"), 0644)

	// Test matcher
	matcher := NewGitIgnoreMatcher(tmpDir)
	if matcher == nil {
		t.Fatal("matcher should not be nil")
	}

	// data.tmp should be ignored
	if !matcher.ShouldIgnore("data.tmp", false) {
		t.Error("data.tmp should be ignored")
	}

	// build/ directory should be ignored
	if !matcher.ShouldIgnore("build", true) {
		t.Error("build/ directory should be ignored")
	}

	// main.go should NOT be ignored
	if matcher.ShouldIgnore("main.go", false) {
		t.Error("main.go should not be ignored")
	}
}
