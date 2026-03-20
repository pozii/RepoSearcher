package search

import (
	"testing"
)

func TestShouldIgnoreFile_NoExtensions(t *testing.T) {
	got := ShouldIgnoreFile("file.go", nil)
	if got {
		t.Error("ShouldIgnoreFile with nil extensions should return false")
	}

	got = ShouldIgnoreFile("file.go", []string{})
	if got {
		t.Error("ShouldIgnoreFile with empty extensions should return false")
	}
}

func TestShouldIgnoreFile_MatchingExtension(t *testing.T) {
	tests := []struct {
		path       string
		extensions []string
		want       bool
	}{
		{"main.go", []string{".go"}, false},
		{"main.py", []string{".go"}, true},
		{"main.go", []string{".go", ".py"}, false},
		{"main.py", []string{".go", ".py"}, false},
		{"main.js", []string{".go", ".py"}, true},
		{"/path/to/file.go", []string{".go"}, false},
		{"file.GO", []string{".go"}, false},
		{"file.go", []string{".GO"}, false},
		{"file.goo", []string{".go"}, true},
	}

	for _, tt := range tests {
		got := ShouldIgnoreFile(tt.path, tt.extensions)
		if got != tt.want {
			t.Errorf("ShouldIgnoreFile(%q, %v) = %v, want %v", tt.path, tt.extensions, got, tt.want)
		}
	}
}

func TestCompilePattern_Keyword(t *testing.T) {
	pattern, err := CompilePattern("hello", false, false)
	if err != nil {
		t.Fatalf("CompilePattern failed: %v", err)
	}

	if !pattern.MatchString("say hello world") {
		t.Error("pattern should match 'say hello world'")
	}
	if pattern.MatchString("say HELLO world") {
		t.Error("case-sensitive pattern should not match 'say HELLO world'")
	}
}

func TestCompilePattern_KeywordIgnoreCase(t *testing.T) {
	pattern, err := CompilePattern("hello", false, true)
	if err != nil {
		t.Fatalf("CompilePattern failed: %v", err)
	}

	if !pattern.MatchString("say HELLO world") {
		t.Error("case-insensitive pattern should match 'say HELLO world'")
	}
	if !pattern.MatchString("say Hello world") {
		t.Error("case-insensitive pattern should match 'say Hello world'")
	}
}

func TestCompilePattern_Regex(t *testing.T) {
	pattern, err := CompilePattern(`func\s+\w+`, true, false)
	if err != nil {
		t.Fatalf("CompilePattern failed: %v", err)
	}

	if !pattern.MatchString("func main()") {
		t.Error("regex pattern should match 'func main()'")
	}
	if pattern.MatchString("function main()") {
		t.Error("regex pattern should not match 'function main()'")
	}
}

func TestCompilePattern_RegexIgnoreCase(t *testing.T) {
	pattern, err := CompilePattern(`func\s+\w+`, true, true)
	if err != nil {
		t.Fatalf("CompilePattern failed: %v", err)
	}

	if !pattern.MatchString("FUNC main()") {
		t.Error("case-insensitive regex should match 'FUNC main()'")
	}
}

func TestCompilePattern_InvalidRegex(t *testing.T) {
	_, err := CompilePattern(`[invalid`, true, false)
	if err == nil {
		t.Error("CompilePattern should return error for invalid regex")
	}
}

func TestCompilePattern_SpecialChars(t *testing.T) {
	pattern, err := CompilePattern("a.b", false, false)
	if err != nil {
		t.Fatalf("CompilePattern failed: %v", err)
	}

	if !pattern.MatchString("a.b") {
		t.Error("keyword pattern should match literal 'a.b'")
	}
	if pattern.MatchString("axb") {
		t.Error("keyword pattern should not match 'axb' (dot is literal)")
	}
}

func TestShouldIgnoreFileByGlobs(t *testing.T) {
	tests := []struct {
		path        string
		include     []string
		exclude     []string
		wantIgnored bool
	}{
		// Exclude patterns
		{"vendor/lib.go", nil, []string{"vendor/**"}, true},
		{"src/main.go", nil, []string{"vendor/**"}, false},
		{"test_foo.go", nil, []string{"test_*"}, true},
		{"main_test.go", nil, []string{"*_test.go"}, true},
		{"src/test_data.go", nil, []string{"**/test_*"}, true},

		// Include patterns
		{"src/main.go", []string{"src/*"}, nil, false},
		{"lib/main.go", []string{"src/*"}, nil, true}, // not in src/
		{"src/main.go", []string{"src/*.go"}, nil, false},
		{"src/main.py", []string{"src/*.go"}, nil, true}, // wrong extension

		// Both include and exclude
		{"src/main.go", []string{"src/**"}, []string{"**/test_*"}, false},
		{"src/test_main.go", []string{"src/**"}, []string{"**/test_*"}, true},

		// No patterns
		{"anything.go", nil, nil, false},
	}

	for _, tt := range tests {
		got := ShouldIgnoreFileByGlobs(tt.path, tt.include, tt.exclude)
		if got != tt.wantIgnored {
			t.Errorf("ShouldIgnoreFileByGlobs(%q, %v, %v) = %v, want %v",
				tt.path, tt.include, tt.exclude, got, tt.wantIgnored)
		}
	}
}
