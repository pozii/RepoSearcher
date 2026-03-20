package search

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
	"sync/atomic"

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

	for _, root := range config.Paths {
		results, err := e.searchPathParallel(root, config)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// searchPathParallel searches a path using parallel goroutines
func (e *ParallelEngine) searchPathParallel(root string, config models.SearchConfig) ([]models.SearchResult, error) {
	// Compile pattern once (cached)
	pattern, err := GetCompiledPattern(config.Query, config.IsRegex, config.IgnoreCase)
	if err != nil {
		return nil, err
	}

	// Collect all files first
	files, err := e.collectFiles(root, config.Extensions)
	if err != nil {
		return nil, err
	}

	// Determine number of workers
	numWorkers := runtime.NumCPU()
	if numWorkers > len(files) {
		numWorkers = len(files)
	}
	if numWorkers < 1 {
		numWorkers = 1
	}

	// Parallel search with channels
	fileChan := make(chan string, len(files))
	resultChan := make(chan []models.SearchResult, len(files))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range fileChan {
				results := e.searchFileFast(filePath, pattern, config)
				if len(results) > 0 {
					resultChan <- results
				}
			}
		}()
	}

	// Send files to workers
	go func() {
		for _, file := range files {
			fileChan <- file
		}
		close(fileChan)
	}()

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var results []models.SearchResult
	for batch := range resultChan {
		results = append(results, batch...)
	}

	return results, nil
}

// collectFiles collects all files to search
func (e *ParallelEngine) collectFiles(root string, extensions []string) ([]string, error) {
	var files []string

	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		if !ShouldIgnoreFile(root, extensions) && !isBinaryFile(root) {
			files = append(files, root)
		}
		return files, nil
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

		// Fast binary check
		if isBinaryFile(path) {
			return nil
		}

		if ShouldIgnoreFile(path, extensions) {
			return nil
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

// searchFileFast searches a file with optimizations
func (e *ParallelEngine) searchFileFast(filePath string, pattern *regexp.Regexp, config models.SearchConfig) []models.SearchResult {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	var results []models.SearchResult
	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Use a larger buffer for faster scanning
	scanner.Buffer(make([]byte, 64*1024), 64*1024)

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

	return results
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

// PerformanceStats tracks search performance
type PerformanceStats struct {
	FilesScanned   int64
	FilesSkipped   int64
	MatchesFound   int64
	BytesScanned   int64
	GoroutinesUsed int
}

var stats struct {
	filesScanned atomic.Int64
	filesSkipped atomic.Int64
	matchesFound atomic.Int64
	bytesScanned atomic.Int64
	mu           sync.Mutex
}

// ResetStats resets the performance stats
func ResetStats() {
	stats.filesScanned.Store(0)
	stats.filesSkipped.Store(0)
	stats.matchesFound.Store(0)
	stats.bytesScanned.Store(0)
}

// GetStats returns the current performance stats
func GetStats() PerformanceStats {
	return PerformanceStats{
		FilesScanned:   stats.filesScanned.Load(),
		FilesSkipped:   stats.filesSkipped.Load(),
		MatchesFound:   stats.matchesFound.Load(),
		BytesScanned:   stats.bytesScanned.Load(),
		GoroutinesUsed: runtime.NumCPU(),
	}
}

// IncrementScanned increments files scanned counter
func IncrementScanned() {
	stats.filesScanned.Add(1)
}

// IncrementSkipped increments files skipped counter
func IncrementSkipped() {
	stats.filesSkipped.Add(1)
}

// IncrementMatches increments matches found counter
func IncrementMatches() {
	stats.matchesFound.Add(1)
}
