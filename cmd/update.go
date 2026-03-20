package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pozii/RepoSearcher/internal/updater"
	"github.com/spf13/cobra"
)

var (
	flagYes bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for and install updates",
	Long: `Check for the latest version of repo-searcher and install updates.

This command will:
  - Check the latest release on GitHub
  - If a newer version is available, prompt for confirmation
  - Download and install the update

You can also pass --yes to skip the confirmation prompt.

Examples:
  # Interactive update
  repo-searcher update

  # Auto-confirm update
  repo-searcher update --yes`,
	RunE: runUpdate,
}

func init() {
	updateCmd.Flags().BoolVarP(&flagYes, "yes", "y", false, "Skip confirmation prompt")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	return CheckAndApplyUpdate(flagYes)
}

// CheckAndApplyUpdate checks for updates and applies them
func CheckAndApplyUpdate(silent bool) error {
	fmt.Println("Checking for updates...")

	release, hasUpdate, err := updater.CheckForUpdate()
	if err != nil {
		return fmt.Errorf("update check failed: %w", err)
	}

	if !hasUpdate {
		fmt.Printf("You are running the latest version (%s)\n", updater.GetCurrentVersion())
		return nil
	}

	fmt.Printf("New version available: %s\n", release.TagName)
	fmt.Printf("Current version: %s\n", updater.GetCurrentVersion())

	// Skip confirmation if --yes or silent mode
	if silent {
		fmt.Println("Downloading update...")
		if err := updater.Update(release); err != nil {
			return fmt.Errorf("update failed: %w", err)
		}
		fmt.Printf("Updated to %s!\n", release.TagName)
		return nil
	}

	// Interactive confirmation
	fmt.Print("Would you like to update? [Y/n]: ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "n" || response == "no" {
		fmt.Println("Update cancelled.")
		return nil
	}

	fmt.Println("Downloading update...")
	if err := updater.Update(release); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	fmt.Printf("Updated to %s!\n", release.TagName)
	return nil
}
