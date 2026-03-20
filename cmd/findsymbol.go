package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/pozii/RepoSearcher/internal/lsp"
	"github.com/spf13/cobra"
)

var flagFindExt string

var findSymbolCmd = &cobra.Command{
	Use:   "find-symbol <name> [path...]",
	Short: "Find symbols matching a name",
	Long: `Find code symbols (functions, structs, variables, types) matching a name.

This command uses pattern matching to find symbols across your codebase
without requiring a Language Server. It works with:
  - Functions and methods
  - Structs and interfaces
  - Variables and constants
  - Types and classes
  - Fields

The search is case-insensitive and supports partial matches.

Examples:
  # Find symbol containing "Search"
  repo-searcher find-symbol "Search" .

  # Find in Go files only
  repo-searcher find-symbol "Engine" ./src --ext .go

  # Find across multiple directories
  repo-searcher find-symbol "Config" ./src ./lib`,
	Args: cobra.MinimumNArgs(2),
	RunE: runFindSymbol,
}

func init() {
	findSymbolCmd.Flags().StringVar(&flagFindExt, "ext", "", "Filter by file extension (e.g. .go, .py)")
	rootCmd.AddCommand(findSymbolCmd)
}

func runFindSymbol(cmd *cobra.Command, args []string) error {
	name := args[0]
	paths := args[1:]

	// Parse extensions
	var extensions []string
	if flagFindExt != "" {
		extensions = strings.Split(flagFindExt, ",")
	}

	// Extract symbols
	extractor := lsp.NewSymbolExtractor()
	index, err := extractor.ExtractSymbols(paths[0], extensions)
	if err != nil {
		return fmt.Errorf("symbol extraction failed: %w", err)
	}

	// Find matching symbols
	symbols := extractor.FindSymbol(index, name)

	if len(symbols) == 0 {
		fmt.Printf("No symbols found matching \"%s\"\n", name)
		return nil
	}

	// Print results with color
	titleColor := color.New(color.FgCyan, color.Bold)
	typeColor := color.New(color.FgYellow)
	fileColor := color.New(color.FgGreen)
	lineColor := color.New(color.FgWhite, color.Faint)

	fmt.Printf("\nFound %d symbol(s) matching \"%s\":\n\n", len(symbols), name)

	for _, sym := range symbols {
		fmt.Printf("  %s %s %s %s\n",
			typeColor.Sprintf("[%s]", sym.Type),
			titleColor.Sprintf("%s", sym.Name),
			fileColor.Sprintf("%s:%d", sym.File, sym.Line),
			lineColor.Sprintf("%s", sym.Signature),
		)
	}

	fmt.Println()
	return nil
}
