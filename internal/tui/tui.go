package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pozii/RepoSearcher/internal/export"
	"github.com/pozii/RepoSearcher/internal/search"
	"github.com/pozii/RepoSearcher/pkg/models"
)

// AppVersion is set by the cmd package at startup
var AppVersion = "dev"

// Model represents the TUI application state
type Model struct {
	query      textinput.Model
	results    []models.SearchResult
	cursor     int
	searching  bool
	paths      []string
	extensions string
	isRegex    bool
	ignoreCase bool
	width      int
	height     int
	quitting   bool
	err        error
}

// NewModel creates a new TUI model
func NewModel(paths []string, extensions string, isRegex, ignoreCase bool) Model {
	ti := textinput.New()
	ti.Placeholder = "Enter search query..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	return Model{
		query:      ti,
		results:    []models.SearchResult{},
		paths:      paths,
		extensions: extensions,
		isRegex:    isRegex,
		ignoreCase: ignoreCase,
		width:      80,
		height:     24,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.query.Width = m.width - 4

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			m.searching = true
			m.results = []models.SearchResult{}
			m.cursor = 0
			return m, searchCmd(m.query.Value(), m.paths, m.extensions, m.isRegex, m.ignoreCase)

		case "j", "down":
			if len(m.results) > 0 && m.cursor < len(m.results)-1 {
				m.cursor++
			}

		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}

		case "g":
			m.cursor = 0

		case "G":
			if len(m.results) > 0 {
				m.cursor = len(m.results) - 1
			}

		case "e":
			// Export functionality placeholder
			if len(m.results) > 0 {
				exportJSON(m.results, "results.json")
			}

		case "f":
			// Filter functionality placeholder
		}

	case searchResultMsg:
		m.searching = false
		m.results = msg.results

	case errMsg:
		m.searching = false
		m.err = msg.err
	}

	// Update query input
	m.query, cmd = m.query.Update(msg)
	return m, cmd
}

// View renders the TUI
func (m Model) View() string {
	if m.quitting {
		return quitStyle.Render("Goodbye!") + "\n"
	}

	var sb strings.Builder

	// Title
	title := titleStyle.Render("RepoSearcher " + AppVersion + " - Interactive Search")
	sb.WriteString(title + "\n\n")

	// Query input
	sb.WriteString(promptStyle.Render("Query: "))
	sb.WriteString(inputStyle.Render(m.query.View()))
	sb.WriteString("\n\n")

	// Status line
	if m.searching {
		sb.WriteString(statusStyle.Render("Searching..."))
		sb.WriteString("\n\n")
	} else if m.err != nil {
		sb.WriteString(quitStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		sb.WriteString("\n\n")
	} else if len(m.results) > 0 {
		sb.WriteString(fileStyle.Render(fmt.Sprintf("Found %d matches", len(m.results))))
		sb.WriteString("\n\n")
	}

	// Results
	sb.WriteString(fileStyle.Render("Matches: "))
	sb.WriteString("\n")

	if len(m.results) == 0 {
		sb.WriteString(lineContentStyle.Render("  No results yet. Press Enter to search."))
	} else {
		// Show results
		visibleResults := m.height - 12
		if visibleResults < 1 {
			visibleResults = 10
		}

		start := 0
		if m.cursor > visibleResults/2 {
			start = m.cursor - visibleResults/2
		}
		if start > len(m.results)-visibleResults {
			start = len(m.results) - visibleResults
		}
		if start < 0 {
			start = 0
		}

		end := start + visibleResults
		if end > len(m.results) {
			end = len(m.results)
		}

		for i := start; i < end; i++ {
			r := m.results[i]
			lineNum := lineNumStyle.Render(fmt.Sprintf("%3d", r.LineNumber))

			// Highlight matched text
			content := r.LineContent
			highlighted := strings.Replace(content, r.MatchText, matchStyle.Render(r.MatchText), 1)
			if len(highlighted) > 60 {
				highlighted = highlighted[:60] + "..."
			}

			cursor := " "
			if i == m.cursor {
				cursor = "▸"
			}

			resultLine := fmt.Sprintf("%s %s:%s %s",
				cursor,
				fileStyle.Render(r.FilePath),
				lineNum,
				lineContentStyle.Render(highlighted),
			)

			if i == m.cursor {
				sb.WriteString(selectedStyle.Render(resultLine))
			} else {
				sb.WriteString(resultLine)
			}
			sb.WriteString("\n")
		}
	}

	// Help bar
	sb.WriteString("\n")
	help := helpStyle.Render("[j/k] Navigate  [Enter] Search  [g/G] Top/Bottom  [e] Export  [q] Quit")
	sb.WriteString(statusStyle.Render(help))

	return sb.String()
}

// Run starts the interactive TUI
func Run(paths []string, extensions string, isRegex, ignoreCase bool) error {
	p := tea.NewProgram(
		NewModel(paths, extensions, isRegex, ignoreCase),
		tea.WithAltScreen(),
	)
	_, err := p.Run()
	return err
}

// Messages
type searchResultMsg struct {
	results []models.SearchResult
}

type errMsg struct {
	err error
}

// Commands
func searchCmd(query string, paths []string, extensions string, isRegex, ignoreCase bool) tea.Cmd {
	return func() tea.Msg {
		// Parse extensions
		var ext []string
		if extensions != "" {
			ext = strings.Split(extensions, ",")
		}

		// Create search config
		config := models.SearchConfig{
			Query:      query,
			Paths:      paths,
			IsRegex:    isRegex,
			IgnoreCase: ignoreCase,
			Extensions: ext,
		}

		// Use ParallelEngine for optimized performance
		engine := search.NewParallelEngine()
		results, err := engine.Search(config)
		if err != nil {
			return errMsg{err}
		}

		return searchResultMsg{results}
	}
}

// exportJSON exports search results to a JSON file
func exportJSON(results []models.SearchResult, filename string) {
	if err := export.ToJSON(results, filename); err != nil {
		return
	}
}
