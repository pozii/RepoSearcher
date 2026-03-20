package search

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pozii/RepoSearcher/internal/fileutil"
	"github.com/pozii/RepoSearcher/pkg/models"
)

// LocalEngine implements SearchEngine for local filesystem search
type LocalEngine struct{}

// NewLocalEngine creates a new LocalEngine instance
func NewLocalEngine() *LocalEngine {
	return &LocalEngine{}
}

// Search performs a search across local filesystem paths
func (e *LocalEngine) Search(config models.SearchConfig) ([]models.SearchResult, error) {
	pattern, err := CompilePattern(config.Query, config.IsRegex, config.IgnoreCase)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern: %w", err)
	}

	var allResults []models.SearchResult

	for _, root := range config.Paths {
		results, err := e.searchPath(root, pattern, config)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// searchPath recursively searches through a directory or file
func (e *LocalEngine) searchPath(root string, pattern *regexp.Regexp, config models.SearchConfig) ([]models.SearchResult, error) {
	var results []models.SearchResult

	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("cannot access path %s: %w", root, err)
	}

	if !info.IsDir() {
		return e.searchFile(root, pattern, config)
	}

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if fileutil.ShouldSkipDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		if ShouldIgnoreFile(path, config.Extensions) {
			return nil
		}

		fileResults, err := e.searchFile(path, pattern, config)
		if err != nil {
			return nil
		}

		results = append(results, fileResults...)
		return nil
	})

	return results, err
}

// searchFile searches within a single file
func (e *LocalEngine) searchFile(filePath string, pattern *regexp.Regexp, config models.SearchConfig) ([]models.SearchResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read all lines for context support
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	var results []models.SearchResult
	for lineIdx, line := range lines {
		lineNum := lineIdx + 1

		if pattern.MatchString(line) {
			content := line
			if config.Context > 0 {
				content = buildContextContent(lines, lineIdx, config.Context)
			}

			results = append(results, models.SearchResult{
				FilePath:    filePath,
				LineNumber:  lineNum,
				LineContent: content,
				MatchText:   pattern.FindString(line),
			})
		}
	}

	return results, nil
}

// buildContextContent builds a content string with surrounding context lines
func buildContextContent(lines []string, matchIdx, contextLines int) string {
	start := matchIdx - contextLines
	if start < 0 {
		start = 0
	}
	end := matchIdx + contextLines + 1
	if end > len(lines) {
		end = len(lines)
	}

	var sb strings.Builder
	for i := start; i < end; i++ {
		prefix := "  "
		if i == matchIdx {
			prefix = "> "
		}
		sb.WriteString(prefix)
		sb.WriteString(lines[i])
		if i < end-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}
