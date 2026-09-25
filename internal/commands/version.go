package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	BuildDate = "unknown"
	GitCommit = "unknown"
)

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long: `Display version information for the scaffold-go CLI tool.

This command shows:
  • Version number (semantic versioning)
  • Build date (when the binary was compiled)
  • Git commit hash (source code revision)

Use this to verify which version of scaffold-go you're running and for troubleshooting.`,
		Example: `  # Display version information
  scaffold-go version

  # Check version in scripts
  scaffold-go version | grep "version"`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("scaffold-go version %s\n", Version)
			fmt.Printf("Build date: %s\n", BuildDate)
			fmt.Printf("Git commit: %s\n", GitCommit)
		},
	}
}
