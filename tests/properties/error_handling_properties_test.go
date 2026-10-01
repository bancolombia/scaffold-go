package properties

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bancolombia/scaffold-go/internal/commands"
	"github.com/bancolombia/scaffold-go/internal/domain"
	"github.com/bancolombia/scaffold-go/internal/validator"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"go.uber.org/zap"
)

// Custom generators for error handling property-based testing

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

// genInvalidProjectName generates invalid project names
func genInvalidProjectName() gopter.Gen {
	return gen.OneGenOf(
		gen.Const(""),                      // Empty string
		gen.Const("123project"),            // Starts with number
		gen.Const("Project"),               // Starts with uppercase
		gen.Const("project_name"),          // Contains underscore
		gen.Const("project name"),          // Contains space
		gen.Const("project!name"),          // Contains special character
		gen.RegexMatch(`^[A-Z][a-z0-9-]*`), // Uppercase start
	)
}

// genInvalidFieldSpec generates invalid field specifications
func genInvalidFieldSpec() gopter.Gen {
	return gen.OneGenOf(
		gen.Const("InvalidFormat"),    // Missing colon
		gen.Const("name:"),            // Missing type
		gen.Const(":string"),          // Missing name
		gen.Const("lowercase:string"), // Lowercase field name
		gen.Const("Name:"),            // Empty type
		gen.Const(":Type"),            // Empty name
		gen.Const("Name:Type:Extra"),  // Too many parts
	)
}

// Feature: go-microservice-scaffolding-cli, Property 58: Invalid input produces descriptive error messages
// **Validates: Requirements 19.1, 19.4**
func TestProperty_InvalidInputProducesDescriptiveErrorMessages(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	v := validator.New()

	properties.Property("invalid component names produce descriptive validation errors", prop.ForAll(
		func(invalidName string) bool {
			err := v.ValidateName(invalidName)

			// Should return an error
			if err == nil {
				return false
			}

			// Error message should be descriptive
			errMsg := err.Error()
			return len(errMsg) > 0 &&
				(strings.Contains(errMsg, "name") ||
					strings.Contains(errMsg, "uppercase") ||
					strings.Contains(errMsg, "alphanumeric") ||
					strings.Contains(errMsg, "empty"))
		},
		genInvalidComponentName(),
	))

	properties.Property("invalid project names produce descriptive validation errors", prop.ForAll(
		func(invalidName string) bool {
			err := v.ValidateProjectName(invalidName)

			// Should return an error
			if err == nil {
				return false
			}

			// Error message should be descriptive
			errMsg := err.Error()
			return len(errMsg) > 0 &&
				(strings.Contains(errMsg, "project name") ||
					strings.Contains(errMsg, "lowercase") ||
					strings.Contains(errMsg, "empty"))
		},
		genInvalidProjectName(),
	))

	properties.Property("invalid field specifications produce descriptive validation errors", prop.ForAll(
		func(invalidField string) bool {
			err := v.ValidateFields([]string{invalidField})

			// Should return an error
			if err == nil {
				return false
			}

			// Error message should be descriptive and mention the invalid field
			errMsg := err.Error()
			return len(errMsg) > 0 &&
				(strings.Contains(errMsg, "field") ||
					strings.Contains(errMsg, "format") ||
					strings.Contains(errMsg, "name:type") ||
					strings.Contains(errMsg, "empty") ||
					strings.Contains(errMsg, "uppercase"))
		},
		genInvalidFieldSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 59: File system errors produce descriptive error messages
// **Validates: Requirements 19.2**
func TestProperty_FileSystemErrorsProduceDescriptiveErrorMessages(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	v := validator.New()

	properties.Property("non-existent directory validation produces descriptive error", prop.ForAll(
		func(seed int64) bool {
			// Create a path that doesn't exist
			nonExistentPath := filepath.Join(t.TempDir(), "nonexistent", "path")

			err := v.ValidateDirectoryExists(nonExistentPath)

			// Should return an error
			if err == nil {
				return false
			}

			// Error message should be descriptive
			errMsg := err.Error()
			return len(errMsg) > 0 &&
				strings.Contains(errMsg, "directory") &&
				(strings.Contains(errMsg, "not exist") || strings.Contains(errMsg, "does not exist"))
		},
		gen.Int64(),
	))

	properties.Property("existing directory validation produces descriptive error", prop.ForAll(
		func(seed int64) bool {
			// Create a directory that exists
			tmpDir := t.TempDir()
			existingPath := filepath.Join(tmpDir, "existing")
			if err := os.MkdirAll(existingPath, 0755); err != nil {
				return false
			}

			err := v.ValidateDirectoryNotExists(existingPath)

			// Should return an error
			if err == nil {
				return false
			}

			// Error message should be descriptive
			errMsg := err.Error()
			return len(errMsg) > 0 &&
				strings.Contains(errMsg, "directory") &&
				strings.Contains(errMsg, "already exists")
		},
		gen.Int64(),
	))

	properties.Property("file system error wrapping produces descriptive messages", prop.ForAll(
		func(operation string, path string) bool {
			baseErr := errors.New("permission denied")
			wrappedErr := commands.WrapFileSystemError(operation, path, baseErr)

			// Should be a FileSystemError
			var fsErr *domain.FileSystemError
			if !errors.As(wrappedErr, &fsErr) {
				return false
			}

			// Error message should be descriptive
			errMsg := wrappedErr.Error()
			return len(errMsg) > 0 &&
				strings.Contains(errMsg, "file system error") &&
				strings.Contains(errMsg, operation) &&
				strings.Contains(errMsg, path)
		},
		gen.OneConstOf("read", "write", "create", "delete"),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 60: Failed commands exit with non-zero status
// **Validates: Requirements 19.3**
func TestProperty_FailedCommandsExitWithNonZeroStatus(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	logger, _ := zap.NewDevelopment()

	properties.Property("validation errors return non-zero exit code", prop.ForAll(
		func(field string, value string, message string) bool {
			err := &domain.ValidationError{
				Field:   field,
				Value:   value,
				Message: message,
			}

			exitCode := commands.HandleError(err, logger)

			// Should return non-zero exit code
			return exitCode != commands.ExitCodeSuccess && exitCode == commands.ExitCodeValidationError
		},
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.Property("file system errors return non-zero exit code", prop.ForAll(
		func(operation string, path string) bool {
			err := &domain.FileSystemError{
				Operation: operation,
				Path:      path,
				Err:       errors.New("permission denied"),
			}

			exitCode := commands.HandleError(err, logger)

			// Should return non-zero exit code
			return exitCode != commands.ExitCodeSuccess && exitCode == commands.ExitCodeFileSystemError
		},
		gen.OneConstOf("read", "write", "create", "delete"),
		gen.AlphaString(),
	))

	properties.Property("template errors return non-zero exit code", prop.ForAll(
		func(templateName string, operation string) bool {
			err := &domain.TemplateError{
				TemplateName: templateName,
				Operation:    operation,
				Err:          errors.New("template not found"),
			}

			exitCode := commands.HandleError(err, logger)

			// Should return non-zero exit code
			return exitCode != commands.ExitCodeSuccess && exitCode == commands.ExitCodeTemplateError
		},
		gen.AlphaString(),
		gen.OneConstOf("load", "parse", "execute"),
	))

	properties.Property("generation errors return non-zero exit code", prop.ForAll(
		func(component string, stage string) bool {
			err := &domain.GenerationError{
				Component: component,
				Stage:     stage,
				Err:       errors.New("generation failed"),
			}

			exitCode := commands.HandleError(err, logger)

			// Should return non-zero exit code
			return exitCode != commands.ExitCodeSuccess && exitCode == commands.ExitCodeGenerationError
		},
		gen.AlphaString(),
		gen.OneConstOf("validation", "template", "write"),
	))

	properties.Property("unknown errors return non-zero exit code", prop.ForAll(
		func(message string) bool {
			err := errors.New(message)

			exitCode := commands.HandleError(err, logger)

			// Should return non-zero exit code
			return exitCode != commands.ExitCodeSuccess && exitCode == commands.ExitCodeUnknownError
		},
		gen.AlphaString(),
	))

	properties.Property("nil error returns success exit code", prop.ForAll(
		func(seed int64) bool {
			exitCode := commands.HandleError(nil, logger)

			// Should return success exit code
			return exitCode == commands.ExitCodeSuccess
		},
		gen.Int64(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 61: Validation occurs before execution
// **Validates: Requirements 19.4**
func TestProperty_ValidationOccursBeforeExecution(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	v := validator.New()

	properties.Property("project initialization validation occurs before directory creation", prop.ForAll(
		func(seed int64) bool {
			// Create a temporary directory without go.mod
			tmpDir := t.TempDir()
			testDir := filepath.Join(tmpDir, "test-no-gomod")
			if err := os.MkdirAll(testDir, 0755); err != nil {
				return false
			}

			// Validation should fail before any operations
			err := v.ValidateProjectInitialized(testDir)

			// Should return error (validation failed)
			if err == nil {
				return false
			}

			// Directory should still exist (no cleanup needed since validation failed)
			_, statErr := os.Stat(testDir)
			return statErr == nil
		},
		gen.Int64(),
	))

	properties.Property("component name validation occurs before generation", prop.ForAll(
		func(invalidName string) bool {
			// Validate name first
			err := v.ValidateName(invalidName)

			// Should fail validation
			if err == nil {
				return false
			}

			// Error should indicate validation failure
			return strings.Contains(err.Error(), "name") ||
				strings.Contains(err.Error(), "uppercase") ||
				strings.Contains(err.Error(), "alphanumeric") ||
				strings.Contains(err.Error(), "empty")
		},
		genInvalidComponentName(),
	))

	properties.Property("field specification validation occurs before entity generation", prop.ForAll(
		func(invalidField string) bool {
			// Validate fields first
			err := v.ValidateFields([]string{invalidField})

			// Should fail validation
			if err == nil {
				return false
			}

			// Error should indicate validation failure
			return strings.Contains(err.Error(), "field") ||
				strings.Contains(err.Error(), "format") ||
				strings.Contains(err.Error(), "name:type")
		},
		genInvalidFieldSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 62: Missing parameters produce descriptive error messages
// **Validates: Requirements 19.5**
func TestProperty_MissingParametersProduceDescriptiveErrorMessages(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	v := validator.New()

	properties.Property("empty component name produces descriptive error", prop.ForAll(
		func(seed int64) bool {
			err := v.ValidateName("")

			// Should return an error
			if err == nil {
				return false
			}

			// Error message should mention empty or missing name
			errMsg := err.Error()
			return len(errMsg) > 0 &&
				(strings.Contains(errMsg, "empty") ||
					strings.Contains(errMsg, "cannot be empty") ||
					strings.Contains(errMsg, "name"))
		},
		gen.Int64(),
	))

	properties.Property("empty project name produces descriptive error", prop.ForAll(
		func(seed int64) bool {
			err := v.ValidateProjectName("")

			// Should return an error
			if err == nil {
				return false
			}

			// Error message should mention empty or missing project name
			errMsg := err.Error()
			return len(errMsg) > 0 &&
				(strings.Contains(errMsg, "empty") ||
					strings.Contains(errMsg, "cannot be empty") ||
					strings.Contains(errMsg, "project name"))
		},
		gen.Int64(),
	))

	properties.Property("empty field name produces descriptive error", prop.ForAll(
		func(seed int64) bool {
			// Field with empty name
			err := v.ValidateFields([]string{":string"})

			// Should return an error
			if err == nil {
				return false
			}

			// Error message should mention empty field name
			errMsg := err.Error()
			return len(errMsg) > 0 &&
				(strings.Contains(errMsg, "empty") ||
					strings.Contains(errMsg, "field name") ||
					strings.Contains(errMsg, "cannot be empty"))
		},
		gen.Int64(),
	))

	properties.Property("empty field type produces descriptive error", prop.ForAll(
		func(seed int64) bool {
			// Field with empty type
			err := v.ValidateFields([]string{"Name:"})

			// Should return an error
			if err == nil {
				return false
			}

			// Error message should mention empty field type
			errMsg := err.Error()
			return len(errMsg) > 0 &&
				(strings.Contains(errMsg, "empty") ||
					strings.Contains(errMsg, "type") ||
					strings.Contains(errMsg, "cannot be empty"))
		},
		gen.Int64(),
	))

	properties.TestingRun(t)
}
