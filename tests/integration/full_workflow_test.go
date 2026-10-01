package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFullProjectGenerationWorkflow tests the complete workflow:
// 1. Initialize a new project
// 2. Generate multiple components (entity, usecase, handler)
// 3. Verify all files are created in correct locations
// 4. Verify the generated project compiles with go build
func TestFullProjectGenerationWorkflow(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()

	// Build the CLI tool first
	cliPath := buildCLI(t)

	// Project configuration
	projectName := "test-microservice"
	modulePath := "github.com/test/microservice"
	projectDir := filepath.Join(tmpDir, projectName)

	// Step 1: Initialize the project
	t.Run("init project", func(t *testing.T) {
		cmd := exec.Command(cliPath, "init", projectName,
			"--module", modulePath,
			"--output", tmpDir)
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "init command failed: %s", string(output))

		// Verify project directory was created
		assert.DirExists(t, projectDir, "project directory should exist")
	})

	// Step 2: Verify initial project structure
	t.Run("verify initial structure", func(t *testing.T) {
		// Verify go.mod
		goModPath := filepath.Join(projectDir, "go.mod")
		assert.FileExists(t, goModPath, "go.mod should exist")

		content, err := os.ReadFile(goModPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), modulePath, "go.mod should contain module path")

		// Verify main.go
		mainGoPath := filepath.Join(projectDir, "cmd", "main.go")
		assert.FileExists(t, mainGoPath, "main.go should exist")

		// Verify config directory
		configDir := filepath.Join(projectDir, "config")
		assert.DirExists(t, configDir, "config directory should exist")

		// Verify README.md
		readmePath := filepath.Join(projectDir, "README.md")
		assert.FileExists(t, readmePath, "README.md should exist")

		// Verify domain directory
		domainDir := filepath.Join(projectDir, "internal", "domain")
		assert.DirExists(t, domainDir, "domain directory should exist")

		// Verify entrypoints directories
		httpDir := filepath.Join(projectDir, "internal", "entrypoints", "http")
		assert.DirExists(t, httpDir, "http entrypoints directory should exist")

		messagingDir := filepath.Join(projectDir, "internal", "entrypoints", "messaging")
		assert.DirExists(t, messagingDir, "messaging entrypoints directory should exist")

		// Verify drivenadapters directories
		dbDir := filepath.Join(projectDir, "internal", "drivenadapters", "database")
		assert.DirExists(t, dbDir, "database drivenadapters directory should exist")

		apiDir := filepath.Join(projectDir, "internal", "drivenadapters", "api")
		assert.DirExists(t, apiDir, "api drivenadapters directory should exist")

		cacheDir := filepath.Join(projectDir, "internal", "drivenadapters", "cache")
		assert.DirExists(t, cacheDir, "cache drivenadapters directory should exist")
	})

	// Step 3: Generate an entity
	t.Run("generate entity", func(t *testing.T) {
		cmd := exec.Command(cliPath, "generate", "entity", "User",
			"--fields", "Name:string,Email:string,Age:int")
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "generate entity command failed: %s", string(output))

		// Verify entity file
		entityPath := filepath.Join(projectDir, "internal", "domain", "user.go")
		assert.FileExists(t, entityPath, "entity file should exist")

		content, err := os.ReadFile(entityPath)
		require.NoError(t, err)
		entityContent := string(content)

		// Verify entity contains expected fields
		assert.Contains(t, entityContent, "type User struct", "entity should have struct definition")
		assert.Contains(t, entityContent, "Name", "entity should have Name field")
		assert.Contains(t, entityContent, "Email", "entity should have Email field")
		assert.Contains(t, entityContent, "Age", "entity should have Age field")
		assert.Contains(t, entityContent, "Validate()", "entity should have Validate method")

		// Verify repository interface file
		repoPath := filepath.Join(projectDir, "internal", "domain", "user_repository.go")
		assert.FileExists(t, repoPath, "repository interface file should exist")

		repoContent, err := os.ReadFile(repoPath)
		require.NoError(t, err)
		repoStr := string(repoContent)

		// Verify repository contains CRUD methods
		assert.Contains(t, repoStr, "type IUserRepository interface", "repository should have interface definition")
		assert.Contains(t, repoStr, "Create", "repository should have Create method")
		assert.Contains(t, repoStr, "GetByID", "repository should have GetByID method")
		assert.Contains(t, repoStr, "Update", "repository should have Update method")
		assert.Contains(t, repoStr, "Delete", "repository should have Delete method")
		assert.Contains(t, repoStr, "List", "repository should have List method")
	})

	// Step 4: Generate a use case
	t.Run("generate usecase", func(t *testing.T) {
		cmd := exec.Command(cliPath, "generate", "usecase", "CreateUser",
			"--dependencies", "userRepository:IUserRepository")
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "generate usecase command failed: %s", string(output))

		// Verify use case file
		usecasePath := filepath.Join(projectDir, "internal", "usecase", "createuser_usecase.go")
		assert.FileExists(t, usecasePath, "usecase file should exist")

		content, err := os.ReadFile(usecasePath)
		require.NoError(t, err)
		usecaseContent := string(content)

		// Verify use case structure (actual template uses "UseCaseName + UseCase" pattern)
		assert.Contains(t, usecaseContent, "type CreateUserUseCase interface", "usecase should have interface definition")
		assert.Contains(t, usecaseContent, "Execute", "usecase should have Execute method")
		assert.Contains(t, usecaseContent, "NewCreateUserUseCase", "usecase should have constructor")
		assert.Contains(t, usecaseContent, "userRepository", "usecase should reference repository dependency")
	})

	// Step 5: Generate an HTTP handler entry-point
	t.Run("generate ep http", func(t *testing.T) {
		cmd := exec.Command(cliPath, "generate", "ep", "http", "UserHandler",
			"--usecases", "createUser:CreateUserUseCase",
			"--routes", "POST:/users:CreateUser",
			"--enable-cors",
			"--enable-rate-limiter",
			"--enable-error-handler")
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "generate ep http command failed: %s", string(output))

		// The generator strips the "Handler" suffix from the name, so the file is user_handler.go
		handlerPath := filepath.Join(projectDir, "internal", "entrypoints", "http", "handler", "user_handler.go")
		assert.FileExists(t, handlerPath, "handler file should exist")

		content, err := os.ReadFile(handlerPath)
		require.NoError(t, err)
		handlerContent := string(content)

		// Verify handler uses Fiber framework
		assert.Contains(t, handlerContent, "github.com/gofiber/fiber/v2", "handler should import Fiber")
		assert.Contains(t, handlerContent, "UserHandler struct", "handler should have struct definition")
		assert.Contains(t, handlerContent, "NewUserHandler", "handler should have constructor")

		// The Fiber app is created in server.go
		serverPath := filepath.Join(projectDir, "internal", "entrypoints", "http", "server.go")
		assert.FileExists(t, serverPath, "server file should exist")
		serverBytes, err := os.ReadFile(serverPath)
		require.NoError(t, err)
		serverContent := string(serverBytes)
		assert.Contains(t, serverContent, "fiber.New", "server should create Fiber app")
		assert.Contains(t, serverContent, "middleware.CorsMiddleware", "server should wire CORS middleware")
		assert.Contains(t, serverContent, "ErrorHandler", "server should reference an ErrorHandler slot")

		// CORS middleware definition
		corsPath := filepath.Join(projectDir, "internal", "entrypoints", "http", "middleware", "cors.go")
		assert.FileExists(t, corsPath, "CORS middleware file should exist")
		corsBytes, err := os.ReadFile(corsPath)
		require.NoError(t, err)
		assert.Contains(t, string(corsBytes), "cors.New", "cors middleware should configure CORS")

		// Route registration lives in the router files
		routerPath := filepath.Join(projectDir, "internal", "entrypoints", "http", "routes", "user_router.go")
		assert.FileExists(t, routerPath, "user router file should exist")
		routerBytes, err := os.ReadFile(routerPath)
		require.NoError(t, err)
		routerContent := string(routerBytes)
		assert.Contains(t, routerContent, "SetupRoutes(router fiber.Router)", "router should expose SetupRoutes")
		assert.Contains(t, routerContent, "userRouter.Post(", "router should register the POST route")

		// Verify fx Module file for the handler package
		modulePath := filepath.Join(projectDir, "internal", "entrypoints", "http", "handler", "module.go")
		assert.FileExists(t, modulePath, "handler module file should exist")

		moduleContent, err := os.ReadFile(modulePath)
		require.NoError(t, err)
		assert.Contains(t, string(moduleContent), "fx.Module", "module file should define an fx.Module")
	})

	// Step 6: Verify the project compiles
	// Run BEFORE messaging generation: messaging components pull in external
	// broker dependencies (bancolombia/reactive-commons-go, commons-mq-go) that
	// we deliberately do not vendor in this test, so we sanity-check the
	// no-broker baseline first. The messaging file-format checks below are
	// content-only and do not require `go build`.
	t.Run("verify project compiles", func(t *testing.T) {
		// First, run go mod tidy to download dependencies
		tidyCmd := exec.Command("go", "mod", "tidy")
		tidyCmd.Dir = projectDir
		tidyOutput, err := tidyCmd.CombinedOutput()
		if err != nil {
			t.Logf("go mod tidy output: %s", string(tidyOutput))
		}
		require.NoError(t, err, "go mod tidy failed: %s", string(tidyOutput))

		// Try to build the project
		buildCmd := exec.Command("go", "build", "-o", "test-app", "./cmd")
		buildCmd.Dir = projectDir
		buildOutput, err := buildCmd.CombinedOutput()
		if err != nil {
			t.Logf("go build output: %s", string(buildOutput))
		}
		require.NoError(t, err, "go build failed: %s", string(buildOutput))

		// Verify the binary was created
		binaryPath := filepath.Join(projectDir, "test-app")
		assert.FileExists(t, binaryPath, "compiled binary should exist")
	})

	// Step 7: Generate a RabbitMQ event publisher (driven adapter).
	t.Run("generate da publisher rabbitmq event", func(t *testing.T) {
		cmd := exec.Command(cliPath, "generate", "da", "publisher", "UserEventPublisher",
			"--broker", "rabbitmq",
			"--pattern", "event",
			"--target", "user.created")
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "generate da publisher command failed: %s", string(output))

		publisherPath := filepath.Join(projectDir, "internal", "drivenadapters", "messaging", "rabbitmq", "usereventpublisher_publisher.go")
		assert.FileExists(t, publisherPath, "publisher file should exist")

		modulePath := filepath.Join(projectDir, "internal", "drivenadapters", "messaging", "rabbitmq", "module.go")
		assert.FileExists(t, modulePath, "per-component publisher module file should exist")

		foundationPath := filepath.Join(projectDir, "internal", "infrastructure", "messaging", "rabbitmq", "module.go")
		assert.FileExists(t, foundationPath, "rabbitmq foundation module file should exist")

		mainGoPath := filepath.Join(projectDir, "cmd", "main.go")
		mainContent, err := os.ReadFile(mainGoPath)
		require.NoError(t, err)
		assert.Contains(t, string(mainContent), "publishers.Module",
			"main.go should be auto-wired with the publisher module")
		assert.Contains(t, string(mainContent), "rabbitmqinfra.Module",
			"main.go should be auto-wired with the rabbitmq foundation module")

		publisherContent, err := os.ReadFile(publisherPath)
		require.NoError(t, err)
		assert.Contains(t, string(publisherContent), "package publishers",
			"publisher file should declare the publishers package")
		assert.Contains(t, string(publisherContent), "reactive-commons-go",
			"publisher should reference the reactive-commons-go library")
	})

	// Step 8: Generate a RabbitMQ event consumer (entry-point).
	t.Run("generate ep consumer rabbitmq event", func(t *testing.T) {
		cmd := exec.Command(cliPath, "generate", "ep", "consumer", "UserEventConsumer",
			"--broker", "rabbitmq",
			"--pattern", "event",
			"--usecases", "createUser:CreateUserUseCase",
			"--queue", "user.created")
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "generate ep consumer command failed: %s", string(output))

		consumerPath := filepath.Join(projectDir, "internal", "entrypoints", "messaging", "usereventconsumer_consumer.go")
		assert.FileExists(t, consumerPath, "consumer file should exist")

		modulePath := filepath.Join(projectDir, "internal", "entrypoints", "messaging", "usereventconsumer_module.go")
		assert.FileExists(t, modulePath, "per-component consumer module file should exist")

		mainGoPath := filepath.Join(projectDir, "cmd", "main.go")
		mainContent, err := os.ReadFile(mainGoPath)
		require.NoError(t, err)
		assert.Contains(t, string(mainContent), "consumers.UserEventConsumerModule",
			"main.go should be auto-wired with the consumer module")

		consumerContent, err := os.ReadFile(consumerPath)
		require.NoError(t, err)
		assert.Contains(t, string(consumerContent), "package consumers",
			"consumer file should declare the consumers package")
		assert.Contains(t, string(consumerContent), "ListenEvent",
			"consumer should register itself via ListenEvent")
	})

	// Step 9: Re-running with the same name must be idempotent (no duplicate
	// imports or fx.New entries in cmd/main.go).
	t.Run("re-generating consumer is idempotent", func(t *testing.T) {
		cmd := exec.Command(cliPath, "generate", "ep", "consumer", "UserEventConsumer",
			"--broker", "rabbitmq",
			"--pattern", "event",
			"--usecases", "createUser:CreateUserUseCase",
			"--queue", "user.created")
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "re-generate consumer failed: %s", string(output))

		mainGoPath := filepath.Join(projectDir, "cmd", "main.go")
		mainContent, err := os.ReadFile(mainGoPath)
		require.NoError(t, err)
		got := strings.Count(string(mainContent), "consumers.UserEventConsumerModule")
		assert.Equal(t, 1, got, "consumer module should be wired exactly once")
	})

	// Step 10: Kafka is reserved as a placeholder; the CLI must reject it.
	t.Run("kafka broker rejected as not implemented", func(t *testing.T) {
		cmd := exec.Command(cliPath, "generate", "ep", "consumer", "OrderConsumer",
			"--broker", "kafka",
			"--pattern", "event")
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		require.Error(t, err, "kafka should be rejected; got success with output: %s", string(output))
	})

	// Step 11: Verify all files are in correct locations.
	t.Run("verify file locations", func(t *testing.T) {
		expectedFiles := []string{
			"go.mod",
			"README.md",
			"Makefile",
			"cmd/main.go",
			"config/config.go",
			"config/config.yaml",
			"internal/domain/user.go",
			"internal/domain/user_repository.go",
			"internal/usecase/createuser_usecase.go",
			"internal/entrypoints/http/handler/user_handler.go",
			"internal/entrypoints/http/handler/module.go",
			"internal/entrypoints/http/server.go",
			"internal/entrypoints/http/middleware/cors.go",
			"internal/entrypoints/http/routes/routes.go",
			"internal/entrypoints/http/routes/user_router.go",
			"internal/entrypoints/messaging/usereventconsumer_consumer.go",
			"internal/entrypoints/messaging/usereventconsumer_module.go",
			"internal/drivenadapters/messaging/rabbitmq/usereventpublisher_publisher.go",
			"internal/drivenadapters/messaging/rabbitmq/module.go",
			"internal/infrastructure/messaging/rabbitmq/module.go",
		}

		for _, file := range expectedFiles {
			filePath := filepath.Join(projectDir, file)
			assert.FileExists(t, filePath, "file %s should exist in correct location", file)
		}
	})
}

// TestMultipleComponentGeneration tests generating multiple components of the same type
func TestMultipleComponentGeneration(t *testing.T) {
	tmpDir := t.TempDir()
	cliPath := buildCLI(t)

	projectName := "multi-component-test"
	modulePath := "github.com/test/multi"
	projectDir := filepath.Join(tmpDir, projectName)

	// Initialize project
	initCmd := exec.Command(cliPath, "init", projectName,
		"--module", modulePath,
		"--output", tmpDir)
	output, err := initCmd.CombinedOutput()
	require.NoError(t, err, "init command failed: %s", string(output))

	// Generate multiple entities
	entities := []struct {
		name   string
		fields string
	}{
		{"User", "Name:string,Email:string"},
		{"Product", "Title:string,Price:float64"},
		{"Order", "UserID:string,ProductID:string,Quantity:int"},
	}

	for _, entity := range entities {
		t.Run("generate entity "+entity.name, func(t *testing.T) {
			cmd := exec.Command(cliPath, "generate", "entity", entity.name,
				"--fields", entity.fields)
			cmd.Dir = projectDir
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "generate entity %s failed: %s", entity.name, string(output))

			// Verify entity file exists
			entityFile := strings.ToLower(entity.name) + ".go"
			entityPath := filepath.Join(projectDir, "internal", "domain", entityFile)
			assert.FileExists(t, entityPath, "entity file for %s should exist", entity.name)

			// Verify repository file exists
			repoFile := strings.ToLower(entity.name) + "_repository.go"
			repoPath := filepath.Join(projectDir, "internal", "domain", repoFile)
			assert.FileExists(t, repoPath, "repository file for %s should exist", entity.name)
		})
	}

	// Verify project still compiles with multiple entities
	t.Run("verify compilation with multiple entities", func(t *testing.T) {
		tidyCmd := exec.Command("go", "mod", "tidy")
		tidyCmd.Dir = projectDir
		tidyOutput, err := tidyCmd.CombinedOutput()
		require.NoError(t, err, "go mod tidy failed: %s", string(tidyOutput))

		buildCmd := exec.Command("go", "build", "-o", "test-app", "./cmd")
		buildCmd.Dir = projectDir
		buildOutput, err := buildCmd.CombinedOutput()
		require.NoError(t, err, "go build failed: %s", string(buildOutput))
	})
}

// buildCLI builds the CLI tool and returns the path to the binary
func buildCLI(t *testing.T) string {
	t.Helper()

	// Get the project root (assuming tests are in tests/integration)
	projectRoot, err := filepath.Abs("../..")
	require.NoError(t, err, "failed to get project root")

	// Build the CLI
	tmpBin := filepath.Join(t.TempDir(), "scaffold-go")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, "./cmd")
	buildCmd.Dir = projectRoot
	output, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "failed to build CLI: %s", string(output))

	return tmpBin
}
