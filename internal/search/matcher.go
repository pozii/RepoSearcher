package search

import (
	"regexp"
	"strings"
)

// Match performs text matching based on configuration
func Match(line, query string, isRegex, ignoreCase bool) ([]string, error) {
	if isRegex {
		return matchRegex(line, query, ignoreCase)
	}
	return matchKeyword(line, query, ignoreCase), nil
}

// matchRegex finds all regex matches in a line
func matchRegex(line, pattern string, ignoreCase bool) ([]string, error) {
	if ignoreCase {
		pattern = "(?i)" + pattern
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	matches := re.FindAllString(line, -1)
	if matches == nil {
		return []string{}, nil
	}
	return matches, nil
}

// matchKeyword finds keyword occurrences in a line
func matchKeyword(line, query string, ignoreCase bool) []string {
	if ignoreCase {
		line = strings.ToLower(line)
		query = strings.ToLower(query)
	}

	var matches []string
	for i := 0; i <= len(line)-len(query); i++ {
		if line[i:i+len(query)] == query {
			matches = append(matches, query)
		}
	}
	return matches
}

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
