package search

import (
	"testing"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s1, s2 string
		want   int
	}{
		{"", "", 0},
		{"abc", "", 3},
		{"", "abc", 3},
		{"abc", "abc", 0},
		{"abc", "abx", 1},
		{"kitten", "sitting", 3},
		{"saturday", "sunday", 3},
		{"function", "fucntion", 2},
		{"hello", "world", 4},
	}

	for _, tt := range tests {
		got := LevenshteinDistance(tt.s1, tt.s2)
		if got != tt.want {
			t.Errorf("LevenshteinDistance(%q, %q) = %d, want %d", tt.s1, tt.s2, got, tt.want)
		}
	}
}

func TestLevenshteinDistance_CaseInsensitive(t *testing.T) {
	got := LevenshteinDistance("Hello", "hello")
	if got != 0 {
		t.Errorf("LevenshteinDistance should be case-insensitive, got %d want 0", got)
	}
}

func TestJaroSimilarity(t *testing.T) {
	tests := []struct {
		s1, s2 string
		min    float64
		max    float64
	}{
		{"abc", "abc", 1.0, 1.0},
		{"", "abc", 0.0, 0.0},
		{"abc", "", 0.0, 0.0},
		{"martha", "marhta", 0.9, 1.0},
		{"dixon", "dicksonx", 0.7, 0.9},
		{"function", "function", 1.0, 1.0},
	}

	for _, tt := range tests {
		got := JaroSimilarity(tt.s1, tt.s2)
		if got < tt.min || got > tt.max {
			t.Errorf("JaroSimilarity(%q, %q) = %f, want [%f, %f]", tt.s1, tt.s2, got, tt.min, tt.max)
		}
	}
}

func TestJaroWinklerSimilarity(t *testing.T) {
	tests := []struct {
		s1, s2 string
		min    float64
		max    float64
	}{
		{"abc", "abc", 1.0, 1.0},
		{"function", "fucntion", 0.8, 1.0},
		{"parse", "parseJSON", 0.5, 1.0},
		{"hello", "world", 0.0, 0.5},
	}

	for _, tt := range tests {
		got := JaroWinklerSimilarity(tt.s1, tt.s2)
		if got < tt.min || got > tt.max {
			t.Errorf("JaroWinklerSimilarity(%q, %q) = %f, want [%f, %f]", tt.s1, tt.s2, got, tt.min, tt.max)
		}
	}
}

func TestJaroWinkler_HigherForCommonPrefix(t *testing.T) {
	jaro := JaroSimilarity("prefixabc", "prefixxyz")
	winkler := JaroWinklerSimilarity("prefixabc", "prefixxyz")
	if winkler <= jaro {
		t.Errorf("Jaro-Winkler should be >= Jaro for strings with common prefix: jaro=%f, winkler=%f", jaro, winkler)
	}
}

func TestFuzzyScore(t *testing.T) {
	tests := []struct {
		query, target string
		minScore      float64
	}{
		{"function", "function", 1.0},
		{"func", "function", 0.8},
		{"fucntion", "function", 0.7},
		{"parse", "parseJSON", 0.8},
		{"abc", "xyz", 0.0},
		{"hello", "hello_world", 0.5},
	}

	for _, tt := range tests {
		got := FuzzyScore(tt.query, tt.target)
		if got < tt.minScore {
			t.Errorf("FuzzyScore(%q, %q) = %f, want >= %f", tt.query, tt.target, got, tt.minScore)
		}
	}
}

func TestFuzzyScore_ExactMatch(t *testing.T) {
	got := FuzzyScore("function", "function")
	if got != 1.0 {
		t.Errorf("FuzzyScore exact match should be 1.0, got %f", got)
	}
}

func TestFuzzyScore_CaseInsensitive(t *testing.T) {
	got := FuzzyScore("Function", "function")
	if got != 1.0 {
		t.Errorf("FuzzyScore should be case-insensitive, got %f", got)
	}
}

func TestFuzzyFindAll(t *testing.T) {
	candidates := []string{"function", "func", "factorial", "format", "xyz"}
	results := FuzzyFindAll("func", candidates, 0.4)

	if len(results) == 0 {
		t.Fatal("FuzzyFindAll should return results")
	}

	// Results should be sorted by score descending
	for i := 1; i < len(results); i++ {
		if results[i].Score > results[i-1].Score {
			t.Errorf("Results not sorted: [%d]=%f > [%d]=%f", i, results[i].Score, i-1, results[i-1].Score)
		}
	}

	// "func" should be the top result
	if results[0].Text != "func" {
		t.Errorf("Top result should be 'func', got %q", results[0].Text)
	}
}

func TestFuzzyFindAll_EmptyQuery(t *testing.T) {
	candidates := []string{"abc", "def"}
	results := FuzzyFindAll("", candidates, 0.5)
	// Empty query should match some things with partial score
	// Just verify it doesn't panic
	_ = results
}

func TestFuzzyFindAll_NoMatch(t *testing.T) {
	candidates := []string{"xyz", "abc"}
	results := FuzzyFindAll("zzzzzzzzzz", candidates, 0.9)
	if len(results) != 0 {
		t.Errorf("Expected no results for very different query, got %d", len(results))
	}
}
