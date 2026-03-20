package updater

import (
	"testing"
)

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		new, old string
		want     bool
	}{
		{"v2.0.0", "v1.0.0", true},
		{"v1.1.0", "v1.0.0", true},
		{"v1.0.1", "v1.0.0", true},
		{"v1.0.0", "v1.0.0", false},
		{"v1.0.0", "v2.0.0", false},
		{"v1.0.0", "v1.1.0", false},
		{"v10.0.0", "v9.0.0", true},
		{"v1.10.0", "v1.9.0", true},
		{"2.0.0", "1.0.0", true},  // without v prefix
		{"v2.0.0", "1.0.0", true}, // mixed
		{"v1.0.0.1", "v1.0.0", true},
		{"v1.0", "v1.0.0", false},
	}

	for _, tt := range tests {
		got := isNewerVersion(tt.new, tt.old)
		if got != tt.want {
			t.Errorf("isNewerVersion(%q, %q) = %v, want %v", tt.new, tt.old, got, tt.want)
		}
	}
}

func TestGetCurrentVersion(t *testing.T) {
	// Default should return a non-empty version
	version := GetCurrentVersion()
	if version == "" {
		t.Error("GetCurrentVersion should not return empty string")
	}
}

func TestSetCurrentVersion(t *testing.T) {
	original := GetCurrentVersion()
	defer SetCurrentVersion(original)

	SetCurrentVersion("v9.9.9")
	if got := GetCurrentVersion(); got != "v9.9.9" {
		t.Errorf("GetCurrentVersion = %q after SetCurrentVersion, want 'v9.9.9'", got)
	}
}

func TestGetAssetURL(t *testing.T) {
	release := &Release{
		TagName: "v1.0.0",
		Assets: []Asset{
			{Name: "repo-searcher-linux-amd64", BrowserDownloadURL: "https://example.com/linux"},
			{Name: "repo-searcher-windows-amd64.exe", BrowserDownloadURL: "https://example.com/windows"},
			{Name: "repo-searcher-darwin-arm64", BrowserDownloadURL: "https://example.com/darwin"},
			{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums"},
		},
	}

	url, err := getAssetURL(release)
	if err != nil {
		t.Fatalf("getAssetURL failed: %v", err)
	}
	if url == "" {
		t.Error("getAssetURL should return non-empty URL")
	}
}

func TestGetAssetURL_NoMatch(t *testing.T) {
	release := &Release{
		TagName: "v1.0.0",
		Assets: []Asset{
			{Name: "unrelated-file.txt", BrowserDownloadURL: "https://example.com/file"},
		},
	}

	_, err := getAssetURL(release)
	if err == nil {
		t.Error("getAssetURL should return error when no matching asset")
	}
}

func TestGetChecksumURL(t *testing.T) {
	release := &Release{
		Assets: []Asset{
			{Name: "repo-searcher-linux-amd64", BrowserDownloadURL: "https://example.com/linux"},
			{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums"},
		},
	}

	url := getChecksumURL(release)
	if url != "https://example.com/checksums" {
		t.Errorf("getChecksumURL = %q, want 'https://example.com/checksums'", url)
	}
}

func TestGetChecksumURL_NoChecksum(t *testing.T) {
	release := &Release{
		Assets: []Asset{
			{Name: "repo-searcher-linux-amd64", BrowserDownloadURL: "https://example.com/linux"},
		},
	}

	url := getChecksumURL(release)
	if url != "" {
		t.Errorf("getChecksumURL should return empty when no checksum file, got %q", url)
	}
}

func TestConstants(t *testing.T) {
	if RepoOwner == "" {
		t.Error("RepoOwner should not be empty")
	}
	if RepoName == "" {
		t.Error("RepoName should not be empty")
	}
	if BaseURL == "" {
		t.Error("BaseURL should not be empty")
	}
}
