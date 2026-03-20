package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/pozii/RepoSearcher/internal/lsp"
	"github.com/spf13/cobra"
)

var findRefCmd = &cobra.Command{
	Use:   "find-references <name> [path...]",
	Short: "Find all references to a symbol",
	Long: `Find all places where a symbol is used/referenced in your codebase.

This command searches for all references to a given symbol across
your code files. Works with:
  - Function calls
  - Variable usage
  - Type references
  - Import statements
  - Field access

The command categorizes references as:
  - definition: Where the symbol is defined
  - usage: Where the symbol is used

Examples:
  # Find all references to SearchEngine
  repo-searcher find-references "SearchEngine" .

  # Find references in Go files
  repo-searcher find-references "config" ./src --ext .go

  # Find across multiple directories
  repo-searcher find-references "NewEngine" ./src ./lib`,
	Args: cobra.MinimumNArgs(2),
	RunE: runFindRef,
}

func init() {
	findRefCmd.Flags().StringVar(&flagFindExt, "ext", "", "Filter by file extension (e.g. .go, .py)")
	rootCmd.AddCommand(findRefCmd)
}

func runFindRef(cmd *cobra.Command, args []string) error {
	name := args[0]
	paths := args[1:]

	// Parse extensions
	var extensions []string
	if flagFindExt != "" {
		extensions = strings.Split(flagFindExt, ",")
	}

	// Find references
	extractor := lsp.NewSymbolExtractor()
	references, err := extractor.FindReferences(paths[0], name, extensions)
	if err != nil {
		return fmt.Errorf("reference search failed: %w", err)
	}

	if len(references) == 0 {
		fmt.Printf("No references found for \"%s\"\n", name)
		return nil
	}

	// Print results with color
	defColor := color.New(color.FgGreen, color.Bold)
	refColor := color.New(color.FgCyan)
	fileColor := color.New(color.FgYellow)
	lineColor := color.New(color.FgWhite, color.Faint)

	fmt.Printf("\nFound %d reference(s) for \"%s\":\n\n", len(references), name)

	for _, ref := range references {
		prefix := refColor.Sprint("[usage]")
		if ref.Category == "definition" {
			prefix = defColor.Sprint("[def]  ")
		}

		fmt.Printf("  %s %s:%d  %s\n",
			prefix,
			fileColor.Sprintf("%s", ref.Symbol.File),
			ref.Line,
			lineColor.Sprintf("%s", truncate(ref.Content, 60)),
		)
	}

	fmt.Println()
	return nil
}

// truncate truncates a string to a maximum length
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
