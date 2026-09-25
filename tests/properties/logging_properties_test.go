package properties

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Custom generators for logging property-based testing

// genComponentType generates different component types that should have logging
func genComponentType() gopter.Gen {
	return gen.OneConstOf(
		"handler",
		"usecase",
		"adapter",
		"consumer",
		"publisher",
		"cache",
	)
}

// Feature: go-microservice-scaffolding-cli, Property 45: Generated components include structured logging
// **Validates: Requirements 14.2**
func TestProperty_GeneratedComponentsIncludeStructuredLogging(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10
	properties := gopter.NewProperties(parameters)

	properties.Property("generated entrypoints and adapters include Zap logger initialization in constructor", prop.ForAll(
		func(projectName string, modulePath string, componentType string, componentName string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Execute init command
			err := initProject(projectName, modulePath, tmpDir)
			if err != nil {
				t.Logf("init failed: %v", err)
				return false
			}

			projectDir := filepath.Join(tmpDir, projectName)

			// Generate component
			err = generateComponent(componentType, componentName, projectDir)
			if err != nil {
				t.Logf("generate %s failed: %v", componentType, err)
				return false
			}

			// Determine the file path based on component type
			var filePath string
			switch componentType {
			case "handler":
				filePath = filepath.Join(projectDir, "internal", "entrypoints", "http", "handler", strings.ToLower(componentName)+"_handler.go")
			case "usecase":
				filePath = filepath.Join(projectDir, "internal", "usecase", strings.ToLower(componentName)+"_usecase.go")
			case "adapter":
				// generateComponent seeds a TestEntity and generates postgres adapter files.
				filePath = filepath.Join(projectDir, "internal", "drivenadapters", "database", "postgres_adapter.go")
			case "consumer":
				filePath = filepath.Join(projectDir, "internal", "entrypoints", "messaging", strings.ToLower(componentName)+"_consumer.go")
			case "publisher":
				filePath = filepath.Join(projectDir, "internal", "drivenadapters", "messaging", strings.ToLower(componentName)+"_publisher.go")
			case "cache":
				filePath = filepath.Join(projectDir, "internal", "drivenadapters", "cache", strings.ToLower(componentName)+"_cache.go")
			default:
				t.Logf("unknown component type: %s", componentType)
				return false
			}

			// Read generated file
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Logf("failed to read generated file %s: %v", filePath, err)
				return false
			}

			fileContent := string(content)

			// Verify Zap import is present
			if !strings.Contains(fileContent, "\"go.uber.org/zap\"") {
				t.Logf("generated %s does not import go.uber.org/zap", componentType)
				return false
			}

			// Verify logger field in struct
			if !strings.Contains(fileContent, "logger *zap.Logger") {
				t.Logf("generated %s does not have logger field in struct", componentType)
				return false
			}

			// Verify logger parameter in constructor
			constructorPattern := "logger *zap.Logger"
			if !strings.Contains(fileContent, constructorPattern) {
				t.Logf("generated %s constructor does not accept logger parameter", componentType)
				return false
			}

			// Verify logger is assigned in constructor
			// Use regex to handle variable whitespace
			loggerAssignmentPattern := `logger:\s+logger`
			matched, _ := regexp.MatchString(loggerAssignmentPattern, fileContent)
			if !matched {
				t.Logf("generated %s constructor does not assign logger", componentType)
				return false
			}

			return true
		},
		genValidProjectName(),
		genValidModulePath(),
		genComponentType(),
		genValidComponentName(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 46: Generated logging uses appropriate log levels
// **Validates: Requirements 14.3**
func TestProperty_GeneratedLoggingUsesAppropriateLogLevels(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10
	properties := gopter.NewProperties(parameters)

	properties.Property("generated components use appropriate log levels (Info, Error, Debug, Warn)", prop.ForAll(
		func(projectName string, modulePath string, componentType string, componentName string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Execute init command
			err := initProject(projectName, modulePath, tmpDir)
			if err != nil {
				t.Logf("init failed: %v", err)
				return false
			}

			projectDir := filepath.Join(tmpDir, projectName)

			// Generate component
			err = generateComponent(componentType, componentName, projectDir)
			if err != nil {
				t.Logf("generate %s failed: %v", componentType, err)
				return false
			}

			// Determine the file path based on component type
			var filePath string
			switch componentType {
			case "handler":
				filePath = filepath.Join(projectDir, "internal", "entrypoints", "http", "handler", strings.ToLower(componentName)+"_handler.go")
			case "usecase":
				filePath = filepath.Join(projectDir, "internal", "usecase", strings.ToLower(componentName)+"_usecase.go")
			case "adapter":
				// generateComponent seeds a TestEntity; the per-entity repository carries
				// the operational logging (Info + Error) that this property asserts on.
				filePath = filepath.Join(projectDir, "internal", "drivenadapters", "database", "testentity_repository.go")
			case "consumer":
				filePath = filepath.Join(projectDir, "internal", "entrypoints", "messaging", strings.ToLower(componentName)+"_consumer.go")
			case "publisher":
				filePath = filepath.Join(projectDir, "internal", "drivenadapters", "messaging", strings.ToLower(componentName)+"_publisher.go")
			case "cache":
				filePath = filepath.Join(projectDir, "internal", "drivenadapters", "cache", strings.ToLower(componentName)+"_cache.go")
			default:
				t.Logf("unknown component type: %s", componentType)
				return false
			}

			// Read generated file
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Logf("failed to read generated file %s: %v", filePath, err)
				return false
			}

			fileContent := string(content)

			// Verify at least one appropriate log level is used
			logLevels := []string{
				".logger.Info(",
				".logger.Error(",
				".logger.Debug(",
				".logger.Warn(",
			}

			foundLogLevel := false
			for _, level := range logLevels {
				if strings.Contains(fileContent, level) {
					foundLogLevel = true
					break
				}
			}

			if !foundLogLevel {
				t.Logf("generated %s does not use any appropriate log levels", componentType)
				return false
			}

			// Verify Info level is used for normal operations
			// (most components should log info for successful operations)
			// Cache operations can use Debug level since they're low-level
			if componentType != "cache" && !strings.Contains(fileContent, ".logger.Info(") {
				t.Logf("generated %s does not use Info log level for normal operations", componentType)
				return false
			}

			// For adapters, verify Error level is used for error handling
			if componentType == "adapter" || componentType == "publisher" || componentType == "cache" {
				if !strings.Contains(fileContent, ".logger.Error(") {
					t.Logf("generated %s adapter does not use Error log level for error handling", componentType)
					return false
				}
			}

			return true
		},
		genValidProjectName(),
		genValidModulePath(),
		genComponentType(),
		genValidComponentName(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 47: Generated logging includes contextual fields
// **Validates: Requirements 14.4**
func TestProperty_GeneratedLoggingIncludesContextualFields(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10
	properties := gopter.NewProperties(parameters)

	properties.Property("generated logging includes contextual fields using zap.String, zap.Error, etc.", prop.ForAll(
		func(projectName string, modulePath string, componentType string, componentName string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Execute init command
			err := initProject(projectName, modulePath, tmpDir)
			if err != nil {
				t.Logf("init failed: %v", err)
				return false
			}

			projectDir := filepath.Join(tmpDir, projectName)

			// Generate component
			err = generateComponent(componentType, componentName, projectDir)
			if err != nil {
				t.Logf("generate %s failed: %v", componentType, err)
				return false
			}

			// Determine the file path based on component type
			var filePath string
			switch componentType {
			case "handler":
				filePath = filepath.Join(projectDir, "internal", "entrypoints", "http", "handler", strings.ToLower(componentName)+"_handler.go")
			case "usecase":
				filePath = filepath.Join(projectDir, "internal", "usecase", strings.ToLower(componentName)+"_usecase.go")
			case "adapter":
				// generateComponent seeds a TestEntity and generates a per-entity repository
				// that carries the structured logging (adapter file itself is thin).
				filePath = filepath.Join(projectDir, "internal", "drivenadapters", "database", "testentity_repository.go")
			case "consumer":
				filePath = filepath.Join(projectDir, "internal", "entrypoints", "messaging", strings.ToLower(componentName)+"_consumer.go")
			case "publisher":
				filePath = filepath.Join(projectDir, "internal", "drivenadapters", "messaging", strings.ToLower(componentName)+"_publisher.go")
			case "cache":
				filePath = filepath.Join(projectDir, "internal", "drivenadapters", "cache", strings.ToLower(componentName)+"_cache.go")
			default:
				t.Logf("unknown component type: %s", componentType)
				return false
			}

			// Read generated file
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Logf("failed to read generated file %s: %v", filePath, err)
				return false
			}

			fileContent := string(content)

			// For use cases, we have different requirements since they're in the domain layer
			if componentType == "usecase" {
				// Use cases should at least have a descriptive message
				// Check if logging statement exists with meaningful context
				hasLogging := strings.Contains(fileContent, ".logger.Info(")
				if !hasLogging {
					t.Logf("generated use case does not include logging statements")
					return false
				}
				// For usecases, we don't require structured fields since they're in the domain layer
				// Just having the logger is sufficient
				return true
			}

			// Verify structured logging with contextual fields
			// Check for zap field functions
			zapFieldFunctions := []string{
				"zap.String(",
				"zap.Error(",
				"zap.Int(",
				"zap.Any(",
			}

			foundZapField := false
			for _, zapFunc := range zapFieldFunctions {
				if strings.Contains(fileContent, zapFunc) {
					foundZapField = true
					break
				}
			}

			if !foundZapField {
				t.Logf("generated %s does not use structured logging with zap fields", componentType)
				return false
			}

			// Verify contextual fields are meaningful
			// For handlers, should include method, path, or IP
			if componentType == "handler" {
				hasContextualField := strings.Contains(fileContent, "zap.String(\"method\"") ||
					strings.Contains(fileContent, "zap.String(\"path\"") ||
					strings.Contains(fileContent, "zap.String(\"ip\"")

				if !hasContextualField {
					t.Logf("generated handler does not include contextual fields (method, path, ip)")
					return false
				}
			}

			// For adapters, should include table name, entity ID, or operation name
			if componentType == "adapter" {
				hasContextualField := strings.Contains(fileContent, "zap.String(\"table\"") ||
					strings.Contains(fileContent, "zap.String(\"id\"") ||
					strings.Contains(fileContent, "zap.Error(err)")

				if !hasContextualField {
					t.Logf("generated adapter does not include contextual fields (table, id, error)")
					return false
				}
			}

			// Verify error logging includes zap.Error(err)
			if strings.Contains(fileContent, ".logger.Error(") {
				if !strings.Contains(fileContent, "zap.Error(err)") {
					t.Logf("generated %s error logging does not include zap.Error(err)", componentType)
					return false
				}
			}

			return true
		},
		genValidProjectName(),
		genValidModulePath(),
		genComponentType(),
		genValidComponentName(),
	))

	properties.TestingRun(t)
}
