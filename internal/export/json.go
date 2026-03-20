package export

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pozii/RepoSearcher/pkg/models"
)

// ToJSON exports search results to a JSON file
func ToJSON(results []models.SearchResult, filePath string) error {
	data := models.SearchResults{
		Matches:    results,
		TotalCount: len(results),
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}

	return nil
}
