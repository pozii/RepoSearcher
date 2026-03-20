package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pozii/RepoSearcher/pkg/models"
)

const (
	githubSearchAPI = "https://api.github.com/search/code"
	maxPages        = 5
	resultsPerPage  = 30
)

// GitHubEngine implements SearchEngine for GitHub Codesearch API
type GitHubEngine struct {
	Token      string
	HTTPClient *http.Client
}

// NewGitHubEngine creates a new GitHubEngine instance
func NewGitHubEngine(token string) *GitHubEngine {
	return &GitHubEngine{
		Token: token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Search performs a search using GitHub Codesearch API
func (e *GitHubEngine) Search(config models.SearchConfig) ([]models.SearchResult, error) {
	var allResults []models.SearchResult

	for _, repo := range config.Paths {
		results, err := e.searchRepo(repo, config.Query, config.IsRegex, config.IgnoreCase)
		if err != nil {
			return nil, fmt.Errorf("searching repo %s: %w", repo, err)
		}
		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// searchRepo searches within a single GitHub repository with pagination
func (e *GitHubEngine) searchRepo(repo, query string, isRegex, ignoreCase bool) ([]models.SearchResult, error) {
	searchQuery := url.QueryEscape(query + " repo:" + repo)
	if !ignoreCase {
		searchQuery += "+is:case-sensitive"
	}

	var results []models.SearchResult

	for page := 1; page <= maxPages; page++ {
		reqURL := fmt.Sprintf("%s?q=%s&per_page=%d&page=%d", githubSearchAPI, searchQuery, resultsPerPage, page)

		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Accept", "application/vnd.github.v3.text-match+json")
		if e.Token != "" {
			req.Header.Set("Authorization", "token "+e.Token)
		}

		resp, err := e.HTTPClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			switch resp.StatusCode {
			case http.StatusForbidden:
				return results, fmt.Errorf("GitHub API rate limit exceeded. Set GITHUB_TOKEN or wait for rate limit reset")
			case http.StatusUnauthorized:
				return results, fmt.Errorf("GitHub API authentication failed. Check your GITHUB_TOKEN")
			case http.StatusNotFound:
				return results, fmt.Errorf("GitHub repository not found: %s", repo)
			default:
				return results, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
			}
		}

		var searchResp models.GitHubCodesearchResponse
		if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
			resp.Body.Close()
			return results, fmt.Errorf("failed to decode GitHub response: %w", err)
		}
		resp.Body.Close()

		for _, item := range searchResp.Items {
			lineContent := ""
			lineNumber := 0

			// Extract fragment from text matches if available
			if len(item.TextMatches) > 0 {
				fragment := item.TextMatches[0].Fragment
				lineContent = strings.TrimSpace(fragment)
				lineNumber = extractLineNumber(fragment, query)
			}

			results = append(results, models.SearchResult{
				FilePath:    item.Repository.FullName + "/" + item.Path,
				LineNumber:  lineNumber,
				LineContent: lineContent,
				MatchText:   item.Name,
			})
		}

		// Stop if we've fetched all results
		if len(results) >= searchResp.TotalCount || len(searchResp.Items) < resultsPerPage {
			break
		}
	}

	return results, nil
}

// extractLineNumber tries to extract a line number from a fragment
func extractLineNumber(fragment, query string) int {
	// GitHub fragments don't typically include line numbers,
	// but we try to find the query in context
	_ = fragment
	_ = query
	return 0
}
