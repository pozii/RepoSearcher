package search

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/pozii/RepoSearcher/pkg/models"
)

func BenchmarkLevenshteinDistance(b *testing.B) {
	pairs := [][2]string{
		{"function", "fucntion"},
		{"kitten", "sitting"},
		{"saturday", "sunday"},
		{"hello", "world"},
		{"a]bcdefghijklmnopqrstuvwxyz", "abcdefghijklmnopqrstuvwxyz"},
	}
	for _, pair := range pairs {
		b.Run(fmt.Sprintf("%s_%s", pair[0], pair[1]), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				LevenshteinDistance(pair[0], pair[1])
			}
		})
	}
}

func BenchmarkJaroSimilarity(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JaroSimilarity("martha", "marhta")
	}
}

func BenchmarkJaroWinklerSimilarity(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JaroWinklerSimilarity("function", "fucntion")
	}
}

func BenchmarkFuzzyScore(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FuzzyScore("func", "function")
	}
}

func BenchmarkFuzzyFindAll(b *testing.B) {
	candidates := make([]string, 1000)
	for i := range candidates {
		candidates[i] = fmt.Sprintf("identifier_%d", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FuzzyFindAll("identifier_5", candidates, 0.4)
	}
}

func BenchmarkCompilePattern_Keyword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CompilePattern("function", false, false)
	}
}

func BenchmarkCompilePattern_Regex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CompilePattern(`func\s+\w+\(`, true, false)
	}
}

func BenchmarkGetCompiledPattern_Cached(b *testing.B) {
	// Warm up cache
	GetCompiledPattern("benchmark_query", false, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetCompiledPattern("benchmark_query", false, false)
	}
}

// === Speed comparison benchmarks: old vs new approach ===

// createTestFile creates a test Go file with many lines
func createTestFile(b *testing.B) string {
	b.Helper()
	tmpDir := b.TempDir()
	file := filepath.Join(tmpDir, "big.go")

	var content []byte
	for i := 0; i < 5000; i++ {
		line := fmt.Sprintf("\tvar x%d = %d\n", i, i)
		content = append(content, []byte(line)...)
	}
	// Add some 'func' matches
	for i := 0; i < 50; i++ {
		line := fmt.Sprintf("\tfunc handler%d() {}\n", i)
		content = append(content, []byte(line)...)
	}
	// More non-matching lines
	for i := 0; i < 5000; i++ {
		line := fmt.Sprintf("\tvar y%d = %d\n", i, i)
		content = append(content, []byte(line)...)
	}

	os.WriteFile(file, content, 0644)
	return file
}

// OLD approach: bufio.Scanner + regexp.MatchString (line by line)
func searchFileOld(filePath string, pattern *regexp.Regexp) int {
	file, _ := os.Open(filePath)
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 64*1024)
	for scanner.Scan() {
		if pattern.MatchString(scanner.Text()) {
			count++
		}
	}
	return count
}

// NEW approach: os.ReadFile + bytes.Index (literal scan)
func searchFileNew(filePath string, pattern *FastPattern) int {
	results := searchFileBytes(filePath, pattern, models.SearchConfig{})
	return len(results)
}

func BenchmarkSearch_Literal_Old(b *testing.B) {
	file := createTestFile(b)
	pattern := regexp.MustCompile(regexp.QuoteMeta("func"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		searchFileOld(file, pattern)
	}
}

func BenchmarkSearch_Literal_New(b *testing.B) {
	file := createTestFile(b)
	pattern, _ := CompileFastPattern("func", false, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		searchFileNew(file, pattern)
	}
}

func BenchmarkSearch_Regex_Old(b *testing.B) {
	file := createTestFile(b)
	pattern := regexp.MustCompile(`func\s+\w+`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		searchFileOld(file, pattern)
	}
}

func BenchmarkSearch_Regex_New(b *testing.B) {
	file := createTestFile(b)
	pattern, _ := CompileFastPattern(`func\s+\w+`, true, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		searchFileNew(file, pattern)
	}
}
