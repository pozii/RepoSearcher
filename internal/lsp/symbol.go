package lsp

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Symbol represents a code symbol (function, struct, variable, type, etc.)
type Symbol struct {
	Name      string
	Type      string // "function", "method", "struct", "interface", "variable", "type", "const"
	File      string
	Line      int
	Column    int
	Signature string
}

// Reference represents a reference to a symbol
type Reference struct {
	Symbol   Symbol
	Line     int
	Column   int
	Content  string
	Category string // "definition", "usage", "implementation", "import"
}

// SymbolIndex is a map of symbol names to their definitions
type SymbolIndex map[string][]Symbol

// SymbolExtractor extracts symbols from code files
type SymbolExtractor struct{}

// NewSymbolExtractor creates a new SymbolExtractor
func NewSymbolExtractor() *SymbolExtractor {
	return &SymbolExtractor{}
}

// ExtractSymbols extracts all symbols from a path
func (e *SymbolExtractor) ExtractSymbols(root string, extensions []string) (SymbolIndex, error) {
	index := make(SymbolIndex)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if shouldSkipDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		if !isSupportedFile(path, extensions) {
			return nil
		}

		symbols, err := e.extractFromFile(path)
		if err != nil {
			return nil
		}

		for _, sym := range symbols {
			index[sym.Name] = append(index[sym.Name], sym)
		}

		return nil
	})

	return index, err
}

// extractFromFile extracts symbols from a single file
func (e *SymbolExtractor) extractFromFile(path string) ([]Symbol, error) {
	// Determine language from extension
	ext := filepath.Ext(path)
	switch ext {
	case ".go":
		return e.extractGoSymbols(path)
	case ".py":
		return e.extractPythonSymbols(path)
	case ".js", ".ts":
		return e.extractJavaScriptSymbols(path)
	default:
		// Generic extraction
		return e.extractGenericSymbols(path)
	}
}

// extractGoSymbols extracts Go-specific symbols
func (e *SymbolExtractor) extractGoSymbols(path string) ([]Symbol, error) {
	var symbols []Symbol

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Patterns for Go
	funcPattern := regexp.MustCompile(`^func\s+(\w+)\s*(\([^)]*\))?\s*(\([^)]*\))?\s*\{`)
	methodPattern := regexp.MustCompile(`^func\s+\([^)]+\)\s+(\w+)\s*(\([^)]*\))?\s*(\([^)]*\))?\s*\{`)
	structPattern := regexp.MustCompile(`^type\s+(\w+)\s+struct\s*\{`)
	interfacePattern := regexp.MustCompile(`^type\s+(\w+)\s+interface\s*\{`)
	typePattern := regexp.MustCompile(`^type\s+(\w+)\s+\w+`)
	varPattern := regexp.MustCompile(`^(var|const)\s+(\w+)\s*`)
	fieldPattern := regexp.MustCompile(`^\s+(\w+)\s+\w+`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Match function
		if match := funcPattern.FindStringSubmatch(line); match != nil {
			symbols = append(symbols, Symbol{
				Name:      match[1],
				Type:      "function",
				File:      path,
				Line:      lineNum,
				Signature: strings.TrimSpace(match[2]),
			})
		}

		// Match method
		if match := methodPattern.FindStringSubmatch(line); match != nil {
			symbols = append(symbols, Symbol{
				Name:      match[1],
				Type:      "method",
				File:      path,
				Line:      lineNum,
				Signature: strings.TrimSpace(match[2]),
			})
		}

		// Match struct
		if match := structPattern.FindStringSubmatch(line); match != nil {
			symbols = append(symbols, Symbol{
				Name: match[1],
				Type: "struct",
				File: path,
				Line: lineNum,
			})
		}

		// Match interface
		if match := interfacePattern.FindStringSubmatch(line); match != nil {
			symbols = append(symbols, Symbol{
				Name: match[1],
				Type: "interface",
				File: path,
				Line: lineNum,
			})
		}

		// Match type alias
		if match := typePattern.FindStringSubmatch(line); match != nil {
			// Skip if already matched as struct or interface
			if !structPattern.MatchString(line) && !interfacePattern.MatchString(line) {
				symbols = append(symbols, Symbol{
					Name: match[1],
					Type: "type",
					File: path,
					Line: lineNum,
				})
			}
		}

		// Match variable/const
		if match := varPattern.FindStringSubmatch(line); match != nil {
			symbols = append(symbols, Symbol{
				Name: match[2],
				Type: match[1],
				File: path,
				Line: lineNum,
			})
		}

		// Match struct fields
		if match := fieldPattern.FindStringSubmatch(line); match != nil {
			symbols = append(symbols, Symbol{
				Name: match[1],
				Type: "field",
				File: path,
				Line: lineNum,
			})
		}
	}

	return symbols, nil
}

// extractPythonSymbols extracts Python-specific symbols
func (e *SymbolExtractor) extractPythonSymbols(path string) ([]Symbol, error) {
	var symbols []Symbol

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Patterns for Python
	funcPattern := regexp.MustCompile(`^def\s+(\w+)\s*\(([^)]*)\)`)
	classPattern := regexp.MustCompile(`^class\s+(\w+)\s*\(?\s*\w*\s*\)?`)
	varPattern := regexp.MustCompile(`^(\w+)\s*=\s*`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Match function
		if match := funcPattern.FindStringSubmatch(line); match != nil {
			symbols = append(symbols, Symbol{
				Name:      match[1],
				Type:      "function",
				File:      path,
				Line:      lineNum,
				Signature: "(" + match[2] + ")",
			})
		}

		// Match class
		if match := classPattern.FindStringSubmatch(line); match != nil {
			symbols = append(symbols, Symbol{
				Name: match[1],
				Type: "class",
				File: path,
				Line: lineNum,
			})
		}

		// Match top-level variable
		if match := varPattern.FindStringSubmatch(line); match != nil {
			if match[1][0] != '_' { // Skip private variables
				symbols = append(symbols, Symbol{
					Name: match[1],
					Type: "variable",
					File: path,
					Line: lineNum,
				})
			}
		}
	}

	return symbols, nil
}

// extractJavaScriptSymbols extracts JS/TS-specific symbols
func (e *SymbolExtractor) extractJavaScriptSymbols(path string) ([]Symbol, error) {
	var symbols []Symbol

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Patterns for JS/TS
	funcPattern := regexp.MustCompile(`(?:function|const|let|var)\s+(\w+)\s*=?\s*(?:function)?\s*\(([^)]*)\)`)
	arrowFuncPattern := regexp.MustCompile(`(?:const|let)\s+(\w+)\s*=\s*(?:async\s+)?\(([^)]*)\)\s*=>`)
	classPattern := regexp.MustCompile(`(?:export\s+)?(?:class|interface)\s+(\w+)`)
	typePattern := regexp.MustCompile(`(?:type|interface)\s+(\w+)\s*(?:=|\{)`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Match function
		if match := funcPattern.FindStringSubmatch(line); match != nil {
			symbols = append(symbols, Symbol{
				Name:      match[1],
				Type:      "function",
				File:      path,
				Line:      lineNum,
				Signature: "(" + match[2] + ")",
			})
		}

		// Match arrow function
		if match := arrowFuncPattern.FindStringSubmatch(line); match != nil {
			symbols = append(symbols, Symbol{
				Name:      match[1],
				Type:      "function",
				File:      path,
				Line:      lineNum,
				Signature: "(" + match[2] + ")",
			})
		}

		// Match class/interface
		if match := classPattern.FindStringSubmatch(line); match != nil {
			symbols = append(symbols, Symbol{
				Name: match[1],
				Type: "class",
				File: path,
				Line: lineNum,
			})
		}

		// Match type
		if match := typePattern.FindStringSubmatch(line); match != nil {
			symbols = append(symbols, Symbol{
				Name: match[1],
				Type: "type",
				File: path,
				Line: lineNum,
			})
		}
	}

	return symbols, nil
}

// extractGenericSymbols extracts symbols from generic code files
func (e *SymbolExtractor) extractGenericSymbols(path string) ([]Symbol, error) {
	var symbols []Symbol

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Generic patterns
	funcPattern := regexp.MustCompile(`(?:(?:public|private|protected)\s+)?(?:static\s+)?(?:\w+\s+)+(\w+)\s*\(([^)]*)\)`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if match := funcPattern.FindStringSubmatch(line); match != nil {
			symbols = append(symbols, Symbol{
				Name:      match[1],
				Type:      "function",
				File:      path,
				Line:      lineNum,
				Signature: "(" + match[2] + ")",
			})
		}
	}

	return symbols, nil
}

// FindSymbol finds symbols matching a query
func (e *SymbolExtractor) FindSymbol(index SymbolIndex, query string) []Symbol {
	query = strings.ToLower(query)
	var results []Symbol

	for name, symbols := range index {
		if strings.Contains(strings.ToLower(name), query) {
			results = append(results, symbols...)
		}
	}

	return results
}

// FindDefinition finds the definition of a symbol
func (e *SymbolExtractor) FindDefinition(index SymbolIndex, name string) []Symbol {
	return index[name]
}

// FindReferences finds all references to a symbol
func (e *SymbolExtractor) FindReferences(root string, name string, extensions []string) ([]Reference, error) {
	var references []Reference

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if shouldSkipDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		if !isSupportedFile(path, extensions) {
			return nil
		}

		refs, err := e.findReferencesInFile(path, name)
		if err != nil {
			return nil
		}

		references = append(references, refs...)

		return nil
	})

	return references, err
}

// findReferencesInFile finds references in a single file
func (e *SymbolExtractor) findReferencesInFile(path string, name string) ([]Reference, error) {
	var references []Reference

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if strings.Contains(line, name) {
			// Determine category
			category := "usage"
			if strings.Contains(line, "func "+name) ||
				strings.Contains(line, "type "+name) ||
				strings.Contains(line, "var "+name) {
				category = "definition"
			}

			references = append(references, Reference{
				Symbol: Symbol{
					Name: name,
					File: path,
					Line: lineNum,
				},
				Line:     lineNum,
				Content:  strings.TrimSpace(line),
				Category: category,
			})
		}
	}

	return references, nil
}

// isSupportedFile checks if a file is supported for symbol extraction
func isSupportedFile(path string, extensions []string) bool {
	if len(extensions) > 0 {
		for _, ext := range extensions {
			if strings.HasSuffix(strings.ToLower(path), strings.ToLower(ext)) {
				return true
			}
		}
		return false
	}

	// Default supported extensions
	supportedExts := []string{".go", ".py", ".js", ".ts", ".java", ".c", ".cpp", ".h", ".rs"}
	for _, ext := range supportedExts {
		if strings.HasSuffix(strings.ToLower(path), ext) {
			return true
		}
	}

	return false
}

// shouldSkipDir checks if a directory should be skipped
func shouldSkipDir(name string) bool {
	skipDirs := map[string]bool{
		".git": true, ".svn": true, ".hg": true,
		"node_modules": true, "vendor": true,
		".vscode": true, ".idea": true,
		"__pycache__": true, ".venv": true,
	}
	return skipDirs[name] || strings.HasPrefix(name, ".")
}
