package fileutil

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestParallelWalk_Basic(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory tree
	os.WriteFile(filepath.Join(tmpDir, "root.txt"), []byte("test"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "sub1"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "sub1", "a.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "sub1", "b.txt"), []byte("test"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "sub2"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "sub2", "c.txt"), []byte("test"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "sub2", "deep"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "sub2", "deep", "d.txt"), []byte("test"), 0644)

	var collected []string
	err := ParallelWalk(tmpDir, func(name, relPath string, isDir bool) bool {
		return false // don't skip anything
	}, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() {
			rel, _ := filepath.Rel(tmpDir, path)
			collected = append(collected, filepath.ToSlash(rel))
		}
		return nil
	})

	if err != nil {
		t.Fatalf("ParallelWalk failed: %v", err)
	}

	sort.Strings(collected)
	expected := []string{"root.txt", "sub1/a.txt", "sub1/b.txt", "sub2/c.txt", "sub2/deep/d.txt"}

	if len(collected) != len(expected) {
		t.Fatalf("Expected %d files, got %d: %v", len(expected), len(collected), collected)
	}

	for i, exp := range expected {
		if collected[i] != exp {
			t.Errorf("File[%d] = %q, want %q", i, collected[i], exp)
		}
	}
}

func TestParallelWalk_WithSkipDir(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "root.txt"), []byte("test"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "skip_me"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "skip_me", "hidden.txt"), []byte("test"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "keep"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "keep", "visible.txt"), []byte("test"), 0644)

	var collected []string
	err := ParallelWalk(tmpDir, func(name, relPath string, isDir bool) bool {
		return name == "skip_me"
	}, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() {
			rel, _ := filepath.Rel(tmpDir, path)
			collected = append(collected, filepath.ToSlash(rel))
		}
		return nil
	})

	if err != nil {
		t.Fatalf("ParallelWalk failed: %v", err)
	}

	for _, f := range collected {
		if f == "skip_me/hidden.txt" {
			t.Error("Should not collect files from skipped directory")
		}
	}

	if len(collected) != 2 {
		t.Errorf("Expected 2 files, got %d: %v", len(collected), collected)
	}
}

func TestParallelWalk_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "single.txt")
	os.WriteFile(file, []byte("test"), 0644)

	var collected []string
	err := ParallelWalk(file, func(name, relPath string, isDir bool) bool {
		return false
	}, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		collected = append(collected, path)
		return nil
	})

	if err != nil {
		t.Fatalf("ParallelWalk failed: %v", err)
	}

	if len(collected) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(collected))
	}
}
