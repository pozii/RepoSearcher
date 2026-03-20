package export

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"

	"github.com/pozii/RepoSearcher/pkg/models"
)

func TestToCSV(t *testing.T) {
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "results.csv")

	results := []models.SearchResult{
		{FilePath: "main.go", LineNumber: 10, LineContent: "func main()", MatchText: "func"},
		{FilePath: "utils.go", LineNumber: 5, LineContent: "func helper()", MatchText: "func"},
	}

	if err := ToCSV(results, outFile); err != nil {
		t.Fatalf("ToCSV failed: %v", err)
	}

	// Read and verify
	file, err := os.Open(outFile)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	// Header + 2 data rows
	if len(records) != 3 {
		t.Errorf("Expected 3 records (header + 2 rows), got %d", len(records))
	}

	// Check header
	if records[0][0] != "file_path" {
		t.Errorf("Header[0] = %q, want 'file_path'", records[0][0])
	}

	// Check first data row
	if records[1][0] != "main.go" {
		t.Errorf("Row[0][0] = %q, want 'main.go'", records[1][0])
	}
	if records[1][1] != "10" {
		t.Errorf("Row[0][1] = %q, want '10'", records[1][1])
	}
}

func TestToCSV_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "empty.csv")

	if err := ToCSV(nil, outFile); err != nil {
		t.Fatalf("ToCSV with nil results failed: %v", err)
	}

	file, err := os.Open(outFile)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	// Only header
	if len(records) != 1 {
		t.Errorf("Expected 1 record (header only), got %d", len(records))
	}
}
