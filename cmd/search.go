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
	flagContext     int
)

var searchCmd = &cobra.Command{
	Use:   "search <query> <path...>",
	Short: "Search for code in repositories",
	Long: `Search for code patterns across local filesystem directories or GitHub repositories.

Supports:
  • Keyword search (default)
  • Regex search (--regex)
  • Case-insensitive search (--ignore-case)
  • File extension filtering (--extensions)
  • JSON/CSV export (--json, --csv)
  • GitHub Codesearch API (--github)

Examples:
  # Search local directory
  repo-searcher search "func main" ./myproject

  # Search with regex
  repo-searcher search "func\s+\w+\(" ./myproject --regex

  # Search GitHub repo
  repo-searcher search "TODO" owner/repo --github --github-token ghp_xxx

  # Export to JSON
  repo-searcher search "error" ./project --json results.json`,
	Args: cobra.MinimumNArgs(2),
	RunE: runSearch,
}

func init() {
	searchCmd.Flags().BoolVar(&flagRegex, "regex", false, "Enable regex mode")
	searchCmd.Flags().BoolVar(&flagIgnoreCase, "ignore-case", false, "Case-insensitive search")
	searchCmd.Flags().BoolVar(&flagGitHub, "github", false, "Search GitHub repositories")
	searchCmd.Flags().StringVar(&flagGitHubToken, "github-token", "", "GitHub API token")
	searchCmd.Flags().StringVar(&flagJSON, "json", "", "Export results to JSON file")
	searchCmd.Flags().StringVar(&flagCSV, "csv", "", "Export results to CSV file")
	searchCmd.Flags().StringVar(&flagExtensions, "extensions", "", "Filter by file extensions (comma-separated, e.g. .go,.py)")
	searchCmd.Flags().IntVar(&flagContext, "context", 0, "Lines of context around matches")
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]
	paths := args[1:]

	// Parse extensions
	var extensions []string
	if flagExtensions != "" {
		extensions = strings.Split(flagExtensions, ",")
	}

	// Create config
	config := models.SearchConfig{
		Query:       query,
		Paths:       paths,
		IsRegex:     flagRegex,
		IgnoreCase:  flagIgnoreCase,
		Context:     flagContext,
		Extensions:  extensions,
		GitHub:      flagGitHub,
		GitHubToken: flagGitHubToken,
	}

	// Select engine
	var engine search.SearchEngine
	if flagGitHub {
		engine = search.NewGitHubEngine(flagGitHubToken)
	} else {
		engine = search.NewLocalEngine()
	}

	// Execute search
	results, err := engine.Search(config)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	// Print results
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
