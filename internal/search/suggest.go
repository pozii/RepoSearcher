package search

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pozii/RepoSearcher/pkg/models"
)

// SuggestEngine provides smart suggestions based on code patterns
type SuggestEngine struct{}

// NewSuggestEngine creates a new SuggestEngine
func NewSuggestEngine() *SuggestEngine {
	return &SuggestEngine{}
}

// Suggest provides suggestions for a query
func (e *SuggestEngine) Suggest(config models.SearchConfig) ([]Suggestion, error) {
	var allSuggestions []Suggestion

	for _, root := range config.Paths {
		suggestions, err := e.suggestPath(root, config)
		if err != nil {
			return nil, err
		}
		allSuggestions = append(allSuggestions, suggestions...)
	}

	// Deduplicate and sort
	return deduplicateSuggestions(allSuggestions), nil
}

// suggestPath generates suggestions from a single path
func (e *SuggestEngine) suggestPath(root string, config models.SearchConfig) ([]Suggestion, error) {
	// Extract all identifiers from code files
	identifiers, err := extractIdentifiers(root, config.Extensions)
	if err != nil {
		return nil, err
	}

	// Find fuzzy matches
	results := FuzzyFindAll(config.Query, identifiers, 0.5)

	var suggestions []Suggestion
	for _, r := range results {
		suggestions = append(suggestions, Suggestion{
			Text:    r.Text,
			Score:   r.Score,
			Context: "identifier",
		})
	}

	// Extract functions with context
	functions, err := extractFunctions(root, config.Extensions)
	if err == nil {
		for _, fn := range functions {
			score := FuzzyScore(config.Query, fn.Name)
			if score >= 0.4 {
				suggestions = append(suggestions, Suggestion{
					Text:    fn.Name,
					Score:   score * 0.9,
					Context: "function",
					Details: fn.Context,
				})
			}
		}
	}

	// Sort by score
	for i := 0; i < len(suggestions); i++ {
		for j := i + 1; j < len(suggestions); j++ {
			if suggestions[j].Score > suggestions[i].Score {
				suggestions[i], suggestions[j] = suggestions[j], suggestions[i]
			}
		}
	}

	// Limit results
	if len(suggestions) > 10 {
		suggestions = suggestions[:10]
	}

	return suggestions, nil
}

// Suggestion represents a search suggestion
type Suggestion struct {
	Text    string
	Score   float64
	Context string // "identifier", "function", "type", "variable"
	Details string // Additional context like function signature
}

// Function represents an extracted function
type Function struct {
	Name    string
	Context string
	Line    int
}

// extractIdentifiers extracts all identifiers from code files
func extractIdentifiers(root string, extensions []string) ([]string, error) {
	identifiers := make(map[string]bool)

	err := walkFiles(root, extensions, func(path string) error {
		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()

			// Extract identifiers (camelCase, PascalCase, snake_case)
			words := regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*`).FindAllString(line, -1)
			for _, w := range words {
				if len(w) > 2 && !isKeyword(w) {
					identifiers[w] = true
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(identifiers))
	for id := range identifiers {
		result = append(result, id)
	}

	return result, nil
}

// extractFunctions extracts function names with their signatures
func extractFunctions(root string, extensions []string) ([]Function, error) {
	var functions []Function

	goFuncPattern := regexp.MustCompile(`func\s+(\w+)\s*(\([^)]*\))`)
	goMethodPattern := regexp.MustCompile(`func\s+\([^)]+\)\s+(\w+)\s*(\([^)]*\))`)

	err := walkFiles(root, extensions, func(path string) error {
		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			// Match Go functions
			if match := goFuncPattern.FindStringSubmatch(line); match != nil {
				functions = append(functions, Function{
					Name:    match[1],
					Context: match[2],
					Line:    lineNum,
				})
			}

			// Match Go methods
			if match := goMethodPattern.FindStringSubmatch(line); match != nil {
				functions = append(functions, Function{
					Name:    match[1],
					Context: match[2],
					Line:    lineNum,
				})
			}
		}

		return nil
	})

	return functions, err
}

// deduplicateSuggestions removes duplicate suggestions
func deduplicateSuggestions(suggestions []Suggestion) []Suggestion {
	seen := make(map[string]bool)
	var result []Suggestion

	for _, s := range suggestions {
		if !seen[s.Text] {
			seen[s.Text] = true
			result = append(result, s)
		}
	}

	return result
}

// isKeyword checks if a string is a reserved keyword
func isKeyword(s string) bool {
	keywords := map[string]bool{
		"if": true, "else": true, "for": true, "return": true,
		"func": true, "var": true, "const": true, "type": true,
		"package": true, "import": true, "nil": true, "true": true,
		"false": true, "int": true, "string": true, "bool": true,
		"float64": true, "error": true, "interface": true, "struct": true,
		"map": true, "chan": true, "go": true, "defer": true,
		"switch": true, "case": true, "default": true, "break": true,
		"continue": true, "select": true, "range": true,
	}
	return keywords[strings.ToLower(s)]
}

// walkFiles walks through files and applies a function
func walkFiles(root string, extensions []string, fn func(path string) error) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if shouldSkipDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		if ShouldIgnoreFile(path, extensions) {
			return nil
		}

		return fn(path)
	})
}
