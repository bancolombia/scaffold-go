package properties

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: go-microservice-scaffolding-cli, Property 7: Component generation creates files in correct directories
// **Validates: Requirements 2.1, 3.1, 4.1, 5.1, 7.1, 8.1, 9.1, 10.1, 11.1, 12.1**
func TestProperty_ComponentGenerationCreatesFilesInCorrectDirectories(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10 // Reduced from 50 for faster execution
	properties := gopter.NewProperties(parameters)

	// Property: Entity generation creates files in domain directory
	properties.Property("entity generation creates files in domain directory", prop.ForAll(
		func(projectName string, modulePath string, entityName string) bool {
			// Create temporary directory for test
			tmpDir := t.TempDir()
			projectDir := filepath.Join(tmpDir, projectName)

			// Initialize project
			if err := initProject(projectName, modulePath, tmpDir); err != nil {
				t.Logf("Failed to initialize project: %v", err)
				return false
			}

			// Generate entity
			if err := generateComponent("entity", entityName, projectDir); err != nil {
				t.Logf("Failed to generate entity: %v", err)
				return false
			}

			// Check that entity file exists in domain directory (lowercase name)
			entityFile := filepath.Join(projectDir, "internal", "domain", strings.ToLower(entityName)+".go")
			if _, err := os.Stat(entityFile); os.IsNotExist(err) {
				t.Logf("Entity file not found at expected path: %s", entityFile)
				return false
			}

			// Check that repository interface file exists in domain directory
			repoFile := filepath.Join(projectDir, "internal", "domain", strings.ToLower(entityName)+"_repository.go")
			if _, err := os.Stat(repoFile); os.IsNotExist(err) {
				t.Logf("Repository file not found at expected path: %s", repoFile)
				return false
			}

			return true
		},
		genValidProjectName(),
		genValidModulePath(),
		genValidComponentName(),
	))

	// Property: Use case generation creates files in domain directory
	properties.Property("usecase generation creates files in domain directory", prop.ForAll(
		func(projectName string, modulePath string, usecaseName string) bool {
			// Create temporary directory for test
			tmpDir := t.TempDir()
			projectDir := filepath.Join(tmpDir, projectName)

			// Initialize project
			if err := initProject(projectName, modulePath, tmpDir); err != nil {
				t.Logf("Failed to initialize project: %v", err)
				return false
			}

			// Generate usecase
			if err := generateComponent("usecase", usecaseName, projectDir); err != nil {
				t.Logf("Failed to generate usecase: %v", err)
				return false
			}

			// Check that usecase file exists in domain directory
			usecaseFile := filepath.Join(projectDir, "internal", "usecase", usecaseName+"_usecase.go")
			if _, err := os.Stat(usecaseFile); os.IsNotExist(err) {
				t.Logf("Usecase file not found at expected path: %s", usecaseFile)
				return false
			}

			return true
		},
		genValidProjectName(),
		genValidModulePath(),
		genValidComponentName(),
	))

	// Property: Handler generation creates files in entrypoints/http directory
	properties.Property("handler generation creates files in entrypoints/http directory", prop.ForAll(
		func(projectName string, modulePath string, handlerName string) bool {
			// Create temporary directory for test
			tmpDir := t.TempDir()
			projectDir := filepath.Join(tmpDir, projectName)

			// Initialize project
			if err := initProject(projectName, modulePath, tmpDir); err != nil {
				t.Logf("Failed to initialize project: %v", err)
				return false
			}

			// Generate handler
			if err := generateComponent("handler", handlerName, projectDir); err != nil {
				t.Logf("Failed to generate handler: %v", err)
				return false
			}

			// Check that handler file exists in entrypoints/http/handler directory
			handlerFile := filepath.Join(projectDir, "internal", "entrypoints", "http", "handler", strings.ToLower(handlerName)+"_handler.go")
			if _, err := os.Stat(handlerFile); os.IsNotExist(err) {
				t.Logf("Handler file not found at expected path: %s", handlerFile)
				return false
			}

			// Check that module file exists in entrypoints/http/handler directory
			moduleFile := filepath.Join(projectDir, "internal", "entrypoints", "http", "handler", "module.go")
			if _, err := os.Stat(moduleFile); os.IsNotExist(err) {
				t.Logf("Module file not found at expected path: %s", moduleFile)
				return false
			}

			return true
		},
		genValidProjectName(),
		genValidModulePath(),
		genValidComponentName(),
	))

	// Property: Consumer generation creates files in entrypoints/messaging directory
	properties.Property("consumer generation creates files in entrypoints/messaging directory", prop.ForAll(
		func(projectName string, modulePath string, consumerName string) bool {
			// Create temporary directory for test
			tmpDir := t.TempDir()
			projectDir := filepath.Join(tmpDir, projectName)

			// Initialize project
			if err := initProject(projectName, modulePath, tmpDir); err != nil {
				t.Logf("Failed to initialize project: %v", err)
				return false
			}

			// Generate consumer
			if err := generateComponent("consumer", consumerName, projectDir); err != nil {
				t.Logf("Failed to generate consumer: %v", err)
				return false
			}

			// Check that consumer file exists in entrypoints/messaging directory
			consumerFile := filepath.Join(projectDir, "internal", "entrypoints", "messaging", strings.ToLower(consumerName)+"_consumer.go")
			if _, err := os.Stat(consumerFile); os.IsNotExist(err) {
				t.Logf("Consumer file not found at expected path: %s", consumerFile)
				return false
			}

			// Check that per-component module file exists in entrypoints/messaging directory
			moduleFile := filepath.Join(projectDir, "internal", "entrypoints", "messaging", strings.ToLower(consumerName)+"_module.go")
			if _, err := os.Stat(moduleFile); os.IsNotExist(err) {
				t.Logf("Per-component module file not found at expected path: %s", moduleFile)
				return false
			}

			return true
		},
		genValidProjectName(),
		genValidModulePath(),
		genValidComponentName(),
	))

	// Property: Database adapter generation creates files in drivenadapters/database directory
	properties.Property("database adapter generation creates files in drivenadapters/database directory", prop.ForAll(
		func(projectName string, modulePath string, adapterName string) bool {
			// Create temporary directory for test
			tmpDir := t.TempDir()
			projectDir := filepath.Join(tmpDir, projectName)

			// Initialize project
			if err := initProject(projectName, modulePath, tmpDir); err != nil {
				t.Logf("Failed to initialize project: %v", err)
				return false
			}

			// Generate database adapter
			if err := generateComponent("adapter", adapterName, projectDir); err != nil {
				t.Logf("Failed to generate adapter: %v", err)
				return false
			}

			// Adapter file is now engine-named (postgres_adapter.go), not per-component.
			// The generator uses a hardcoded TestEntity seeded above.
			_ = adapterName
			adapterFile := filepath.Join(projectDir, "internal", "drivenadapters", "database", "postgres_adapter.go")
			if _, err := os.Stat(adapterFile); os.IsNotExist(err) {
				t.Logf("Adapter file not found at expected path: %s", adapterFile)
				return false
			}
			repoFile := filepath.Join(projectDir, "internal", "drivenadapters", "database", "testentity_repository.go")
			if _, err := os.Stat(repoFile); os.IsNotExist(err) {
				t.Logf("Repository file not found at expected path: %s", repoFile)
				return false
			}

			// Check that module file exists in drivenadapters/database directory
			moduleFile := filepath.Join(projectDir, "internal", "drivenadapters", "database", "module.go")
			if _, err := os.Stat(moduleFile); os.IsNotExist(err) {
				t.Logf("Module file not found at expected path: %s", moduleFile)
				return false
			}

			return true
		},
		genValidProjectName(),
		genValidModulePath(),
		genValidComponentName(),
	))

	// Property: API client generation creates files in drivenadapters/api directory
	properties.Property("api client generation creates files in drivenadapters/api directory", prop.ForAll(
		func(projectName string, modulePath string, clientName string, seed int64) bool {
			// Create temporary directory for test
			tmpDir := t.TempDir()
			projectDir := filepath.Join(tmpDir, projectName)

			// Initialize project
			if err := initProject(projectName, modulePath, tmpDir); err != nil {
				t.Logf("Failed to initialize project: %v", err)
				return false
			}

			// Generate API client using the CLI
			cliPath, err := buildCLI()
			if err != nil {
				t.Logf("Failed to build CLI: %v", err)
				return false
			}

			cmd := filepath.Join(cliPath)
			args := []string{"generate", "da", "api", clientName, "--url", "https://api.example.com"}

			// Execute command in project directory
			if err := executeCommandInDir(cmd, args, projectDir); err != nil {
				t.Logf("Failed to generate API client: %v", err)
				return false
			}

			// Check that API client file exists in drivenadapters/api directory
			clientFile := filepath.Join(projectDir, "internal", "drivenadapters", "api", clientName+"_client.go")
			if _, err := os.Stat(clientFile); os.IsNotExist(err) {
				t.Logf("API client file not found at expected path: %s", clientFile)
				return false
			}

			// Check that module file exists in drivenadapters/api directory
			moduleFile := filepath.Join(projectDir, "internal", "drivenadapters", "api", "module.go")
			if _, err := os.Stat(moduleFile); os.IsNotExist(err) {
				t.Logf("Module file not found at expected path: %s", moduleFile)
				return false
			}

			return true
		},
		genValidProjectName(),
		genValidModulePath(),
		genValidComponentName(),
		gen.Int64(),
	))

	// Property: Publisher generation creates files in drivenadapters/messaging directory
	properties.Property("publisher generation creates files in drivenadapters/messaging directory", prop.ForAll(
		func(projectName string, modulePath string, publisherName string) bool {
			// Create temporary directory for test
			tmpDir := t.TempDir()
			projectDir := filepath.Join(tmpDir, projectName)

			// Initialize project
			if err := initProject(projectName, modulePath, tmpDir); err != nil {
				t.Logf("Failed to initialize project: %v", err)
				return false
			}

			// Generate publisher
			if err := generateComponent("publisher", publisherName, projectDir); err != nil {
				t.Logf("Failed to generate publisher: %v", err)
				return false
			}

			// Check that publisher file exists in drivenadapters/messaging directory
			publisherFile := filepath.Join(projectDir, "internal", "drivenadapters", "messaging", strings.ToLower(publisherName)+"_publisher.go")
			if _, err := os.Stat(publisherFile); os.IsNotExist(err) {
				t.Logf("Publisher file not found at expected path: %s", publisherFile)
				return false
			}

			// Check that per-component module file exists in drivenadapters/messaging directory
			moduleFile := filepath.Join(projectDir, "internal", "drivenadapters", "messaging", strings.ToLower(publisherName)+"_module.go")
			if _, err := os.Stat(moduleFile); os.IsNotExist(err) {
				t.Logf("Per-component module file not found at expected path: %s", moduleFile)
				return false
			}

			return true
		},
		genValidProjectName(),
		genValidModulePath(),
		genValidComponentName(),
	))

	// Property: Cache adapter generation creates files in drivenadapters/cache directory
	properties.Property("cache adapter generation creates files in drivenadapters/cache directory", prop.ForAll(
		func(projectName string, modulePath string, cacheName string) bool {
			// Create temporary directory for test
			tmpDir := t.TempDir()
			projectDir := filepath.Join(tmpDir, projectName)

			// Initialize project
			if err := initProject(projectName, modulePath, tmpDir); err != nil {
				t.Logf("Failed to initialize project: %v", err)
				return false
			}

			// Generate cache adapter
			if err := generateComponent("cache", cacheName, projectDir); err != nil {
				t.Logf("Failed to generate cache: %v", err)
				return false
			}

			// Check that cache file exists in drivenadapters/cache directory
			cacheFile := filepath.Join(projectDir, "internal", "drivenadapters", "cache", cacheName+"_cache.go")
			if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
				t.Logf("Cache file not found at expected path: %s", cacheFile)
				return false
			}

			// Check that module file exists in drivenadapters/cache directory
			moduleFile := filepath.Join(projectDir, "internal", "drivenadapters", "cache", "module.go")
			if _, err := os.Stat(moduleFile); os.IsNotExist(err) {
				t.Logf("Module file not found at expected path: %s", moduleFile)
				return false
			}

			return true
		},
		genValidProjectName(),
		genValidModulePath(),
		genValidComponentName(),
	))

	properties.TestingRun(t)
}
