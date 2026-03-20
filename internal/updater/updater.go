package updater

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
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

// currentVersion holds the binary version, set by the cmd package at startup
var currentVersion = "v1.0.0"

// SetCurrentVersion sets the current binary version (called by cmd at init)
func SetCurrentVersion(v string) {
	currentVersion = v
}

// GetCurrentVersion returns the current binary version
func GetCurrentVersion() string {
	return currentVersion
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

	// Get expected checksum
	checksumURL := getChecksumURL(release)
	var expectedChecksum string
	if checksumURL != "" {
		expectedChecksum, err = downloadExpectedChecksum(checksumURL, assetURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not verify checksum: %v\n", err)
		}
	}

	// Download new binary
	tmpFile, err := downloadBinary(assetURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer os.Remove(tmpFile)

	// Verify checksum if available
	if expectedChecksum != "" {
		actualChecksum, err := computeSHA256(tmpFile)
		if err != nil {
			return fmt.Errorf("failed to compute checksum: %w", err)
		}
		if actualChecksum != expectedChecksum {
			return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
		}
	}

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

// computeSHA256 computes the SHA256 hash of a file
func computeSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// getChecksumURL returns the URL of the checksums file asset in a release
func getChecksumURL(release *Release) string {
	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)
		if strings.Contains(name, "checksums") || strings.Contains(name, "sha256") {
			return asset.BrowserDownloadURL
		}
	}
	return ""
}

// downloadExpectedChecksum downloads the checksums file and extracts the expected hash
// for the given asset URL. The checksums file format: "<hash>  <filename>"
func downloadExpectedChecksum(checksumURL, assetURL string) (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(checksumURL)
	if err != nil {
		return "", fmt.Errorf("failed to download checksums: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksums download failed with status %d", resp.StatusCode)
	}

	// Extract filename from asset URL
	assetFilename := assetURL[strings.LastIndex(assetURL, "/")+1:]

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			hash := parts[0]
			filename := parts[len(parts)-1]
			if strings.Contains(filename, assetFilename) || assetFilename == filename {
				return strings.ToLower(hash), nil
			}
		}
	}

	return "", fmt.Errorf("checksum for %s not found in checksums file", assetFilename)
}

func isNewerVersion(new, old string) bool {
	new = strings.TrimPrefix(new, "v")
	old = strings.TrimPrefix(old, "v")

	newParts := strings.Split(new, ".")
	oldParts := strings.Split(old, ".")

	for i := 0; i < len(newParts) && i < len(oldParts); i++ {
		n, _ := strconv.Atoi(newParts[i])
		o, _ := strconv.Atoi(oldParts[i])
		if n > o {
			return true
		}
		if n < o {
			return false
		}
	}

	return len(newParts) > len(oldParts)
}
