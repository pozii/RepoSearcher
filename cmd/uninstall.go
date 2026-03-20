package cmd

import (
	"fmt"

	"github.com/pozii/RepoSearcher/internal/installer"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove repo-searcher from PATH",
	Long: `Uninstall repo-searcher from your system.

This command will:
  - Remove the binary from the install directory
  - Remove the install directory from PATH

After uninstalling, 'repo-searcher' will no longer be available
as a global command.`,
	RunE: runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall(cmd *cobra.Command, args []string) error {
	fmt.Println("Uninstalling repo-searcher...")

	if err := installer.Uninstall(); err != nil {
		return fmt.Errorf("uninstallation failed: %w", err)
	}

	fmt.Println("")
	fmt.Println("Uninstallation complete!")
	return nil
}
