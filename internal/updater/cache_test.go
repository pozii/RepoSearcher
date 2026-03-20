package updater

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestSaveAndLoadCache(t *testing.T) {
	// Create a temp dir and manually test round-trip
	cacheDir := t.TempDir()
	cacheFile := cacheDir + "/update-check.json"

	cache := &updateCheckCache{
		LastCheck: time.Now().Truncate(time.Second),
		HasUpdate: true,
		TagName:   "v2.0.0",
	}

	data, err := json.Marshal(cache)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		t.Fatal(err)
	}

	readData, err := os.ReadFile(cacheFile)
	if err != nil {
		t.Fatal(err)
	}
	var loaded updateCheckCache
	if err := json.Unmarshal(readData, &loaded); err != nil {
		t.Fatal(err)
	}

	if loaded.HasUpdate != cache.HasUpdate {
		t.Errorf("HasUpdate mismatch: got %v, want %v", loaded.HasUpdate, cache.HasUpdate)
	}
	if loaded.TagName != cache.TagName {
		t.Errorf("TagName mismatch: got %q, want %q", loaded.TagName, cache.TagName)
	}
}

func TestCacheInterval(t *testing.T) {
	if cacheInterval != 24*time.Hour {
		t.Errorf("cacheInterval = %v, want 24h", cacheInterval)
	}
}

func TestUpdateCheckCache_FreshCache(t *testing.T) {
	cache := &updateCheckCache{
		LastCheck: time.Now().Add(-1 * time.Hour),
		HasUpdate: true,
		TagName:   "v2.0.0",
	}

	if time.Since(cache.LastCheck) >= cacheInterval {
		t.Error("1 hour old cache should be considered fresh")
	}
}

func TestUpdateCheckCache_StaleCache(t *testing.T) {
	cache := &updateCheckCache{
		LastCheck: time.Now().Add(-25 * time.Hour),
		HasUpdate: false,
	}

	if time.Since(cache.LastCheck) < cacheInterval {
		t.Error("25 hour old cache should be considered stale")
	}
}
