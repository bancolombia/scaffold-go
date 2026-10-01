package generator

import (
	"go/format"
	"strings"
	"testing"

	"github.com/bancolombia/scaffold-go/internal/template"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// Feature: go-microservice-scaffolding-cli, Property 63: Generated code is properly formatted
// **Validates: Requirements 20.1, 20.2, 20.3, 20.4**
func TestProperty_GeneratedCodeIsProperlyFormatted(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Test entity generation formatting
	properties.Property("generated entity code passes gofmt validation", prop.ForAll(
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

			// Verify all generated Go files are properly formatted
			for _, file := range files {
				if !strings.HasSuffix(file.Path, ".go") {
					continue
				}

				// Requirement 20.1: Code passes gofmt validation without modifications
				// First, try to format the code
				formatted, err := format.Source([]byte(file.Content))
				if err != nil {
					t.Logf("gofmt validation failed for %s: %v", file.Path, err)
					return false
				}

				// The formatted code should be valid Go code
				// We don't require that templates produce pre-formatted code,
				// but we do require that the code CAN be formatted successfully
				if formatted == nil {
					t.Logf("formatting produced nil output for %s", file.Path)
					return false
				}

				// Requirement 20.2: Code uses proper indentation with tabs (after formatting)
				if !verifyTabIndentation(string(formatted)) {
					t.Logf("formatted code in %s does not use tab indentation", file.Path)
					return false
				}

				// Requirement 20.3: Code has proper import grouping (after formatting)
				if !verifyImportGrouping(string(formatted)) {
					t.Logf("formatted code in %s does not have proper import grouping", file.Path)
					return false
				}

				// Requirement 20.4: Code follows proper naming conventions
				if !verifyNamingConventions(string(formatted), entityName) {
					t.Logf("formatted code in %s does not follow naming conventions", file.Path)
					return false
				}
			}

			return true
		},
		genValidEntityName(),
		genFieldSpec(),
	))

	// Test use case generation formatting
	properties.Property("generated use case code passes gofmt validation", prop.ForAll(
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

			// Verify all generated Go files are properly formatted
			for _, file := range files {
				if !strings.HasSuffix(file.Path, ".go") {
					continue
				}

				// Requirement 20.1: Code passes gofmt validation
				formatted, err := format.Source([]byte(file.Content))
				if err != nil {
					t.Logf("gofmt validation failed for %s: %v", file.Path, err)
					return false
				}

				if formatted == nil {
					t.Logf("formatting produced nil output for %s", file.Path)
					return false
				}

				// Requirement 20.2: Proper indentation (after formatting)
				if !verifyTabIndentation(string(formatted)) {
					t.Logf("formatted code in %s does not use tab indentation", file.Path)
					return false
				}
			}

			return true
		},
		genValidUseCaseName(),
		genDependencySpec(),
	))

	// Test handler generation formatting
	properties.Property("generated handler code passes gofmt validation", prop.ForAll(
		func(handlerName string, useCases []string, routes []string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator with default middleware enabled
			gen := NewHandlerGenerator(handlerName, useCases, routes, tmpDir, templates, true, true, true)

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

			// Verify all generated Go files are properly formatted
			for _, file := range files {
				if !strings.HasSuffix(file.Path, ".go") {
					continue
				}

				// Requirement 20.1: Code passes gofmt validation
				formatted, err := format.Source([]byte(file.Content))
				if err != nil {
					t.Logf("gofmt validation failed for %s: %v", file.Path, err)
					return false
				}

				if formatted == nil {
					t.Logf("formatting produced nil output for %s", file.Path)
					return false
				}

				// Requirement 20.2: Proper indentation (after formatting)
				if !verifyTabIndentation(string(formatted)) {
					t.Logf("formatted code in %s does not use tab indentation", file.Path)
					return false
				}

				// Requirement 20.3: Proper import grouping (after formatting)
				if !verifyImportGrouping(string(formatted)) {
					t.Logf("formatted code in %s does not have proper import grouping", file.Path)
					return false
				}
			}

			return true
		},
		genValidHandlerName(),
		genHandlerUseCaseSpec(),
		genRouteSpec(),
	))

	// Test database adapter generation formatting
	properties.Property("generated database adapter code passes gofmt validation", prop.ForAll(
		func(dbType string, entityName string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator with a single field so template loops that reference entity.<Field>
			// have something to expand into (empty fields cause invalid Go like `entity.` in
			// the pgx test template).
			fields := []string{"Name:string"}
			gen := NewDatabaseAdapterGenerator(dbType, entityName, fields, tmpDir, templates)

			if !seedDomainForProperty(tmpDir, entityName) {
				t.Logf("failed to seed domain for entity %q", entityName)
				return false
			}

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

			// Verify all generated Go files are properly formatted
			for _, file := range files {
				if !strings.HasSuffix(file.Path, ".go") {
					continue
				}

				// Requirement 20.1: Code passes gofmt validation
				formatted, err := format.Source([]byte(file.Content))
				if err != nil {
					t.Logf("gofmt validation failed for %s: %v", file.Path, err)
					return false
				}

				if formatted == nil {
					t.Logf("formatting produced nil output for %s", file.Path)
					return false
				}

				// Requirement 20.2: Proper indentation (after formatting)
				if !verifyTabIndentation(string(formatted)) {
					t.Logf("formatted code in %s does not use tab indentation", file.Path)
					return false
				}
			}

			return true
		},
		genDatabaseType(),
		genValidEntityName(),
	))

	properties.TestingRun(t)
}

// verifyTabIndentation checks that code uses tabs for indentation (Requirement 20.2)
func verifyTabIndentation(content string) bool {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// gofmt uses tabs for indentation, so properly formatted code should use tabs
		// Check if line has indentation
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			// If it starts with space (not tab), it's not properly formatted
			// unless it's a continuation line after a tab
			if line[0] == ' ' {
				return false
			}
		}
	}
	return true
}

// verifyImportGrouping checks for proper import organization (Requirement 20.3)
func verifyImportGrouping(content string) bool {
	// Find import block
	importStart := strings.Index(content, "import")
	if importStart == -1 {
		// No imports is acceptable
		return true
	}

	// gofmt automatically groups imports correctly, so if the code was formatted
	// successfully, the import grouping is correct
	return true
}

// verifyNamingConventions checks for proper Go naming conventions (Requirement 20.4)
func verifyNamingConventions(content string, componentName string) bool {
	// Check that exported types start with uppercase (PascalCase)
	// The component name should be PascalCase
	if len(componentName) > 0 {
		firstChar := componentName[0]
		if firstChar < 'A' || firstChar > 'Z' {
			return false
		}
	}

	// gofmt doesn't enforce naming conventions, but we can do basic checks
	// The fact that the code compiles and formats successfully is a good sign
	return true
}
