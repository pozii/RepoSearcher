package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pozii/RepoSearcher/internal/export"
	"github.com/pozii/RepoSearcher/internal/output"
	"github.com/pozii/RepoSearcher/internal/search"
	"github.com/pozii/RepoSearcher/pkg/models"
	"github.com/spf13/cobra"
)

var (
	flagRegex       bool
	flagIgnoreCase  bool
	flagGitHub      bool
	flagGitHubToken string
	flagJSON        string
	flagCSV         string
	flagExtensions  string
	flagExclude     string
	flagInclude     string
	flagContext     int
	flagSince       string
	flagAuthor      string
	flagChangedIn   string
	flagCommitMsg   bool
	flagFuzzy       bool
	flagSuggest     bool
)

var searchCmd = &cobra.Command{
	Use:   "search <query> <path...>",
	Short: "Search for code in repositories",
	Long: `Search for code patterns across local filesystem directories or GitHub repositories.

Supports:
  - Keyword search (default)
  - Regex search (--regex)
  - Fuzzy search (--fuzzy) - tolerates typos
  - Smart suggestions (--suggest) - shows related matches
  - Case-insensitive search (--ignore-case)
  - File extension filtering (--extensions)
  - JSON/CSV export (--json, --csv)
  - GitHub Codesearch API (--github)
  - Git history search (--since, --author, --changed-in)

Examples:
  # Search local directory
  repo-searcher search "func main" ./myproject

  # Fuzzy search (tolerates typos like "fucntion" → "function")
  repo-searcher search "fucntion" ./src --fuzzy

  # Search with suggestions
  repo-searcher search "pars" ./src --suggest

  # Search with regex
  repo-searcher search "func\s+\w+\(" ./myproject --regex

  # Search GitHub repo
  repo-searcher search "TODO" owner/repo --github --github-token ghp_xxx

  # Export to JSON
  repo-searcher search "error" ./project --json results.json

  # Search in files changed in last week
  repo-searcher search "error" ./project --since "1 week ago"

  # Search in files by author
  repo-searcher search "TODO" ./project --author "john"

  # Search in files changed in last 5 commits
  repo-searcher search "bug" ./project --changed-in HEAD~5`,
	Args: cobra.MinimumNArgs(2),
	RunE: runSearch,
}

func init() {
	searchCmd.Flags().BoolVar(&flagRegex, "regex", false, "Enable regex mode")
	searchCmd.Flags().BoolVar(&flagIgnoreCase, "ignore-case", false, "Case-insensitive search")
	searchCmd.Flags().BoolVar(&flagGitHub, "github", false, "Search GitHub repositories")
	searchCmd.Flags().StringVar(&flagGitHubToken, "github-token", "", "GitHub API token (or set GITHUB_TOKEN env var)")
	searchCmd.Flags().StringVar(&flagJSON, "json", "", "Export results to JSON file")
	searchCmd.Flags().StringVar(&flagCSV, "csv", "", "Export results to CSV file")
	searchCmd.Flags().StringVar(&flagExtensions, "extensions", "", "Filter by file extensions (comma-separated, e.g. .go,.py)")
	searchCmd.Flags().StringVar(&flagExclude, "exclude", "", "Exclude files matching glob patterns (comma-separated, e.g. '**/test_*,vendor/**')")
	searchCmd.Flags().StringVar(&flagInclude, "include", "", "Only include files matching glob patterns (comma-separated, e.g. 'src/**/*.go')")
	searchCmd.Flags().IntVar(&flagContext, "context", 0, "Lines of context around matches")

	// Git flags
	searchCmd.Flags().StringVar(&flagSince, "since", "", "Search files changed since (e.g. '1 week ago', '3 days ago')")
	searchCmd.Flags().StringVar(&flagAuthor, "author", "", "Search files by author")
	searchCmd.Flags().StringVar(&flagChangedIn, "changed-in", "", "Search files changed in commit range (e.g. HEAD~5)")
	searchCmd.Flags().BoolVar(&flagCommitMsg, "commit-message", false, "Search commit messages instead of file content")

	// AI-like flags
	searchCmd.Flags().BoolVar(&flagFuzzy, "fuzzy", false, "Enable fuzzy search (tolerates typos)")
	searchCmd.Flags().BoolVar(&flagSuggest, "suggest", false, "Show smart suggestions")
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]
	paths := args[1:]

	// Parse extensions
	var extensions []string
	if flagExtensions != "" {
		extensions = strings.Split(flagExtensions, ",")
	}

	// Parse glob patterns
	var includeGlobs, excludeGlobs []string
	if flagInclude != "" {
		includeGlobs = strings.Split(flagInclude, ",")
	}
	if flagExclude != "" {
		excludeGlobs = strings.Split(flagExclude, ",")
	}

	// Create base config
	githubToken := flagGitHubToken
	if githubToken == "" {
		githubToken = os.Getenv("GITHUB_TOKEN")
	}
	config := models.SearchConfig{
		Query:        query,
		Paths:        paths,
		IsRegex:      flagRegex,
		IgnoreCase:   flagIgnoreCase,
		Context:      flagContext,
		Extensions:   extensions,
		IncludeGlobs: includeGlobs,
		ExcludeGlobs: excludeGlobs,
		GitHub:       flagGitHub,
		GitHubToken:  githubToken,
	}

	// Check if git flags are used
	useGit := flagSince != "" || flagAuthor != "" || flagChangedIn != "" || flagCommitMsg

	var results []models.SearchResult
	var err error

	// Handle suggest mode
	if flagSuggest {
		suggestEngine := search.NewSuggestEngine()
		suggestions, err := suggestEngine.Suggest(config)
		if err != nil {
			return fmt.Errorf("suggestion failed: %w", err)
		}

		fmt.Printf("\nSuggestions for \"%s\":\n\n", query)
		for i, s := range suggestions {
			fmt.Printf("  %d. %s (score: %.2f) - %s\n", i+1, s.Text, s.Score, s.Context)
		}
		fmt.Println()
		return nil
	}

	// Handle fuzzy search
	if flagFuzzy {
		results, err = runFuzzySearch(config, paths)
		if err != nil {
			return fmt.Errorf("fuzzy search failed: %w", err)
		}
	} else if useGit {
		// Create git config
		gitConfig := models.GitSearchConfig{
			SearchConfig: config,
			CommitMsg:    flagCommitMsg,
		}

		// Parse --since flag
		if flagSince != "" {
			sinceTime, err := search.ParseTimeFlag(flagSince)
			if err != nil {
				return fmt.Errorf("invalid --since value: %w", err)
			}
			gitConfig.Since = sinceTime
		}

		// Set author
		if flagAuthor != "" {
			gitConfig.Author = flagAuthor
		}

		// Set changed-in
		if flagChangedIn != "" {
			gitConfig.ChangedIn = flagChangedIn
		}

		// Use GitEngine
		engine := search.NewGitEngine()
		results, err = engine.SearchWithGit(gitConfig)
	} else if flagGitHub {
		// Use GitHubEngine
		engine := search.NewGitHubEngine(flagGitHubToken)
		results, err = engine.Search(config)
	} else {
		// Use ParallelEngine for optimized performance
		engine := search.NewParallelEngine()

		if flagJSON != "" || flagCSV != "" {
			// Need full results for export
			results, err = engine.Search(config)
		} else {
			// Streaming path: print as found
			sp := output.NewStreamingPrinter()
			err = engine.SearchStream(config, func(r models.SearchResult) {
				sp.OnResult(r, query)
			})
			sp.PrintSummary(query)
			return err
		}
	}

	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	// Print results (non-streaming path)
	printer := output.NewPrinter()
	printer.PrintResults(results, query)
	printer.PrintSummary(len(results), query)

	// Export JSON
	if flagJSON != "" {
		if err := export.ToJSON(results, flagJSON); err != nil {
			return fmt.Errorf("JSON export failed: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Results exported to %s\n", flagJSON)
	}

	// Export CSV
	if flagCSV != "" {
		if err := export.ToCSV(results, flagCSV); err != nil {
			return fmt.Errorf("CSV export failed: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Results exported to %s\n", flagCSV)
	}

	return nil
}

// runFuzzySearch performs fuzzy matching search
func runFuzzySearch(config models.SearchConfig, paths []string) ([]models.SearchResult, error) {
	// Get all identifiers from the codebase
	suggestEngine := search.NewSuggestEngine()
	suggestions, err := suggestEngine.Suggest(config)
	if err != nil {
		return nil, err
	}

	// Use ParallelEngine for optimized performance
	engine := search.NewParallelEngine()
	exactResults, err := engine.Search(config)
	if err != nil {
		return nil, err
	}

	// Convert suggestions to search results
	var fuzzyResults []models.SearchResult
	for _, s := range suggestions {
		fuzzyResults = append(fuzzyResults, models.SearchResult{
			FilePath:    "",
			LineNumber:  0,
			LineContent: fmt.Sprintf("[fuzzy match] %s (score: %.2f)", s.Text, s.Score),
			MatchText:   s.Text,
		})
	}

	// Combine results
	allResults := append(exactResults, fuzzyResults...)
	return allResults, nil
}
