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

// Custom generators for property-based testing

// genValidEntityName generates valid entity names (PascalCase)
func genValidEntityName() gopter.Gen {
	return gen.RegexMatch(`^[A-Z][a-zA-Z0-9]{0,20}`)
}

// genFieldSpec generates valid field specifications in "Name:Type" format
func genFieldSpec() gopter.Gen {
	// Generate a slice of field specifications
	return gen.SliceOf(gen.OneGenOf(
		gen.Const("Name:string"),
		gen.Const("Email:string"),
		gen.Const("Age:int"),
		gen.Const("ID:string"),
		gen.Const("Count:int64"),
		gen.Const("Price:float64"),
		gen.Const("Active:bool"),
		gen.Const("CreatedAt:time.Time"),
		gen.Const("UpdatedAt:time.Time"),
	))
}

// Feature: go-microservice-scaffolding-cli, Property 8: Entity generation includes specified fields
// **Validates: Requirements 2.2**
func TestProperty_EntityGenerationIncludesSpecifiedFields(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping property-based test in short mode")
	}

	parameters := GetTestParameters()
	properties := gopter.NewProperties(parameters)

	properties.Property("generated entity contains all specified fields", prop.ForAll(
		func(entityName string, fields []string) bool {
			// Skip if no fields
			if len(fields) == 0 {
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
			gen := NewEntityGenerator(entityName, fields, tmpDir, templates)

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

			// Find entity file (not repository)
			var entityContent string
			for _, file := range files {
				if !strings.Contains(file.Path, "_repository.go") {
					entityContent = file.Content
					break
				}
			}

			if entityContent == "" {
				t.Logf("entity file not found")
				return false
			}

			// Verify all fields are present
			for _, field := range fields {
				parts := strings.Split(field, ":")
				fieldName := strings.TrimSpace(parts[0])
				fieldType := strings.TrimSpace(parts[1])

				// Check field name is present
				if !strings.Contains(entityContent, fieldName) {
					t.Logf("field name %s not found in entity", fieldName)
					return false
				}

				// Check field type is present
				if !strings.Contains(entityContent, fieldType) {
					t.Logf("field type %s not found in entity", fieldType)
					return false
				}
			}

			return true
		},
		genValidEntityName(),
		genFieldSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 9: Entity generation includes validation methods
// **Validates: Requirements 2.4**
func TestProperty_EntityGenerationIncludesValidationMethods(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping property-based test in short mode")
	}

	parameters := GetTestParameters()
	properties := gopter.NewProperties(parameters)

	properties.Property("generated entity contains Validate method", prop.ForAll(
		func(entityName string, fields []string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator
			gen := NewEntityGenerator(entityName, fields, tmpDir, templates)

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

			// Find entity file (not repository)
			var entityContent string
			for _, file := range files {
				if !strings.Contains(file.Path, "_repository.go") {
					entityContent = file.Content
					break
				}
			}

			if entityContent == "" {
				t.Logf("entity file not found")
				return false
			}

			// Check for Validate method
			validateMethodSignature := "func (e *" + entityName + ") Validate() error"
			if !strings.Contains(entityContent, validateMethodSignature) {
				t.Logf("Validate method not found in entity")
				return false
			}

			return true
		},
		genValidEntityName(),
		genFieldSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 10: Entity generation creates repository interface
// **Validates: Requirements 2.5**
func TestProperty_EntityGenerationCreatesRepositoryInterface(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping property-based test in short mode")
	}

	parameters := GetTestParameters()
	properties := gopter.NewProperties(parameters)

	properties.Property("entity generation creates both entity and repository files", prop.ForAll(
		func(entityName string, fields []string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator
			gen := NewEntityGenerator(entityName, fields, tmpDir, templates)

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

			// Should generate exactly 3 files: entity, repository, and test
			if len(files) != 3 {
				t.Logf("expected 3 files, got %d", len(files))
				return false
			}

			// Check that we have entity, repository, and test files
			hasEntityFile := false
			hasRepoFile := false
			hasTestFile := false

			for _, file := range files {
				if strings.Contains(file.Path, "_test.go") {
					hasTestFile = true
					// Verify it's in the domain directory
					if !strings.Contains(file.Path, filepath.Join("internal", "domain")) {
						t.Logf("test file not in domain directory")
						return false
					}
				} else if strings.Contains(file.Path, "_repository.go") {
					hasRepoFile = true
					// Verify it's in the domain directory
					if !strings.Contains(file.Path, filepath.Join("internal", "domain")) {
						t.Logf("repository file not in domain directory")
						return false
					}
				} else if strings.HasSuffix(file.Path, ".go") {
					hasEntityFile = true
					// Verify it's in the domain directory
					if !strings.Contains(file.Path, filepath.Join("internal", "domain")) {
						t.Logf("entity file not in domain directory")
						return false
					}
				}
			}

			if !hasEntityFile || !hasRepoFile || !hasTestFile {
				t.Logf("missing entity, repository, or test file")
				return false
			}

			return true
		},
		genValidEntityName(),
		genFieldSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 11: Repository interface includes CRUD methods
// **Validates: Requirements 2.6**
func TestProperty_RepositoryInterfaceIncludesCRUDMethods(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping property-based test in short mode")
	}

	parameters := GetTestParameters()
	properties := gopter.NewProperties(parameters)

	properties.Property("generated repository interface contains CRUD methods", prop.ForAll(
		func(entityName string, fields []string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator
			gen := NewEntityGenerator(entityName, fields, tmpDir, templates)

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

			// Find repository file
			var repoContent string
			for _, file := range files {
				if strings.Contains(file.Path, "_repository.go") {
					repoContent = file.Content
					break
				}
			}

			if repoContent == "" {
				t.Logf("repository file not found")
				return false
			}

			// Check for repository interface definition
			repoInterfaceName := "I" + entityName + "Repository"
			if !strings.Contains(repoContent, "type "+repoInterfaceName+" interface") {
				t.Logf("repository interface definition not found")
				return false
			}

			// Check for CRUD methods
			requiredMethods := []string{
				"Create",
				"GetByID",
				"Update",
				"Delete",
				"List",
			}

			for _, method := range requiredMethods {
				if !strings.Contains(repoContent, method) {
					t.Logf("CRUD method %s not found in repository", method)
					return false
				}
			}

			return true
		},
		genValidEntityName(),
		genFieldSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 12: Entity files have no external dependencies
// **Validates: Requirements 2.8**
func TestProperty_EntityFilesHaveNoExternalDependencies(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping property-based test in short mode")
	}

	parameters := GetTestParameters()
	properties := gopter.NewProperties(parameters)

	properties.Property("generated entity file imports only standard library", prop.ForAll(
		func(entityName string, fields []string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator
			gen := NewEntityGenerator(entityName, fields, tmpDir, templates)

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

			// Find entity file (not repository)
			var entityContent string
			for _, file := range files {
				if !strings.Contains(file.Path, "_repository.go") {
					entityContent = file.Content
					break
				}
			}

			if entityContent == "" {
				t.Logf("entity file not found")
				return false
			}

			// Check imports - should only have standard library imports
			// Standard library imports don't have dots in their paths (except for special cases)
			// External imports typically have domain names like "github.com/..."

			// Extract import section
			importStart := strings.Index(entityContent, "import")
			if importStart == -1 {
				// No imports is fine (no external dependencies)
				return true
			}

			// Find the import block
			importSection := entityContent[importStart:]
			importEnd := strings.Index(importSection, ")")
			if importEnd == -1 {
				// Single import line
				importEnd = strings.Index(importSection, "\n")
			}

			if importEnd != -1 {
				importSection = importSection[:importEnd]
			}

			// Check for external dependencies (contain dots like github.com, gopkg.in, etc.)
			// But allow standard library packages
			lines := strings.Split(importSection, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				// Skip empty lines and the import keyword
				if line == "" || line == "import" || line == "import (" {
					continue
				}

				// Remove quotes
				line = strings.Trim(line, "\"")

				// Skip if it's a standard library import (no dots or known stdlib packages)
				if !strings.Contains(line, ".") {
					continue
				}

				// Check if it's a known standard library package
				// Standard library packages with dots: time, encoding/json, etc.
				stdlibPackages := []string{"time", "fmt", "errors", "context", "encoding/", "net/", "os/"}
				isStdlib := false
				for _, pkg := range stdlibPackages {
					if strings.HasPrefix(line, pkg) || line == pkg {
						isStdlib = true
						break
					}
				}

				if isStdlib {
					continue
				}

				// If it contains a domain-like pattern (e.g., github.com, gopkg.in), it's external
				if strings.Contains(line, "github.com") || strings.Contains(line, "gopkg.in") ||
					strings.Contains(line, "golang.org") || strings.Contains(line, "go.uber.org") {
					t.Logf("found external dependency: %s", line)
					return false
				}
			}

			return true
		},
		genValidEntityName(),
		genFieldSpec(),
	))

	properties.TestingRun(t)
}
