package lsp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractGoSymbols(t *testing.T) {
	tmpDir := t.TempDir()
	code := `package main

import "fmt"

type Config struct {
	Name    string
	Timeout int
}

type Engine interface {
	Search(query string) error
}

func main() {
	fmt.Println("hello")
}

func NewEngine() {
	fmt.Println("created")
}

func Helper(a int, b string) {
}

var globalVar = "test"

const MaxResults = 100
`
	file := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(file, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	extractor := NewSymbolExtractor()
	symbols, err := extractor.extractGoSymbols(file)
	if err != nil {
		t.Fatalf("extractGoSymbols failed: %v", err)
	}

	if len(symbols) == 0 {
		t.Fatal("Should extract symbols from Go file")
	}

	// Check for specific symbols
	foundTypes := make(map[string]string)
	for _, sym := range symbols {
		foundTypes[sym.Name] = sym.Type
	}

	if typ, ok := foundTypes["Config"]; !ok || typ != "struct" {
		t.Errorf("Expected 'Config' struct, got: %v", foundTypes["Config"])
	}
	if typ, ok := foundTypes["Engine"]; !ok || typ != "interface" {
		t.Errorf("Expected 'Engine' interface, got: %v", foundTypes["Engine"])
	}
	if typ, ok := foundTypes["main"]; !ok || typ != "function" {
		t.Errorf("Expected 'main' function, got: %v", foundTypes["main"])
	}
	if typ, ok := foundTypes["NewEngine"]; !ok || typ != "function" {
		t.Errorf("Expected 'NewEngine' function, got: %v", foundTypes["NewEngine"])
	}
}

func TestExtractPythonSymbols(t *testing.T) {
	tmpDir := t.TempDir()
	code := `class MyClass:
    def __init__(self):
        pass

    def process(self, data):
        return data

def helper():
    pass

global_var = 42
`
	file := filepath.Join(tmpDir, "main.py")
	if err := os.WriteFile(file, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	extractor := NewSymbolExtractor()
	symbols, err := extractor.extractPythonSymbols(file)
	if err != nil {
		t.Fatalf("extractPythonSymbols failed: %v", err)
	}

	if len(symbols) == 0 {
		t.Fatal("Should extract symbols from Python file")
	}

	foundTypes := make(map[string]string)
	for _, sym := range symbols {
		foundTypes[sym.Name] = sym.Type
	}

	if typ, ok := foundTypes["MyClass"]; !ok || typ != "class" {
		t.Errorf("Expected 'MyClass' class, got: %v", foundTypes["MyClass"])
	}
}

func TestExtractJavaScriptSymbols(t *testing.T) {
	tmpDir := t.TempDir()
	code := `function processData(data) {
    return data;
}

const fetchUser = async (id) => {
    return id;
};

class UserService {
    constructor() {}
}

type Config = {
    name: string;
};
`
	file := filepath.Join(tmpDir, "main.ts")
	if err := os.WriteFile(file, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	extractor := NewSymbolExtractor()
	symbols, err := extractor.extractJavaScriptSymbols(file)
	if err != nil {
		t.Fatalf("extractJavaScriptSymbols failed: %v", err)
	}

	if len(symbols) == 0 {
		t.Fatal("Should extract symbols from JS/TS file")
	}
}

func TestFindSymbol(t *testing.T) {
	extractor := NewSymbolExtractor()
	index := SymbolIndex{
		"SearchEngine": []Symbol{{Name: "SearchEngine", Type: "struct", File: "engine.go", Line: 10}},
		"Search":       []Symbol{{Name: "Search", Type: "method", File: "engine.go", Line: 20}},
		"Config":       []Symbol{{Name: "Config", Type: "struct", File: "config.go", Line: 5}},
	}

	results := extractor.FindSymbol(index, "Search")
	if len(results) != 2 {
		t.Errorf("FindSymbol('Search') returned %d results, want 2", len(results))
	}

	results = extractor.FindSymbol(index, "Config")
	if len(results) != 1 {
		t.Errorf("FindSymbol('Config') returned %d results, want 1", len(results))
	}

	results = extractor.FindSymbol(index, "NonExistent")
	if len(results) != 0 {
		t.Errorf("FindSymbol('NonExistent') returned %d results, want 0", len(results))
	}
}

func TestFindDefinition(t *testing.T) {
	extractor := NewSymbolExtractor()
	index := SymbolIndex{
		"Engine": []Symbol{
			{Name: "Engine", Type: "struct", File: "engine.go", Line: 10},
			{Name: "Engine", Type: "interface", File: "types.go", Line: 5},
		},
	}

	defs := extractor.FindDefinition(index, "Engine")
	if len(defs) != 2 {
		t.Errorf("FindDefinition('Engine') returned %d results, want 2", len(defs))
	}

	defs = extractor.FindDefinition(index, "Unknown")
	if len(defs) != 0 {
		t.Errorf("FindDefinition('Unknown') should return 0 results")
	}
}

func TestIsSupportedFile(t *testing.T) {
	tests := []struct {
		path string
		exts []string
		want bool
	}{
		{"main.go", nil, true},
		{"main.py", nil, true},
		{"main.js", nil, true},
		{"main.ts", nil, true},
		{"main.java", nil, true},
		{"main.rs", nil, true},
		{"main.txt", nil, false},
		{"main.md", nil, false},
		{"main.go", []string{".go"}, true},
		{"main.py", []string{".go"}, false},
		{"main.go", []string{".go", ".py"}, true},
	}

	for _, tt := range tests {
		got := isSupportedFile(tt.path, tt.exts)
		if got != tt.want {
			t.Errorf("isSupportedFile(%q, %v) = %v, want %v", tt.path, tt.exts, got, tt.want)
		}
	}
}

func TestFindReferences_WordBoundary(t *testing.T) {
	tmpDir := t.TempDir()
	code := `package test
func Run() {}
func RunE() {}
func Running() {}
func cronRun() {}
var Run = 1
func main() { Run(); RunE() }
`
	os.WriteFile(filepath.Join(tmpDir, "code.go"), []byte(code), 0644)

	extractor := NewSymbolExtractor()
	refs, err := extractor.FindReferences(tmpDir, "Run", nil)
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	// Should match standalone "Run" only, not RunE, Running, cronRun
	for _, ref := range refs {
		if ref.Symbol.Name != "Run" {
			t.Errorf("Unexpected reference: %s", ref.Symbol.Name)
		}
	}

	// Should find Run in: func Run(), var Run, Run()
	// Should NOT find: RunE, Running, cronRun
	if len(refs) < 3 {
		t.Errorf("Expected at least 3 references for standalone 'Run', got %d", len(refs))
	}

	// Verify no references from lines that ONLY contain RunE/Running/cronRun
	for _, ref := range refs {
		line := strings.TrimSpace(ref.Content)
		if line == "func RunE() {}" || line == "func Running() {}" || line == "func cronRun() {}" {
			t.Errorf("Should not match non-standalone references: %q", line)
		}
	}
}

func TestFindReferences_DefinitionCategory(t *testing.T) {
	tmpDir := t.TempDir()
	code := `package test
func MyFunc() {}
var MyFunc = 1
type MyFunc struct {}
x := MyFunc()
`
	os.WriteFile(filepath.Join(tmpDir, "code.go"), []byte(code), 0644)

	extractor := NewSymbolExtractor()
	refs, err := extractor.FindReferences(tmpDir, "MyFunc", nil)
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	defCount := 0
	usageCount := 0
	for _, ref := range refs {
		if ref.Category == "definition" {
			defCount++
		} else {
			usageCount++
		}
	}

	if defCount < 2 {
		t.Errorf("Expected at least 2 definition references, got %d", defCount)
	}
}

func TestExtractSymbols_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a Go file
	goCode := `package test
func Hello() {}
type MyStruct struct {}
`
	os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte(goCode), 0644)

	// Create a Python file
	pyCode := `def hello():
    pass

class MyClass:
    pass
`
	os.WriteFile(filepath.Join(tmpDir, "test.py"), []byte(pyCode), 0644)

	extractor := NewSymbolExtractor()
	index, err := extractor.ExtractSymbols(tmpDir, nil)
	if err != nil {
		t.Fatalf("ExtractSymbols failed: %v", err)
	}

	if len(index) == 0 {
		t.Error("Should extract symbols from multiple files")
	}

	// Should find symbols from both files
	if _, ok := index["Hello"]; !ok {
		t.Error("Should find 'Hello' symbol from Go file")
	}
	if _, ok := index["hello"]; !ok {
		t.Error("Should find 'hello' symbol from Python file")
	}
}
