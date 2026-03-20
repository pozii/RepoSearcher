package search

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
			if shouldSkipDir(info.Name()) {
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

	var results []models.SearchResult
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if pattern.MatchString(line) {
			results = append(results, models.SearchResult{
				FilePath:    filePath,
				LineNumber:  lineNum,
				LineContent: line,
				MatchText:   pattern.FindString(line),
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// shouldSkipDir checks if a directory should be skipped during search
func shouldSkipDir(name string) bool {
	skipDirs := map[string]bool{
		".git": true, ".svn": true, ".hg": true,
		"node_modules": true, "vendor": true,
		".vscode": true, ".idea": true,
	}
	return skipDirs[name] || strings.HasPrefix(name, ".")
}
