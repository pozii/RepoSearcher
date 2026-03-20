package search

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pozii/RepoSearcher/pkg/models"
)

// GitEngine implements SearchEngine for git-aware search
type GitEngine struct{}

// NewGitEngine creates a new GitEngine instance
func NewGitEngine() *GitEngine {
	return &GitEngine{}
}

// Search performs a git-aware search
func (e *GitEngine) Search(config models.SearchConfig) ([]models.SearchResult, error) {
	var allResults []models.SearchResult

	for _, root := range config.Paths {
		results, err := e.searchPath(root, config)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// SearchWithGit performs search with git filters
func (e *GitEngine) SearchWithGit(config models.GitSearchConfig) ([]models.SearchResult, error) {
	var allResults []models.SearchResult

	for _, root := range config.Paths {
		repo, err := e.findGitRepo(root)
		if err != nil {
			results, err2 := e.searchPath(root, config.SearchConfig)
			if err2 != nil {
				return nil, err2
			}
			allResults = append(allResults, results...)
			continue
		}

		changedFiles, err := e.getChangedFiles(repo, config)
		if err != nil {
			return nil, fmt.Errorf("failed to get changed files: %w", err)
		}

		pattern, err := CompilePattern(config.Query, config.IsRegex, config.IgnoreCase)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern: %w", err)
		}

		for _, filePath := range changedFiles {
			absPath := filepath.Join(root, filePath)
			if ShouldIgnoreFile(absPath, config.Extensions) {
				continue
			}
			results, err := e.searchFile(absPath, filePath, pattern, config.SearchConfig)
			if err != nil {
				continue
			}
			allResults = append(allResults, results...)
		}
	}

	return allResults, nil
}

// searchPath searches through a single path
func (e *GitEngine) searchPath(root string, config models.SearchConfig) ([]models.SearchResult, error) {
	pattern, err := CompilePattern(config.Query, config.IsRegex, config.IgnoreCase)
	if err != nil {
		return nil, err
	}

	var results []models.SearchResult

	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("cannot access path %s: %w", root, err)
	}

	if !info.IsDir() {
		return e.searchFile(root, root, pattern, config)
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

		fileResults, err := e.searchFile(path, path, pattern, config)
		if err != nil {
			return nil
		}

		results = append(results, fileResults...)
		return nil
	})

	return results, err
}

// searchFile searches within a single file
func (e *GitEngine) searchFile(absPath, relPath string, pattern *regexp.Regexp, config models.SearchConfig) ([]models.SearchResult, error) {
	file, err := os.Open(absPath)
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
				FilePath:    relPath,
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

// findGitRepo finds the git repository for a given path
func (e *GitEngine) findGitRepo(path string) (*git.Repository, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		parent := filepath.Dir(path)
		for parent != path {
			repo, err = git.PlainOpen(parent)
			if err == nil {
				return repo, nil
			}
			path = parent
			parent = filepath.Dir(path)
		}
		return nil, fmt.Errorf("no git repository found")
	}
	return repo, nil
}

// getChangedFiles returns files that match the git criteria
func (e *GitEngine) getChangedFiles(repo *git.Repository, config models.GitSearchConfig) ([]string, error) {
	fileSet := make(map[string]bool)

	commits, err := e.getFilteredCommits(repo, config)
	if err != nil {
		return nil, err
	}

	for _, commit := range commits {
		files, err := e.getFilesInCommit(repo, commit)
		if err != nil {
			continue
		}
		for _, f := range files {
			fileSet[f] = true
		}
	}

	files := make([]string, 0, len(fileSet))
	for f := range fileSet {
		files = append(files, f)
	}

	return files, nil
}

// getFilteredCommits returns commits matching the criteria
func (e *GitEngine) getFilteredCommits(repo *git.Repository, config models.GitSearchConfig) ([]object.Commit, error) {
	commitIter, err := repo.Log(&git.LogOptions{All: true})
	if err != nil {
		return nil, err
	}

	var commits []object.Commit
	commitIter.ForEach(func(commit *object.Commit) error {
		match := true

		if !config.Since.IsZero() {
			if commit.Author.When.Before(config.Since) {
				match = false
			}
		}

		if config.Author != "" {
			if !strings.Contains(strings.ToLower(commit.Author.Name), strings.ToLower(config.Author)) &&
				!strings.Contains(strings.ToLower(commit.Author.Email), strings.ToLower(config.Author)) {
				match = false
			}
		}

		if match {
			commits = append(commits, *commit)
		}

		return nil
	})

	return commits, nil
}

// getFilesInCommit returns the list of files changed in a commit
func (e *GitEngine) getFilesInCommit(repo *git.Repository, commit object.Commit) ([]string, error) {
	var files []string

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	if commit.NumParents() > 0 {
		parentIter := commit.Parents()
		parent, err := parentIter.Next()
		if err != nil {
			return nil, err
		}

		parentTree, err := parent.Tree()
		if err != nil {
			return nil, err
		}

		changes, err := object.DiffTree(parentTree, tree)
		if err != nil {
			return nil, err
		}

		for _, change := range changes {
			if change.To.Name != "" {
				files = append(files, change.To.Name)
			}
			if change.From.Name != "" {
				files = append(files, change.From.Name)
			}
		}
	} else {
		tree.Files().ForEach(func(f *object.File) error {
			files = append(files, f.Name)
			return nil
		})
	}

	return files, nil
}

// ParseTimeFlag parses time-related flags like "1 week ago", "2 days ago"
func ParseTimeFlag(since string) (time.Time, error) {
	parts := strings.Split(since, " ")
	if len(parts) < 3 {
		return time.Time{}, fmt.Errorf("invalid time format: %s (use: 'N days/weeks/months ago')", since)
	}

	amount := 1
	if n, err := fmt.Sscanf(parts[0], "%d", &amount); err != nil || n != 1 {
		return time.Time{}, fmt.Errorf("invalid number: %s", parts[0])
	}

	unit := strings.ToLower(parts[1])
	switch unit {
	case "day", "days":
		return time.Now().AddDate(0, 0, -amount), nil
	case "week", "weeks":
		return time.Now().AddDate(0, 0, -amount*7), nil
	case "month", "months":
		return time.Now().AddDate(0, -amount, 0), nil
	case "year", "years":
		return time.Now().AddDate(-amount, 0, 0), nil
	default:
		return time.Time{}, fmt.Errorf("unknown time unit: %s", unit)
	}
}
