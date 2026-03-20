package search

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"

	"github.com/pozii/RepoSearcher/internal/fileutil"
	"github.com/pozii/RepoSearcher/pkg/models"
)

// ParallelEngine is an optimized search engine with parallel processing
type ParallelEngine struct{}

// NewParallelEngine creates a new ParallelEngine instance
func NewParallelEngine() *ParallelEngine {
	return &ParallelEngine{}
}

// PatternCache caches compiled regex patterns
var patternCache sync.Map

// GetCompiledPattern returns a cached compiled pattern
func GetCompiledPattern(query string, isRegex, ignoreCase bool) (*regexp.Regexp, error) {
	cacheKey := query
	if isRegex {
		cacheKey += ":regex"
	}
	if ignoreCase {
		cacheKey += ":ignore"
	}

	if cached, ok := patternCache.Load(cacheKey); ok {
		return cached.(*regexp.Regexp), nil
	}

	compiled, err := CompilePattern(query, isRegex, ignoreCase)
	if err != nil {
		return nil, err
	}

	patternCache.Store(cacheKey, compiled)
	return compiled, nil
}

// Search performs a parallel search with performance optimizations
func (e *ParallelEngine) Search(config models.SearchConfig) ([]models.SearchResult, error) {
	var allResults []models.SearchResult
	err := e.SearchStream(config, func(r models.SearchResult) {
		allResults = append(allResults, r)
	})
	return allResults, err
}

// SearchStream performs a parallel search, calling callback for each result as found
func (e *ParallelEngine) SearchStream(config models.SearchConfig, callback func(models.SearchResult)) error {
	for _, root := range config.Paths {
		if err := e.searchPathStream(root, config, callback); err != nil {
			return err
		}
	}
	return nil
}

// searchPathStream searches a path and streams results via callback
func (e *ParallelEngine) searchPathStream(root string, config models.SearchConfig, callback func(models.SearchResult)) error {
	pattern, err := GetFastPattern(config.Query, config.IsRegex, config.IgnoreCase)
	if err != nil {
		return err
	}

	files, err := e.collectFiles(root, config)
	if err != nil {
		return err
	}

	numWorkers := runtime.NumCPU()
	if numWorkers > len(files) {
		numWorkers = len(files)
	}
	if numWorkers < 1 {
		numWorkers = 1
	}

	fileChan := make(chan string, len(files))
	resultChan := make(chan models.SearchResult, 256)

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range fileChan {
				results := searchFileBytes(filePath, pattern, config)
				for _, r := range results {
					resultChan <- r
				}
			}
		}()
	}

	go func() {
		for _, file := range files {
			fileChan <- file
		}
		close(fileChan)
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		callback(result)
	}

	return nil
}

// collectFiles collects all files to search using parallel directory walking
func (e *ParallelEngine) collectFiles(root string, config models.SearchConfig) ([]string, error) {
	var files []string

	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		if !ShouldIgnoreFile(root, config.Extensions) && !ShouldIgnoreFileByGlobs(root, config.IncludeGlobs, config.ExcludeGlobs) && !isBinaryFile(root) {
			files = append(files, root)
		}
		return files, nil
	}

	matcher := fileutil.NewGitIgnoreMatcher(root)

	shouldSkip := func(name, relPath string, isDir bool) bool {
		if fileutil.ShouldSkipDir(name) {
			return true
		}
		if matcher != nil && matcher.ShouldIgnore(relPath, isDir) {
			return true
		}
		return false
	}

	var mu sync.Mutex

	err = fileutil.ParallelWalk(root, shouldSkip, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}

		// Check .gitignore for files
		if matcher != nil {
			relPath, relErr := filepath.Rel(root, path)
			if relErr == nil && matcher.ShouldIgnore(relPath, false) {
				return nil
			}
		}

		if isBinaryFile(path) {
			return nil
		}

		if ShouldIgnoreFile(path, config.Extensions) {
			return nil
		}

		if ShouldIgnoreFileByGlobs(path, config.IncludeGlobs, config.ExcludeGlobs) {
			return nil
		}

		mu.Lock()
		files = append(files, path)
		mu.Unlock()
		return nil
	})

	return files, err
}

// searchFileBytes searches a file using []byte-based fast path.
// Uses os.ReadFile + bytes.Index for literal queries (SIMD-accelerated).
// Falls back to regex only for candidate lines with literal prefix matches.
func searchFileBytes(filePath string, pattern *FastPattern, config models.SearchConfig) []models.SearchResult {
	// Read entire file into memory — fastest for code-sized files (< 100MB)
	data, err := os.ReadFile(filePath)
	if err != nil || len(data) == 0 {
		return nil
	}

	// Early exit: pre-scan file for literal — skip entire file with one SIMD scan
	if pattern.HasLiteral() && !pattern.PreScanFile(data) {
		return nil
	}

	// Pure literal, case-sensitive — fastest possible path
	if pattern.isFullLiteral && !pattern.isCI {
		return searchLiteralBytes(data, filePath, pattern.literal)
	}

	// Regex or case-insensitive — []byte-based line scanning with regex
	return searchRegexBytes(data, filePath, pattern, config)
}

// searchLiteralBytes does pure literal matching on []byte — zero regex, zero string alloc for non-matches
func searchLiteralBytes(data []byte, filePath string, literal []byte) []models.SearchResult {
	var results []models.SearchResult
	lineNum := 0
	pos := 0

	for pos < len(data) {
		lineEnd := pos
		for lineEnd < len(data) && data[lineEnd] != '\n' {
			lineEnd++
		}
		lineNum++

		if bytes.Index(data[pos:lineEnd], literal) >= 0 {
			results = append(results, models.SearchResult{
				FilePath:    filePath,
				LineNumber:  lineNum,
				LineContent: string(data[pos:lineEnd]),
				MatchText:   string(literal),
			})
		}

		pos = lineEnd + 1
	}

	return results
}

// searchRegexBytes does regex matching on []byte with optional two-phase literal pre-scan
func searchRegexBytes(data []byte, filePath string, pattern *FastPattern, config models.SearchConfig) []models.SearchResult {
	// Two-phase search: scan for literal prefix first, then regex only on candidates
	if len(pattern.regexPrefix) > 0 {
		return searchTwoPhaseBytes(data, filePath, pattern, config)
	}

	// No literal prefix — line-by-line regex on []byte
	var results []models.SearchResult
	lineNum := 0
	pos := 0

	for pos < len(data) {
		lineEnd := pos
		for lineEnd < len(data) && data[lineEnd] != '\n' {
			lineEnd++
		}
		lineNum++
		line := data[pos:lineEnd]

		if pattern.regex.Match(line) {
			matchText := pattern.FindMatch(line)
			results = append(results, models.SearchResult{
				FilePath:    filePath,
				LineNumber:  lineNum,
				LineContent: string(line),
				MatchText:   string(matchText),
			})
		}

		pos = lineEnd + 1
	}

	return results
}

// searchTwoPhaseBytes uses literal prefix scan + regex on candidate lines
func searchTwoPhaseBytes(data []byte, filePath string, pattern *FastPattern, config models.SearchConfig) []models.SearchResult {
	var results []models.SearchResult
	lineNum := 0
	pos := 0

	searchPrefix := pattern.regexPrefix
	searchData := data
	if pattern.isCI {
		searchData = bytes.ToLower(data)
	}

	for pos < len(data) {
		// Find next literal prefix match
		relIdx := bytes.Index(searchData[pos:], searchPrefix)
		if relIdx < 0 {
			break
		}
		matchPos := pos + relIdx

		// Find line boundaries
		lineStart := matchPos
		for lineStart > 0 && data[lineStart-1] != '\n' {
			lineStart--
		}
		lineEnd := matchPos
		for lineEnd < len(data) && data[lineEnd] != '\n' {
			lineEnd++
		}

		// Count line number
		lineNum += countNewlines(data[pos:lineStart])
		lineNum++
		pos = lineEnd + 1

		line := data[lineStart:lineEnd]

		// Run regex only on this candidate line
		if pattern.regex.Match(line) {
			matchText := pattern.FindMatch(line)
			results = append(results, models.SearchResult{
				FilePath:    filePath,
				LineNumber:  lineNum,
				LineContent: string(line),
				MatchText:   string(matchText),
			})
		}
	}

	return results
}

// countNewlines counts newline bytes in data
func countNewlines(data []byte) int {
	return bytes.Count(data, []byte{'\n'})
}

// isBinaryFile checks if a file is binary (fast check)
func isBinaryFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		return false
	}

	// Check for null bytes (binary indicator)
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true
		}
	}

	return false
}
