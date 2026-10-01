package properties

import (
	"strings"
	"testing"

	"github.com/bancolombia/scaffold-go/internal/commands"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/spf13/pflag"
)

// **Validates: Requirements 17.2, 17.3, 17.4**
//
// Feature: go-microservice-scaffolding-cli, Property 57: Command help displays detailed usage and examples
//
// This property test verifies that when a developer executes a command with the help flag,
// the CLI tool displays:
// 1. Detailed usage information for that specific command (Requirement 17.2)
// 2. Examples for each command in the help output (Requirement 17.3)
// 3. Available flags and their descriptions for each command (Requirement 17.4)
func TestProperty_CommandHelpDisplaysDetailedUsageAndExamples(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for command paths
	genCommandPath := gen.OneGenOf(
		gen.Const([]string{}),                              // root command
		gen.Const([]string{"init"}),                        // init command
		gen.Const([]string{"version"}),                     // version command
		gen.Const([]string{"generate"}),                    // generate command
		gen.Const([]string{"generate", "entity"}),          // generate entity
		gen.Const([]string{"generate", "usecase"}),         // generate usecase
		gen.Const([]string{"generate", "ep"}),              // generate ep
		gen.Const([]string{"generate", "ep", "http"}),      // generate ep http
		gen.Const([]string{"generate", "ep", "consumer"}),  // generate ep consumer
		gen.Const([]string{"generate", "da"}),              // generate da
		gen.Const([]string{"generate", "da", "database"}),  // generate da database
		gen.Const([]string{"generate", "da", "api"}),       // generate da api
		gen.Const([]string{"generate", "da", "publisher"}), // generate da publisher
		gen.Const([]string{"generate", "da", "cache"}),     // generate da cache
	)

	properties.Property("all commands display detailed usage information", prop.ForAll(
		func(cmdPath []string) bool {
			rootCmd := commands.NewRootCommand()

			// Navigate to the target command
			cmd, _, err := rootCmd.Find(cmdPath)
			if err != nil {
				t.Logf("Failed to find command %v: %v", cmdPath, err)
				return false
			}

			// Requirement 17.2: Detailed usage information
			// Every command must have a Use field
			if cmd.Use == "" {
				t.Logf("Command %v missing Use field", cmdPath)
				return false
			}

			// Every command must have a Short description
			if cmd.Short == "" {
				t.Logf("Command %v missing Short description", cmdPath)
				return false
			}

			// Every command must have a Long description with detailed information
			if cmd.Long == "" {
				t.Logf("Command %v missing Long description", cmdPath)
				return false
			}

			// Long description should be more detailed than Short (at least 2x longer)
			if len(cmd.Long) < len(cmd.Short)*2 {
				t.Logf("Command %v Long description not detailed enough (Short: %d chars, Long: %d chars)",
					cmdPath, len(cmd.Short), len(cmd.Long))
				return false
			}

			return true
		},
		genCommandPath,
	))

	properties.Property("all commands display examples in help output", prop.ForAll(
		func(cmdPath []string) bool {
			rootCmd := commands.NewRootCommand()

			// Navigate to the target command
			cmd, _, err := rootCmd.Find(cmdPath)
			if err != nil {
				t.Logf("Failed to find command %v: %v", cmdPath, err)
				return false
			}

			// Requirement 17.3: Examples for each command
			// Every command must have an Example field
			if cmd.Example == "" {
				t.Logf("Command %v missing Example field", cmdPath)
				return false
			}

			// Examples should contain actual command usage
			// Check for the presence of "scaffold-go" in examples
			if !strings.Contains(cmd.Example, "scaffold-go") {
				t.Logf("Command %v Example doesn't contain 'scaffold-go'", cmdPath)
				return false
			}

			// Examples should have multiple lines (at least 2 example commands)
			exampleLines := strings.Split(strings.TrimSpace(cmd.Example), "\n")
			nonEmptyLines := 0
			for _, line := range exampleLines {
				if strings.TrimSpace(line) != "" {
					nonEmptyLines++
				}
			}
			if nonEmptyLines < 2 {
				t.Logf("Command %v has insufficient examples (only %d non-empty lines)", cmdPath, nonEmptyLines)
				return false
			}

			return true
		},
		genCommandPath,
	))

	properties.Property("all commands display available flags and descriptions", prop.ForAll(
		func(cmdPath []string) bool {
			rootCmd := commands.NewRootCommand()

			// Navigate to the target command
			cmd, _, err := rootCmd.Find(cmdPath)
			if err != nil {
				t.Logf("Failed to find command %v: %v", cmdPath, err)
				return false
			}

			// Requirement 17.4: Available flags and their descriptions
			// Check that all flags have usage descriptions
			hasFlags := false
			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				hasFlags = true
				if flag.Usage == "" {
					t.Logf("Command %v flag --%s missing usage description", cmdPath, flag.Name)
				}
			})

			// Also check persistent flags (inherited from parent commands)
			cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
				hasFlags = true
				if flag.Usage == "" {
					t.Logf("Command %v persistent flag --%s missing usage description", cmdPath, flag.Name)
				}
			})

			// Some commands should have flags (like init, generate commands)
			// Root command should have persistent flags (config, verbose)
			if len(cmdPath) == 0 {
				// Root command must have persistent flags
				if !hasFlags {
					t.Logf("Root command has no flags")
					return false
				}
			}

			// Verify all flags have descriptions by checking they're not empty
			allFlagsHaveDescriptions := true
			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				if flag.Usage == "" {
					allFlagsHaveDescriptions = false
				}
			})
			cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
				if flag.Usage == "" {
					allFlagsHaveDescriptions = false
				}
			})

			return allFlagsHaveDescriptions
		},
		genCommandPath,
	))

	properties.TestingRun(t)
}
