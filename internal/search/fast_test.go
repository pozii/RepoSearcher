package search

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pozii/RepoSearcher/pkg/models"
)

func TestFastPattern_Literal(t *testing.T) {
	fp, err := CompileFastPattern("func", false, false)
	if err != nil {
		t.Fatal(err)
	}
	if !fp.isFullLiteral {
		t.Error("non-regex pattern should be full literal")
	}
	if !fp.MatchLine([]byte("func main()")) {
		t.Error("should match 'func main()'")
	}
	if fp.MatchLine([]byte("hello world")) {
		t.Error("should not match 'hello world'")
	}
}

func TestFastPattern_CaseInsensitive(t *testing.T) {
	fp, err := CompileFastPattern("func", false, true)
	if err != nil {
		t.Fatal(err)
	}
	if !fp.MatchLine([]byte("FUNC main()")) {
		t.Error("should match 'FUNC main()' case-insensitively")
	}
}

func TestFastPattern_RegexWithPrefix(t *testing.T) {
	fp, err := CompileFastPattern(`func\s+\w+`, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if string(fp.literal) != "func" {
		t.Errorf("expected prefix 'func', got %q", string(fp.literal))
	}
	if !fp.MatchLine([]byte("func main()")) {
		t.Error("should match 'func main()'")
	}
}

func TestFastPattern_PreScanFile(t *testing.T) {
	fp, _ := CompileFastPattern("func", false, false)

	data := []byte("package main\nfunc main() {}\n")
	if !fp.PreScanFile(data) {
		t.Error("should find 'func' in data")
	}

	noMatch := []byte("package main\nvar x = 1\n")
	if fp.PreScanFile(noMatch) {
		t.Error("should not find 'func' in data")
	}
}

func TestSearchLiteralBytes(t *testing.T) {
	data := []byte("line one\nfunc main()\nline three\nfunc helper()\n")
	literal := []byte("func")

	results := searchLiteralBytes(data, "test.go", literal)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].LineNumber != 2 {
		t.Errorf("first match at line 2, got %d", results[0].LineNumber)
	}
	if results[1].LineNumber != 4 {
		t.Errorf("second match at line 4, got %d", results[1].LineNumber)
	}
}

func TestSearchFileBytes_Literal(t *testing.T) {
	tmpDir := t.TempDir()
	content := "package main\nimport \"fmt\"\nfunc main() {\nfmt.Println(\"hello\")\n}\n"
	file := filepath.Join(tmpDir, "main.go")
	os.WriteFile(file, []byte(content), 0644)

	pattern, _ := CompileFastPattern("func", false, false)
	config := models.SearchConfig{}

	results := searchFileBytes(file, pattern, config)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].LineNumber != 3 {
		t.Errorf("match at line 3, got %d", results[0].LineNumber)
	}
}

func TestSearchFileBytes_EarlyExit(t *testing.T) {
	tmpDir := t.TempDir()
	content := "package main\nvar x = 1\nvar y = 2\n"
	file := filepath.Join(tmpDir, "main.go")
	os.WriteFile(file, []byte(content), 0644)

	pattern, _ := CompileFastPattern("func", false, false)
	config := models.SearchConfig{}

	results := searchFileBytes(file, pattern, config)
	if len(results) != 0 {
		t.Errorf("expected 0 results (early exit), got %d", len(results))
	}
}

func TestSearchFileBytes_RegexTwoPhase(t *testing.T) {
	tmpDir := t.TempDir()
	content := "package main\nfunc main() {}\nfunc helper() {}\nvar x = 1\n"
	file := filepath.Join(tmpDir, "main.go")
	os.WriteFile(file, []byte(content), 0644)

	pattern, _ := CompileFastPattern(`func\s+\w+`, true, false)
	config := models.SearchConfig{}

	results := searchFileBytes(file, pattern, config)
	if len(results) != 2 {
		t.Fatalf("expected 2 regex results, got %d", len(results))
	}
}

func TestCountNewlines(t *testing.T) {
	tests := []struct {
		data []byte
		want int
	}{
		{[]byte("hello\nworld\nfoo"), 2},
		{[]byte("no newlines"), 0},
		{[]byte(""), 0},
		{[]byte("\n\n\n"), 3},
		{[]byte("one\n"), 1},
	}
	for _, tt := range tests {
		got := countNewlines(tt.data)
		if got != tt.want {
			t.Errorf("countNewlines(%q) = %d, want %d", tt.data, got, tt.want)
		}
	}
}
