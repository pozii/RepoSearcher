package output

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/pozii/RepoSearcher/pkg/models"
)

// Printer handles colored terminal output
type Printer struct {
	FileColor      *color.Color
	LineNumColor   *color.Color
	HighlightColor *color.Color
	SeparatorColor *color.Color
}

// NewPrinter creates a new Printer with default colors
func NewPrinter() *Printer {
	return &Printer{
		FileColor:      color.New(color.FgCyan, color.Bold),
		LineNumColor:   color.New(color.FgGreen),
		HighlightColor: color.New(color.FgRed, color.Bold, color.Underline),
		SeparatorColor: color.New(color.FgWhite, color.Faint),
	}
}

// PrintResults displays search results with colored output
func (p *Printer) PrintResults(results []models.SearchResult, query string) {
	if len(results) == 0 {
		fmt.Println("No matches found.")
		return
	}

	lastFile := ""
	for _, r := range results {
		if r.FilePath != lastFile {
			if lastFile != "" {
				p.SeparatorColor.Println(strings.Repeat("─", 60))
			}
			p.FileColor.Println(r.FilePath)
			lastFile = r.FilePath
		}
		p.printMatch(r, query)
	}
}

// printMatch prints a single match with color highlighting
func (p *Printer) printMatch(r models.SearchResult, query string) {
	lineNum := p.LineNumColor.Sprintf("%d", r.LineNumber)
	content := r.LineContent

	highlighted := strings.ReplaceAll(
		content,
		r.MatchText,
		p.HighlightColor.Sprint(r.MatchText),
	)

	fmt.Printf("  %s │ %s\n", lineNum, highlighted)
}

// PrintSummary prints a summary of search results
func (p *Printer) PrintSummary(total int, query string) {
	fmt.Println()
	p.SeparatorColor.Println(strings.Repeat("═", 60))
	fmt.Printf("Found %d matches for \"%s\"\n", total, query)
	p.SeparatorColor.Println(strings.Repeat("═", 60))
}

// StreamingPrinter prints results as they arrive (for real-time output)
type StreamingPrinter struct {
	*Printer
	lastFile string
	count    int
}

// NewStreamingPrinter creates a new StreamingPrinter
func NewStreamingPrinter() *StreamingPrinter {
	return &StreamingPrinter{Printer: NewPrinter()}
}

// OnResult prints a single result as it arrives
func (sp *StreamingPrinter) OnResult(r models.SearchResult, query string) {
	if r.FilePath != sp.lastFile {
		if sp.lastFile != "" {
			sp.SeparatorColor.Println(strings.Repeat("─", 60))
		}
		sp.FileColor.Println(r.FilePath)
		sp.lastFile = r.FilePath
	}
	sp.printMatch(r, query)
	sp.count++
}

// PrintSummary prints the final summary after streaming
func (sp *StreamingPrinter) PrintSummary(query string) {
	fmt.Println()
	sp.SeparatorColor.Println(strings.Repeat("═", 60))
	fmt.Printf("Found %d matches for \"%s\"\n", sp.count, query)
	sp.SeparatorColor.Println(strings.Repeat("═", 60))
}
