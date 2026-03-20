package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pozii/RepoSearcher/internal/tui"
	"github.com/spf13/cobra"
)

var (
	flagTUIExtensions string
	flagTUIRegex      bool
	flagTUIIgnoreCase bool
)

var interactiveCmd = &cobra.Command{
	Use:     "interactive [path...]",
	Aliases: []string{"i"},
	Short:   "Launch interactive search mode",
	Long: `Launch an interactive terminal-based search interface.

This command opens a full-screen TUI (Terminal User Interface) for
interactive code searching. Features:

  - Real-time search with vim-style navigation
  - Search result preview
  - File path highlighting
  - Export functionality (press 'e')
  - Keyboard shortcuts for efficient workflow

Keyboard shortcuts:
  j/k or Up/Down    Navigate results
  Enter             Execute search
  g/G               Jump to top/bottom
  e                 Export results to JSON
  q/Esc             Quit

Examples:
  # Launch in current directory
  repo-searcher interactive .

  # Launch with regex mode
  repo-searcher i ./src --regex

  # Launch with file extensions
  repo-searcher i ./project --extensions .go,.py`,
	Args: cobra.MinimumNArgs(1),
	RunE: runInteractive,
}

func init() {
	interactiveCmd.Flags().StringVar(&flagTUIExtensions, "extensions", "", "Filter by file extensions (comma-separated)")
	interactiveCmd.Flags().BoolVar(&flagTUIRegex, "regex", false, "Enable regex mode")
	interactiveCmd.Flags().BoolVar(&flagTUIIgnoreCase, "ignore-case", false, "Case-insensitive search")
	rootCmd.AddCommand(interactiveCmd)
}

func runInteractive(cmd *cobra.Command, args []string) error {
	// Validate paths exist
	for _, path := range args {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
	}

	// Parse extensions
	extensions := flagTUIExtensions
	if extensions != "" && !strings.HasPrefix(extensions, ".") {
		extensions = "." + extensions
	}

	fmt.Println("Launching Interactive Search...")
	fmt.Println("Press 'q' or Esc to quit, 'Enter' to search")
	fmt.Println()

	// Run TUI
	if err := tui.Run(args, extensions, flagTUIRegex, flagTUIIgnoreCase); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}
