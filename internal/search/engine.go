package search

import (
	"github.com/pozii/RepoSearcher/pkg/models"
)

// SearchEngine is the interface for all search implementations
type SearchEngine interface {
	Search(config models.SearchConfig) ([]models.SearchResult, error)
}
