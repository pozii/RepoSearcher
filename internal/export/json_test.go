package export

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/pozii/RepoSearcher/pkg/models"
)

func TestToJSON(t *testing.T) {
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "results.json")

	results := []models.SearchResult{
		{FilePath: "main.go", LineNumber: 10, LineContent: "func main()", MatchText: "func"},
		{FilePath: "utils.go", LineNumber: 5, LineContent: "func helper()", MatchText: "func"},
	}

	if err := ToJSON(results, outFile); err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Read and verify
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var parsed models.SearchResults
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if parsed.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", parsed.TotalCount)
	}
	if len(parsed.Matches) != 2 {
		t.Errorf("Matches count = %d, want 2", len(parsed.Matches))
	}
	if parsed.Matches[0].FilePath != "main.go" {
		t.Errorf("First match FilePath = %q, want 'main.go'", parsed.Matches[0].FilePath)
	}
}

func TestToJSON_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "empty.json")

	if err := ToJSON(nil, outFile); err != nil {
		t.Fatalf("ToJSON with nil results failed: %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}

	var parsed models.SearchResults
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.TotalCount != 0 {
		t.Errorf("TotalCount = %d, want 0", parsed.TotalCount)
	}
}

func TestToJSON_InvalidPath(t *testing.T) {
	err := ToJSON(nil, "/nonexistent/path/results.json")
	if err == nil {
		t.Error("ToJSON should fail for invalid path")
	}
}
