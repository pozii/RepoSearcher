package cmd

import (
	"fmt"

	"github.com/pozii/RepoSearcher/internal/installer"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install repo-searcher to PATH",
	Long: `Install repo-searcher to your system PATH for global access.

This command will:
  - Copy the binary to the install directory
  - Add the install directory to PATH
  - Verify installation

After installation, you can run 'repo-searcher' from anywhere.

Platform-specific behavior:
  - Windows: Adds to %USERPROFILE%\AppData\Local\repo-searcher\bin
  - macOS/Linux: Creates symlink in /usr/local/bin`,
	RunE: runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func runInstall(cmd *cobra.Command, args []string) error {
	fmt.Println("Installing repo-searcher...")

	if err := installer.Install(); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	fmt.Println("")
	fmt.Println("Installation complete!")
	fmt.Println("You can now run 'repo-searcher' from anywhere.")
	return nil
}
