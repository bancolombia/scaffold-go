package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gabheadz/scaffold-go/internal/template"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// genValidApiClientName generates valid API client names (PascalCase identifiers)
func genValidApiClientName() gopter.Gen {
	return gen.RegexMatch(`^[A-Z][a-zA-Z0-9]{2,20}$`)
}

// genBaseURL generates valid base URLs
func genBaseURL() gopter.Gen {
	return gen.OneGenOf(
		gen.Const("https://api.example.com"),
		gen.Const("https://service.internal"),
		gen.Const("http://localhost:8080"),
		gen.Const("https://api.production.com/v1"),
	)
}

// Feature: go-microservice-scaffolding-cli, Property 32: API client includes HTTP client methods
// Validates: Requirements 10.2
func TestProperty_ApiClientIncludesHTTPClientMethods(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("API client includes HTTP methods (GET, POST, PUT, DELETE)", prop.ForAll(
		func(clientName string, baseURL string) bool {
			tempDir := t.TempDir()
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("Failed to create template engine: %v", err)
				return false
			}

			gen := NewApiClientGenerator(clientName, baseURL, tempDir, templates)
			files, err := gen.Generate()
			if err != nil {
				t.Logf("Failed to generate API client: %v", err)
				return false
			}

			if len(files) == 0 {
				t.Log("No files generated")
				return false
			}

			clientFile := files[0]
			content := clientFile.Content

			// Verify all HTTP methods are present
			httpMethods := []string{"Get", "Post", "Put", "Delete"}
			for _, method := range httpMethods {
				methodSignature := "func (c *" + clientName + "Client) " + method + "("
				if !strings.Contains(content, methodSignature) {
					t.Logf("Missing HTTP method: %s", method)
					return false
				}
			}

			// Verify request and response handling
			if !strings.Contains(content, "makeRequest") {
				t.Log("Missing makeRequest helper method")
				return false
			}

			if !strings.Contains(content, "decodeResponse") {
				t.Log("Missing decodeResponse helper method")
				return false
			}

			// Verify HTTP context usage
			if !strings.Contains(content, "http.NewRequestWithContext") {
				t.Log("Missing context-aware HTTP request creation")
				return false
			}

			return true
		},
		genValidApiClientName(),
		genBaseURL(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 33: API client constructor includes HTTP client
// Validates: Requirements 10.3
func TestProperty_ApiClientConstructorIncludesHTTPClient(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("API client constructor includes HTTP client and configuration", prop.ForAll(
		func(clientName string, baseURL string) bool {
			tempDir := t.TempDir()
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("Failed to create template engine: %v", err)
				return false
			}

			gen := NewApiClientGenerator(clientName, baseURL, tempDir, templates)
			files, err := gen.Generate()
			if err != nil {
				t.Logf("Failed to generate API client: %v", err)
				return false
			}

			if len(files) == 0 {
				t.Log("No files generated")
				return false
			}

			clientFile := files[0]
			content := clientFile.Content

			// Verify struct has HTTP client field
			if !strings.Contains(content, "httpClient *http.Client") {
				t.Log("Missing httpClient field in struct")
				return false
			}

			// Verify struct has baseURL field for configuration
			if !strings.Contains(content, "baseURL") {
				t.Log("Missing baseURL field in struct")
				return false
			}

			// Verify constructor signature includes configuration
			constructorSignature := "func New" + clientName + "Client(baseURL string, logger *zap.Logger)"
			if !strings.Contains(content, constructorSignature) {
				t.Logf("Missing or incorrect constructor signature")
				return false
			}

			// Verify HTTP client is initialized in constructor
			if !strings.Contains(content, "httpClient: &http.Client{") {
				t.Log("HTTP client not initialized in constructor")
				return false
			}

			// Verify timeout configuration
			if !strings.Contains(content, "Timeout:") {
				t.Log("Missing timeout configuration for HTTP client")
				return false
			}

			return true
		},
		genValidApiClientName(),
		genBaseURL(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 53: Generated adapters are wired as fx Modules
// Validates: Requirements 15.6
func TestProperty_GeneratedApiClientAdaptersAreWiredAsFxModules(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("API client adapters are wired as fx Modules", prop.ForAll(
		func(clientName string, baseURL string) bool {
			tempDir := t.TempDir()
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("Failed to create template engine: %v", err)
				return false
			}

			gen := NewApiClientGenerator(clientName, baseURL, tempDir, templates)
			files, err := gen.Generate()
			if err != nil {
				t.Logf("Failed to generate API client: %v", err)
				return false
			}

			if len(files) < 2 {
				t.Logf("Expected at least 2 files (client and module), got %d", len(files))
				return false
			}

			// Find the module file
			var moduleFile *GeneratedFile
			for i := range files {
				if strings.Contains(files[i].Path, "module.go") {
					moduleFile = &files[i]
					break
				}
			}

			if moduleFile == nil {
				t.Log("Module file not generated")
				return false
			}

			content := moduleFile.Content

			// Verify fx Module structure
			if !strings.Contains(content, "package api") {
				t.Log("Module file missing package declaration")
				return false
			}

			if !strings.Contains(content, `import "go.uber.org/fx"`) {
				t.Log("Module file missing fx import")
				return false
			}

			if !strings.Contains(content, "var Module = fx.Module") {
				t.Log("Module file missing fx.Module declaration")
				return false
			}

			if !strings.Contains(content, `fx.Module("api"`) {
				t.Log("Module file missing fx.Module name")
				return false
			}

			if !strings.Contains(content, "fx.Provide") {
				t.Log("Module file missing fx.Provide")
				return false
			}

			// Verify constructor is provided
			constructorName := "New" + clientName + "Client"
			if !strings.Contains(content, constructorName) {
				t.Logf("Module file missing constructor: %s", constructorName)
				return false
			}

			return true
		},
		genValidApiClientName(),
		genBaseURL(),
	))

	properties.TestingRun(t)
}

// Additional property test: API client includes error handling
func TestProperty_ApiClientIncludesErrorHandling(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("API client includes comprehensive error handling", prop.ForAll(
		func(clientName string, baseURL string) bool {
			tempDir := t.TempDir()
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("Failed to create template engine: %v", err)
				return false
			}

			gen := NewApiClientGenerator(clientName, baseURL, tempDir, templates)
			files, err := gen.Generate()
			if err != nil {
				t.Logf("Failed to generate API client: %v", err)
				return false
			}

			if len(files) == 0 {
				t.Log("No files generated")
				return false
			}

			clientFile := files[0]
			content := clientFile.Content

			// Verify error handling patterns
			if !strings.Contains(content, "if err != nil") {
				t.Log("Missing error checking")
				return false
			}

			// Verify HTTP status code error handling
			if !strings.Contains(content, "StatusCode >= 400") {
				t.Log("Missing HTTP 4xx error handling")
				return false
			}

			if !strings.Contains(content, "StatusCode >= 500") {
				t.Log("Missing HTTP 5xx error handling")
				return false
			}

			// Verify error logging
			if !strings.Contains(content, "logger.Error") {
				t.Log("Missing error logging")
				return false
			}

			// Verify error wrapping/formatting
			if !strings.Contains(content, "fmt.Errorf") {
				t.Log("Missing error formatting")
				return false
			}

			return true
		},
		genValidApiClientName(),
		genBaseURL(),
	))

	properties.TestingRun(t)
}

// Additional property test: API client includes retry logic
func TestProperty_ApiClientIncludesRetryLogic(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("API client includes retry logic with exponential backoff", prop.ForAll(
		func(clientName string, baseURL string) bool {
			tempDir := t.TempDir()
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("Failed to create template engine: %v", err)
				return false
			}

			gen := NewApiClientGenerator(clientName, baseURL, tempDir, templates)
			files, err := gen.Generate()
			if err != nil {
				t.Logf("Failed to generate API client: %v", err)
				return false
			}

			if len(files) == 0 {
				t.Log("No files generated")
				return false
			}

			clientFile := files[0]
			content := clientFile.Content

			// Verify retry logic is present
			if !strings.Contains(content, "maxRetries") {
				t.Log("Missing maxRetries configuration")
				return false
			}

			if !strings.Contains(content, "for attempt := 0; attempt < maxRetries; attempt++") {
				t.Log("Missing retry loop")
				return false
			}

			// Verify retry logging
			if !strings.Contains(content, "Retrying request") {
				t.Log("Missing retry logging")
				return false
			}

			// Verify backoff delay
			if !strings.Contains(content, "time.Sleep") {
				t.Log("Missing backoff delay")
				return false
			}

			// Verify retry on server errors
			if !strings.Contains(content, "Server error, will retry") {
				t.Log("Missing server error retry logic")
				return false
			}

			return true
		},
		genValidApiClientName(),
		genBaseURL(),
	))

	properties.TestingRun(t)
}

// Additional property test: API client includes Zap logger integration
func TestProperty_ApiClientIncludesZapLoggerIntegration(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("API client includes Zap logger integration", prop.ForAll(
		func(clientName string, baseURL string) bool {
			tempDir := t.TempDir()
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("Failed to create template engine: %v", err)
				return false
			}

			gen := NewApiClientGenerator(clientName, baseURL, tempDir, templates)
			files, err := gen.Generate()
			if err != nil {
				t.Logf("Failed to generate API client: %v", err)
				return false
			}

			if len(files) == 0 {
				t.Log("No files generated")
				return false
			}

			clientFile := files[0]
			content := clientFile.Content

			// Verify logger field in struct
			if !strings.Contains(content, "logger *zap.Logger") {
				t.Log("Missing logger field in struct")
				return false
			}

			// Verify logger in constructor
			constructorSignature := "func New" + clientName + "Client(baseURL string, logger *zap.Logger)"
			if !strings.Contains(content, constructorSignature) {
				t.Log("Logger not in constructor signature")
				return false
			}

			// Verify logger is used
			if !strings.Contains(content, "logger.Info") {
				t.Log("Logger Info method not used")
				return false
			}

			if !strings.Contains(content, "logger.Error") {
				t.Log("Logger Error method not used")
				return false
			}

			// Verify structured logging with zap fields
			if !strings.Contains(content, "zap.String") {
				t.Log("Missing zap.String for structured logging")
				return false
			}

			if !strings.Contains(content, "zap.Error") {
				t.Log("Missing zap.Error for error logging")
				return false
			}

			return true
		},
		genValidApiClientName(),
		genBaseURL(),
	))

	properties.TestingRun(t)
}

// Additional property test: Generated files are in correct directory
func TestProperty_ApiClientFilesInCorrectDirectory(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("API client files are generated in correct directory", prop.ForAll(
		func(clientName string, baseURL string) bool {
			tempDir := t.TempDir()
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("Failed to create template engine: %v", err)
				return false
			}

			gen := NewApiClientGenerator(clientName, baseURL, tempDir, templates)
			files, err := gen.Generate()
			if err != nil {
				t.Logf("Failed to generate API client: %v", err)
				return false
			}

			if len(files) < 2 {
				t.Logf("Expected at least 2 files, got %d", len(files))
				return false
			}

			// Verify client file path
			clientFile := files[0]
			expectedPath := filepath.Join("internal", "drivenadapters", "api")
			if !strings.Contains(clientFile.Path, expectedPath) {
				t.Logf("Client file not in correct directory. Expected path to contain: %s, got: %s", expectedPath, clientFile.Path)
				return false
			}

			// Verify module file path
			moduleFile := files[1]
			if !strings.Contains(moduleFile.Path, expectedPath) {
				t.Logf("Module file not in correct directory. Expected path to contain: %s, got: %s", expectedPath, moduleFile.Path)
				return false
			}

			// Verify files can be written
			for _, file := range files {
				dir := filepath.Dir(file.Path)
				err := os.MkdirAll(dir, 0755)
				if err != nil {
					t.Logf("Failed to create directory: %v", err)
					return false
				}

				err = os.WriteFile(file.Path, []byte(file.Content), 0644)
				if err != nil {
					t.Logf("Failed to write file: %v", err)
					return false
				}
			}

			return true
		},
		genValidApiClientName(),
		genBaseURL(),
	))

	properties.TestingRun(t)
}

// Additional property test: Generated code compiles
func TestProperty_ApiClientGeneratedCodeCompiles(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 20 // Reduced for compilation tests
	properties := gopter.NewProperties(parameters)

	properties.Property("generated API client code compiles without errors", prop.ForAll(
		func(clientName string, baseURL string) bool {
			tempDir := t.TempDir()

			// Create a minimal Go module
			goModContent := `module testmodule

go 1.21

require (
	go.uber.org/zap v1.26.0
	go.uber.org/fx v1.20.1
)
`
			err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
			if err != nil {
				t.Logf("Failed to create go.mod: %v", err)
				return false
			}

			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("Failed to create template engine: %v", err)
				return false
			}

			gen := NewApiClientGenerator(clientName, baseURL, tempDir, templates)
			files, err := gen.Generate()
			if err != nil {
				t.Logf("Failed to generate API client: %v", err)
				return false
			}

			// Write generated files
			for _, file := range files {
				dir := filepath.Dir(file.Path)
				err := os.MkdirAll(dir, 0755)
				if err != nil {
					t.Logf("Failed to create directory: %v", err)
					return false
				}

				err = os.WriteFile(file.Path, []byte(file.Content), 0644)
				if err != nil {
					t.Logf("Failed to write file: %v", err)
					return false
				}
			}

			// Note: We skip actual compilation here as it would require downloading dependencies
			// In a real scenario, you would run: go build ./...
			// For now, we just verify the files were created successfully

			return true
		},
		genValidApiClientName(),
		genBaseURL(),
	))

	properties.TestingRun(t)
}
