package models

import "time"

// GitSearchConfig extends SearchConfig with git-specific options
type GitSearchConfig struct {
	SearchConfig
	Since     time.Time
	Author    string
	ChangedIn string // e.g. "HEAD~5", "abc123..def456"
	CommitMsg bool   // Search commit messages instead of file content
}

// GitCommit represents a git commit
type GitCommit struct {
	Hash    string
	Author  string
	Date    time.Time
	Message string
	Files   []string
}

// GitSearchResult extends SearchResult with git metadata
type GitSearchResult struct {
	SearchResult
	LastCommit GitCommit `json:"last_commit,omitempty"`
	Author     string    `json:"author,omitempty"`
}
