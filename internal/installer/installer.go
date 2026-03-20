package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	AppName    = "repo-searcher"
	BinaryName = "repo-searcher"
)

// GetInstallDir returns the install directory based on platform
func GetInstallDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, "AppData", "Local", "repo-searcher", "bin"), nil
	default:
		return "/usr/local/bin", nil
	}
}

// GetBinaryPath returns the current binary path
func GetBinaryPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

// Install adds the binary to PATH
func Install() error {
	binaryPath, err := GetBinaryPath()
	if err != nil {
		return fmt.Errorf("failed to get binary path: %w", err)
	}

	installDir, err := GetInstallDir()
	if err != nil {
		return fmt.Errorf("failed to get install dir: %w", err)
	}

	// Create install directory
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install dir: %w", err)
	}

	targetPath := filepath.Join(installDir, BinaryName)
	if runtime.GOOS == "windows" {
		targetPath += ".exe"
	}

	// Copy binary
	if err := copyFile(binaryPath, targetPath); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	// Add to PATH
	if err := addToPATH(installDir); err != nil {
		return fmt.Errorf("failed to add to PATH: %w", err)
	}

	fmt.Printf("Installed to: %s\n", installDir)
	fmt.Printf("Binary path: %s\n", targetPath)
	return nil
}

// Uninstall removes the binary from PATH and deletes installed files
func Uninstall() error {
	installDir, err := GetInstallDir()
	if err != nil {
		return fmt.Errorf("failed to get install dir: %w", err)
	}

	targetPath := filepath.Join(installDir, BinaryName)
	if runtime.GOOS == "windows" {
		targetPath += ".exe"
	}

	// Remove binary
	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove binary: %w", err)
	}

	// Remove from PATH
	if err := removeFromPATH(installDir); err != nil {
		return fmt.Errorf("failed to remove from PATH: %w", err)
	}

	// Remove install directory if empty
	os.Remove(installDir)

	fmt.Printf("Uninstalled from: %s\n", installDir)
	return nil
}

// IsInstalled checks if the binary is installed in the install directory
func IsInstalled() bool {
	installDir, err := GetInstallDir()
	if err != nil {
		return false
	}

	targetPath := filepath.Join(installDir, BinaryName)
	if runtime.GOOS == "windows" {
		targetPath += ".exe"
	}

	_, err = os.Stat(targetPath)
	return err == nil
}

// IsInPATH checks if install directory is in PATH
func IsInPATH() bool {
	installDir, err := GetInstallDir()
	if err != nil {
		return false
	}

	pathEnv := os.Getenv("PATH")
	for _, p := range filepath.SplitList(pathEnv) {
		if p == installDir {
			return true
		}
	}
	return false
}

// Platform-specific PATH management
func addToPATH(dir string) error {
	switch runtime.GOOS {
	case "windows":
		return addWindowsPath(dir)
	case "darwin", "linux":
		return addUnixPath(dir)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func removeFromPATH(dir string) error {
	switch runtime.GOOS {
	case "windows":
		return removeWindowsPath(dir)
	case "darwin", "linux":
		return removeUnixPath(dir)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// Windows PATH management
func addWindowsPath(dir string) error {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf("[Environment]::SetEnvironmentVariable('PATH', '%s;' + [Environment]::GetEnvironmentVariable('PATH', 'User'), 'User')", dir))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func removeWindowsPath(dir string) error {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf("[Environment]::SetEnvironmentVariable('PATH', ([Environment]::GetEnvironmentVariable('PATH', 'User')).Replace('%s;', ''), 'User')", dir))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Unix PATH management
func addUnixPath(dir string) error {
	// Create symlink
	binPath := filepath.Join(dir, BinaryName)
	currentPath, err := os.Executable()
	if err != nil {
		return err
	}

	// Remove existing symlink if any
	os.Remove(binPath)

	return os.Symlink(currentPath, binPath)
}

func removeUnixPath(dir string) error {
	binPath := filepath.Join(dir, BinaryName)
	return os.Remove(binPath)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0755)
}
