package fileutil

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

// gitIgnoreRule represents a single parsed .gitignore pattern
type gitIgnoreRule struct {
	pattern  string
	negate   bool
	dirOnly  bool
	compiled []glob.Glob // multiple compiled patterns for flexible matching
}

// GitIgnoreMatcher holds rules from .gitignore files
type GitIgnoreMatcher struct {
	rules []gitIgnoreRule
}

// NewGitIgnoreMatcher reads and parses .gitignore from root directory.
// Returns nil if no .gitignore exists.
func NewGitIgnoreMatcher(root string) *GitIgnoreMatcher {
	ignoreFile := filepath.Join(root, ".gitignore")
	rules := parseGitIgnoreFile(ignoreFile)
	if len(rules) == 0 {
		return nil
	}
	return &GitIgnoreMatcher{rules: rules}
}

// parseGitIgnoreFile reads a .gitignore file and returns parsed rules
func parseGitIgnoreFile(path string) []gitIgnoreRule {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	var rules []gitIgnoreRule
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		rule := gitIgnoreRule{}

		if strings.HasPrefix(line, "!") {
			rule.negate = true
			line = line[1:]
		}

		if strings.HasSuffix(line, "/") {
			rule.dirOnly = true
			line = strings.TrimSuffix(line, "/")
		}

		compiled := compilePatternVariants(line)
		if len(compiled) == 0 {
			continue
		}

		rule.pattern = line
		rule.compiled = compiled
		rules = append(rules, rule)
	}

	return rules
}

// compilePatternVariants compiles multiple variants of a .gitignore pattern
// to handle both basename and full-path matching
func compilePatternVariants(pattern string) []glob.Glob {
	anchored := strings.HasPrefix(pattern, "/")
	if anchored {
		pattern = strings.TrimPrefix(pattern, "/")
	}

	hasSlash := strings.Contains(pattern, "/")
	var variants []string

	if hasSlash || anchored {
		variants = append(variants, pattern)
	} else {
		variants = append(variants, pattern)       // direct basename match
		variants = append(variants, "**/"+pattern) // nested match
	}

	var compiled []glob.Glob
	for _, v := range variants {
		g, err := glob.Compile(v)
		if err == nil {
			compiled = append(compiled, g)
		}
	}
	return compiled
}

// ShouldIgnore checks whether a path should be ignored per .gitignore rules
func (m *GitIgnoreMatcher) ShouldIgnore(relPath string, isDir bool) bool {
	if m == nil {
		return false
	}

	relPath = filepath.ToSlash(relPath)
	basename := filepath.Base(relPath)

	ignored := false
	for _, rule := range m.rules {
		if rule.dirOnly && !isDir {
			continue
		}

		matched := false
		for _, g := range rule.compiled {
			if g.Match(relPath) || g.Match(basename) {
				matched = true
				break
			}
		}

		if matched {
			if rule.negate {
				ignored = false
			} else {
				ignored = true
			}
		}
	}

	return ignored
}
