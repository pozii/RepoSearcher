package search

import (
	"bytes"
	"regexp"
	"strings"
	"sync"
)

// FastPattern holds a pre-compiled search pattern optimized for fast matching.
// For literal queries, it uses bytes.Index (SIMD-accelerated) instead of regexp.
type FastPattern struct {
	literal       []byte         // extracted literal string
	isFullLiteral bool           // entire pattern is this literal (non-regex query)
	isCI          bool           // case-insensitive
	regex         *regexp.Regexp // fallback regex (nil for pure literals)
	regexPrefix   []byte         // extracted from regex.LiteralPrefix()
}

// CompileFastPattern compiles a pattern with literal extraction optimization
func CompileFastPattern(query string, isRegex, ignoreCase bool) (*FastPattern, error) {
	fp := &FastPattern{isCI: ignoreCase}

	if !isRegex {
		// Pure literal — no regex needed at all
		fp.literal = []byte(query)
		fp.isFullLiteral = true
		if ignoreCase {
			// Still need regex fallback for case-insensitive matching
			re, err := CompilePattern(query, false, true)
			if err != nil {
				return nil, err
			}
			fp.regex = re
			// Try to get prefix for two-phase search
			prefix, complete := re.LiteralPrefix()
			if prefix != "" {
				fp.regexPrefix = []byte(strings.ToLower(prefix))
				fp.isFullLiteral = complete
			}
		}
		return fp, nil
	}

	// Regex query — compile and try to extract literal prefix
	re, err := CompilePattern(query, true, ignoreCase)
	if err != nil {
		return nil, err
	}
	fp.regex = re

	// Extract prefix for two-phase search
	prefix, complete := re.LiteralPrefix()
	if prefix != "" {
		fp.literal = []byte(prefix)
		fp.isFullLiteral = complete
		if ignoreCase {
			fp.regexPrefix = []byte(strings.ToLower(prefix))
		} else {
			fp.regexPrefix = fp.literal
		}
	}

	return fp, nil
}

// MatchLine checks if a line matches the pattern using the fastest method available
func (fp *FastPattern) MatchLine(line []byte) bool {
	if fp.isFullLiteral && !fp.isCI {
		// Pure literal, case-sensitive — fastest path (SIMD)
		return bytes.Contains(line, fp.literal)
	}
	if fp.isFullLiteral && fp.isCI {
		// Pure literal, case-insensitive — use regex (already compiled with (?i))
		return fp.regex.Match(line)
	}
	// Regex — use compiled regex
	return fp.regex.Match(line)
}

// FindMatch returns the matching text for the line
func (fp *FastPattern) FindMatch(line []byte) []byte {
	if fp.isFullLiteral && !fp.isCI {
		idx := bytes.Index(line, fp.literal)
		if idx >= 0 {
			return line[idx : idx+len(fp.literal)]
		}
		return nil
	}
	return fp.regex.Find(line)
}

// HasLiteral returns true if a literal prefix was extracted for pre-scanning
func (fp *FastPattern) HasLiteral() bool {
	return len(fp.literal) > 0 || len(fp.regexPrefix) > 0
}

// ScanLiteral performs a fast SIMD scan of data for the literal prefix.
// Returns the byte offset of the first match, or -1 if not found.
func (fp *FastPattern) ScanLiteral(data []byte) int {
	if fp.isFullLiteral && !fp.isCI {
		return bytes.Index(data, fp.literal)
	}
	if len(fp.regexPrefix) > 0 {
		return bytes.Index(data, fp.regexPrefix)
	}
	return 0 // no literal to scan for, must run regex on everything
}

// PreScanFile quickly checks if the file contains the literal anywhere.
// Returns true if the literal is found (file is worth searching), false to skip.
func (fp *FastPattern) PreScanFile(data []byte) bool {
	if fp.isFullLiteral && !fp.isCI {
		return bytes.Index(data, fp.literal) >= 0
	}
	if len(fp.regexPrefix) > 0 {
		searchData := data
		if fp.isCI {
			searchData = bytes.ToLower(data)
		}
		return bytes.Index(searchData, fp.regexPrefix) >= 0
	}
	return true // no literal to pre-scan, must search
}

// fastPatternCache caches compiled FastPatterns
var fastPatternCache sync.Map

// GetFastPattern returns a cached compiled FastPattern
func GetFastPattern(query string, isRegex, ignoreCase bool) (*FastPattern, error) {
	cacheKey := query
	if isRegex {
		cacheKey += ":regex"
	}
	if ignoreCase {
		cacheKey += ":ignore"
	}
	cacheKey += ":fast"

	if cached, ok := fastPatternCache.Load(cacheKey); ok {
		return cached.(*FastPattern), nil
	}

	compiled, err := CompileFastPattern(query, isRegex, ignoreCase)
	if err != nil {
		return nil, err
	}

	fastPatternCache.Store(cacheKey, compiled)
	return compiled, nil
}
