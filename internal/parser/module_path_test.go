package parser

import (
	"os"
	"path/filepath"
	"testing"
)

// TestParse_ValidStandardModulePath tests parsing a valid go.mod with a standard module path
func TestParse_ValidStandardModulePath(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create go.mod with standard module path
	goModContent := "module myproject\n\ngo 1.21\n"
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to create test go.mod: %v", err)
	}

	// Parse module path
	parser := NewModulePathParser(tmpDir)
	modulePath, err := parser.Parse()

	// Verify results
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if modulePath != "myproject" {
		t.Errorf("expected module path 'myproject', got: %s", modulePath)
	}
}

// TestParse_ValidMultiSegmentPath tests parsing a go.mod with multi-segment path
func TestParse_ValidMultiSegmentPath(t *testing.T) {
	testCases := []struct {
		name         string
		moduleDecl   string
		expectedPath string
	}{
		{
			name:         "two segments",
			moduleDecl:   "module github.com/myproject",
			expectedPath: "github.com/myproject",
		},
		{
			name:         "three segments",
			moduleDecl:   "module github.com/user/repo",
			expectedPath: "github.com/user/repo",
		},
		{
			name:         "with hyphens",
			moduleDecl:   "module github.com/user/my-project",
			expectedPath: "github.com/user/my-project",
		},
		{
			name:         "with underscores",
			moduleDecl:   "module github.com/user/my_project",
			expectedPath: "github.com/user/my_project",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create go.mod
			goModContent := tc.moduleDecl + "\n\ngo 1.21\n"
			goModPath := filepath.Join(tmpDir, "go.mod")
			if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
				t.Fatalf("failed to create test go.mod: %v", err)
			}

			// Parse module path
			parser := NewModulePathParser(tmpDir)
			modulePath, err := parser.Parse()

			// Verify results
			if err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
			if modulePath != tc.expectedPath {
				t.Errorf("expected module path '%s', got: %s", tc.expectedPath, modulePath)
			}
		})
	}
}

// TestParse_MissingGoMod tests error handling when go.mod file doesn't exist
func TestParse_MissingGoMod(t *testing.T) {
	// Create temporary directory without go.mod
	tmpDir := t.TempDir()

	// Attempt to parse
	parser := NewModulePathParser(tmpDir)
	modulePath, err := parser.Parse()

	// Verify error is returned
	if err == nil {
		t.Error("expected error for missing go.mod, got nil")
	}
	if modulePath != "" {
		t.Errorf("expected empty module path, got: %s", modulePath)
	}
	// Verify error message mentions go.mod
	if err != nil && !contains(err.Error(), "go.mod") {
		t.Errorf("expected error message to mention 'go.mod', got: %v", err)
	}
}

// TestParse_NoModuleDeclaration tests error handling when go.mod has no module declaration
func TestParse_NoModuleDeclaration(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create go.mod without module declaration
	goModContent := "go 1.21\n\nrequire (\n\tgithub.com/example/pkg v1.0.0\n)\n"
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to create test go.mod: %v", err)
	}

	// Attempt to parse
	parser := NewModulePathParser(tmpDir)
	modulePath, err := parser.Parse()

	// Verify error is returned
	if err == nil {
		t.Error("expected error for missing module declaration, got nil")
	}
	if modulePath != "" {
		t.Errorf("expected empty module path, got: %s", modulePath)
	}
	// Verify error message mentions module declaration
	if err != nil && !contains(err.Error(), "module declaration not found") {
		t.Errorf("expected error message about missing module declaration, got: %v", err)
	}
}

// TestParse_ModuleWithOnlyWhitespace tests behavior when module line has only whitespace after keyword
// Note: Due to TrimSpace being called before checking the prefix, a line like "module   "
// becomes "module" after trimming, which won't match the "module " prefix (with space).
// This test documents that such cases are treated as "module declaration not found"
func TestParse_ModuleWithOnlyWhitespace(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create go.mod with "module" followed by only whitespace
	// After TrimSpace, this becomes just "module" which won't match "module " prefix
	goModContent := "module   \n\ngo 1.21\n"
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to create test go.mod: %v", err)
	}

	// Attempt to parse
	parser := NewModulePathParser(tmpDir)
	modulePath, err := parser.Parse()

	// Verify error is returned
	if err == nil {
		t.Error("expected error, got nil")
	}
	if modulePath != "" {
		t.Errorf("expected empty module path, got: %s", modulePath)
	}
	// Due to implementation, this returns "module declaration not found" not "empty module path"
	if err != nil && !contains(err.Error(), "module declaration not found") {
		t.Errorf("expected 'module declaration not found' error, got: %v", err)
	}
}

// TestParse_CommentsBeforeModuleDeclaration tests parsing with comments before module
func TestParse_CommentsBeforeModuleDeclaration(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create go.mod with comments before module declaration
	goModContent := `// This is a comment
// Another comment
module github.com/user/project

go 1.21
`
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to create test go.mod: %v", err)
	}

	// Parse module path
	parser := NewModulePathParser(tmpDir)
	modulePath, err := parser.Parse()

	// Verify results
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if modulePath != "github.com/user/project" {
		t.Errorf("expected module path 'github.com/user/project', got: %s", modulePath)
	}
}

// TestParse_WhitespaceAroundModulePath tests parsing with extra whitespace
func TestParse_WhitespaceAroundModulePath(t *testing.T) {
	testCases := []struct {
		name         string
		moduleDecl   string
		expectedPath string
	}{
		{
			name:         "leading spaces",
			moduleDecl:   "module    github.com/user/project",
			expectedPath: "github.com/user/project",
		},
		{
			name:         "trailing spaces",
			moduleDecl:   "module github.com/user/project   ",
			expectedPath: "github.com/user/project",
		},
		{
			name:         "mixed whitespace",
			moduleDecl:   "module  \t github.com/user/project",
			expectedPath: "github.com/user/project",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create go.mod
			goModContent := tc.moduleDecl + "\n\ngo 1.21\n"
			goModPath := filepath.Join(tmpDir, "go.mod")
			if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
				t.Fatalf("failed to create test go.mod: %v", err)
			}

			// Parse module path
			parser := NewModulePathParser(tmpDir)
			modulePath, err := parser.Parse()

			// Verify results
			if err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
			if modulePath != tc.expectedPath {
				t.Errorf("expected module path '%s', got: %s", tc.expectedPath, modulePath)
			}
		})
	}
}

// TestParse_ModuleInMiddleOfFile tests parsing when module declaration is not first
func TestParse_ModuleInMiddleOfFile(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create go.mod with module declaration after other content
	goModContent := `// Package comment
// More comments

module github.com/user/project

go 1.21

require (
	github.com/example/pkg v1.0.0
)
`
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to create test go.mod: %v", err)
	}

	// Parse module path
	parser := NewModulePathParser(tmpDir)
	modulePath, err := parser.Parse()

	// Verify results
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if modulePath != "github.com/user/project" {
		t.Errorf("expected module path 'github.com/user/project', got: %s", modulePath)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
