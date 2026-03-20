package search

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/gobwas/glob"
)

// ShouldIgnoreFile checks if a file should be excluded from search
func ShouldIgnoreFile(path string, extensions []string) bool {
	if len(extensions) == 0 {
		return false
	}

	for _, ext := range extensions {
		if strings.HasSuffix(strings.ToLower(path), strings.ToLower(ext)) {
			return false
		}
	}
	return true
}

// CompilePattern compiles a regex pattern with optional case-insensitive flag
func CompilePattern(query string, isRegex, ignoreCase bool) (*regexp.Regexp, error) {
	if isRegex {
		if ignoreCase {
			return regexp.Compile("(?i)" + query)
		}
		return regexp.Compile(query)
	}

	escaped := regexp.QuoteMeta(query)
	if ignoreCase {
		escaped = "(?i)" + escaped
	}
	return regexp.Compile(escaped)
}

// globCache caches compiled glob patterns
var globCache sync.Map

func compileGlob(pattern string) (glob.Glob, error) {
	if cached, ok := globCache.Load(pattern); ok {
		return cached.(glob.Glob), nil
	}
	compiled, err := glob.Compile(pattern)
	if err != nil {
		return nil, err
	}
	globCache.Store(pattern, compiled)
	return compiled, nil
}

// ShouldIgnoreFileByGlobs checks if a file path matches any exclude glob
// or fails to match any include glob.
func ShouldIgnoreFileByGlobs(path string, includeGlobs, excludeGlobs []string) bool {
	// Normalize path separators for glob matching
	normalized := strings.ReplaceAll(path, string(os.PathSeparator), "/")

	// Check excludes first
	for _, pattern := range excludeGlobs {
		g, err := compileGlob(pattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: invalid glob pattern %q: %v\n", pattern, err)
			continue
		}
		if g.Match(normalized) {
			return true
		}
	}

	// Check includes (if specified, file must match at least one)
	if len(includeGlobs) > 0 {
		matched := false
		for _, pattern := range includeGlobs {
			g, err := compileGlob(pattern)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: invalid glob pattern %q: %v\n", pattern, err)
				continue
			}
			if g.Match(normalized) {
				matched = true
				break
			}
		}
		if !matched {
			return true
		}
	}

	return false
}
