package generator

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/gabheadz/scaffold-go/internal/template"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Custom generators for use case property-based testing

// genValidUseCaseName generates valid use case names (PascalCase)
func genValidUseCaseName() gopter.Gen {
	return gen.RegexMatch(`^[A-Z][a-zA-Z0-9]{0,30}`)
}

// genDependencySpec generates valid dependency specifications in "name:Type" format
func genDependencySpec() gopter.Gen {
	return gen.SliceOf(gen.OneGenOf(
		gen.Const("customerRepo:CustomerRepository"),
		gen.Const("orderRepo:OrderRepository"),
		gen.Const("productService:ProductService"),
		gen.Const("paymentService:PaymentService"),
		gen.Const("notificationService:NotificationService"),
		gen.Const("userRepo:UserRepository"),
		gen.Const("inventoryService:InventoryService"),
	))
}

// Feature: go-microservice-scaffolding-cli, Property 13: Use case generation includes struct and interface
// **Validates: Requirements 3.1**
func TestProperty_UseCaseGenerationIncludesStructAndInterface(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated use case contains both interface and implementation struct", prop.ForAll(
		func(useCaseName string, dependencies []string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator
			gen := NewUseCaseGenerator(useCaseName, dependencies, tmpDir, templates)

			// Validate
			if err := gen.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			// Generate
			files, err := gen.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			// Should generate exactly 2 files: usecase and test
			if len(files) != 3 {
				t.Logf("expected 3 files, got %d", len(files))
				return false
			}

			content := files[0].Content

			// Check for interface definition
			interfaceName := useCaseName + "UseCase"
			if !strings.Contains(content, "type "+interfaceName+" interface") {
				t.Logf("interface definition not found for %s", interfaceName)
				return false
			}

			// Check for implementation struct
			// The implementation struct should be fully lowercase version of use case name + "Impl"
			implStructName := strings.ToLower(useCaseName) + "Impl"
			if !strings.Contains(content, "type "+implStructName+" struct") {
				t.Logf("implementation struct not found for %s", implStructName)
				return false
			}

			// Check for constructor function
			constructorName := "func New" + useCaseName + "UseCase"
			if !strings.Contains(content, constructorName) {
				t.Logf("constructor function not found: %s", constructorName)
				return false
			}

			return true
		},
		genValidUseCaseName(),
		genDependencySpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 14: Use case constructor includes specified dependencies
// **Validates: Requirements 3.2**
func TestProperty_UseCaseConstructorIncludesSpecifiedDependencies(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated use case constructor includes all specified dependencies", prop.ForAll(
		func(useCaseName string, dependencies []string) bool {
			// Skip if no dependencies
			if len(dependencies) == 0 {
				return true
			}

			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator
			gen := NewUseCaseGenerator(useCaseName, dependencies, tmpDir, templates)

			// Validate
			if err := gen.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			// Generate
			files, err := gen.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			if len(files) != 3 {
				t.Logf("expected 3 files, got %d", len(files))
				return false
			}

			content := files[0].Content

			// Verify all dependencies are present in the struct
			for _, dep := range dependencies {
				parts := strings.Split(dep, ":")
				if len(parts) != 2 {
					continue
				}

				depName := strings.TrimSpace(parts[0])
				depType := strings.TrimSpace(parts[1])

				// Check dependency field in struct
				if !strings.Contains(content, depName+" domain."+depType) {
					t.Logf("dependency field %s %s not found in struct", depName, depType)
					return false
				}

				// Check dependency in constructor parameters
				constructorName := "func New" + useCaseName + "UseCase"
				constructorIdx := strings.Index(content, constructorName)
				if constructorIdx == -1 {
					t.Logf("constructor not found")
					return false
				}

				// Extract constructor signature (up to the opening brace)
				constructorSection := content[constructorIdx:]
				braceIdx := strings.Index(constructorSection, "{")
				if braceIdx == -1 {
					t.Logf("constructor body not found")
					return false
				}
				constructorSignature := constructorSection[:braceIdx]

				// Check dependency parameter in constructor
				if !strings.Contains(constructorSignature, depName+" domain."+depType) {
					t.Logf("dependency parameter %s %s not found in constructor", depName, depType)
					return false
				}
			}

			return true
		},
		genValidUseCaseName(),
		genDependencySpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 15: Use case includes Execute method
// **Validates: Requirements 3.3**
func TestProperty_UseCaseIncludesExecuteMethod(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated use case contains Execute method", prop.ForAll(
		func(useCaseName string, dependencies []string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator
			gen := NewUseCaseGenerator(useCaseName, dependencies, tmpDir, templates)

			// Validate
			if err := gen.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			// Generate
			files, err := gen.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			if len(files) != 3 {
				t.Logf("expected 3 files, got %d", len(files))
				return false
			}

			content := files[0].Content

			// Check for Execute method in interface
			if !strings.Contains(content, "Execute(ctx context.Context, input any) (any, error)") {
				t.Logf("Execute method not found in interface")
				return false
			}

			// Check for Execute method implementation
			implStructName := strings.ToLower(useCaseName) + "Impl"
			executeMethodSignature := "func (uc *" + implStructName + ") Execute(ctx context.Context, input any) (any, error)"
			if !strings.Contains(content, executeMethodSignature) {
				t.Logf("Execute method implementation not found")
				return false
			}

			// Verify Execute method accepts context.Context
			if !strings.Contains(content, "context.Context") {
				t.Logf("Execute method does not accept context.Context")
				return false
			}

			return true
		},
		genValidUseCaseName(),
		genDependencySpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 16: Use case depends only on domain interfaces
// **Validates: Requirements 3.5**
func TestProperty_UseCaseDependsOnlyOnDomainInterfaces(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated use case file imports only standard library and domain packages", prop.ForAll(
		func(useCaseName string, dependencies []string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator
			gen := NewUseCaseGenerator(useCaseName, dependencies, tmpDir, templates)

			// Validate
			if err := gen.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			// Generate
			files, err := gen.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			if len(files) != 3 {
				t.Logf("expected 3 files, got %d", len(files))
				return false
			}

			content := files[0].Content

			// Verify file is in domain package
			if !strings.Contains(content, "package usecase") {
				t.Logf("use case not in usecase package")
				return false
			}

			// Check imports - should only have standard library imports
			// Use cases should not import external frameworks or infrastructure packages
			importStart := strings.Index(content, "import")
			if importStart == -1 {
				// No imports is fine
				return true
			}

			// Find the import block
			importSection := content[importStart:]
			importEnd := strings.Index(importSection, ")")
			if importEnd == -1 {
				// Single import line
				importEnd = strings.Index(importSection, "\n")
			}

			if importEnd != -1 {
				importSection = importSection[:importEnd]
			}

			// Check for external dependencies (should not have framework dependencies)
			// Allowed: standard library (context, fmt, errors, time, etc.)
			// Not allowed: external frameworks (fiber, gin, gorm, etc.)
			lines := strings.Split(importSection, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				// Skip empty lines and the import keyword
				if line == "" || line == "import" || line == "import (" {
					continue
				}

				// Remove quotes
				line = strings.Trim(line, "\"")

				// Check for framework dependencies (these should not be in domain layer)
				forbiddenPackages := []string{
					"github.com/gofiber/fiber",
					"github.com/gin-gonic/gin",
					"gorm.io/gorm",
					"github.com/jmoiron/sqlx",
					"go.mongodb.org/mongo-driver",
					"github.com/aws/aws-sdk-go",
					"github.com/redis/go-redis",
				}

				for _, forbidden := range forbiddenPackages {
					if strings.Contains(line, forbidden) {
						t.Logf("found forbidden framework dependency: %s", line)
						return false
					}
				}

				// Check for infrastructure layer imports (drivenadapters, entrypoints)
				if strings.Contains(line, "/drivenadapters/") || strings.Contains(line, "/entrypoints/") {
					t.Logf("found infrastructure layer import: %s", line)
					return false
				}
			}

			// Verify use case is in the correct directory (domain)
			if !strings.Contains(files[0].Path, filepath.Join("internal", "usecase")) {
				t.Logf("use case file not in usecase directory")
				return false
			}

			return true
		},
		genValidUseCaseName(),
		genDependencySpec(),
	))

	properties.TestingRun(t)
}
