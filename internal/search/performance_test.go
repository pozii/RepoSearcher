package search

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pozii/RepoSearcher/pkg/models"
)

func TestGetCompiledPattern_Cache(t *testing.T) {
	// First call should compile
	p1, err := GetCompiledPattern("hello", false, false)
	if err != nil {
		t.Fatalf("GetCompiledPattern failed: %v", err)
	}

	// Second call with same args should return cached
	p2, err := GetCompiledPattern("hello", false, false)
	if err != nil {
		t.Fatalf("GetCompiledPattern failed: %v", err)
	}

	// Should be the same pointer (cached)
	if p1 != p2 {
		t.Error("GetCompiledPattern should return cached pattern")
	}

	// Different flags should compile new pattern
	p3, err := GetCompiledPattern("hello", false, true)
	if err != nil {
		t.Fatalf("GetCompiledPattern failed: %v", err)
	}
	if p1 == p3 {
		t.Error("Different flags should produce different patterns")
	}
}

func TestIsBinaryFile(t *testing.T) {
	// Create a temporary text file
	tmpDir := t.TempDir()
	textFile := filepath.Join(tmpDir, "text.txt")
	if err := os.WriteFile(textFile, []byte("hello world\nfoo bar\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if isBinaryFile(textFile) {
		t.Error("text file should not be detected as binary")
	}

	// Create a binary file (with null bytes)
	binFile := filepath.Join(tmpDir, "binary.bin")
	data := []byte{0x00, 0x01, 0x02, 0x03, 0x00, 0xFF}
	if err := os.WriteFile(binFile, data, 0644); err != nil {
		t.Fatal(err)
	}

	if !isBinaryFile(binFile) {
		t.Error("file with null bytes should be detected as binary")
	}
}

func TestIsBinaryFile_NonExistent(t *testing.T) {
	if isBinaryFile("/nonexistent/path/file.txt") {
		t.Error("nonexistent file should return false")
	}
}

func TestParallelEngine_Search(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()

	// Write test files
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "utils.go"), []byte("package main\n\nfunc helper() error {\n\treturn nil\n}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# Hello World\nThis is a test.\n"), 0644); err != nil {
		t.Fatal(err)
	}

	engine := NewParallelEngine()
	config := models.SearchConfig{
		Query:   "func",
		Paths:   []string{tmpDir},
		IsRegex: false,
	}

	results, err := engine.Search(config)
	if err != nil {
		t.Fatalf("ParallelEngine.Search failed: %v", err)
	}

	// Should find "func" in main.go and utils.go, not in readme.md
	if len(results) < 2 {
		t.Errorf("Expected at least 2 results, got %d", len(results))
	}

	for _, r := range results {
		if r.LineNumber == 0 {
			t.Error("LineNumber should not be 0")
		}
		if r.LineContent == "" {
			t.Error("LineContent should not be empty")
		}
	}
}

func TestParallelEngine_SearchWithExtensions(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\nfunc test() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "app.py"), []byte("def test():\n    pass\n"), 0644); err != nil {
		t.Fatal(err)
	}

	engine := NewParallelEngine()

	// Search only .go files
	config := models.SearchConfig{
		Query:      "test",
		Paths:      []string{tmpDir},
		Extensions: []string{".go"},
	}

	results, err := engine.Search(config)
	if err != nil {
		t.Fatalf("ParallelEngine.Search failed: %v", err)
	}

	for _, r := range results {
		if filepath.Ext(r.FilePath) != ".go" {
			t.Errorf("Expected .go file, got %s", r.FilePath)
		}
	}
}

func TestParallelEngine_SearchRegex(t *testing.T) {
	tmpDir := t.TempDir()

	content := "func main() {}\nfunc test(a int) string {}\nfunc helper() {}\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "code.go"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	engine := NewParallelEngine()
	config := models.SearchConfig{
		Query:   `func\s+\w+\(`,
		Paths:   []string{tmpDir},
		IsRegex: true,
	}

	results, err := engine.Search(config)
	if err != nil {
		t.Fatalf("ParallelEngine.Search regex failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 regex matches, got %d", len(results))
	}
}

func TestParallelEngine_SearchIgnoreCase(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "code.go"), []byte("func HELLO() {}\nfunc hello() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	engine := NewParallelEngine()
	config := models.SearchConfig{
		Query:      "hello",
		Paths:      []string{tmpDir},
		IgnoreCase: true,
	}

	results, err := engine.Search(config)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 case-insensitive matches, got %d", len(results))
	}
}

func TestCollectFiles(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "a.go"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "b.py"), []byte("test"), 0644)
	os.Mkdir(filepath.Join(tmpDir, ".git"), 0755)
	os.WriteFile(filepath.Join(tmpDir, ".git", "config"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "binary.bin"), []byte{0x00, 0x01}, 0644)

	engine := NewParallelEngine()
	files, err := engine.collectFiles(tmpDir, models.SearchConfig{})
	if err != nil {
		t.Fatalf("collectFiles failed: %v", err)
	}

	// Should have a.go and b.py, not .git/config or binary.bin
	for _, f := range files {
		base := filepath.Base(f)
		if base == "config" {
			t.Error("Should not collect files from .git directory")
		}
		if base == "binary.bin" {
			t.Error("Should not collect binary files")
		}
	}
}
