package fileutil

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// WalkFunc is the callback for parallel walking
type WalkFunc func(path string, d os.DirEntry, err error) error

// ParallelWalk walks the directory tree, calling fn for each entry.
// Subdirectories at the same level are walked in parallel.
func ParallelWalk(root string, shouldSkipDir func(name, relPath string, isDir bool) bool, fn WalkFunc) error {
	info, err := os.Lstat(root)
	if err != nil {
		return fn(root, nil, err)
	}

	d := dirEntryFromInfo(info)
	if !info.IsDir() {
		return fn(root, d, nil)
	}

	return walkDir(root, "", shouldSkipDir, fn)
}

func walkDir(absDir, relDir string, shouldSkipDir func(string, string, bool) bool, fn WalkFunc) error {
	entries, err := os.ReadDir(absDir)
	if err != nil {
		return fn(absDir, nil, err)
	}

	// Process files in this directory
	for _, entry := range entries {
		if !entry.IsDir() {
			absPath := filepath.Join(absDir, entry.Name())
			if err := fn(absPath, entry, nil); err != nil {
				return err
			}
		}
	}

	// Process subdirectories in parallel
	var (
		mu       sync.Mutex
		firstErr error
		wg       sync.WaitGroup
		sem      = make(chan struct{}, runtime.NumCPU())
	)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		relPath := filepath.Join(relDir, name)
		absPath := filepath.Join(absDir, name)

		if shouldSkipDir(name, relPath, true) {
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(abs, rel string) {
			defer wg.Done()
			defer func() { <-sem }()

			if err := walkDir(abs, rel, shouldSkipDir, fn); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}(absPath, relPath)
	}

	wg.Wait()
	return firstErr
}

// dirEntryFromInfo wraps os.FileInfo as os.DirEntry
type fileInfoEntry struct {
	info os.FileInfo
}

func (e fileInfoEntry) Name() string               { return e.info.Name() }
func (e fileInfoEntry) IsDir() bool                { return e.info.IsDir() }
func (e fileInfoEntry) Type() os.FileMode          { return e.info.Mode().Type() }
func (e fileInfoEntry) Info() (os.FileInfo, error) { return e.info, nil }

func dirEntryFromInfo(info os.FileInfo) os.DirEntry {
	return fileInfoEntry{info: info}
}
