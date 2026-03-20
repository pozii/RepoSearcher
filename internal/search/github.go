package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pozii/RepoSearcher/pkg/models"
)

const (
	githubSearchAPI = "https://api.github.com/search/code"
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

// searchRepo searches within a single GitHub repository
func (e *GitHubEngine) searchRepo(repo, query string, isRegex, ignoreCase bool) ([]models.SearchResult, error) {
	searchQuery := url.QueryEscape(query + " repo:" + repo)
	if ignoreCase {
		searchQuery += "+is:case-sensitive"
	}

	req, err := http.NewRequest("GET", githubSearchAPI+"?q="+searchQuery, nil)
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned status %d", resp.StatusCode)
	}

	var searchResp models.GitHubCodesearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	var results []models.SearchResult
	for _, item := range searchResp.Items {
		results = append(results, models.SearchResult{
			FilePath:   item.Repository.FullName + "/" + item.Path,
			LineNumber: 0,
			MatchText:  item.Name,
		})
	}

	return results, nil
}
