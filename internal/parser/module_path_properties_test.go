package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Custom generators for module path property-based testing

// genSingleSegmentModulePath generates single-segment module paths
func genSingleSegmentModulePath() gopter.Gen {
	return gen.RegexMatch(`^[a-z][a-z0-9_-]{2,20}$`)
}

// genTwoSegmentModulePath generates two-segment module paths (e.g., github.com/project)
func genTwoSegmentModulePath() gopter.Gen {
	return gen.RegexMatch(`^[a-z][a-z0-9_-]{2,15}\.[a-z]{2,5}/[a-z][a-z0-9_-]{2,20}$`)
}

// genThreeSegmentModulePath generates three-segment module paths (e.g., github.com/user/project)
func genThreeSegmentModulePath() gopter.Gen {
	return gen.RegexMatch(`^[a-z][a-z0-9_-]{2,15}\.[a-z]{2,5}/[a-z][a-z0-9_-]{2,20}/[a-z][a-z0-9_-]{2,20}$`)
}

// genModulePathWithHyphens generates module paths with hyphens
func genModulePathWithHyphens() gopter.Gen {
	return gen.OneConstOf(
		"github.com/user/my-project",
		"gitlab.com/org/my-service",
		"example.com/my-app",
		"github.com/my-org/my-microservice",
		"bitbucket.org/team/my-api",
	)
}

// genModulePathWithUnderscores generates module paths with underscores
func genModulePathWithUnderscores() gopter.Gen {
	return gen.OneConstOf(
		"github.com/user/my_project",
		"gitlab.com/org/my_service",
		"example.com/my_app",
		"github.com/my_org/my_microservice",
		"bitbucket.org/team/my_api",
	)
}

// genValidModulePath generates any valid module path
func genValidModulePath() gopter.Gen {
	return gen.OneGenOf(
		genSingleSegmentModulePath(),
		genTwoSegmentModulePath(),
		genThreeSegmentModulePath(),
		genModulePathWithHyphens(),
		genModulePathWithUnderscores(),
	)
}

// genGoModContent generates go.mod file content with the given module path
func genGoModContent(modulePath string) string {
	return "module " + modulePath + "\n\ngo 1.21\n"
}

// genGoModContentWithComments generates go.mod file content with comments before module declaration
func genGoModContentWithComments(modulePath string) string {
	return "// This is a comment\n// Another comment\nmodule " + modulePath + "\n\ngo 1.21\n"
}

// genGoModContentWithWhitespace generates go.mod file content with extra whitespace
func genGoModContentWithWhitespace(modulePath string) string {
	return "\n\n  module   " + modulePath + "  \n\ngo 1.21\n"
}

// Feature: domain-module-management, Property 2: Module Path Parsing
// **Validates: Requirements 2.1, 5.2**
func TestProperty_ModulePathParsing(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("extracted module path matches expected value for valid go.mod files", prop.ForAll(
		func(modulePath string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create go.mod file with the module path
			goModPath := filepath.Join(tmpDir, "go.mod")
			content := genGoModContent(modulePath)
			if err := os.WriteFile(goModPath, []byte(content), 0644); err != nil {
				t.Logf("failed to write go.mod: %v", err)
				return false
			}

			// Parse module path
			parser := NewModulePathParser(tmpDir)
			extractedPath, err := parser.Parse()
			if err != nil {
				t.Logf("failed to parse module path: %v", err)
				return false
			}

			// Verify extracted path matches expected value
			if extractedPath != modulePath {
				t.Logf("expected module path %s, got %s", modulePath, extractedPath)
				return false
			}

			return true
		},
		genValidModulePath(),
	))

	properties.TestingRun(t)
}

// Feature: domain-module-management, Property 2: Module Path Parsing (Single-Segment)
// **Validates: Requirements 2.1, 5.2**
func TestProperty_ModulePathParsing_SingleSegment(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("correctly parses single-segment module paths", prop.ForAll(
		func(modulePath string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create go.mod file
			goModPath := filepath.Join(tmpDir, "go.mod")
			content := genGoModContent(modulePath)
			if err := os.WriteFile(goModPath, []byte(content), 0644); err != nil {
				t.Logf("failed to write go.mod: %v", err)
				return false
			}

			// Parse module path
			parser := NewModulePathParser(tmpDir)
			extractedPath, err := parser.Parse()
			if err != nil {
				t.Logf("failed to parse module path: %v", err)
				return false
			}

			// Verify it's a single segment (no slashes)
			if strings.Contains(extractedPath, "/") {
				t.Logf("expected single-segment path, got %s", extractedPath)
				return false
			}

			// Verify extracted path matches expected value
			if extractedPath != modulePath {
				t.Logf("expected module path %s, got %s", modulePath, extractedPath)
				return false
			}

			return true
		},
		genSingleSegmentModulePath(),
	))

	properties.TestingRun(t)
}

// Feature: domain-module-management, Property 2: Module Path Parsing (Multi-Segment)
// **Validates: Requirements 2.1, 5.2**
func TestProperty_ModulePathParsing_MultiSegment(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("correctly parses multi-segment module paths", prop.ForAll(
		func(modulePath string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create go.mod file
			goModPath := filepath.Join(tmpDir, "go.mod")
			content := genGoModContent(modulePath)
			if err := os.WriteFile(goModPath, []byte(content), 0644); err != nil {
				t.Logf("failed to write go.mod: %v", err)
				return false
			}

			// Parse module path
			parser := NewModulePathParser(tmpDir)
			extractedPath, err := parser.Parse()
			if err != nil {
				t.Logf("failed to parse module path: %v", err)
				return false
			}

			// Verify it's multi-segment (has slashes)
			if !strings.Contains(extractedPath, "/") {
				t.Logf("expected multi-segment path, got %s", extractedPath)
				return false
			}

			// Verify extracted path matches expected value
			if extractedPath != modulePath {
				t.Logf("expected module path %s, got %s", modulePath, extractedPath)
				return false
			}

			return true
		},
		gen.OneGenOf(genTwoSegmentModulePath(), genThreeSegmentModulePath()),
	))

	properties.TestingRun(t)
}

// Feature: domain-module-management, Property 2: Module Path Parsing (With Hyphens)
// **Validates: Requirements 2.1, 5.2**
func TestProperty_ModulePathParsing_WithHyphens(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("correctly parses module paths with hyphens", prop.ForAll(
		func(modulePath string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create go.mod file
			goModPath := filepath.Join(tmpDir, "go.mod")
			content := genGoModContent(modulePath)
			if err := os.WriteFile(goModPath, []byte(content), 0644); err != nil {
				t.Logf("failed to write go.mod: %v", err)
				return false
			}

			// Parse module path
			parser := NewModulePathParser(tmpDir)
			extractedPath, err := parser.Parse()
			if err != nil {
				t.Logf("failed to parse module path: %v", err)
				return false
			}

			// Verify path contains hyphens
			if !strings.Contains(extractedPath, "-") {
				t.Logf("expected path with hyphens, got %s", extractedPath)
				return false
			}

			// Verify extracted path matches expected value
			if extractedPath != modulePath {
				t.Logf("expected module path %s, got %s", modulePath, extractedPath)
				return false
			}

			return true
		},
		genModulePathWithHyphens(),
	))

	properties.TestingRun(t)
}

// Feature: domain-module-management, Property 2: Module Path Parsing (With Underscores)
// **Validates: Requirements 2.1, 5.2**
func TestProperty_ModulePathParsing_WithUnderscores(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("correctly parses module paths with underscores", prop.ForAll(
		func(modulePath string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create go.mod file
			goModPath := filepath.Join(tmpDir, "go.mod")
			content := genGoModContent(modulePath)
			if err := os.WriteFile(goModPath, []byte(content), 0644); err != nil {
				t.Logf("failed to write go.mod: %v", err)
				return false
			}

			// Parse module path
			parser := NewModulePathParser(tmpDir)
			extractedPath, err := parser.Parse()
			if err != nil {
				t.Logf("failed to parse module path: %v", err)
				return false
			}

			// Verify path contains underscores
			if !strings.Contains(extractedPath, "_") {
				t.Logf("expected path with underscores, got %s", extractedPath)
				return false
			}

			// Verify extracted path matches expected value
			if extractedPath != modulePath {
				t.Logf("expected module path %s, got %s", modulePath, extractedPath)
				return false
			}

			return true
		},
		genModulePathWithUnderscores(),
	))

	properties.TestingRun(t)
}

// Feature: domain-module-management, Property 2: Module Path Parsing (With Comments)
// **Validates: Requirements 2.1, 5.2**
func TestProperty_ModulePathParsing_WithComments(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("correctly parses module paths when comments precede module declaration", prop.ForAll(
		func(modulePath string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create go.mod file with comments
			goModPath := filepath.Join(tmpDir, "go.mod")
			content := genGoModContentWithComments(modulePath)
			if err := os.WriteFile(goModPath, []byte(content), 0644); err != nil {
				t.Logf("failed to write go.mod: %v", err)
				return false
			}

			// Parse module path
			parser := NewModulePathParser(tmpDir)
			extractedPath, err := parser.Parse()
			if err != nil {
				t.Logf("failed to parse module path: %v", err)
				return false
			}

			// Verify extracted path matches expected value
			if extractedPath != modulePath {
				t.Logf("expected module path %s, got %s", modulePath, extractedPath)
				return false
			}

			return true
		},
		genValidModulePath(),
	))

	properties.TestingRun(t)
}

// Feature: domain-module-management, Property 2: Module Path Parsing (With Whitespace)
// **Validates: Requirements 2.1, 5.2**
func TestProperty_ModulePathParsing_WithWhitespace(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("correctly parses module paths with extra whitespace", prop.ForAll(
		func(modulePath string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create go.mod file with extra whitespace
			goModPath := filepath.Join(tmpDir, "go.mod")
			content := genGoModContentWithWhitespace(modulePath)
			if err := os.WriteFile(goModPath, []byte(content), 0644); err != nil {
				t.Logf("failed to write go.mod: %v", err)
				return false
			}

			// Parse module path
			parser := NewModulePathParser(tmpDir)
			extractedPath, err := parser.Parse()
			if err != nil {
				t.Logf("failed to parse module path: %v", err)
				return false
			}

			// Verify extracted path matches expected value (whitespace should be trimmed)
			if extractedPath != modulePath {
				t.Logf("expected module path %s, got %s", modulePath, extractedPath)
				return false
			}

			return true
		},
		genValidModulePath(),
	))

	properties.TestingRun(t)
}
