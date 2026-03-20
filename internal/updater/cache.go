package updater

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const cacheInterval = 24 * time.Hour

// updateCheckCache stores the last update check result
type updateCheckCache struct {
	LastCheck time.Time `json:"last_check"`
	HasUpdate bool      `json:"has_update"`
	TagName   string    `json:"tag_name"`
}

// getCacheDir returns ~/.repo-searcher/
func getCacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".repo-searcher")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// cacheFilePath returns the path to the cache file
func cacheFilePath() (string, error) {
	dir, err := getCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "update-check.json"), nil
}

// loadCache reads the cached update check result
func loadCache() (*updateCheckCache, error) {
	path, err := cacheFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cache updateCheckCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}
	return &cache, nil
}

// saveCache writes the update check result to disk
func saveCache(cache *updateCheckCache) error {
	path, err := cacheFilePath()
	if err != nil {
		return err
	}
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// CheckForUpdateCached checks for updates using a daily cache
// to avoid repeated network requests on every CLI invocation.
func CheckForUpdateCached() (*Release, bool, error) {
	cache, err := loadCache()
	if err != nil {
		cache = &updateCheckCache{}
	}

	// If cache is fresh (< 24 hours), use it
	if !cache.LastCheck.IsZero() && time.Since(cache.LastCheck) < cacheInterval {
		if cache.HasUpdate {
			return &Release{TagName: cache.TagName}, true, nil
		}
		return nil, false, nil
	}

	// Cache expired or missing — do the network check
	release, hasUpdate, err := CheckForUpdate()
	if err != nil {
		// On error, return silently
		return nil, false, nil
	}

	// Save result to cache
	newCache := &updateCheckCache{
		LastCheck: time.Now(),
		HasUpdate: hasUpdate,
	}
	if hasUpdate && release != nil {
		newCache.TagName = release.TagName
	}
	_ = saveCache(newCache)

	return release, hasUpdate, nil
}
