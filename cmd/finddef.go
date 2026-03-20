package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/pozii/RepoSearcher/internal/lsp"
	"github.com/spf13/cobra"
)

var findDefCmd = &cobra.Command{
	Use:   "find-definition <name> [path...]",
	Short: "Find the definition of a symbol",
	Long: `Find the definition of a symbol (function, struct, type, variable).

This command locates where a symbol is defined in your codebase.
Works with:
  - Function definitions
  - Struct/interface definitions
  - Type definitions
  - Variable/constant declarations

Examples:
  # Find definition of SearchEngine
  repo-searcher find-definition "SearchEngine" .

  # Find in Go files only
  repo-searcher find-definition "ParseJSON" ./src --ext .go`,
	Args: cobra.MinimumNArgs(2),
	RunE: runFindDef,
}

func init() {
	findDefCmd.Flags().StringVar(&flagFindExt, "ext", "", "Filter by file extension (e.g. .go, .py)")
	rootCmd.AddCommand(findDefCmd)
}

func runFindDef(cmd *cobra.Command, args []string) error {
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

	// Find definition
	definitions := extractor.FindDefinition(index, name)

	if len(definitions) == 0 {
		fmt.Printf("No definition found for \"%s\"\n", name)
		return nil
	}

	// Print results with color
	titleColor := color.New(color.FgCyan, color.Bold)
	typeColor := color.New(color.FgYellow)
	fileColor := color.New(color.FgGreen)
	lineColor := color.New(color.FgWhite, color.Faint)

	fmt.Printf("\nDefinition of \"%s\":\n\n", name)

	for _, def := range definitions {
		fmt.Printf("  %s %s %s %s\n",
			typeColor.Sprintf("[%s]", def.Type),
			titleColor.Sprintf("%s", def.Name),
			fileColor.Sprintf("%s:%d", def.File, def.Line),
			lineColor.Sprintf("%s", def.Signature),
		)
	}

	fmt.Println()
	return nil
}
