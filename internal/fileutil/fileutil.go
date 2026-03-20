package fileutil

import "strings"

// skipDirs is the canonical set of directories to skip during file traversal
var skipDirs = map[string]bool{
	".git":         true,
	".svn":         true,
	".hg":          true,
	"node_modules": true,
	"vendor":       true,
	".vscode":      true,
	".idea":        true,
	"__pycache__":  true,
	".venv":        true,
}

// ShouldSkipDir checks if a directory should be skipped during file traversal
func ShouldSkipDir(name string) bool {
	return skipDirs[name] || strings.HasPrefix(name, ".")
}
