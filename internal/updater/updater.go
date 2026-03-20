package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	RepoOwner = "pozii"
	RepoName  = "RepoSearcher"
	BaseURL   = "https://api.github.com"
)

// Release represents a GitHub release
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// GetCurrentVersion returns the current binary version
func GetCurrentVersion() string {
	return "v1.0.0"
}

// CheckForUpdate checks if a newer version is available
func CheckForUpdate() (*Release, bool, error) {
	release, err := getLatestRelease()
	if err != nil {
		return nil, false, err
	}

	currentVersion := GetCurrentVersion()
	if isNewerVersion(release.TagName, currentVersion) {
		return release, true, nil
	}

	return release, false, nil
}

// Update downloads and replaces the current binary
func Update(release *Release) error {
	// Find the correct asset for current platform
	assetURL, err := getAssetURL(release)
	if err != nil {
		return err
	}

	// Download new binary
	tmpFile, err := downloadBinary(assetURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer os.Remove(tmpFile)

	// Replace current binary
	if err := replaceBinary(tmpFile); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	return nil
}

func getLatestRelease() (*Release, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", BaseURL, RepoOwner, RepoName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func getAssetURL(release *Release) (string, error) {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	// Platform-specific file patterns
	patterns := []string{
		fmt.Sprintf("repo-searcher-%s-%s", osName, archName),
	}

	for _, pattern := range patterns {
		for _, asset := range release.Assets {
			if strings.Contains(strings.ToLower(asset.Name), pattern) {
				return asset.BrowserDownloadURL, nil
			}
		}
	}

	return "", fmt.Errorf("no asset found for %s/%s", osName, archName)
}

func downloadBinary(url string) (string, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "repo-searcher-update-*")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func replaceBinary(newBinaryPath string) error {
	currentPath, err := os.Executable()
	if err != nil {
		return err
	}

	// Read new binary
	newBinary, err := os.ReadFile(newBinaryPath)
	if err != nil {
		return err
	}

	// Write new binary over current
	if err := os.WriteFile(currentPath, newBinary, 0755); err != nil {
		return err
	}

	return nil
}

func isNewerVersion(new, old string) bool {
	new = strings.TrimPrefix(new, "v")
	old = strings.TrimPrefix(old, "v")

	newParts := strings.Split(new, ".")
	oldParts := strings.Split(old, ".")

	for i := 0; i < len(newParts) && i < len(oldParts); i++ {
		if newParts[i] > oldParts[i] {
			return true
		}
		if newParts[i] < oldParts[i] {
			return false
		}
	}

	return len(newParts) > len(oldParts)
}
