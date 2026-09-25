package manager

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewGenericModuleManager tests the constructor
func TestNewGenericModuleManager(t *testing.T) {
	projectDir := "/test/project"
	manager := NewGenericModuleManager(projectDir)

	if manager == nil {
		t.Fatal("NewGenericModuleManager returned nil")
	}

	if manager.projectDir != projectDir {
		t.Errorf("Expected projectDir %s, got %s", projectDir, manager.projectDir)
	}
}

// TestModuleExists tests checking if module file exists
func TestModuleExists(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(string) error
		expected bool
	}{
		{
			name: "module exists",
			setup: func(dir string) error {
				usecaseDir := filepath.Join(dir, "internal", "usecase")
				if err := os.MkdirAll(usecaseDir, 0755); err != nil {
					return err
				}
				modulePath := filepath.Join(usecaseDir, "module.go")
				return os.WriteFile(modulePath, []byte("package usecase"), 0644)
			},
			expected: true,
		},
		{
			name: "module does not exist",
			setup: func(dir string) error {
				return nil
			},
			expected: false,
		},
		{
			name: "domain directory exists but no module.go",
			setup: func(dir string) error {
				domainDir := filepath.Join(dir, "internal", "domain")
				return os.MkdirAll(domainDir, 0755)
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			if err := tt.setup(tmpDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			manager := NewGenericModuleManager(tmpDir)
			modulePath := filepath.Join(tmpDir, "internal", "usecase", "module.go")
			result := manager.ModuleExists(modulePath)

			if result != tt.expected {
				t.Errorf("Expected ModuleExists() = %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestCreateModule tests creating a new module file for first use case
func TestCreateModule(t *testing.T) {
	tests := []struct {
		name        string
		useCaseName string
		wantErr     bool
	}{
		{
			name:        "create module with simple use case name",
			useCaseName: "CreateUser",
			wantErr:     false,
		},
		{
			name:        "create module with compound use case name",
			useCaseName: "GetUserProfile",
			wantErr:     false,
		},
		{
			name:        "create module with single word use case",
			useCaseName: "Login",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			manager := NewGenericModuleManager(tmpDir)

			moduleDir := filepath.Join(tmpDir, "internal", "usecase")
			modulePath := filepath.Join(moduleDir, "module.go")

			err := manager.CreateModule(moduleDir, modulePath, `package usecase
import "go.uber.org/fx"

// Module provides use cases dependencies as an fx Module
var Module = fx.Module("usecase",
	fx.Provide(
		%s,
	),
)
`, "New"+tt.useCaseName+"UseCase")

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateModule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify file was created
			if _, err := os.Stat(modulePath); os.IsNotExist(err) {
				t.Errorf("module.go was not created at %s", modulePath)
				return
			}

			// Read and verify content
			content, err := os.ReadFile(modulePath)
			if err != nil {
				t.Fatalf("Failed to read created module.go: %v", err)
			}

			contentStr := string(content)

			// Verify package declaration
			if !strings.Contains(contentStr, "package usecase") {
				t.Error("module.go does not contain 'package usecase'")
			}

			// Verify fx import
			if !strings.Contains(contentStr, `import "go.uber.org/fx"`) {
				t.Error("module.go does not contain fx import")
			}

			// Verify Module variable
			if !strings.Contains(contentStr, "var Module = fx.Module") {
				t.Error("module.go does not contain Module variable")
			}

			// Verify module name is "usecase"
			if !strings.Contains(contentStr, `fx.Module("usecase"`) {
				t.Error("module.go does not contain correct module name")
			}

			// Verify provider function
			expectedProvider := "New" + tt.useCaseName + "UseCase"
			if !strings.Contains(contentStr, expectedProvider) {
				t.Errorf("module.go does not contain provider %s", expectedProvider)
			}

			// Verify fx.Provide
			if !strings.Contains(contentStr, "fx.Provide") {
				t.Error("module.go does not contain fx.Provide")
			}

			// Verify the file is valid Go code
			fset := token.NewFileSet()
			_, err = parser.ParseFile(fset, modulePath, content, parser.ParseComments)
			if err != nil {
				t.Errorf("Created module.go is not valid Go code: %v", err)
			}
		})
	}
}

// TestCreateModuleCreatesDirectory tests that CreateModule creates the domain directory if it doesn't exist
func TestCreateModuleCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewGenericModuleManager(tmpDir)

	// Verify domain directory doesn't exist
	usecaseDir := filepath.Join(tmpDir, "internal", "usecase")
	if _, err := os.Stat(usecaseDir); !os.IsNotExist(err) {
		t.Fatal("Usecase directory should not exist before CreateModule")
	}

	modulePath := filepath.Join(usecaseDir, "module.go")
	err := manager.CreateModule(usecaseDir, modulePath, `package usecase
import "go.uber.org/fx"

// Module provides use cases dependencies as an fx Module
var Module = fx.Module("usecase",
	fx.Provide(
		%s,
	),
)
`, "NewTestUseCase")

	if err != nil {
		t.Fatalf("CreateModule() failed: %v", err)
	}

	// Verify domain directory was created
	if _, err := os.Stat(usecaseDir); os.IsNotExist(err) {
		t.Error("Usecase directory was not created")
	}

	// Verify module.go was created
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		t.Error("module.go was not created")
	}
}

// TestAddProvider tests adding a provider to existing module
func TestAddProvider(t *testing.T) {
	tests := []struct {
		name            string
		initialContent  string
		useCaseName     string
		wantErr         bool
		verifyProviders []string
	}{
		{
			name: "add provider to module with one existing provider",
			initialContent: `package usecase

import "go.uber.org/fx"

var Module = fx.Module("usecase",
	fx.Provide(
		NewCreateUserUseCase,
	),
)
`,
			useCaseName: "GetUser",
			wantErr:     false,
			verifyProviders: []string{
				"NewCreateUserUseCase",
				"NewGetUserUseCase",
			},
		},
		{
			name: "add provider to module with multiple existing providers",
			initialContent: `package usecase

import "go.uber.org/fx"

var Module = fx.Module("usecase",
	fx.Provide(
		NewCreateUserUseCase,
		NewUpdateUserUseCase,
		NewDeleteUserUseCase,
	),
)
`,
			useCaseName: "GetUser",
			wantErr:     false,
			verifyProviders: []string{
				"NewCreateUserUseCase",
				"NewUpdateUserUseCase",
				"NewDeleteUserUseCase",
				"NewGetUserUseCase",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			usecaseDir := filepath.Join(tmpDir, "internal", "usecase")
			if err := os.MkdirAll(usecaseDir, 0755); err != nil {
				t.Fatalf("Failed to create usecase directory: %v", err)
			}

			modulePath := filepath.Join(usecaseDir, "module.go")
			if err := os.WriteFile(modulePath, []byte(tt.initialContent), 0644); err != nil {
				t.Fatalf("Failed to write initial module.go: %v", err)
			}

			manager := NewGenericModuleManager(tmpDir)
			// manager.CreateModule(usecaseDir, modulePath, tt.initialContent, tt.useCaseName)

			err := manager.AddProvider(modulePath, "New"+tt.useCaseName+"UseCase")

			if (err != nil) != tt.wantErr {
				t.Errorf("AddProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Read and verify updated content
			content, err := os.ReadFile(modulePath)
			if err != nil {
				t.Fatalf("Failed to read updated module.go: %v", err)
			}

			contentStr := string(content)

			// Verify all expected providers are present
			for _, provider := range tt.verifyProviders {
				if !strings.Contains(contentStr, provider) {
					t.Errorf("Updated module.go does not contain provider %s", provider)
				}
			}

			// Verify the file is still valid Go code
			fset := token.NewFileSet()
			_, err = parser.ParseFile(fset, modulePath, content, parser.ParseComments)
			if err != nil {
				t.Errorf("Updated module.go is not valid Go code: %v", err)
			}
		})
	}
}

// TestAddProviderIdempotent tests that adding a provider that already exists is idempotent
func TestAddProviderIdempotent(t *testing.T) {
	initialContent := `package usecase

import "go.uber.org/fx"

var Module = fx.Module("usecase",
	fx.Provide(
		NewCreateUserUseCase,
		NewGetUserUseCase,
	),
)
`

	tmpDir := t.TempDir()
	usecaseDir := filepath.Join(tmpDir, "internal", "usecase")
	if err := os.MkdirAll(usecaseDir, 0755); err != nil {
		t.Fatalf("Failed to create usecase directory: %v", err)
	}

	modulePath := filepath.Join(usecaseDir, "module.go")
	if err := os.WriteFile(modulePath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to write initial module.go: %v", err)
	}

	manager := NewGenericModuleManager(tmpDir)

	// Add provider that already exists
	err := manager.AddProvider(modulePath, "GetUser")
	if err != nil {
		t.Fatalf("AddProvider() failed: %v", err)
	}

	// Read content after first add
	content, err := os.ReadFile(modulePath)
	if err != nil {
		t.Fatalf("Failed to read module.go: %v", err)
	}

	// Count occurrences of the provider
	count := strings.Count(string(content), "NewGetUserUseCase")
	if count != 1 {
		t.Errorf("Expected NewGetUserUseCase to appear exactly once, found %d occurrences", count)
	}

	// Add the same provider again
	err = manager.AddProvider(modulePath, "GetUser")
	if err != nil {
		t.Fatalf("Second AddProvider() failed: %v", err)
	}

	// Read content after second add
	content, err = os.ReadFile(modulePath)
	if err != nil {
		t.Fatalf("Failed to read module.go after second add: %v", err)
	}

	// Verify it still appears only once
	count = strings.Count(string(content), "NewGetUserUseCase")
	if count != 1 {
		t.Errorf("After second add, expected NewGetUserUseCase to appear exactly once, found %d occurrences", count)
	}
}

// TestAddProviderPreservesExisting tests that adding a provider preserves all existing providers
func TestAddProviderPreservesExisting(t *testing.T) {
	initialContent := `package usecase

import "go.uber.org/fx"

var Module = fx.Module("usecase",
	fx.Provide(
		NewCreateUserUseCase,
		NewUpdateUserUseCase,
		NewDeleteUserUseCase,
		NewListUsersUseCase,
	),
)
`

	tmpDir := t.TempDir()
	usecaseDir := filepath.Join(tmpDir, "internal", "usecase")
	if err := os.MkdirAll(usecaseDir, 0755); err != nil {
		t.Fatalf("Failed to create usecase directory: %v", err)
	}

	modulePath := filepath.Join(usecaseDir, "module.go")
	if err := os.WriteFile(modulePath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to write initial module.go: %v", err)
	}

	manager := NewGenericModuleManager(tmpDir)

	// Add a new provider
	err := manager.AddProvider(modulePath, "NewGetUserUseCase")
	if err != nil {
		t.Fatalf("AddProvider() failed: %v", err)
	}

	// Read updated content
	content, err := os.ReadFile(modulePath)
	if err != nil {
		t.Fatalf("Failed to read updated module.go: %v", err)
	}

	contentStr := string(content)

	// Verify all original providers are still present
	originalProviders := []string{
		"NewCreateUserUseCase",
		"NewUpdateUserUseCase",
		"NewDeleteUserUseCase",
		"NewListUsersUseCase",
	}

	for _, provider := range originalProviders {
		if !strings.Contains(contentStr, provider) {
			t.Errorf("Original provider %s was not preserved", provider)
		}
	}

	// Verify new provider was added
	if !strings.Contains(contentStr, "NewGetUserUseCase") {
		t.Error("New provider NewGetUserUseCase was not added")
	}
}

// TestAddProviderMaintainsStructure tests that adding a provider maintains module structure and name
func TestAddProviderMaintainsStructure(t *testing.T) {
	initialContent := `package usecase

import "go.uber.org/fx"

// Module provides use cases dependencies as an fx Module
var Module = fx.Module("usecase",
	fx.Provide(
		NewCreateUserUseCase,
	),
)
`

	tmpDir := t.TempDir()
	usecaseDir := filepath.Join(tmpDir, "internal", "usecase")
	if err := os.MkdirAll(usecaseDir, 0755); err != nil {
		t.Fatalf("Failed to create usecase directory: %v", err)
	}

	modulePath := filepath.Join(usecaseDir, "module.go")
	if err := os.WriteFile(modulePath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to write initial module.go: %v", err)
	}

	manager := NewGenericModuleManager(tmpDir)

	// Add a new provider
	err := manager.AddProvider(modulePath, "GetUser")
	if err != nil {
		t.Fatalf("AddProvider() failed: %v", err)
	}

	// Read updated content
	content, err := os.ReadFile(modulePath)
	if err != nil {
		t.Fatalf("Failed to read updated module.go: %v", err)
	}

	contentStr := string(content)

	// Verify package declaration is preserved
	if !strings.Contains(contentStr, "package usecase") {
		t.Error("Package declaration was not preserved")
	}

	// Verify fx import is preserved
	if !strings.Contains(contentStr, `"go.uber.org/fx"`) {
		t.Error("fx import was not preserved")
	}

	// Verify Module variable declaration is preserved
	if !strings.Contains(contentStr, "var Module = fx.Module") {
		t.Error("Module variable declaration was not preserved")
	}

	// Verify module name "usecase" is preserved
	if !strings.Contains(contentStr, `fx.Module("usecase"`) {
		t.Error("Module name 'usecase' was not preserved")
	}

	// Verify fx.Provide structure is preserved
	if !strings.Contains(contentStr, "fx.Provide") {
		t.Error("fx.Provide structure was not preserved")
	}

	// Verify comment is preserved
	if !strings.Contains(contentStr, "Module provides use cases dependencies") {
		t.Error("Comment was not preserved")
	}
}

// TestAddProviderErrorUnparseableModule tests error handling for unparseable module.go
func TestAddProviderErrorUnparseableModule(t *testing.T) {
	tests := []struct {
		name           string
		invalidContent string
		wantErrContain string
	}{
		{
			name:           "syntax error in module.go",
			invalidContent: `package usecase this is invalid syntax`,
			wantErrContain: "failed to parse module.go",
		},
		{
			name: "incomplete module.go",
			invalidContent: `package usecase

import "go.uber.org/fx"

var Module = fx.Module("usecase",
	fx.Provide(
`,
			wantErrContain: "failed to parse module.go",
		},
		{
			name:           "empty file",
			invalidContent: ``,
			wantErrContain: "failed to parse module.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			usecaseDir := filepath.Join(tmpDir, "internal", "usecase")
			if err := os.MkdirAll(usecaseDir, 0755); err != nil {
				t.Fatalf("Failed to create usecase directory: %v", err)
			}

			modulePath := filepath.Join(usecaseDir, "module.go")
			if err := os.WriteFile(modulePath, []byte(tt.invalidContent), 0644); err != nil {
				t.Fatalf("Failed to write invalid module.go: %v", err)
			}

			manager := NewGenericModuleManager(tmpDir)
			err := manager.AddProvider(modulePath, "TestUseCase")

			if err == nil {
				t.Error("AddProvider() should have returned an error for unparseable module.go")
				return
			}

			if !strings.Contains(err.Error(), tt.wantErrContain) {
				t.Errorf("Expected error to contain '%s', got: %v", tt.wantErrContain, err)
			}
		})
	}
}

// TestAddProviderErrorMissingModule tests error handling when module.go doesn't exist
func TestAddProviderErrorMissingModule(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewGenericModuleManager(tmpDir)

	err := manager.AddProvider(filepath.Join(tmpDir, "foo", "bar", "module.go"), "TestUseCase")

	if err == nil {
		t.Error("AddProvider() should have returned an error when module.go doesn't exist")
		return
	}

	if !strings.Contains(err.Error(), "failed to parse module.go") {
		t.Errorf("Expected error about parsing module.go, got: %v", err)
	}
}
