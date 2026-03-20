package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set at build time
var Version = "v1.0.0"

// BuildDate is set at build time
var BuildDate = "unknown"

// GitCommit is set at build time
var GitCommit = "unknown"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("repo-searcher %s\n", Version)
		fmt.Printf("Commit: %s\n", GitCommit)
		fmt.Printf("Built: %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
