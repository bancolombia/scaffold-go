package properties

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
)

// Common helper functions for property-based testing

// buildCLI builds the CLI tool and returns the path to the binary
var cliBinaryPath string
var cliBuildOnce sync.Once

func buildCLI() (string, error) {
	var buildErr error
	cliBuildOnce.Do(func() {
		// Build the CLI tool to a temporary location
		tmpDir := os.TempDir()
		cliBinaryPath = filepath.Join(tmpDir, "scaffold-go-test")

		cmd := exec.Command("go", "build", "-o", cliBinaryPath, "../../cmd/main.go")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		buildErr = cmd.Run()
	})

	if buildErr != nil {
		return "", fmt.Errorf("failed to build CLI: %w", buildErr)
	}

	return cliBinaryPath, nil
}

// genValidProjectName generates valid project names (lowercase, alphanumeric, hyphens)
func genValidProjectName() gopter.Gen {
	return gen.RegexMatch(`^[a-z][a-z0-9-]{0,30}`)
}

// genValidModulePath generates valid Go module paths
func genValidModulePath() gopter.Gen {
	return gen.OneGenOf(
		gen.Const("github.com/user/project"),
		gen.Const("gitlab.com/org/service"),
		gen.Const("example.com/app/backend"),
		gen.Const("mycompany.com/services/api"),
		gen.RegexMatch(`^[a-z][a-z0-9-]{0,20}\.[a-z]{2,6}/[a-z][a-z0-9-]{0,20}/[a-z][a-z0-9-]{0,20}`),
	)
}

// genValidComponentName generates valid component names
func genValidComponentName() gopter.Gen {
	return gen.OneConstOf(
		"User",
		"Order",
		"Product",
		"Customer",
		"Payment",
		"Inventory",
		"Notification",
	)
}

// initProject is a helper function to initialize a project for testing
func initProject(projectName, modulePath, outputDir string) error {
	// Use the pre-built CLI binary instead of go run for better performance
	cliPath, err := buildCLI()
	if err != nil {
		return err
	}

	cmd := exec.Command(cliPath, "init", projectName, "--module", modulePath, "--output", outputDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// generateComponent is a helper function to generate a component for testing
func generateComponent(componentType, componentName, projectDir string) error {
	// Get the CLI binary path
	cliPath, err := buildCLI()
	if err != nil {
		return err
	}

	var cmd *exec.Cmd

	switch componentType {
	case "entity":
		cmd = exec.Command(cliPath, "generate", "entity", componentName,
			"--fields", "Name:string,Age:int")
	case "handler":
		cmd = exec.Command(cliPath, "generate", "ep", "http", componentName,
			"--usecases", "testUseCase:TestUseCase",
			"--routes", "GET:/test:HandleTest")
	case "usecase":
		cmd = exec.Command(cliPath, "generate", "usecase", componentName,
			"--dependencies", "testRepo:TestRepository")
	case "adapter":
		// The database adapter must satisfy domain.<Entity>Repository, so the
		// entity has to exist first.
		seed := exec.Command(cliPath, "generate", "entity", "TestEntity",
			"--fields", "Name:string,Age:int")
		seed.Dir = projectDir
		seed.Stdout = os.Stdout
		seed.Stderr = os.Stderr
		if err := seed.Run(); err != nil {
			return err
		}
		_ = componentName
		cmd = exec.Command(cliPath, "generate", "da", "database",
			"--type", "postgres",
			"--entity", "TestEntity",
			"--fields", "Name:string,Age:int")
	case "consumer":
		cmd = exec.Command(cliPath, "generate", "ep", "consumer", componentName,
			"--broker", "rabbitmq", "--pattern", "event",
			"--usecases", "testUseCase:TestUseCase",
			"--queue", "test.event")
	case "publisher":
		cmd = exec.Command(cliPath, "generate", "da", "publisher", componentName,
			"--broker", "rabbitmq", "--pattern", "event",
			"--target", "test.event")
	case "cache":
		cmd = exec.Command(cliPath, "generate", "da", "cache", componentName)
	default:
		return nil
	}

	// Set working directory to the generated project so the CLI can find it
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// executeCommandInDir executes a command in a specific directory
func executeCommandInDir(command string, args []string, dir string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
