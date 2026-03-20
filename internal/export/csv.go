package export

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/pozii/RepoSearcher/pkg/models"
)

// ToCSV exports search results to a CSV file
func ToCSV(results []models.SearchResult, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"file_path", "line_number", "line_content", "match_text"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("writing header: %w", err)
	}

	// Write rows
	for _, r := range results {
		row := []string{
			r.FilePath,
			strconv.Itoa(r.LineNumber),
			r.LineContent,
			r.MatchText,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("writing row: %w", err)
		}
	}

	return nil
}
