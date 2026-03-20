package models

// SearchResult represents a single match found in the codebase
type SearchResult struct {
	FilePath    string `json:"file_path" csv:"file_path"`
	LineNumber  int    `json:"line_number" csv:"line_number"`
	LineContent string `json:"line_content" csv:"line_content"`
	MatchText   string `json:"match_text" csv:"match_text"`
}

// SearchResults holds a collection of search matches
type SearchResults struct {
	Matches    []SearchResult `json:"matches"`
	TotalCount int            `json:"total_count"`
	Query      string         `json:"query"`
	IsRegex    bool           `json:"is_regex"`
}

// SearchConfig holds configuration for a search operation
type SearchConfig struct {
	Query       string
	Paths       []string
	IsRegex     bool
	IgnoreCase  bool
	Context     int
	Extensions  []string
	GitHub      bool
	GitHubToken string
}

// GitHubCodesearchResponse represents the GitHub Codesearch API response
type GitHubCodesearchResponse struct {
	TotalCount        int              `json:"total_count"`
	IncompleteResults bool             `json:"incomplete_results"`
	Items             []GitHubCodeItem `json:"items"`
}

// GitHubCodeItem represents a single code search result from GitHub
type GitHubCodeItem struct {
	Name       string     `json:"name"`
	Path       string     `json:"path"`
	Repository GitHubRepo `json:"repository"`
	HTMLURL    string     `json:"html_url"`
}

// GitHubRepo represents a GitHub repository in search results
type GitHubRepo struct {
	FullName string `json:"full_name"`
}

// GitHubCodeMatch represents a single line match from GitHub API
type GitHubCodeMatch struct {
	TextMatches []GitHubTextMatch `json:"text_matches"`
}

// GitHubTextMatch represents a text match fragment
type GitHubTextMatch struct {
	Fragment string `json:"fragment"`
}
