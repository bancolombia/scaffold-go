package validator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Custom generators for property-based testing

// genInvalidComponentName generates names with invalid characters
func genInvalidComponentName() gopter.Gen {
	return gen.OneGenOf(
		gen.Const("123Invalid"),                              // Starts with number
		gen.Const("invalid-name"),                            // Contains hyphen
		gen.Const("invalid_name"),                            // Contains underscore
		gen.Const("invalid name"),                            // Contains space
		gen.Const("invalid!name"),                            // Contains special character
		gen.Const(""),                                        // Empty string
		gen.Const("lowercase"),                               // Starts with lowercase
		gen.RegexMatch(`^[a-z][a-zA-Z0-9]*`),                 // Lowercase start
		gen.RegexMatch(`^[A-Z][a-zA-Z0-9]*[_-][a-zA-Z0-9]*`), // Contains invalid chars
	)
}

// genValidComponentName generates valid component names (PascalCase)
func genValidComponentName() gopter.Gen {
	return gen.RegexMatch(`^[A-Z][a-zA-Z0-9]{0,20}`)
}

// genValidProjectName generates valid project names (lowercase with hyphens)
func genValidProjectName() gopter.Gen {
	return gen.RegexMatch(`^[a-z][a-z0-9-]{0,20}`)
}

// Feature: go-microservice-scaffolding-cli, Property 6: Component name validation rejects invalid characters
// **Validates: Requirements 1.6, 2.7, 3.4, 4.3, 5.6, 7.6, 8.6, 9.7, 10.6, 11.6, 12.6**
func TestProperty_ComponentNameValidationRejectsInvalidCharacters(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	v := New()

	properties.Property("ValidateName rejects invalid characters", prop.ForAll(
		func(invalidName string) bool {
			err := v.ValidateName(invalidName)
			// Should return an error for invalid names
			return err != nil
		},
		genInvalidComponentName(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 64: Generate commands validate project initialization
// **Validates: Requirements 22.1, 22.2**
func TestProperty_GenerateCommandsValidateProjectInitialization(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	v := New()

	properties.Property("ValidateProjectInitialized returns error when no go.mod exists", prop.ForAll(
		func(seed int64) bool {
			// Create a temporary directory without go.mod
			tmpDir := t.TempDir()
			testDir := filepath.Join(tmpDir, "test-no-gomod")
			if err := os.MkdirAll(testDir, 0755); err != nil {
				return false
			}

			// Validate should fail
			err := v.ValidateProjectInitialized(testDir)

			// Should return error indicating project not initialized
			return err != nil &&
				strings.Contains(err.Error(), "project not initialized") &&
				strings.Contains(err.Error(), "go.mod")
		},
		gen.Int64(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 65: Generate commands proceed when project is initialized
// **Validates: Requirements 22.3**
func TestProperty_GenerateCommandsProceedWhenProjectIsInitialized(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	v := New()

	properties.Property("ValidateProjectInitialized succeeds when go.mod exists in current directory", prop.ForAll(
		func(projectName string, seed int64) bool {
			// Create a temporary directory with go.mod
			tmpDir := t.TempDir()
			projectDir := filepath.Join(tmpDir, projectName)
			if err := os.MkdirAll(projectDir, 0755); err != nil {
				return false
			}

			// Create go.mod file
			goModPath := filepath.Join(projectDir, "go.mod")
			goModContent := "module " + projectName + "\n\ngo 1.25.0\n"
			if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
				return false
			}

			// Validate should succeed
			err := v.ValidateProjectInitialized(projectDir)
			return err == nil
		},
		genValidProjectName(),
		gen.Int64(),
	))

	properties.Property("ValidateProjectInitialized succeeds when go.mod exists in parent directory", prop.ForAll(
		func(projectName string, seed int64) bool {
			// Create a temporary directory structure with go.mod in parent
			tmpDir := t.TempDir()
			projectRoot := filepath.Join(tmpDir, projectName)
			if err := os.MkdirAll(projectRoot, 0755); err != nil {
				return false
			}

			// Create go.mod in project root
			goModPath := filepath.Join(projectRoot, "go.mod")
			goModContent := "module " + projectName + "\n\ngo 1.25.0\n"
			if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
				return false
			}

			// Create subdirectory (simulating internal/domain or similar)
			subDir := filepath.Join(projectRoot, "internal", "domain")
			if err := os.MkdirAll(subDir, 0755); err != nil {
				return false
			}

			// Validate from subdirectory should succeed (finds go.mod in parent)
			err := v.ValidateProjectInitialized(subDir)
			return err == nil
		},
		genValidProjectName(),
		gen.Int64(),
	))

	properties.TestingRun(t)
}
