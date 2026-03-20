package search

import (
	"testing"
)

func TestIsKeyword(t *testing.T) {
	keywords := []string{
		"if", "else", "for", "return", "func", "var", "const",
		"type", "package", "import", "nil", "true", "false",
		"int", "string", "bool", "float64", "error", "interface",
		"struct", "map", "chan", "go", "defer", "switch", "case",
		"default", "break", "continue", "select", "range",
	}
	for _, kw := range keywords {
		if !isKeyword(kw) {
			t.Errorf("isKeyword(%q) = false, want true", kw)
		}
	}

	nonKeywords := []string{
		"function", "parse", "hello", "world", "myFunc", "MyStruct",
		"config", "handler", "engine", "result",
	}
	for _, kw := range nonKeywords {
		if isKeyword(kw) {
			t.Errorf("isKeyword(%q) = true, want false", kw)
		}
	}
}

func TestIsKeyword_CaseInsensitive(t *testing.T) {
	if !isKeyword("IF") {
		t.Error("isKeyword should be case-insensitive")
	}
	if !isKeyword("Func") {
		t.Error("isKeyword should be case-insensitive")
	}
}

func TestDeduplicateSuggestions(t *testing.T) {
	suggestions := []Suggestion{
		{Text: "function", Score: 0.9, Context: "identifier"},
		{Text: "func", Score: 0.8, Context: "identifier"},
		{Text: "function", Score: 0.7, Context: "function"},
		{Text: "format", Score: 0.6, Context: "identifier"},
	}

	result := deduplicateSuggestions(suggestions)

	if len(result) != 3 {
		t.Errorf("deduplicateSuggestions returned %d results, want 3", len(result))
	}

	// First occurrence should be kept
	found := false
	for _, s := range result {
		if s.Text == "function" && s.Score == 0.9 {
			found = true
		}
	}
	if !found {
		t.Error("deduplicateSuggestions should keep first occurrence with score 0.9")
	}
}

func TestDeduplicateSuggestions_Empty(t *testing.T) {
	result := deduplicateSuggestions(nil)
	if len(result) != 0 {
		t.Errorf("deduplicateSuggestions(nil) should return empty, got %d", len(result))
	}
}

func TestSuggestionStruct(t *testing.T) {
	s := Suggestion{
		Text:    "parseJSON",
		Score:   0.85,
		Context: "function",
		Details: "(data []byte) (interface{}, error)",
	}

	if s.Text != "parseJSON" {
		t.Errorf("Suggestion.Text = %q, want %q", s.Text, "parseJSON")
	}
	if s.Score != 0.85 {
		t.Errorf("Suggestion.Score = %f, want 0.85", s.Score)
	}
}
