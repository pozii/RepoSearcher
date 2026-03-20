package main

import (
	"os"

	"github.com/pozii/RepoSearcher/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
