package main

import (
	"os"

	"github.com/gabheadz/scaffold-go/internal/commands"
)

func main() {
	rootCmd := commands.NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		logger := commands.GetLogger()
		exitCode := commands.HandleError(err, logger)
		os.Exit(exitCode)
	}
}
