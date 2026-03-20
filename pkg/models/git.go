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
