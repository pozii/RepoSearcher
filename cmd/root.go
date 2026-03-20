package cmd

import (
	"fmt"
	"os"

	"github.com/pozii/RepoSearcher/internal/updater"
	"github.com/spf13/cobra"
)

var (
	flagNoUpdate bool
)

var rootCmd = &cobra.Command{
	Use:   "repo-searcher",
	Short: "A powerful code search tool for local and GitHub repositories",
	Long: `RepoSearcher is a fast, professional CLI tool for searching code across 
multiple repositories with beautiful terminal output and export options.

Search with keywords or regex patterns in:
  - Local filesystem directories
  - GitHub repositories (via Codesearch API)

Features:
  - Regex and keyword matching
  - Case-insensitive search
  - File extension filtering
  - JSON and CSV export
  - Colored terminal output
  - Auto-update support`,
	PersistentPreRun: checkForUpdates,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&flagNoUpdate, "no-update-check", false, "Skip automatic update check")
	rootCmd.AddCommand(searchCmd)
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

// checkForUpdates runs an automatic update check before any command
func checkForUpdates(cmd *cobra.Command, args []string) {
	// Skip update check if --no-update-check is set
	// or if this is an install/update/uninstall command
	if flagNoUpdate || cmd.Name() == "install" || cmd.Name() == "update" || cmd.Name() == "uninstall" || cmd.Name() == "completion" {
		return
	}

	// Check for update (non-blocking, just show message)
	release, hasUpdate, err := updater.CheckForUpdate()
	if err != nil {
		// Silent fail - don't block the user
		return
	}

	if hasUpdate {
		fmt.Fprintf(os.Stderr, "\nNew version available: %s (current: %s)\n", release.TagName, updater.GetCurrentVersion())
		fmt.Fprintf(os.Stderr, "Run 'repo-searcher update' to update.\n\n")
	}
}
