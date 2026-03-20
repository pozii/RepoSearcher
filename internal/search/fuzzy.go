package search

import (
	"math"
	"strings"
)

// LevenshteinDistance calculates the minimum edit distance between two strings
func LevenshteinDistance(s1, s2 string) int {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	m := len(s1)
	n := len(s2)

	if m == 0 {
		return n
	}
	if n == 0 {
		return m
	}

	// Create matrix
	matrix := make([][]int, m+1)
	for i := range matrix {
		matrix[i] = make([]int, n+1)
	}

	// Initialize first row and column
	for i := 0; i <= m; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= n; j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[m][n]
}

// JaroSimilarity calculates the Jaro similarity between two strings
func JaroSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	len1 := len(s1)
	len2 := len(s2)

	if len1 == 0 || len2 == 0 {
		return 0.0
	}

	// Maximum distance for matching
	matchDistance := int(math.Max(float64(len1), float64(len2))/2.0 - 1)

	// Match flags
	s1Matches := make([]bool, len1)
	s2Matches := make([]bool, len2)

	matches := 0
	transpositions := 0

	// Find matches
	for i := 0; i < len1; i++ {
		low := int(math.Max(0, float64(i-matchDistance)))
		high := int(math.Min(float64(len2-1), float64(i+matchDistance)))

		for j := low; j <= high; j++ {
			if s2Matches[j] || s1[i] != s2[j] {
				continue
			}
			s1Matches[i] = true
			s2Matches[j] = true
			matches++
			break
		}
	}

	if matches == 0 {
		return 0.0
	}

	// Count transpositions
	k := 0
	for i := 0; i < len1; i++ {
		if !s1Matches[i] {
			continue
		}
		for !s2Matches[k] {
			k++
		}
		if s1[i] != s2[k] {
			transpositions++
		}
		k++
	}

	jaro := (float64(matches)/float64(len1) +
		float64(matches)/float64(len2) +
		float64(matches-transpositions/2)/float64(matches)) / 3.0

	return jaro
}

// JaroWinklerSimilarity calculates the Jaro-Winkler similarity
func JaroWinklerSimilarity(s1, s2 string) float64 {
	jaro := JaroSimilarity(s1, s2)

	// Find common prefix
	prefix := 0
	minLen := int(math.Min(float64(len(s1)), float64(len(s2))))
	for i := 0; i < minLen && i < 4; i++ {
		if s1[i] == s2[i] {
			prefix++
		} else {
			break
		}
	}

	winkler := jaro + float64(prefix)*0.1*(1.0-jaro)
	return winkler
}

// FuzzyMatch checks if two strings match with fuzzy tolerance
func FuzzyMatch(query, target string, threshold float64) bool {
	// Exact match first
	if strings.EqualFold(query, target) {
		return true
	}

	// Contains match
	if strings.Contains(strings.ToLower(target), strings.ToLower(query)) {
		return true
	}

	// Levenshtein distance check
	distance := LevenshteinDistance(query, target)
	maxLen := int(math.Max(float64(len(query)), float64(len(target))))
	similarity := 1.0 - float64(distance)/float64(maxLen)

	return similarity >= threshold
}

// FuzzyScore returns a similarity score between 0 and 1
func FuzzyScore(query, target string) float64 {
	query = strings.ToLower(query)
	target = strings.ToLower(target)

	if query == target {
		return 1.0
	}

	if strings.Contains(target, query) {
		return 0.9
	}

	if strings.HasPrefix(target, query) {
		return 0.85
	}

	jaroWinkler := JaroWinklerSimilarity(query, target)

	levenshtein := LevenshteinDistance(query, target)
	maxLen := int(math.Max(float64(len(query)), float64(len(target))))
	levSimilarity := 1.0 - float64(levenshtein)/float64(maxLen)

	// Weighted average
	return (jaroWinkler*0.6 + levSimilarity*0.4)
}

// FuzzyFindAll finds all matches above the threshold and returns sorted results
func FuzzyFindAll(query string, candidates []string, threshold float64) []FuzzyResult {
	var results []FuzzyResult

	for _, candidate := range candidates {
		score := FuzzyScore(query, candidate)
		if score >= threshold {
			results = append(results, FuzzyResult{
				Text:  candidate,
				Score: score,
			})
		}
	}

	// Sort by score descending
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

// FuzzyResult represents a fuzzy match result
type FuzzyResult struct {
	Text  string
	Score float64
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}
